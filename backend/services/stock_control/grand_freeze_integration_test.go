package stockcontrol

import (
	"context"
	"strings"
	"testing"

	"gorm.io/gorm"
	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/stock_control"
	master "wms-api/models/master"
	inventoryrepo "wms-api/repository/inventory"
	repository "wms-api/repository/stock_control"
	inventory "wms-api/services/inventory"
)

// Called by both existing-schema and fresh-schema integration fixtures. All rows
// belong to the fixture transaction and are rolled back with it.
func testGrandOpnameFreeze(t *testing.T, tx *gorm.DB, suffix string, ownerID, warehouseID, otherWarehouseID, accountID, approverID, itemID, sourceLocationID, targetLocationID, statusID string) {
	t.Helper()
	ctx := context.Background()
	service := NewService(repository.NewRepositories(tx))
	inv := inventory.NewService(inventoryrepo.NewRepositories(tx))
	balances := inventoryrepo.NewInventoryBalanceRepository(tx)
	var originalItem master.Item
	okay(t, tx.Where("item_id=?", itemID).Take(&originalItem).Error)
	countOwner := master.Organization{Code: "FREEZE_" + suffix, Name: "Freeze test customer", TimezoneName: "Asia/Jakarta"}
	okay(t, tx.Create(&countOwner).Error)
	for _, warehouse := range []string{warehouseID, otherWarehouseID} {
		okay(t, tx.Create(&master.WarehouseOwner{OwnerID: countOwner.ID, WarehouseID: warehouse}).Error)
	}
	for _, actor := range []string{accountID, approverID} {
		okay(t, tx.Create(&master.AccountOwnerAccess{AccountID: actor, OwnerID: countOwner.ID}).Error)
	}
	countItem := master.Item{OwnerID: countOwner.ID, Code: "FREEZE_" + suffix, Name: "Count freeze item", BaseUOMID: originalItem.BaseUOMID}
	newItem := master.Item{OwnerID: countOwner.ID, Code: "FREEZE_NEW_" + suffix, Name: "Unsnapshotted item", BaseUOMID: originalItem.BaseUOMID}
	serialItem := master.Item{OwnerID: countOwner.ID, Code: "FREEZE_SERIAL_" + suffix, Name: "Count freeze serialized item", BaseUOMID: originalItem.BaseUOMID, SerialControlled: true}
	for _, item := range []*master.Item{&countItem, &newItem, &serialItem} {
		okay(t, tx.Create(item).Error)
	}
	posting := func(key, owner, warehouse, item, location, quantity string) inventorydto.PostingRequest {
		return inventorydto.PostingRequest{OperationKey: "freeze-" + key + "-" + suffix, MovementTypeCode: "RECEIVE", OwnerID: owner, WarehouseID: warehouse, BusinessDate: "2026-09-07", ItemID: item, To: &inventorydto.BalanceDimension{LocationID: location, InventoryStatusID: statusID}, Quantity: quantity, SourceDocumentID: "FREEZE-TEST"}
	}
	first, err := inv.PostMovement(ctx, posting("first", countOwner.ID, warehouseID, countItem.ID, sourceLocationID, "5"), accountID)
	okay(t, err)
	second, err := inv.PostMovement(ctx, posting("second", countOwner.ID, warehouseID, countItem.ID, targetLocationID, "3"), accountID)
	okay(t, err)
	serialOne, err := inv.CreateSerial(ctx, inventorydto.CreateSerialRequest{OwnerID: countOwner.ID, ItemID: serialItem.ID, SerialNo: "FREEZE-ONE-" + suffix}, accountID)
	okay(t, err)
	serialTwo, err := inv.CreateSerial(ctx, inventorydto.CreateSerialRequest{OwnerID: countOwner.ID, ItemID: serialItem.ID, SerialNo: "FREEZE-TWO-" + suffix}, accountID)
	okay(t, err)
	serialPosting := posting("serial-one", countOwner.ID, warehouseID, serialItem.ID, sourceLocationID, "1")
	serialPosting.SerialIDs = []string{serialOne.ID}
	serialized, err := inv.PostMovement(ctx, serialPosting, accountID)
	okay(t, err)
	serialPosting.OperationKey = "freeze-serial-two-" + suffix
	serialPosting.SerialIDs = []string{serialTwo.ID}
	serialPosting.ExpectedDestinationVersion = &serialized.ToBalance.VersionNo
	serialized, err = inv.PostMovement(ctx, serialPosting, accountID)
	okay(t, err)

	okay(t, balances.SetQuantities(ctx, first.ToBalance.ID, "5", "1"))
	_, err = service.CreateGrandStockOpname(ctx, dto.CreateGrandStockOpnameRequest{OwnerID: countOwner.ID, WarehouseID: warehouseID, BusinessDate: "2026-09-07", ToleranceQty: "0"}, accountID)
	wants(t, err, ErrInvalidInput)
	if !strings.Contains(strings.ToLower(err.Error()), "reserv") {
		t.Fatalf("reserved stock did not explain why the grand count cannot start: %v", err)
	}
	okay(t, balances.SetQuantities(ctx, first.ToBalance.ID, "5", "0"))
	firstBalance, err := inv.GetBalance(ctx, first.ToBalance.ID)
	okay(t, err)
	first.ToBalance = &firstBalance

	grand, err := service.CreateGrandStockOpname(ctx, dto.CreateGrandStockOpnameRequest{OwnerID: countOwner.ID, WarehouseID: warehouseID, BusinessDate: "2026-09-07", ToleranceQty: "0"}, accountID)
	okay(t, err)
	if grand.TotalLines != 3 {
		t.Fatalf("freeze fixture expected three snapshot lines, got %d", grand.TotalLines)
	}
	byBalance := make(map[string]string)
	for _, line := range grand.Lines {
		byBalance[*line.BalanceID] = line.ID
	}
	assertFrozen := func(key string) {
		t.Helper()
		_, postErr := inv.PostMovement(ctx, posting(key, countOwner.ID, warehouseID, newItem.ID, sourceLocationID, "1"), accountID)
		wants(t, postErr, inventoryrepo.ErrStockFrozen)
	}
	assertFrozen("new-balance-blocked")
	_, err = inv.PostMovement(ctx, posting("existing-inbound-blocked", countOwner.ID, warehouseID, countItem.ID, sourceLocationID, "1"), accountID)
	wants(t, err, inventoryrepo.ErrStockFrozen)
	forged := posting("forged-count-document", countOwner.ID, warehouseID, countItem.ID, sourceLocationID, "1")
	forged.MovementTypeCode = "COUNT_CORRECTION"
	forged.SourceDocumentID = grand.ID
	forged.SourceLineID = pointer(byBalance[first.ToBalance.ID])
	forged.ExpectedDestinationVersion = &first.ToBalance.VersionNo
	_, err = inv.PostMovement(ctx, forged, approverID)
	wants(t, err, inventoryrepo.ErrStockFrozen)
	outgoing := forged
	outgoing.OperationKey = "freeze-outbound-blocked-" + suffix
	outgoing.From, outgoing.To = forged.To, nil
	outgoing.ExpectedSourceVersion, outgoing.ExpectedDestinationVersion = &first.ToBalance.VersionNo, nil
	_, err = inv.PostMovement(ctx, outgoing, accountID)
	wants(t, err, inventoryrepo.ErrStockFrozen)
	okayBefore, err := balances.Get(ctx, first.ToBalance.ID)
	okay(t, err)
	wants(t, balances.SetQuantities(ctx, first.ToBalance.ID, "5", "1"), inventoryrepo.ErrStockFrozen)
	okayAfter, err := balances.Get(ctx, first.ToBalance.ID)
	okay(t, err)
	if okayAfter.OnHandQty != okayBefore.OnHandQty || okayAfter.ReservedQty != okayBefore.ReservedQty || okayAfter.VersionNo != okayBefore.VersionNo {
		t.Fatalf("blocked reservation changed the counted balance: before=%+v after=%+v", okayBefore, okayAfter)
	}
	newBalances, total, err := balances.List(ctx, inventoryrepo.BalanceFilter{OwnerID: countOwner.ID, WarehouseID: warehouseID, ItemID: newItem.ID, IncludeZero: true, Page: 1, PageSize: 10})
	okay(t, err)
	if total != 0 || len(newBalances) != 0 {
		t.Fatal("rejected receipt created an unsnapshotted balance")
	}

	// An owner/warehouse freeze leaves both neighboring scopes usable.
	_, err = inv.PostMovement(ctx, posting("other-owner", ownerID, warehouseID, itemID, sourceLocationID, "1"), accountID)
	okay(t, err)
	var otherLocation master.WarehouseLocation
	okay(t, tx.Where("warehouse_id=?", otherWarehouseID).Order("code").Take(&otherLocation).Error)
	_, err = inv.PostMovement(ctx, posting("other-warehouse", countOwner.ID, otherWarehouseID, countItem.ID, otherLocation.ID, "1"), accountID)
	okay(t, err)

	entries := []dto.CycleCountEntryRequest{
		{LineID: byBalance[first.ToBalance.ID], CountedQty: "4"},
		{LineID: byBalance[second.ToBalance.ID], CountedQty: "4"},
		{LineID: byBalance[serialized.ToBalance.ID], CountedQty: "1"},
	}
	grand, err = service.RecordCycleCount(ctx, grand.ID, dto.RecordCycleCountRequest{ExpectedVersion: grand.VersionNo, Lines: entries}, accountID)
	okay(t, err)
	assertFrozen("counting-blocked")
	grand, err = service.RecordCycleCount(ctx, grand.ID, dto.RecordCycleCountRequest{ExpectedVersion: grand.VersionNo, Lines: entries}, accountID)
	okay(t, err)
	if grand.StatusCode != "REVIEW" {
		t.Fatalf("grand freeze count did not reach review: %s", grand.StatusCode)
	}
	_, err = service.ApproveCycleCount(ctx, grand.ID, dto.CycleCountDecisionRequest{ExpectedVersion: grand.VersionNo, LineIDs: []string{entries[0].LineID}}, accountID)
	wants(t, err, ErrInvalidInput)
	_, err = service.ReconcileCycleCountSerials(ctx, grand.ID, dto.ReconcileCycleCountSerialsRequest{ExpectedVersion: grand.VersionNo, LineID: entries[2].LineID, MissingSerialIDs: []string{serialTwo.ID}}, accountID)
	wants(t, err, ErrInvalidInput)
	grand, err = service.ApproveCycleCount(ctx, grand.ID, dto.CycleCountDecisionRequest{ExpectedVersion: grand.VersionNo, LineIDs: []string{entries[0].LineID}}, approverID)
	okay(t, err)
	if grand.StatusCode != "PARTIALLY_POSTED" {
		t.Fatalf("partially approved grand count status=%s", grand.StatusCode)
	}
	assertFrozen("partially-posted-blocked")
	grand, err = service.ReconcileCycleCountSerials(ctx, grand.ID, dto.ReconcileCycleCountSerialsRequest{ExpectedVersion: grand.VersionNo, LineID: entries[2].LineID, MissingSerialIDs: []string{serialTwo.ID}}, approverID)
	okay(t, err)
	if grand.StatusCode != "PARTIALLY_POSTED" {
		t.Fatalf("serial reconciliation released the count before its last line: %s", grand.StatusCode)
	}
	assertFrozen("serial-partial-blocked")
	grand, err = service.ApproveCycleCount(ctx, grand.ID, dto.CycleCountDecisionRequest{ExpectedVersion: grand.VersionNo, LineIDs: []string{entries[1].LineID}}, approverID)
	okay(t, err)
	if grand.StatusCode != "POSTED" || grand.FinalLines != 3 {
		t.Fatalf("grand count completion=%+v", grand)
	}
	for id, want := range map[string]string{first.ToBalance.ID: "4.000000", second.ToBalance.ID: "4.000000", serialized.ToBalance.ID: "1.000000"} {
		balance, balanceErr := balances.Get(ctx, id)
		okay(t, balanceErr)
		if balance.OnHandQty != want {
			t.Fatalf("approved correction balance %s = %s, want %s", id, balance.OnHandQty, want)
		}
	}
	_, err = inv.GetSerialState(ctx, serialTwo.ID)
	wants(t, err, inventoryrepo.ErrNotFound)
	_, err = inv.PostMovement(ctx, posting("completed-released", countOwner.ID, warehouseID, newItem.ID, sourceLocationID, "1"), accountID)
	okay(t, err)

	grand, err = service.CreateGrandStockOpname(ctx, dto.CreateGrandStockOpnameRequest{OwnerID: countOwner.ID, WarehouseID: warehouseID, BusinessDate: "2026-09-07"}, accountID)
	okay(t, err)
	assertFrozen("second-grand-blocked")
	// Simulate an active grand document from before the freeze migration, then
	// prove migration backfill restores enforcement without changing the count.
	okay(t, tx.Exec("DELETE FROM inventory_stock_freeze WHERE owner_id=? AND warehouse_id=?", countOwner.ID, warehouseID).Error)
	okay(t, repository.MigrateGrandStockOpnameFreeze(tx))
	assertFrozen("backfilled-grand-blocked")
	grand, err = service.CancelCycleCount(ctx, grand.ID, dto.CancelCycleCountRequest{ExpectedVersion: grand.VersionNo, Reason: "Freeze test complete"}, accountID)
	okay(t, err)
	if grand.StatusCode != "CANCELLED" {
		t.Fatalf("cancelled grand status=%s", grand.StatusCode)
	}
	_, err = inv.PostMovement(ctx, posting("cancelled-released", countOwner.ID, warehouseID, countItem.ID, sourceLocationID, "1"), accountID)
	okay(t, err)
}
