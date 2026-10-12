package stockcontrol

import (
	"context"
	"gorm.io/gorm"
	"testing"
	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/stock_control"
	inventorymodel "wms-api/models/inventory"
	master "wms-api/models/master"
	inventoryrepo "wms-api/repository/inventory"
	repository "wms-api/repository/stock_control"
	inventory "wms-api/services/inventory"
)

func testCountFoundStock(t *testing.T, tx *gorm.DB, suffix, warehouseID, counterID, reviewerID, storageID, otherStorageID, pickFaceID, receivingID, statusID, uomID string) {
	t.Helper()
	ctx := context.Background()
	s := NewService(repository.NewRepositories(tx))
	inv := inventory.NewService(inventoryrepo.NewRepositories(tx))
	owner := master.Organization{Code: "FOUND_" + suffix, Name: "Found stock customer", TimezoneName: "Asia/Jakarta"}
	okay(t, tx.Create(&owner).Error)
	okay(t, tx.Create(&master.WarehouseOwner{OwnerID: owner.ID, WarehouseID: warehouseID}).Error)
	for _, actor := range []string{counterID, reviewerID} {
		okay(t, tx.Create(&master.AccountOwnerAccess{AccountID: actor, OwnerID: owner.ID}).Error)
	}
	bulk := master.Item{OwnerID: owner.ID, Code: "FOUND_B_" + suffix, Name: "Found bulk", BaseUOMID: uomID}
	serial := master.Item{OwnerID: owner.ID, Code: "FOUND_S_" + suffix, Name: "Found serial", BaseUOMID: uomID, SerialControlled: true}
	lotItem := master.Item{OwnerID: owner.ID, Code: "FOUND_L_" + suffix, Name: "Found lot", BaseUOMID: uomID, LotControlled: true}
	for _, item := range []*master.Item{&bulk, &serial, &lotItem} {
		okay(t, tx.Create(item).Error)
	}
	lot := inventorymodel.InventoryLot{ID: "FOUND-LOT-" + suffix, OwnerID: owner.ID, ItemID: lotItem.ID, LotNumber: "FOUND-LOT-NO-" + suffix}
	okay(t, tx.Create(&lot).Error)
	count, err := s.CreateGrandStockOpname(ctx, dto.CreateGrandStockOpnameRequest{OwnerID: owner.ID, WarehouseID: warehouseID, BusinessDate: "2026-10-06"}, counterID)
	okay(t, err)
	if count.TotalLines != 0 || !count.StockFrozen {
		t.Fatalf("empty customer must support found-stock counting: %+v", count)
	}
	options, err := s.GetFoundStockOptions(ctx, count.ID, repository.FoundStockOptionsFilter{ItemID: lotItem.ID}, counterID)
	okay(t, err)
	if len(options.Items) != 3 || len(options.Lots) != 1 {
		t.Fatalf("found-stock options leaked or omitted customer data: %+v", options)
	}
	for _, entry := range options.Locations {
		if entry.ID == receivingID {
			t.Fatal("found-stock lookup offered receiving location")
		}
	}
	notes := "Found behind the storage rack"
	request := dto.AddFoundStockRequest{ExpectedVersion: count.VersionNo, ItemID: bulk.ID, LocationID: storageID, InventoryStatusID: statusID, CountedQty: "3", Notes: &notes}
	invalidLocation := request
	invalidLocation.LocationID = receivingID
	_, err = s.AddFoundStock(ctx, count.ID, invalidLocation, counterID)
	wants(t, err, ErrInvalidInput)
	missingLot := request
	missingLot.ItemID = lotItem.ID
	_, err = s.AddFoundStock(ctx, count.ID, missingLot, counterID)
	wants(t, err, ErrInvalidInput)
	fractional := request
	fractional.ItemID = serial.ID
	fractional.CountedQty = "1.5"
	_, err = s.AddFoundStock(ctx, count.ID, fractional, counterID)
	wants(t, err, ErrInvalidInput)
	count, err = s.AddFoundStock(ctx, count.ID, request, counterID)
	okay(t, err)
	bulkLine := count.Lines[0]
	if !bulkLine.FoundStock || bulkLine.BalanceID != nil || !bulkLine.RequiresRecount {
		t.Fatalf("found stock must be unposted and need a recount: %+v", bulkLine)
	}
	_, err = s.AddFoundStock(ctx, count.ID, request, counterID)
	wants(t, err, inventoryrepo.ErrConflict)
	request.ExpectedVersion = count.VersionNo
	_, err = s.AddFoundStock(ctx, count.ID, request, counterID)
	wants(t, err, ErrInvalidInput)
	var balances int64
	okay(t, tx.Model(&inventorymodel.InventoryBalance{}).Where("owner_id=?", owner.ID).Count(&balances).Error)
	if balances != 0 {
		t.Fatal("recording found stock created inventory before approval")
	}
	serialRequest := request
	serialRequest.ItemID = serial.ID
	serialRequest.LocationID = pickFaceID
	serialRequest.CountedQty = "2"
	count, err = s.AddFoundStock(ctx, count.ID, serialRequest, counterID)
	okay(t, err)
	serialLine := count.Lines[1]
	lotRequest := request
	lotRequest.ExpectedVersion = count.VersionNo
	lotRequest.ItemID = lotItem.ID
	lotRequest.LotID = &lot.ID
	lotRequest.CountedQty = "1"
	count, err = s.AddFoundStock(ctx, count.ID, lotRequest, counterID)
	okay(t, err)
	lotLine := count.Lines[2]
	count, err = s.RecordCycleCount(ctx, count.ID, dto.RecordCycleCountRequest{ExpectedVersion: count.VersionNo, Lines: []dto.CycleCountEntryRequest{{LineID: bulkLine.ID, CountedQty: "3"}, {LineID: serialLine.ID, CountedQty: "2"}, {LineID: lotLine.ID, CountedQty: "1"}}}, counterID)
	okay(t, err)
	if count.StatusCode != "REVIEW" || !count.StockFrozen {
		t.Fatalf("found stock recount did not reach frozen review: %+v", count)
	}
	if count.Lines[0].DiscoveryNotes == nil || *count.Lines[0].DiscoveryNotes != notes {
		t.Fatal("recount lost the original discovery notes")
	}
	_, err = s.ApproveCycleCount(ctx, count.ID, dto.CycleCountDecisionRequest{ExpectedVersion: count.VersionNo, LineIDs: []string{bulkLine.ID}}, counterID)
	wants(t, err, ErrInvalidInput)
	_, err = s.ApproveCycleCount(ctx, count.ID, dto.CycleCountDecisionRequest{ExpectedVersion: count.VersionNo, LineIDs: []string{serialLine.ID}}, reviewerID)
	wants(t, err, ErrInvalidInput)
	count, err = s.ApproveCycleCount(ctx, count.ID, dto.CycleCountDecisionRequest{ExpectedVersion: count.VersionNo, LineIDs: []string{bulkLine.ID}}, reviewerID)
	okay(t, err)
	if count.StatusCode != "PARTIALLY_POSTED" || !count.StockFrozen {
		t.Fatal("partial found-stock approval released freeze")
	}
	if count.Lines[0].ResultingBalanceID == nil {
		t.Fatal("found-stock approval did not link resulting balance")
	}
	posted, err := inv.GetBalance(ctx, *count.Lines[0].ResultingBalanceID)
	okay(t, err)
	if posted.OnHandQty != "3.000000" {
		t.Fatalf("found quantity=%s", posted.OnHandQty)
	}
	serialNumbers := []string{"FOUND-A-" + suffix, "FOUND-B-" + suffix}
	count, err = s.ReconcileCycleCountSerials(ctx, count.ID, dto.ReconcileCycleCountSerialsRequest{ExpectedVersion: count.VersionNo, LineID: serialLine.ID, FoundSerialNumbers: serialNumbers}, reviewerID)
	okay(t, err)
	if !count.StockFrozen || len(count.Lines[1].SerialReconciliations) != 2 {
		t.Fatal("found serial audit/freeze was not retained")
	}
	count, err = s.RejectCycleCount(ctx, count.ID, dto.RejectCycleCountRequest{ExpectedVersion: count.VersionNo, LineIDs: []string{lotLine.ID}, Reason: "Stock belongs in another count"}, reviewerID)
	okay(t, err)
	if count.StockFrozen || count.FinalLines != 3 {
		t.Fatal("final found-stock rejection did not release count")
	}
	request.ExpectedVersion = count.VersionNo
	request.LocationID = otherStorageID
	_, err = s.AddFoundStock(ctx, count.ID, request, counterID)
	wants(t, err, ErrInvalidInput)
	// In a normal count, a concurrent receipt into an unsnapshotted identity
	// must prevent found stock from being added a second time at approval.
	normal, err := s.CreateCycleCount(ctx, dto.CreateCycleCountRequest{BusinessDate: "2026-10-06", ToleranceQty: "10", BalanceIDs: []string{posted.ID}}, counterID)
	okay(t, err)
	request.ExpectedVersion = normal.VersionNo
	request.ItemID = lotItem.ID
	request.LotID = &lot.ID
	request.CountedQty = "1"
	normal, err = s.AddFoundStock(ctx, normal.ID, request, reviewerID)
	okay(t, err)
	newLine := normal.Lines[1]
	normal, err = s.RecordCycleCount(ctx, normal.ID, dto.RecordCycleCountRequest{ExpectedVersion: normal.VersionNo, Lines: []dto.CycleCountEntryRequest{{LineID: normal.Lines[0].ID, CountedQty: "3"}}}, counterID)
	okay(t, err)
	_, err = s.ApproveCycleCount(ctx, normal.ID, dto.CycleCountDecisionRequest{ExpectedVersion: normal.VersionNo, LineIDs: []string{newLine.ID}}, reviewerID)
	wants(t, err, ErrInvalidInput)
	_, err = s.CancelCycleCount(ctx, normal.ID, dto.CancelCycleCountRequest{ExpectedVersion: normal.VersionNo, Reason: "Found-stock test complete"}, counterID)
	okay(t, err)
	normal, err = s.CreateCycleCount(ctx, dto.CreateCycleCountRequest{BusinessDate: "2026-10-06", ToleranceQty: "10", BalanceIDs: []string{posted.ID}}, counterID)
	okay(t, err)
	request.ExpectedVersion = normal.VersionNo
	request.ItemID = bulk.ID
	request.LotID = nil
	normal, err = s.AddFoundStock(ctx, normal.ID, request, counterID)
	okay(t, err)
	newLine = normal.Lines[1]
	normal, err = s.RecordCycleCount(ctx, normal.ID, dto.RecordCycleCountRequest{ExpectedVersion: normal.VersionNo, Lines: []dto.CycleCountEntryRequest{{LineID: normal.Lines[0].ID, CountedQty: "3"}}}, counterID)
	okay(t, err)
	_, err = inv.PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "found-concurrent-" + suffix, MovementTypeCode: "RECEIVE", OwnerID: owner.ID, WarehouseID: warehouseID, BusinessDate: "2026-10-06", ItemID: bulk.ID, To: &inventorydto.BalanceDimension{LocationID: otherStorageID, InventoryStatusID: statusID}, Quantity: "1", SourceDocumentID: "CONCURRENT-RECEIPT"}, counterID)
	okay(t, err)
	_, err = s.ApproveCycleCount(ctx, normal.ID, dto.CycleCountDecisionRequest{ExpectedVersion: normal.VersionNo, LineIDs: []string{newLine.ID}}, reviewerID)
	wants(t, err, ErrInvalidInput)
	_, err = s.CancelCycleCount(ctx, normal.ID, dto.CancelCycleCountRequest{ExpectedVersion: normal.VersionNo, Reason: "Concurrent stock must be counted again"}, counterID)
	okay(t, err)
}
