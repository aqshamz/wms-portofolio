package master

import (
	"context"
	"time"
	model "wms-api/models/master"
	repository "wms-api/repository/master"
)

// SeedOperational mirrors the operational subset of database/master/00_bootstrap_reference_data.sql.
// Existing records, disabled rules and user-selected initial states are preserved.
func (s *OperationalService) SeedOperational(ctx context.Context) error {
	return s.transaction(ctx, func(local *OperationalService) error {
		repos := local.repositories
		for _, value := range []model.AppModule{
			{Code: "AUTH", Name: "Authentication and access", DisplayOrder: 10, IsActive: true},
			{Code: "SECURITY", Name: "Security administration", DisplayOrder: 15, IsActive: true},
			{Code: "MASTER", Name: "Master data", DisplayOrder: 20, IsActive: true},
			{Code: "INBOUND", Name: "Inbound operations", DisplayOrder: 30, IsActive: true},
			{Code: "INVENTORY", Name: "Inventory control", DisplayOrder: 40, IsActive: true},
			{Code: "OUTBOUND", Name: "Outbound operations", DisplayOrder: 50, IsActive: true},
			{Code: "BILLING", Name: "Billing and receivables", DisplayOrder: 60, IsActive: true},
			{Code: "REPORTING", Name: "Reports and analytics", DisplayOrder: 70, IsActive: true},
			{Code: "GENERAL", Name: "General/shared", DisplayOrder: 90, IsActive: true},
		} {
			if err := repos.AppModule.SeedOne(ctx, &value); err != nil {
				return err
			}
		}
		for _, value := range []model.TaskType{
			{Code: "PUTAWAY", Name: "Putaway", Description: operationalString("Move received stock into storage."), IsActive: true},
			{Code: "PICK", Name: "Pick", Description: operationalString("Pick stock for outbound fulfillment."), IsActive: true},
			{Code: "REPLENISHMENT", Name: "Replenishment", Description: operationalString("Move stock into a picking location."), IsActive: true},
			{Code: "STOCK_COUNT", Name: "Stock count", Description: operationalString("Count stock at a warehouse location."), IsActive: true},
			{Code: "REWORK", Name: "Rework", Description: operationalString("Rework quarantined inbound stock."), IsActive: true},
		} {
			if err := repos.TaskType.SeedOne(ctx, &value); err != nil {
				return err
			}
		}
		for _, value := range []model.TaskPriority{
			{Code: "LOW", Name: "Low", PriorityValue: 100, IsActive: true},
			{Code: "NORMAL", Name: "Normal", PriorityValue: 200, IsActive: true},
			{Code: "HIGH", Name: "High", PriorityValue: 300, IsActive: true},
			{Code: "URGENT", Name: "Urgent", PriorityValue: 400, IsActive: true},
		} {
			if err := repos.TaskPriority.SeedOne(ctx, &value); err != nil {
				return err
			}
		}
		for _, value := range []model.PickingSortMethod{
			{Code: "FEFO", Name: "First expiry, first out", Description: operationalString("Pick the stock with the earliest expiry date first."), IsActive: true},
			{Code: "FIFO", Name: "First in, first out", Description: operationalString("Pick the oldest received stock first."), IsActive: true},
			{Code: "LOCATION", Name: "Location sequence", Description: operationalString("Pick according to warehouse location sequence."), IsActive: true},
			{Code: "LOT", Name: "Lot sequence", Description: operationalString("Pick according to lot-number sequence."), IsActive: true},
		} {
			if err := repos.PickingSortMethod.SeedOne(ctx, &value); err != nil {
				return err
			}
		}
		for _, value := range []model.DocumentType{
			{Code: "PURCHASE_ORDER", Name: "Purchase order", ModuleCode: "INBOUND", Description: operationalString("Client purchase order expected by the 3PL."), IsActive: true},
			{Code: "INBOUND", Name: "Inbound order", ModuleCode: "INBOUND", Description: operationalString("Expected inbound goods."), IsActive: true},
			{Code: "RECEIPT", Name: "Receipt", ModuleCode: "INBOUND", Description: operationalString("Physical receipt of goods."), IsActive: true},
			{Code: "QUALITY_INSPECTION", Name: "Quality inspection", ModuleCode: "INBOUND", Description: operationalString("Inbound quality inspection."), IsActive: true},
			{Code: "PUTAWAY_TASK", Name: "Putaway task", ModuleCode: "INBOUND", Description: operationalString("Inbound putaway work."), IsActive: true},
			{Code: "OUTBOUND", Name: "Outbound order", ModuleCode: "OUTBOUND", Description: operationalString("Outbound fulfillment order."), IsActive: true},
			{Code: "RESERVATION", Name: "Reservation", ModuleCode: "OUTBOUND", Description: operationalString("Inventory reservation."), IsActive: true},
			{Code: "PICK_TASK", Name: "Pick task", ModuleCode: "OUTBOUND", Description: operationalString("Outbound picking work."), IsActive: true},
			{Code: "PACKING", Name: "Packing", ModuleCode: "OUTBOUND", Description: operationalString("Packed outbound goods."), IsActive: true},
			{Code: "SHIPMENT", Name: "Shipment", ModuleCode: "OUTBOUND", Description: operationalString("Shipment from a warehouse."), IsActive: true},
			{Code: "OUTBOUND_VALIDATION", Name: "Outbound validation", ModuleCode: "OUTBOUND", Description: operationalString("Recorded validation of a client delivery order."), IsActive: true},
			{Code: "OUTBOUND_WAVE", Name: "Outbound wave", ModuleCode: "OUTBOUND", Description: operationalString("Group of allocated orders released for picking."), IsActive: true},
			{Code: "OUTBOUND_STAGING", Name: "Outbound staging", ModuleCode: "OUTBOUND", Description: operationalString("Picked stock staged by order."), IsActive: true},
			{Code: "OUTBOUND_CHECK", Name: "Outbound check", ModuleCode: "OUTBOUND", Description: operationalString("Quantity and item verification before packing."), IsActive: true},
			{Code: "DELIVERY", Name: "Store delivery", ModuleCode: "OUTBOUND", Description: operationalString("Shipment delivery and proof of delivery."), IsActive: true},
			{Code: "BILLING_CONTRACT", Name: "Billing contract", ModuleCode: "BILLING", Description: operationalString("Commercial billing agreement."), IsActive: true},
			{Code: "RATE_CARD", Name: "Rate card", ModuleCode: "BILLING", Description: operationalString("Versioned service pricing."), IsActive: true},
			{Code: "BILLABLE_EVENT", Name: "Billable event", ModuleCode: "BILLING", Description: operationalString("Immutable operational billing event."), IsActive: true},
			{Code: "BILLING_RUN", Name: "Billing run", ModuleCode: "BILLING", Description: operationalString("Rated events for a billing period."), IsActive: true},
			{Code: "BILLING_CHARGE", Name: "Billing charge", ModuleCode: "BILLING", Description: operationalString("Calculated event charge snapshot."), IsActive: true},
			{Code: "INVOICE", Name: "Invoice", ModuleCode: "BILLING", Description: operationalString("Client billing invoice."), IsActive: true},
			{Code: "CREDIT_NOTE", Name: "Credit note", ModuleCode: "BILLING", Description: operationalString("Issued invoice credit."), IsActive: true},
			{Code: "PAYMENT", Name: "Payment", ModuleCode: "BILLING", Description: operationalString("Client payment receipt."), IsActive: true},
			{Code: "TRANSFER", Name: "Transfer order", ModuleCode: "INVENTORY", Description: operationalString("Inter-warehouse transfer."), IsActive: true},
			{Code: "INTERNAL_MOVE", Name: "Internal movement", ModuleCode: "INVENTORY", Description: operationalString("Movement inside one warehouse."), IsActive: true},
			{Code: "TRANSFER_DISPATCH", Name: "Transfer dispatch", ModuleCode: "INVENTORY", Description: operationalString("Dispatch from a source warehouse."), IsActive: true},
			{Code: "TRANSFER_RECEIPT", Name: "Transfer receipt", ModuleCode: "INVENTORY", Description: operationalString("Receipt at a target warehouse."), IsActive: true},
			{Code: "INVENTORY_STATUS_CHANGE", Name: "Inventory status change", ModuleCode: "INVENTORY", Description: operationalString("Authorized stock status change."), IsActive: true},
			{Code: "INVENTORY_ADJUSTMENT", Name: "Inventory adjustment", ModuleCode: "INVENTORY", Description: operationalString("Authorized on-hand quantity adjustment."), IsActive: true},
			{Code: "STOCK_COUNT", Name: "Stock count", ModuleCode: "INVENTORY", Description: operationalString("Physical inventory count."), IsActive: true},
			{Code: "MOVEMENT", Name: "Inventory movement", ModuleCode: "INVENTORY", Description: operationalString("Inventory ledger movement."), IsActive: true},
			{Code: "INVENTORY_BALANCE", Name: "Inventory balance", ModuleCode: "INVENTORY", Description: operationalString("Current stock balance identity."), IsActive: true},
			{Code: "HANDLING_UNIT", Name: "Handling unit", ModuleCode: "INVENTORY", Description: operationalString("Pallet, carton, tote, or bin."), IsActive: true},
			{Code: "LOT", Name: "Inventory lot", ModuleCode: "INVENTORY", Description: operationalString("Lot-controlled stock identity."), IsActive: true},
			{Code: "SERIAL", Name: "Serial number", ModuleCode: "INVENTORY", Description: operationalString("Serial-controlled stock identity."), IsActive: true},
			{Code: "QUARANTINE_CASE", Name: "Quarantine case", ModuleCode: "INBOUND", Description: operationalString("Client disposition case for rejected stock."), IsActive: true},
			{Code: "QUARANTINE_DISPOSITION", Name: "Quarantine disposition", ModuleCode: "INBOUND", Description: operationalString("Client decision for quarantined stock."), IsActive: true},
		} {
			if err := repos.DocumentType.SeedOne(ctx, &value); err != nil {
				return err
			}
		}
		if err := repos.TaskStatus.LockConfiguration(ctx); err != nil {
			return err
		}
		for _, value := range []model.TaskStatus{
			{Code: "OPEN", Name: "Open", IsInitial: true, IsFinal: false, IsCancelled: false, IsActive: true},
			{Code: "ASSIGNED", Name: "Assigned", IsInitial: false, IsFinal: false, IsCancelled: false, IsActive: true},
			{Code: "IN_PROGRESS", Name: "In progress", IsInitial: false, IsFinal: false, IsCancelled: false, IsActive: true},
			{Code: "COMPLETED", Name: "Completed", IsInitial: false, IsFinal: true, IsCancelled: false, IsActive: true},
			{Code: "CANCELLED", Name: "Cancelled", IsInitial: false, IsFinal: true, IsCancelled: true, IsActive: true},
			{Code: "REVERSED", Name: "Reversed", IsInitial: false, IsFinal: true, IsCancelled: false, IsActive: true},
		} {
			if value.IsInitial {
				exists, err := repos.TaskStatus.InitialExists(ctx)
				if err != nil {
					return err
				}
				if exists {
					value.IsInitial = false
				}
			}
			if err := repos.TaskStatus.SeedOne(ctx, &value); err != nil {
				return err
			}
		}
		for _, row := range [][2]string{{"OPEN", "ASSIGNED"},
			{"OPEN", "IN_PROGRESS"},
			{"OPEN", "CANCELLED"},
			{"ASSIGNED", "IN_PROGRESS"},
			{"ASSIGNED", "CANCELLED"},
			{"IN_PROGRESS", "COMPLETED"},
			{"IN_PROGRESS", "CANCELLED"}} {
			from, err := repos.TaskStatus.ByCode(ctx, row[0])
			if err != nil {
				return err
			}
			to, err := repos.TaskStatus.ByCode(ctx, row[1])
			if err != nil {
				return err
			}
			if !from.IsActive || !to.IsActive || from.IsFinal {
				continue
			}
			value := model.TaskStatusTransition{FromStatusID: from.ID, ToStatusID: to.ID, IsActive: true}
			if err := repos.TaskStatusTransition.SeedOne(ctx, &value); err != nil {
				return err
			}
		}
		today, _ := parseOperationalDate(time.Now().In(local.location).Format("2006-01-02"))
		for _, row := range []struct {
			Type, Prefix       string
			Partner, Warehouse bool
		}{
			{"PURCHASE_ORDER", "PO", true, true},
			{"INBOUND", "INB", true, true},
			{"RECEIPT", "RCV", true, true},
			{"QUALITY_INSPECTION", "QIN", false, true},
			{"PUTAWAY_TASK", "PUT", false, true},
			{"OUTBOUND", "OUT", true, true},
			{"RESERVATION", "RSV", false, true},
			{"PICK_TASK", "PCK", false, true},
			{"PACKING", "PKG", false, true},
			{"SHIPMENT", "SHP", false, true},
			{"OUTBOUND_VALIDATION", "OVL", false, true},
			{"OUTBOUND_WAVE", "WAV", false, true},
			{"OUTBOUND_STAGING", "STG", false, true},
			{"OUTBOUND_CHECK", "CHK", false, true},
			{"DELIVERY", "DLV", false, true},
			{"BILLING_CONTRACT", "BCT", false, false},
			{"RATE_CARD", "RAT", false, false},
			{"BILLABLE_EVENT", "BEV", false, true},
			{"BILLING_RUN", "BRN", false, false},
			{"BILLING_CHARGE", "CHG", false, false},
			{"INVOICE", "INV", false, false},
			{"CREDIT_NOTE", "CRN", false, false},
			{"PAYMENT", "PAY", false, false},
			{"TRANSFER", "TRF", false, true},
			{"INTERNAL_MOVE", "IMV", false, true},
			{"TRANSFER_DISPATCH", "TDS", false, true},
			{"TRANSFER_RECEIPT", "TRC", false, true},
			{"INVENTORY_STATUS_CHANGE", "ISC", false, true},
			{"INVENTORY_ADJUSTMENT", "ADJ", false, true},
			{"STOCK_COUNT", "CNT", false, true},
			{"MOVEMENT", "MOV", false, true},
			{"INVENTORY_BALANCE", "BAL", false, true},
			{"HANDLING_UNIT", "HU", false, true},
			{"LOT", "LOT", false, false},
			{"SERIAL", "SER", false, false},
			{"QUARANTINE_CASE", "QCS", false, true},
			{"QUARANTINE_DISPOSITION", "QDS", false, true},
		} {
			kind, err := repos.DocumentType.ByCode(ctx, row.Type)
			if err != nil {
				return err
			}
			if _, err := repos.DocumentType.Lock(ctx, kind.ID); err != nil {
				return err
			}
			history, _, err := repos.DocumentNumberRule.List(ctx, repository.OperationalFilter{ParentID: kind.ID})
			if err != nil {
				return err
			}
			if len(history) == 0 {
				value := model.DocumentNumberRule{DocumentTypeID: kind.ID, Prefix: row.Prefix, Separator: "-", SequenceLength: 6, IncludePartnerCode: row.Partner, IncludeWarehouseCode: row.Warehouse, EffectiveFrom: today, IsActive: true}
				if err := repos.DocumentNumberRule.Create(ctx, &value); err != nil {
					return err
				}
			}
		}
		for _, row := range []struct {
			Type, Code, Name          string
			Initial, Final, Cancelled bool
			Order                     int
		}{
			{"PURCHASE_ORDER", "DRAFT", "Draft", true, false, false, 10},
			{"PURCHASE_ORDER", "APPROVED", "Approved", false, false, false, 20},
			{"PURCHASE_ORDER", "PARTIALLY_RECEIVED", "Partially received", false, false, false, 30},
			{"PURCHASE_ORDER", "RECEIVED", "Received", false, true, false, 40},
			{"PURCHASE_ORDER", "CLOSED", "Closed short", false, true, false, 50},
			{"PURCHASE_ORDER", "CANCELLED", "Cancelled", false, true, true, 90},
			{"INBOUND", "DRAFT", "Draft", true, false, false, 10},
			{"INBOUND", "RELEASED", "Released", false, false, false, 20},
			{"INBOUND", "PARTIALLY_RECEIVED", "Partially received", false, false, false, 30},
			{"INBOUND", "RECEIVED", "Received", false, true, false, 40},
			{"INBOUND", "CLOSED", "Closed short", false, true, false, 50},
			{"INBOUND", "CANCELLED", "Cancelled", false, true, true, 90},
			{"RECEIPT", "OPEN", "Open", true, false, false, 10},
			{"RECEIPT", "COMPLETED", "Completed", false, true, false, 20},
			{"RECEIPT", "REVERSED", "Reversed", false, true, false, 30},
			{"RECEIPT", "CANCELLED", "Cancelled", false, true, true, 90},
			{"OUTBOUND", "DRAFT", "Draft", true, false, false, 10},
			{"OUTBOUND", "VALIDATED", "Validated", false, false, false, 15},
			{"OUTBOUND", "RELEASED", "Released", false, false, false, 20},
			{"OUTBOUND", "PARTIALLY_ALLOCATED", "Partially allocated", false, false, false, 25},
			{"OUTBOUND", "ALLOCATED", "Allocated", false, false, false, 30},
			{"OUTBOUND", "WAVED", "Waved", false, false, false, 35},
			{"OUTBOUND", "PICKING", "Picking", false, false, false, 40},
			{"OUTBOUND", "STAGED", "Staged", false, false, false, 45},
			{"OUTBOUND", "CHECKING", "Checking", false, false, false, 50},
			{"OUTBOUND", "CHECK_FAILED", "Check failed", false, false, false, 55},
			{"OUTBOUND", "CHECKED", "Checked", false, false, false, 60},
			{"OUTBOUND", "PACKING", "Packing", false, false, false, 65},
			{"OUTBOUND", "PACKED", "Packed", false, false, false, 70},
			{"OUTBOUND", "SHIPPED", "Shipped", false, false, false, 75},
			{"OUTBOUND", "DELIVERED", "Delivered", false, true, false, 80},
			{"OUTBOUND", "DELIVERY_FAILED", "Delivery failed", false, false, false, 85},
			{"OUTBOUND", "PARTIALLY_DELIVERED", "Partially delivered", false, true, false, 87},
			{"OUTBOUND", "RETURNED", "Returned", false, true, false, 88},
			{"OUTBOUND", "CANCELLED", "Cancelled", false, true, true, 90},
			{"OUTBOUND_VALIDATION", "PENDING", "Pending", true, false, false, 10},
			{"OUTBOUND_VALIDATION", "PASSED", "Passed", false, true, false, 20},
			{"OUTBOUND_VALIDATION", "FAILED", "Failed", false, true, false, 30},
			{"RESERVATION", "ACTIVE", "Active", true, false, false, 10},
			{"RESERVATION", "PARTIALLY_PICKED", "Partially picked", false, false, false, 20},
			{"RESERVATION", "CONSUMED", "Consumed", false, true, false, 30},
			{"RESERVATION", "RELEASED", "Released", false, true, true, 90},
			{"OUTBOUND_WAVE", "DRAFT", "Draft", true, false, false, 10},
			{"OUTBOUND_WAVE", "RELEASED", "Released", false, false, false, 20},
			{"OUTBOUND_WAVE", "IN_PROGRESS", "In progress", false, false, false, 30},
			{"OUTBOUND_WAVE", "COMPLETED", "Completed", false, true, false, 40},
			{"OUTBOUND_WAVE", "CANCELLED", "Cancelled", false, true, true, 90},
			{"OUTBOUND_STAGING", "OPEN", "Open", true, false, false, 10},
			{"OUTBOUND_STAGING", "COMPLETED", "Completed", false, true, false, 20},
			{"OUTBOUND_STAGING", "CANCELLED", "Cancelled", false, true, true, 90},
			{"OUTBOUND_CHECK", "OPEN", "Open", true, false, false, 10},
			{"OUTBOUND_CHECK", "PASSED", "Passed", false, true, false, 20},
			{"OUTBOUND_CHECK", "FAILED", "Failed", false, true, false, 30},
			{"OUTBOUND_CHECK", "CANCELLED", "Cancelled", false, true, true, 90},
			{"PACKING", "OPEN", "Open", true, false, false, 10},
			{"PACKING", "COMPLETED", "Completed", false, true, false, 20},
			{"PACKING", "CANCELLED", "Cancelled", false, true, true, 90},
			{"SHIPMENT", "PLANNED", "Planned", true, false, false, 10},
			{"SHIPMENT", "SHIPPED", "Shipped", false, true, false, 20},
			{"SHIPMENT", "CANCELLED", "Cancelled", false, true, true, 90},
			{"DELIVERY", "PLANNED", "Planned", true, false, false, 10},
			{"DELIVERY", "IN_TRANSIT", "In transit", false, false, false, 20},
			{"DELIVERY", "ARRIVED", "Arrived", false, false, false, 30},
			{"DELIVERY", "FAILED", "Failed", false, false, false, 40},
			{"DELIVERY", "DELIVERED", "Delivered", false, true, false, 50},
			{"DELIVERY", "PARTIALLY_DELIVERED", "Partially delivered", false, true, false, 55},
			{"DELIVERY", "RETURNED", "Returned", false, true, false, 60},
			{"DELIVERY", "CANCELLED", "Cancelled", false, true, true, 90},
			{"BILLING_CONTRACT", "DRAFT", "Draft", true, false, false, 10},
			{"BILLING_CONTRACT", "ACTIVE", "Active", false, false, false, 20},
			{"BILLING_CONTRACT", "SUSPENDED", "Suspended", false, false, false, 30},
			{"BILLING_CONTRACT", "EXPIRED", "Expired", false, true, false, 40},
			{"BILLING_CONTRACT", "TERMINATED", "Terminated", false, true, true, 90},
			{"RATE_CARD", "DRAFT", "Draft", true, false, false, 10},
			{"RATE_CARD", "APPROVED", "Approved", false, false, false, 20},
			{"RATE_CARD", "ACTIVE", "Active", false, false, false, 30},
			{"RATE_CARD", "EXPIRED", "Expired", false, true, false, 40},
			{"RATE_CARD", "CANCELLED", "Cancelled", false, true, true, 90},
			{"BILLABLE_EVENT", "PENDING", "Pending", true, false, false, 10},
			{"BILLABLE_EVENT", "RATED", "Rated", false, true, false, 20},
			{"BILLABLE_EVENT", "EXCLUDED", "Excluded", false, true, true, 80},
			{"BILLABLE_EVENT", "CANCELLED", "Cancelled", false, true, true, 90},
			{"BILLING_RUN", "DRAFT", "Draft", true, false, false, 10},
			{"BILLING_RUN", "CALCULATED", "Calculated", false, false, false, 20},
			{"BILLING_RUN", "REVIEWED", "Reviewed", false, false, false, 30},
			{"BILLING_RUN", "INVOICED", "Invoiced", false, true, false, 40},
			{"BILLING_RUN", "CANCELLED", "Cancelled", false, true, true, 90},
			{"INVOICE", "DRAFT", "Draft", true, false, false, 10},
			{"INVOICE", "REVIEWED", "Reviewed", false, false, false, 20},
			{"INVOICE", "ISSUED", "Issued", false, false, false, 30},
			{"INVOICE", "PARTIALLY_PAID", "Partially paid", false, false, false, 40},
			{"INVOICE", "PAID", "Paid", false, true, false, 50},
			{"INVOICE", "SETTLED", "Settled by credit", false, true, false, 60},
			{"INVOICE", "VOID", "Void", false, true, true, 90},
			{"CREDIT_NOTE", "DRAFT", "Draft", true, false, false, 10},
			{"CREDIT_NOTE", "ISSUED", "Issued", false, true, false, 20},
			{"CREDIT_NOTE", "VOID", "Void", false, true, true, 90},
			{"PAYMENT", "RECEIVED", "Received", true, false, false, 10},
			{"PAYMENT", "PARTIALLY_ALLOCATED", "Partially allocated", false, false, false, 20},
			{"PAYMENT", "ALLOCATED", "Allocated", false, true, false, 30},
			{"PAYMENT", "VOID", "Void", false, true, true, 90},
			{"TRANSFER", "DRAFT", "Draft", true, false, false, 10},
			{"TRANSFER", "APPROVED", "Approved", false, false, false, 20},
			{"TRANSFER", "PARTIALLY_DISPATCHED", "Partially dispatched", false, false, false, 30},
			{"TRANSFER", "IN_TRANSIT", "In transit", false, false, false, 40},
			{"TRANSFER", "PARTIALLY_RECEIVED", "Partially received", false, false, false, 50},
			{"TRANSFER", "RECEIVED", "Received", false, true, false, 60},
			{"TRANSFER", "CANCELLED", "Cancelled", false, true, true, 90},
			{"INTERNAL_MOVE", "DRAFT", "Draft", true, false, false, 10},
			{"INTERNAL_MOVE", "APPROVED", "Approved", false, false, false, 20},
			{"INTERNAL_MOVE", "IN_PROGRESS", "In progress", false, false, false, 30},
			{"INTERNAL_MOVE", "COMPLETED", "Completed", false, true, false, 40},
			{"INTERNAL_MOVE", "CANCELLED", "Cancelled", false, true, true, 90},
			{"TRANSFER_DISPATCH", "DRAFT", "Draft", true, false, false, 10},
			{"TRANSFER_DISPATCH", "DISPATCHED", "Dispatched", false, true, false, 20},
			{"TRANSFER_DISPATCH", "CANCELLED", "Cancelled", false, true, true, 90},
			{"TRANSFER_RECEIPT", "DRAFT", "Draft", true, false, false, 10},
			{"TRANSFER_RECEIPT", "RECEIVED", "Received", false, true, false, 20},
			{"TRANSFER_RECEIPT", "CANCELLED", "Cancelled", false, true, true, 90},
			{"INVENTORY_STATUS_CHANGE", "DRAFT", "Draft", true, false, false, 10},
			{"INVENTORY_STATUS_CHANGE", "APPROVED", "Approved", false, false, false, 20},
			{"INVENTORY_STATUS_CHANGE", "POSTED", "Posted", false, true, false, 30},
			{"INVENTORY_STATUS_CHANGE", "CANCELLED", "Cancelled", false, true, true, 90},
			{"INVENTORY_ADJUSTMENT", "DRAFT", "Draft", true, false, false, 10},
			{"INVENTORY_ADJUSTMENT", "APPROVED", "Approved", false, false, false, 20},
			{"INVENTORY_ADJUSTMENT", "POSTED", "Posted", false, true, false, 30},
			{"INVENTORY_ADJUSTMENT", "CANCELLED", "Cancelled", false, true, true, 90},
			{"STOCK_COUNT", "DRAFT", "Draft", true, false, false, 10},
			{"STOCK_COUNT", "COUNTING", "Counting", false, false, false, 20},
			{"STOCK_COUNT", "REVIEW", "Review", false, false, false, 30},
			{"STOCK_COUNT", "POSTED", "Posted", false, true, false, 40},
			{"STOCK_COUNT", "CANCELLED", "Cancelled", false, true, true, 90},
			{"QUARANTINE_CASE", "OPEN", "Open", true, false, false, 10},
			{"QUARANTINE_CASE", "PARTIALLY_DECIDED", "Partially decided", false, false, false, 20},
			{"QUARANTINE_CASE", "CLOSED", "Closed", false, true, false, 30},
			{"QUARANTINE_DISPOSITION", "DECIDED", "Decided", true, false, false, 10},
			{"QUARANTINE_DISPOSITION", "PROCESSED", "Processed", false, true, false, 20},
			{"QUARANTINE_DISPOSITION", "CANCELLED", "Cancelled", false, true, true, 90},
		} {
			kind, err := repos.DocumentType.ByCode(ctx, row.Type)
			if err != nil {
				return err
			}
			if _, err := repos.DocumentType.Lock(ctx, kind.ID); err != nil {
				return err
			}
			if row.Initial {
				exists, err := repos.DocumentStatus.InitialExists(ctx, kind.ID)
				if err != nil {
					return err
				}
				if exists {
					row.Initial = false
				}
			}
			value := model.DocumentStatus{DocumentTypeID: kind.ID, Code: row.Code, Name: row.Name, IsInitial: row.Initial, IsFinal: row.Final, IsCancelled: row.Cancelled, DisplayOrder: row.Order, IsActive: true}
			if err := repos.DocumentStatus.SeedOne(ctx, &value); err != nil {
				return err
			}
		}
		for _, row := range [][3]string{{"PURCHASE_ORDER", "DRAFT", "APPROVED"},
			{"PURCHASE_ORDER", "DRAFT", "CANCELLED"},
			{"PURCHASE_ORDER", "APPROVED", "PARTIALLY_RECEIVED"},
			{"PURCHASE_ORDER", "APPROVED", "RECEIVED"},
			{"PURCHASE_ORDER", "APPROVED", "CANCELLED"},
			{"PURCHASE_ORDER", "PARTIALLY_RECEIVED", "RECEIVED"},
			{"PURCHASE_ORDER", "APPROVED", "CLOSED"},
			{"PURCHASE_ORDER", "PARTIALLY_RECEIVED", "CLOSED"},
			{"INBOUND", "DRAFT", "RELEASED"},
			{"INBOUND", "DRAFT", "CANCELLED"},
			{"INBOUND", "RELEASED", "PARTIALLY_RECEIVED"},
			{"INBOUND", "RELEASED", "RECEIVED"},
			{"INBOUND", "RELEASED", "CANCELLED"},
			{"INBOUND", "PARTIALLY_RECEIVED", "RECEIVED"},
			{"INBOUND", "PARTIALLY_RECEIVED", "CANCELLED"},
			{"INBOUND", "RELEASED", "CLOSED"},
			{"INBOUND", "PARTIALLY_RECEIVED", "CLOSED"},
			{"RECEIPT", "OPEN", "COMPLETED"},
			{"RECEIPT", "OPEN", "CANCELLED"},
			{"OUTBOUND", "DRAFT", "VALIDATED"},
			{"OUTBOUND", "DRAFT", "CANCELLED"},
			{"OUTBOUND", "VALIDATED", "RELEASED"},
			{"OUTBOUND", "VALIDATED", "CANCELLED"},
			{"OUTBOUND", "RELEASED", "PARTIALLY_ALLOCATED"},
			{"OUTBOUND", "RELEASED", "ALLOCATED"},
			{"OUTBOUND", "RELEASED", "CANCELLED"},
			{"OUTBOUND", "PARTIALLY_ALLOCATED", "ALLOCATED"},
			{"OUTBOUND", "PARTIALLY_ALLOCATED", "CANCELLED"},
			{"OUTBOUND", "ALLOCATED", "WAVED"},
			{"OUTBOUND", "ALLOCATED", "CANCELLED"},
			{"OUTBOUND", "WAVED", "PICKING"},
			{"OUTBOUND", "PICKING", "STAGED"},
			{"OUTBOUND", "STAGED", "CHECKING"},
			{"OUTBOUND", "CHECKING", "CHECKED"},
			{"OUTBOUND", "CHECKING", "CHECK_FAILED"},
			{"OUTBOUND", "CHECK_FAILED", "CHECKING"},
			{"OUTBOUND", "CHECKED", "PACKING"},
			{"OUTBOUND", "PACKING", "PACKED"},
			{"OUTBOUND", "PACKED", "SHIPPED"},
			{"OUTBOUND", "SHIPPED", "DELIVERED"},
			{"OUTBOUND", "SHIPPED", "DELIVERY_FAILED"},
			{"OUTBOUND", "DELIVERY_FAILED", "DELIVERED"},
			{"OUTBOUND", "DELIVERY_FAILED", "PARTIALLY_DELIVERED"},
			{"OUTBOUND", "DELIVERY_FAILED", "RETURNED"},
			{"OUTBOUND_VALIDATION", "PENDING", "PASSED"},
			{"OUTBOUND_VALIDATION", "PENDING", "FAILED"},
			{"RESERVATION", "ACTIVE", "PARTIALLY_PICKED"},
			{"RESERVATION", "ACTIVE", "CONSUMED"},
			{"RESERVATION", "ACTIVE", "RELEASED"},
			{"RESERVATION", "PARTIALLY_PICKED", "CONSUMED"},
			{"RESERVATION", "PARTIALLY_PICKED", "RELEASED"},
			{"OUTBOUND_WAVE", "DRAFT", "RELEASED"},
			{"OUTBOUND_WAVE", "DRAFT", "CANCELLED"},
			{"OUTBOUND_WAVE", "RELEASED", "IN_PROGRESS"},
			{"OUTBOUND_WAVE", "RELEASED", "COMPLETED"},
			{"OUTBOUND_WAVE", "IN_PROGRESS", "COMPLETED"},
			{"OUTBOUND_STAGING", "OPEN", "COMPLETED"},
			{"OUTBOUND_STAGING", "OPEN", "CANCELLED"},
			{"OUTBOUND_CHECK", "OPEN", "PASSED"},
			{"OUTBOUND_CHECK", "OPEN", "FAILED"},
			{"OUTBOUND_CHECK", "OPEN", "CANCELLED"},
			{"PACKING", "OPEN", "COMPLETED"},
			{"PACKING", "OPEN", "CANCELLED"},
			{"SHIPMENT", "PLANNED", "SHIPPED"},
			{"SHIPMENT", "PLANNED", "CANCELLED"},
			{"DELIVERY", "PLANNED", "IN_TRANSIT"},
			{"DELIVERY", "PLANNED", "CANCELLED"},
			{"DELIVERY", "IN_TRANSIT", "ARRIVED"},
			{"DELIVERY", "IN_TRANSIT", "FAILED"},
			{"DELIVERY", "ARRIVED", "DELIVERED"},
			{"DELIVERY", "ARRIVED", "FAILED"},
			{"DELIVERY", "FAILED", "IN_TRANSIT"},
			{"DELIVERY", "FAILED", "DELIVERED"},
			{"DELIVERY", "FAILED", "PARTIALLY_DELIVERED"},
			{"DELIVERY", "FAILED", "RETURNED"},
			{"BILLING_CONTRACT", "DRAFT", "ACTIVE"},
			{"BILLING_CONTRACT", "DRAFT", "TERMINATED"},
			{"BILLING_CONTRACT", "ACTIVE", "SUSPENDED"},
			{"BILLING_CONTRACT", "ACTIVE", "EXPIRED"},
			{"BILLING_CONTRACT", "ACTIVE", "TERMINATED"},
			{"BILLING_CONTRACT", "SUSPENDED", "ACTIVE"},
			{"BILLING_CONTRACT", "SUSPENDED", "TERMINATED"},
			{"RATE_CARD", "DRAFT", "APPROVED"},
			{"RATE_CARD", "DRAFT", "CANCELLED"},
			{"RATE_CARD", "APPROVED", "ACTIVE"},
			{"RATE_CARD", "APPROVED", "CANCELLED"},
			{"RATE_CARD", "ACTIVE", "EXPIRED"},
			{"BILLABLE_EVENT", "PENDING", "RATED"},
			{"BILLABLE_EVENT", "PENDING", "EXCLUDED"},
			{"BILLABLE_EVENT", "PENDING", "CANCELLED"},
			{"BILLING_RUN", "DRAFT", "CALCULATED"},
			{"BILLING_RUN", "DRAFT", "CANCELLED"},
			{"BILLING_RUN", "CALCULATED", "REVIEWED"},
			{"BILLING_RUN", "CALCULATED", "DRAFT"},
			{"BILLING_RUN", "CALCULATED", "CANCELLED"},
			{"BILLING_RUN", "REVIEWED", "INVOICED"},
			{"INVOICE", "DRAFT", "REVIEWED"},
			{"INVOICE", "DRAFT", "VOID"},
			{"INVOICE", "REVIEWED", "ISSUED"},
			{"INVOICE", "REVIEWED", "DRAFT"},
			{"INVOICE", "REVIEWED", "VOID"},
			{"INVOICE", "ISSUED", "PARTIALLY_PAID"},
			{"INVOICE", "ISSUED", "PAID"},
			{"INVOICE", "ISSUED", "VOID"},
			{"INVOICE", "ISSUED", "SETTLED"},
			{"INVOICE", "PARTIALLY_PAID", "PAID"},
			{"INVOICE", "PARTIALLY_PAID", "SETTLED"},
			{"CREDIT_NOTE", "DRAFT", "ISSUED"},
			{"CREDIT_NOTE", "DRAFT", "VOID"},
			{"PAYMENT", "RECEIVED", "PARTIALLY_ALLOCATED"},
			{"PAYMENT", "RECEIVED", "ALLOCATED"},
			{"PAYMENT", "RECEIVED", "VOID"},
			{"PAYMENT", "PARTIALLY_ALLOCATED", "ALLOCATED"},
			{"PAYMENT", "PARTIALLY_ALLOCATED", "VOID"},
			{"TRANSFER", "DRAFT", "APPROVED"},
			{"TRANSFER", "DRAFT", "CANCELLED"},
			{"TRANSFER", "APPROVED", "PARTIALLY_DISPATCHED"},
			{"TRANSFER", "APPROVED", "IN_TRANSIT"},
			{"TRANSFER", "APPROVED", "CANCELLED"},
			{"TRANSFER", "PARTIALLY_DISPATCHED", "IN_TRANSIT"},
			{"TRANSFER", "PARTIALLY_DISPATCHED", "PARTIALLY_RECEIVED"},
			{"TRANSFER", "PARTIALLY_DISPATCHED", "CANCELLED"},
			{"TRANSFER", "IN_TRANSIT", "PARTIALLY_RECEIVED"},
			{"TRANSFER", "IN_TRANSIT", "RECEIVED"},
			{"TRANSFER", "PARTIALLY_RECEIVED", "RECEIVED"},
			{"INTERNAL_MOVE", "DRAFT", "APPROVED"},
			{"INTERNAL_MOVE", "DRAFT", "CANCELLED"},
			{"INTERNAL_MOVE", "APPROVED", "IN_PROGRESS"},
			{"INTERNAL_MOVE", "APPROVED", "COMPLETED"},
			{"INTERNAL_MOVE", "APPROVED", "CANCELLED"},
			{"INTERNAL_MOVE", "IN_PROGRESS", "COMPLETED"},
			{"INTERNAL_MOVE", "IN_PROGRESS", "CANCELLED"},
			{"TRANSFER_DISPATCH", "DRAFT", "DISPATCHED"},
			{"TRANSFER_DISPATCH", "DRAFT", "CANCELLED"},
			{"TRANSFER_RECEIPT", "DRAFT", "RECEIVED"},
			{"TRANSFER_RECEIPT", "DRAFT", "CANCELLED"},
			{"INVENTORY_STATUS_CHANGE", "DRAFT", "APPROVED"},
			{"INVENTORY_STATUS_CHANGE", "DRAFT", "CANCELLED"},
			{"INVENTORY_STATUS_CHANGE", "APPROVED", "POSTED"},
			{"INVENTORY_STATUS_CHANGE", "APPROVED", "CANCELLED"},
			{"INVENTORY_ADJUSTMENT", "DRAFT", "APPROVED"},
			{"INVENTORY_ADJUSTMENT", "DRAFT", "CANCELLED"},
			{"INVENTORY_ADJUSTMENT", "APPROVED", "POSTED"},
			{"INVENTORY_ADJUSTMENT", "APPROVED", "CANCELLED"},
			{"STOCK_COUNT", "DRAFT", "COUNTING"},
			{"STOCK_COUNT", "DRAFT", "CANCELLED"},
			{"STOCK_COUNT", "COUNTING", "REVIEW"},
			{"STOCK_COUNT", "COUNTING", "CANCELLED"},
			{"STOCK_COUNT", "REVIEW", "POSTED"},
			{"STOCK_COUNT", "REVIEW", "COUNTING"},
			{"STOCK_COUNT", "REVIEW", "CANCELLED"},
			{"QUARANTINE_CASE", "OPEN", "PARTIALLY_DECIDED"},
			{"QUARANTINE_CASE", "OPEN", "CLOSED"},
			{"QUARANTINE_CASE", "PARTIALLY_DECIDED", "CLOSED"},
			{"QUARANTINE_DISPOSITION", "DECIDED", "PROCESSED"},
			{"QUARANTINE_DISPOSITION", "DECIDED", "CANCELLED"}} {
			kind, err := repos.DocumentType.ByCode(ctx, row[0])
			if err != nil {
				return err
			}
			states, _, err := repos.DocumentStatus.List(ctx, repository.OperationalFilter{ParentID: kind.ID})
			if err != nil {
				return err
			}
			var from, to model.DocumentStatus
			for _, state := range states {
				if state.Code == row[1] {
					from = state
				}
				if state.Code == row[2] {
					to = state
				}
			}
			if from.ID == "" || to.ID == "" || !from.IsActive || !to.IsActive || from.IsFinal {
				continue
			}
			value := model.DocumentStatusTransition{DocumentTypeID: kind.ID, FromStatusID: from.ID, ToStatusID: to.ID, IsActive: true}
			if err := repos.DocumentStatusTransition.SeedOne(ctx, &value); err != nil {
				return err
			}
		}
		defaultPicking := model.PickingStrategy{Code: "DEFAULT_FEFO", Name: "Default FEFO", Description: operationalString("Prefer allocatable stock with the earliest expiry date."), IsActive: true}
		if err := repos.PickingStrategy.SeedOne(ctx, &defaultPicking); err != nil {
			return err
		}
		strategies, _, err := repos.PickingStrategy.List(ctx, repository.OperationalFilter{})
		if err != nil {
			return err
		}
		for _, strategy := range strategies {
			if strategy.Code != "DEFAULT_FEFO" || strategy.OwnerID != nil || strategy.WarehouseID != nil || !strategy.IsActive {
				continue
			}
			if _, err := repos.PickingStrategy.Lock(ctx, strategy.ID); err != nil {
				return err
			}
			method, err := repos.PickingSortMethod.ByCode(ctx, "FEFO")
			if err != nil {
				return err
			}
			statuses, _, err := repos.Catalog.InventoryStatus.List(ctx, repository.CatalogFilter{Search: "AVAILABLE"})
			if err != nil {
				return err
			}
			for _, status := range statuses {
				if status.Code != "AVAILABLE" || !status.IsActive || !status.IsAllocatable || !status.IsPickable || !method.IsActive {
					continue
				}
				value := model.PickingStrategyRule{PickingStrategyID: strategy.ID, SequenceNo: 10, InventoryStatusID: &status.ID, PickingSortMethodID: method.ID, IsActive: true}
				if err := repos.PickingStrategyRule.SeedOne(ctx, &value); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
func operationalString(value string) *string { return &value }
