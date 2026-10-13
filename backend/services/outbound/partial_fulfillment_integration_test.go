package outbound

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"testing"
	"time"
	dto "wms-api/dto/outbound"
	inventorymodel "wms-api/models/inventory"
	mastermodel "wms-api/models/master"
	repository "wms-api/repository/outbound"
)

type partialFixture struct{ actor, owner, warehouse, customer, shipTo, uom, status, strategy, pick, stage, pack, date, suffix string }

func testPartialFulfillment(t *testing.T, db *gorm.DB, service *Service, f partialFixture) {
	ctx := context.Background()
	items := []mastermodel.Item{{OwnerID: f.owner, Code: "PART-A-" + f.suffix, Name: "Partial A", BaseUOMID: f.uom}, {OwnerID: f.owner, Code: "PART-B-" + f.suffix, Name: "Partial B", BaseUOMID: f.uom}}
	for i := range items {
		outboundOK(t, db.Create(&items[i]).Error)
	}
	balances := []inventorymodel.InventoryBalance{}
	for i, item := range items {
		qty := "0"
		if i == 0 {
			qty = "6"
		}
		balance := inventorymodel.InventoryBalance{ID: fmt.Sprintf("PART-BAL-%d-%s", i, f.suffix), OwnerID: f.owner, WarehouseID: f.warehouse, LocationID: f.pick, ItemID: item.ID, InventoryStatusID: f.status, OnHandQty: qty, ReservedQty: "0", UOMID: f.uom, VersionNo: 1}
		outboundOK(t, db.Create(&balance).Error)
		balances = append(balances, balance)
	}
	create := func(ref string) dto.OutboundOrderResponse {
		order, err := service.CreateOutboundOrder(ctx, dto.CreateOutboundOrderRequest{OwnerID: f.owner, WarehouseID: f.warehouse, CustomerID: f.customer, ShipToPartnerID: f.shipTo, BusinessDate: f.date, ClientDeliveryOrderNo: ref + f.suffix, RequestedShipAt: f.date + "T15:00:00+07:00", Lines: []dto.OutboundOrderLineRequest{{ItemID: items[0].ID, OrderedQty: "10"}, {ItemID: items[1].ID, OrderedQty: "2"}}}, f.actor)
		outboundOK(t, err)
		_, err = service.ValidateOutboundOrder(ctx, order.ID, dto.ValidateOutboundRequest{ExpectedVersion: order.VersionNo}, f.actor)
		outboundOK(t, err)
		order, err = service.GetOutboundOrder(ctx, order.ID)
		outboundOK(t, err)
		order, err = service.ReleaseOutboundOrder(ctx, order.ID, dto.TransitionRequest{ExpectedVersion: order.VersionNo}, f.actor)
		outboundOK(t, err)
		return order
	}
	allocate := func(order dto.OutboundOrderResponse) dto.AllocationResponse {
		result, err := service.AllocateOutboundOrder(ctx, order.ID, dto.AllocateOutboundRequest{ExpectedVersion: order.VersionNo, PickingStrategyID: &f.strategy, AllowPartial: true}, f.actor)
		outboundOK(t, err)
		return result
	}
	// Each invocation owns one wave -> staging -> passed check -> packing ->
	// shipment, while the DO counters remain cumulative across invocations.
	cycle := func(order dto.OutboundOrderResponse, tag string) (dto.ShipmentResponse, dto.PackingResponse, dto.StagingResponse) {
		wave, err := service.CreateWave(ctx, dto.CreateWaveRequest{OwnerID: f.owner, WarehouseID: f.warehouse, BusinessDate: f.date, WaveTypeCode: "SINGLE_ORDER", PickingStrategyID: &f.strategy, OutboundIDs: []string{order.ID}}, f.actor)
		outboundOK(t, err)
		wave, err = service.ReleaseWave(ctx, wave.ID, dto.ReleaseWaveRequest{ExpectedVersion: wave.VersionNo, StagingLocationID: f.stage, PriorityCode: "NORMAL"}, f.actor)
		outboundOK(t, err)
		tasks, err := service.ListPickTasks(ctx, repository.ListFilter{OwnerID: f.owner, WarehouseID: f.warehouse, Page: 1, PageSize: 100}, "")
		outboundOK(t, err)
		for _, task := range tasks.Items {
			if task.WaveID != wave.ID {
				continue
			}
			task, err = service.StartPickTask(ctx, task.ID, f.actor)
			outboundOK(t, err)
			_, err = service.ConfirmPick(ctx, task.ID, dto.ConfirmPickRequest{PickedQty: task.PlannedQty, ExpectedBalanceVersion: task.BalanceVersion, BusinessDate: f.date, OperationKey: tag + task.ID}, f.actor)
			outboundOK(t, err)
		}
		staging, err := service.CreateStaging(ctx, dto.CreateStagingRequest{OutboundID: order.ID, WaveID: wave.ID}, f.actor)
		outboundOK(t, err)
		staging, err = service.CompleteStaging(ctx, staging.ID, f.actor)
		outboundOK(t, err)
		check, err := service.CreateCheck(ctx, dto.CreateCheckRequest{StagingID: staging.ID}, f.actor)
		outboundOK(t, err)
		for _, line := range check.Lines {
			check, err = service.RecordCheckLine(ctx, check.ID, line.ID, dto.RecordCheckLineRequest{CheckedQty: line.ExpectedQty, ResultCode: "PASS"}, f.actor)
			outboundOK(t, err)
		}
		check, err = service.CompleteCheck(ctx, check.ID, f.actor)
		outboundOK(t, err)
		packing, err := service.CreatePacking(ctx, dto.CreatePackingRequest{OutboundCheckID: check.ID, PackingLocationID: f.pack}, f.actor)
		outboundOK(t, err)
		if len(packing.Lines) != len(check.Lines) {
			t.Fatal("packing borrowed a previous cycle's lines")
		}
		for _, line := range packing.Lines {
			packing, err = service.GetPacking(ctx, packing.ID)
			outboundOK(t, err)
			for _, fresh := range packing.Lines {
				if fresh.OutboundCheckLineID == line.OutboundCheckLineID {
					line = fresh
					break
				}
			}
			packing, err = service.PackLine(ctx, packing.ID, line.OutboundCheckLineID, dto.PackLineRequest{ExpectedBalanceVersion: line.SourceBalanceVersion, BusinessDate: f.date, OperationKey: tag + "pack" + line.OutboundCheckLineID}, f.actor)
			outboundOK(t, err)
		}
		packing, err = service.CompletePacking(ctx, packing.ID, f.actor)
		outboundOK(t, err)
		shipment, err := service.CreateShipment(ctx, dto.CreateShipmentRequest{OwnerID: f.owner, WarehouseID: f.warehouse, BusinessDate: f.date, PackingIDs: []string{packing.ID}}, f.actor)
		outboundOK(t, err)
		for _, line := range shipment.Lines {
			shipment, err = service.GetShipment(ctx, shipment.ID)
			outboundOK(t, err)
			for _, fresh := range shipment.Lines {
				if fresh.PackingLineID == line.PackingLineID {
					line = fresh
					break
				}
			}
			shipment, err = service.DispatchShipmentLine(ctx, shipment.ID, line.PackingLineID, dto.DispatchShipmentLineRequest{ExpectedBalanceVersion: line.SourceBalanceVersion, BusinessDate: f.date, OperationKey: tag + "ship" + line.PackingLineID}, f.actor)
			outboundOK(t, err)
		}
		shipment, err = service.CompleteShipment(ctx, shipment.ID, f.actor)
		outboundOK(t, err)
		return shipment, packing, staging
	}
	createDelivery := func(orderID string, shipment dto.ShipmentResponse) dto.DeliveryResponse {
		delivery, err := service.CreateDelivery(ctx, dto.CreateDeliveryRequest{ShipmentID: shipment.ID, OutboundID: orderID, BusinessDate: f.date}, f.actor)
		outboundOK(t, err)
		return delivery
	}
	deliver := func(delivery dto.DeliveryResponse) {
		at := time.Now().Format(time.RFC3339)
		delivery, err := service.DepartDelivery(ctx, delivery.ID, dto.DeliveryEventRequest{EventAt: at}, f.actor)
		outboundOK(t, err)
		delivery, err = service.ArriveDelivery(ctx, delivery.ID, dto.DeliveryEventRequest{EventAt: at}, f.actor)
		outboundOK(t, err)
		for _, line := range delivery.Lines {
			delivery, err = service.DeliverLine(ctx, delivery.ID, line.ID, dto.DeliverLineRequest{DeliveredQty: line.PlannedQty, EventAt: at, RecipientName: "Receiver", ProofReference: "POD-" + delivery.ID}, f.actor)
			outboundOK(t, err)
		}
		_, err = service.CompleteDelivery(ctx, delivery.ID, dto.CompleteDeliveryRequest{DeliveredAt: at}, f.actor)
		outboundOK(t, err)
	}
	order := create("PART-FULL-")
	allocation := allocate(order)
	if allocation.AllocatedQty != "6.000000" || allocation.ShortageQty != "6.000000" {
		t.Fatalf("partial allocation: %+v", allocation)
	}
	_, err := service.CloseShortOutboundOrder(ctx, order.ID, dto.CloseShortOutboundRequest{ExpectedVersion: allocation.Order.VersionNo, Reason: "Cannot close reserved stock"}, f.actor)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("reserved demand closure got %v", err)
	}
	first, firstPack, firstStage := cycle(allocation.Order, "part-first-")
	order, err = service.GetOutboundOrder(ctx, order.ID)
	outboundOK(t, err)
	if order.StatusCode != "PARTIALLY_FULFILLED" || order.Lines[0].ShippedQty != "6.000000" || order.Lines[0].OpenQty != "4.000000" || order.Lines[1].OpenQty != "2.000000" {
		t.Fatalf("partial shipment closed demand: %+v", order)
	}
	noStock := allocate(order)
	if noStock.Order.VersionNo != order.VersionNo || noStock.Order.StatusCode != "PARTIALLY_FULFILLED" || noStock.AllocatedQty != "0.000000" {
		t.Fatal("empty allocation changed the order")
	}
	oldDelivery := createDelivery(order.ID, first)
	// Delivery completion while the next cycle has allocations must not set a
	// final status or block its next wave.
	for i, balance := range balances {
		qty := "2"
		if i == 0 {
			qty = "4"
		}
		outboundOK(t, db.Model(&balance).Updates(map[string]interface{}{"on_hand_qty": qty, "version_no": gorm.Expr("version_no+1")}).Error)
	}
	allocation = allocate(order)
	deliver(oldDelivery)
	order, err = service.GetOutboundOrder(ctx, order.ID)
	outboundOK(t, err)
	if order.StatusCode != "ALLOCATED" {
		t.Fatal("old delivery overwrote the new fulfillment cycle")
	}
	second, _, _ := cycle(order, "part-second-")
	order, err = service.GetOutboundOrder(ctx, order.ID)
	outboundOK(t, err)
	if order.StatusCode != "SHIPPED" || order.Lines[0].ShippedQty != "10.000000" || order.Lines[1].ShippedQty != "2.000000" {
		t.Fatalf("second shipment did not finish demand: %+v", order)
	}
	outboundOK(t, service.refreshDeliveryOutcome(ctx, order.ID, f.actor, time.Now()))
	order, err = service.GetOutboundOrder(ctx, order.ID)
	outboundOK(t, err)
	if order.StatusCode != "SHIPPED" {
		t.Fatal("a shipment without final delivery was ignored in the aggregate outcome")
	}
	history, err := service.GetPacking(ctx, firstPack.ID)
	outboundOK(t, err)
	if len(history.Lines) != 1 {
		t.Fatal("historical packing includes new lines")
	}
	_, err = service.CreateCheck(ctx, dto.CreateCheckRequest{StagingID: firstStage.ID}, f.actor)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("old staging recheck got %v", err)
	}
	deliver(createDelivery(order.ID, second))
	order, err = service.GetOutboundOrder(ctx, order.ID)
	outboundOK(t, err)
	if order.StatusCode != "DELIVERED" {
		t.Fatalf("aggregate delivery not final: %+v", order)
	}

	// A separate two-line DO demonstrates irreversible, versioned demand
	// closure without touching stock, reservations, or original quantity snapshots.
	outboundOK(t, db.Model(&balances[0]).Updates(map[string]interface{}{"on_hand_qty": "6", "version_no": gorm.Expr("version_no+1")}).Error)
	short := create("PART-CLOSE-")
	shortAllocation := allocate(short)
	shortShipment, _, _ := cycle(shortAllocation.Order, "part-short-")
	short, err = service.GetOutboundOrder(ctx, short.ID)
	outboundOK(t, err)
	_, err = service.CloseShortOutboundOrder(ctx, short.ID, dto.CloseShortOutboundRequest{ExpectedVersion: short.VersionNo, Reason: "  "}, f.actor)
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("blank close reason got %v", err)
	}
	_, err = service.CloseShortOutboundOrder(ctx, short.ID, dto.CloseShortOutboundRequest{ExpectedVersion: short.VersionNo - 1, Reason: "Customer agreed"}, f.actor)
	if !errors.Is(err, repository.ErrConcurrentWrite) {
		t.Fatalf("stale close got %v", err)
	}
	short, err = service.CloseShortOutboundOrder(ctx, short.ID, dto.CloseShortOutboundRequest{ExpectedVersion: short.VersionNo, Reason: "Customer agreed to cancel remainder"}, f.actor)
	outboundOK(t, err)
	if short.StatusCode != "CLOSED_SHORT" || short.ClosedShortBy == nil || *short.ClosedShortBy != f.actor || short.ClosedShortReason == nil || short.Lines[0].OrderedBaseQty != "10.000000" || short.Lines[0].ClosedShortQty != "4.000000" || short.Lines[1].ClosedShortQty != "2.000000" || short.Lines[0].OpenQty != "0.000000" {
		t.Fatalf("close audit/quantity incorrect: %+v", short)
	}
	_, err = service.AllocateOutboundOrder(ctx, short.ID, dto.AllocateOutboundRequest{ExpectedVersion: short.VersionNo, PickingStrategyID: &f.strategy, AllowPartial: true}, f.actor)
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("closed demand allocation got %v", err)
	}
	deliver(createDelivery(short.ID, shortShipment))
	short, err = service.GetOutboundOrder(ctx, short.ID)
	outboundOK(t, err)
	if short.StatusCode != "CLOSED_SHORT" || short.Lines[0].DeliveredQty != "6.000000" {
		t.Fatal("delivery destroyed closed-short audit")
	}
	// Applying migration 28 again preserves completed membership and closure audit.
	outboundOK(t, repository.MigratePartialFulfillment(db))
	links, err := service.repositories.ShipmentOrder.List(ctx, shortShipment.ID)
	outboundOK(t, err)
	if len(links) != 1 || links[0].FulfilledAt == nil || links[0].RemovedAt != nil {
		t.Fatal("completed shipment membership was lost on repeat migration")
	}
}
