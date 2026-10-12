package stockcontrol

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/stock_control"
	model "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
	repository "wms-api/repository/stock_control"
	inventory "wms-api/services/inventory"
)

const maxGrandStockOpnameLines = 5000

func cycleCountID(prefix string) (string, error) {
	data := make([]byte, 10)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return prefix + "-" + strings.ToUpper(hex.EncodeToString(data)), nil
}
func mapCycleCount(row repository.CycleCountRow) dto.CycleCountResponse {
	return dto.CycleCountResponse{ID: row.ID, CountTypeCode: row.CountTypeCode, StockFrozen: row.CountTypeCode == "GRAND" && row.CompletedAt == nil && row.CancelledAt == nil, StatusCode: row.StatusCode, OwnerID: row.OwnerID, OwnerCode: row.OwnerCode, OwnerName: row.OwnerName, WarehouseID: row.WarehouseID, WarehouseCode: row.WarehouseCode, WarehouseName: row.WarehouseName, BusinessDate: row.BusinessDate.Format("2006-01-02"), ToleranceQty: row.ToleranceQty, BlindCount: row.BlindCount, Notes: row.Notes, TotalLines: row.TotalLines, OpenLines: row.OpenLines, CountedLines: row.CountedLines, RecountLines: row.RecountLines, FinalLines: row.FinalLines, Lines: []dto.CycleCountLineResponse{}, CompletedAt: row.CompletedAt, CancelledAt: row.CancelledAt, CancellationReason: row.CancellationReason, CreatedAt: row.CreatedAt, CreatedBy: row.CreatedBy, CreatedByDisplayName: row.CreatedByDisplayName, VersionNo: row.VersionNo}
}
func mapCycleCountLine(row repository.CycleCountLineRow, reveal bool) dto.CycleCountLineResponse {
	var system *string
	if reveal {
		value := row.SystemQty
		system = &value
	}
	return dto.CycleCountLineResponse{ID: row.ID, LineNo: row.LineNo, DecisionCode: row.DecisionCode, BalanceID: row.BalanceID, SnapshotVersionNo: row.SnapshotVersionNo, CurrentBalanceVersionNo: row.CurrentBalanceVersionNo, ItemCode: row.ItemCode, ItemName: row.ItemName, LocationCode: row.LocationCode, InventoryStatusCode: row.InventoryStatusCode, UOMCode: row.UOMCode, LotNumber: row.LotNumber, HandlingUnitBarcode: row.HandlingUnitBarcode, SerialControlled: row.SerialControlled, SystemQty: system, CountedQty: row.CountedQty, VarianceQty: func() *string {
		if reveal {
			return row.VarianceQty
		}
		return nil
	}(), CountAttempts: row.CountAttempts, RequiresRecount: row.RequiresRecount, CountedAt: row.CountedAt, CountedByDisplayName: row.CountedByDisplayName, CountNotes: row.CountNotes, DecidedAt: row.DecidedAt, DecidedByDisplayName: row.DecidedByDisplayName, DecisionReason: row.DecisionReason, InventoryMovementID: row.InventoryMovementID, ResultingBalanceID: row.ResultingBalanceID, CurrentSerials: []dto.CycleCountSerialResponse{}, SerialReconciliations: []dto.CycleCountSerialReconciliationResponse{}}
}
func (s *Service) GetCycleCount(ctx context.Context, id, actor string) (dto.CycleCountResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" || len(id) > 140 {
		return dto.CycleCountResponse{}, invalid("invalid cycle_count_id")
	}
	row, err := s.repositories.CycleCounts.Get(ctx, id)
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	if err = authorize(ctx, s.repositories, actor, scopeBalance(row.OwnerID, row.WarehouseID)); err != nil {
		return dto.CycleCountResponse{}, err
	}
	lines, err := s.repositories.CycleCounts.Lines(ctx, id)
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	response := mapCycleCount(row)
	reveal := row.StatusCode == "REVIEW" || row.StatusCode == "PARTIALLY_POSTED" || row.StatusCode == "POSTED" || row.StatusCode == "CANCELLED"
	reconciliations, err := s.repositories.CycleCounts.SerialReconciliations(ctx, id)
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	reconciledByLine := make(map[string][]dto.CycleCountSerialReconciliationResponse)
	for _, reconciliation := range reconciliations {
		reconciledByLine[reconciliation.CycleCountLineID] = append(reconciledByLine[reconciliation.CycleCountLineID], dto.CycleCountSerialReconciliationResponse{ID: reconciliation.ID, SerialID: reconciliation.SerialID, SerialNo: reconciliation.SerialNo, DirectionCode: reconciliation.DirectionCode, InventoryMovementID: reconciliation.InventoryMovementID, CreatedAt: reconciliation.CreatedAt, CreatedByDisplayName: reconciliation.CreatedByDisplayName})
	}
	for _, line := range lines {
		mapped := mapCycleCountLine(line, reveal)
		mapped.SerialReconciliations = reconciledByLine[line.ID]
		if mapped.SerialReconciliations == nil {
			mapped.SerialReconciliations = []dto.CycleCountSerialReconciliationResponse{}
		}
		mapped.FoundStock = line.FoundStock
		mapped.CreatedBy = line.CreatedBy
		mapped.CountedBy = line.CountedBy
		mapped.CreatedByDisplayName = line.CreatedByDisplayName
		mapped.DiscoveryNotes = line.DiscoveryNotes
		if reveal && line.SerialControlled && line.DecisionCode == "COUNTED" && line.BalanceID != nil {
			serials, serialErr := s.repositories.CycleCounts.SerialsByBalance(ctx, *line.BalanceID)
			if serialErr != nil {
				return dto.CycleCountResponse{}, serialErr
			}
			for _, serial := range serials {
				mapped.CurrentSerials = append(mapped.CurrentSerials, dto.CycleCountSerialResponse{SerialID: serial.ID, SerialNo: serial.SerialNo})
			}
		}
		response.Lines = append(response.Lines, mapped)
	}
	return response, nil
}

func normalizeFoundSerialNumber(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || !utf8.ValidString(value) || utf8.RuneCountInString(value) > 120 {
		return "", invalid("found serial numbers must be non-empty and at most 120 characters")
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return "", invalid("found serial numbers cannot contain control characters")
		}
	}
	return value, nil
}

func serialVarianceUnits(value string) (int, int, error) {
	variance, ok := new(big.Rat).SetString(value)
	if !ok || variance.Sign() == 0 || !variance.IsInt() {
		return 0, 0, invalid("serialized count variance must be a non-zero whole number")
	}
	direction := variance.Sign()
	if direction < 0 {
		variance.Abs(variance)
	}
	if !variance.Num().IsInt64() || variance.Num().Int64() > 5000 {
		return 0, 0, invalid("serialized count variance exceeds the 5000-unit reconciliation limit")
	}
	return direction, int(variance.Num().Int64()), nil
}
func (s *Service) ListCycleCounts(ctx context.Context, f repository.CycleCountFilter, actor string) (dto.PageResponse[dto.CycleCountResponse], error) {
	f.OwnerID = strings.ToLower(strings.TrimSpace(f.OwnerID))
	f.WarehouseID = strings.ToLower(strings.TrimSpace(f.WarehouseID))
	f.StatusCode = strings.ToUpper(strings.TrimSpace(f.StatusCode))
	f.CountTypeCode = strings.ToUpper(strings.TrimSpace(f.CountTypeCode))
	f.Search = strings.TrimSpace(f.Search)
	if f.OwnerID == "" || f.WarehouseID == "" || f.Page < 1 || f.PageSize < 1 || f.PageSize > 100 {
		return dto.PageResponse[dto.CycleCountResponse]{}, invalid("invalid cycle count filters")
	}
	if f.CountTypeCode != "" && f.CountTypeCode != "CYCLE" && f.CountTypeCode != "GRAND" {
		return dto.PageResponse[dto.CycleCountResponse]{}, invalid("count_type_code must be CYCLE or GRAND")
	}
	allowed, err := s.repositories.Scope.Allowed(ctx, actor, f.OwnerID, f.WarehouseID)
	if err != nil {
		return dto.PageResponse[dto.CycleCountResponse]{}, err
	}
	if !allowed {
		return dto.PageResponse[dto.CycleCountResponse]{}, ErrForbidden
	}
	rows, total, err := s.repositories.CycleCounts.List(ctx, f)
	if err != nil {
		return dto.PageResponse[dto.CycleCountResponse]{}, err
	}
	items := make([]dto.CycleCountResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapCycleCount(row))
	}
	return replenishmentPage(items, f.Page, f.PageSize, total), nil
}
func (s *Service) CreateCycleCount(ctx context.Context, q dto.CreateCycleCountRequest, actor string) (dto.CycleCountResponse, error) {
	if len(q.BalanceIDs) == 0 || len(q.BalanceIDs) > 100 {
		return dto.CycleCountResponse{}, invalid("cycle count requires 1 to 100 balances")
	}
	tolerance := q.ToleranceQty
	if strings.TrimSpace(tolerance) == "" {
		tolerance = "0"
	}
	tolerance, _, err := number(tolerance, false)
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	date, err := time.Parse("2006-01-02", q.BusinessDate)
	if err != nil {
		return dto.CycleCountResponse{}, invalid("business_date must be YYYY-MM-DD")
	}
	id, err := cycleCountID("CC")
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		doc, err := r.CycleCounts.DocumentType(ctx)
		if err != nil {
			return invalid("STOCK_COUNT document type is not configured")
		}
		status, err := r.CycleCounts.Status(ctx, doc.ID, "DRAFT")
		if err != nil {
			return invalid("STOCK_COUNT DRAFT status is not configured")
		}
		seen := map[string]bool{}
		lines := make([]model.CycleCountLine, 0, len(q.BalanceIDs))
		var ownerID, warehouseID string
		for index, balanceID := range q.BalanceIDs {
			balanceID = strings.TrimSpace(balanceID)
			if balanceID == "" || seen[balanceID] {
				return invalid("balance_ids must be unique and non-empty")
			}
			seen[balanceID] = true
			balance, err := r.Balances.Get(ctx, balanceID)
			if err != nil {
				return err
			}
			if err = validateStorageLocation(ctx, ir, balance.WarehouseID, balance.LocationID, "cycle count"); err != nil {
				return err
			}
			if index == 0 {
				ownerID, warehouseID = balance.OwnerID, balance.WarehouseID
				if err = authorize(ctx, r, actor, balance); err != nil {
					return err
				}
			} else if balance.OwnerID != ownerID || balance.WarehouseID != warehouseID {
				return invalid("all cycle count balances must belong to one owner and warehouse")
			}
			lines = append(lines, model.CycleCountLine{ID: fmt.Sprintf("%s-L%04d", id, index+1), CycleCountID: id, LineNo: index + 1, BalanceID: &balance.ID, SnapshotVersionNo: balance.VersionNo, ItemID: balance.ItemID, LotID: balance.LotID, HandlingUnitID: balance.HandlingUnitID, LocationID: balance.LocationID, InventoryStatusID: balance.InventoryStatusID, UOMID: balance.UOMID, SystemQty: balance.OnHandQty, DecisionCode: "OPEN", CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1})
		}
		header := &model.CycleCount{ID: id, CountTypeCode: "CYCLE", DocumentTypeID: doc.ID, StatusID: status.ID, OwnerID: ownerID, WarehouseID: warehouseID, BusinessDate: date, ToleranceQty: tolerance, BlindCount: true, Notes: q.Notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		return r.CycleCounts.Create(ctx, header, lines)
	})
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	return s.GetCycleCount(ctx, id, actor)
}

func (s *Service) CreateGrandStockOpname(ctx context.Context, q dto.CreateGrandStockOpnameRequest, actor string) (dto.CycleCountResponse, error) {
	ownerID := strings.ToLower(strings.TrimSpace(q.OwnerID))
	warehouseID := strings.ToLower(strings.TrimSpace(q.WarehouseID))
	if ownerID == "" || warehouseID == "" {
		return dto.CycleCountResponse{}, invalid("owner_id and warehouse_id are required")
	}
	tolerance := strings.TrimSpace(q.ToleranceQty)
	if tolerance == "" {
		tolerance = "0"
	}
	tolerance, _, err := number(tolerance, false)
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	date, err := time.Parse("2006-01-02", q.BusinessDate)
	if err != nil {
		return dto.CycleCountResponse{}, invalid("business_date must be YYYY-MM-DD")
	}
	id, err := cycleCountID("GSO")
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, _ *inventoryrepo.Repositories) error {
		served, err := r.Scope.WarehouseServesOwner(ctx, ownerID, warehouseID)
		if err != nil {
			return err
		}
		if !served {
			return invalid("warehouse does not actively serve the selected owner")
		}
		if err = authorize(ctx, r, actor, scopeBalance(ownerID, warehouseID)); err != nil {
			return err
		}
		// Exclude concurrent stock writers before checking reservations and taking
		// the snapshot. Keep this lock until the document and freeze commit together.
		if err = r.Balances.LockCountScope(ctx, ownerID, warehouseID); err != nil {
			return err
		}
		active, err := r.CycleCounts.ActiveGrandExists(ctx, ownerID, warehouseID)
		if err != nil {
			return err
		}
		if active {
			return invalid("an active grand stock opname already exists for this owner and warehouse")
		}
		reserved, err := r.Balances.HasReservedStock(ctx, ownerID, warehouseID)
		if err != nil {
			return err
		}
		if reserved {
			return invalid("finish or cancel stock reservations and replenishments for this customer and warehouse before starting grand stock opname")
		}
		balances, err := r.CycleCounts.GrandBalances(ctx, ownerID, warehouseID, maxGrandStockOpnameLines+1)
		if err != nil {
			return err
		}
		if len(balances) > maxGrandStockOpnameLines {
			return invalid("grand stock opname exceeds the 5000-line document limit")
		}
		doc, err := r.CycleCounts.DocumentType(ctx)
		if err != nil {
			return invalid("STOCK_COUNT document type is not configured")
		}
		status, err := r.CycleCounts.Status(ctx, doc.ID, "DRAFT")
		if err != nil {
			return invalid("STOCK_COUNT DRAFT status is not configured")
		}
		lines := make([]model.CycleCountLine, 0, len(balances))
		for index, balance := range balances {
			lines = append(lines, model.CycleCountLine{ID: fmt.Sprintf("%s-L%04d", id, index+1), CycleCountID: id, LineNo: index + 1, BalanceID: &balance.ID, SnapshotVersionNo: balance.VersionNo, ItemID: balance.ItemID, LotID: balance.LotID, HandlingUnitID: balance.HandlingUnitID, LocationID: balance.LocationID, InventoryStatusID: balance.InventoryStatusID, UOMID: balance.UOMID, SystemQty: balance.OnHandQty, DecisionCode: "OPEN", CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1})
		}
		header := &model.CycleCount{ID: id, CountTypeCode: "GRAND", DocumentTypeID: doc.ID, StatusID: status.ID, OwnerID: ownerID, WarehouseID: warehouseID, BusinessDate: date, ToleranceQty: tolerance, BlindCount: true, Notes: q.Notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		if err = r.CycleCounts.Create(ctx, header, lines); err == inventoryrepo.ErrConflict {
			return invalid("an active grand stock opname already exists for this owner and warehouse")
		}
		if err != nil {
			return err
		}
		return r.Balances.BeginCountFreeze(ctx, ownerID, warehouseID, id)
	})
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	return s.GetCycleCount(ctx, id, actor)
}
func cycleCountHeaderStatus(open, counted, recount, final int64) string {
	if open > 0 || recount > 0 {
		return "COUNTING"
	}
	if counted > 0 && final > 0 {
		return "PARTIALLY_POSTED"
	}
	if counted > 0 {
		return "REVIEW"
	}
	return "POSTED"
}
func updateCycleCountHeader(ctx context.Context, r *repository.Repositories, header model.CycleCount, actor string, extra map[string]any) error {
	open, counted, recount, final, err := r.CycleCounts.Progress(ctx, header.ID)
	if err != nil {
		return err
	}
	code := cycleCountHeaderStatus(open, counted, recount, final)
	status, err := r.CycleCounts.Status(ctx, header.DocumentTypeID, code)
	if err != nil {
		return invalid("STOCK_COUNT " + code + " status is not configured")
	}
	values := map[string]any{"status_id": status.ID, "updated_by": actor}
	for k, v := range extra {
		values[k] = v
	}
	if open == 0 && counted == 0 {
		values["completed_at"] = time.Now()
	}
	if err = r.CycleCounts.UpdateHeader(ctx, header.ID, header.VersionNo, values); err != nil {
		return err
	}
	if header.CountTypeCode == "GRAND" && open == 0 && counted == 0 {
		return r.Balances.ReleaseCountFreeze(ctx, header.ID)
	}
	return nil
}

func countPostingContext(ctx context.Context, header model.CycleCount) context.Context {
	if header.CountTypeCode == "GRAND" {
		return inventoryrepo.WithCountReconciliation(ctx, header.ID)
	}
	return ctx
}
func (s *Service) RecordCycleCount(ctx context.Context, id string, q dto.RecordCycleCountRequest, actor string) (dto.CycleCountResponse, error) {
	if len(q.Lines) == 0 {
		return dto.CycleCountResponse{}, invalid("at least one count line is required")
	}
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		header, err := r.CycleCounts.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := r.CycleCounts.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "DRAFT" && row.StatusCode != "COUNTING" {
			return invalid("only an active counting document accepts count entries")
		}
		if header.VersionNo != q.ExpectedVersion {
			return inventoryrepo.ErrConflict
		}
		if err = authorize(ctx, r, actor, scopeBalance(header.OwnerID, header.WarehouseID)); err != nil {
			return err
		}
		ids := make([]string, len(q.Lines))
		entries := map[string]dto.CycleCountEntryRequest{}
		for i, entry := range q.Lines {
			if _, ok := entries[entry.LineID]; ok {
				return invalid("duplicate cycle count line")
			}
			entries[entry.LineID] = entry
			ids[i] = entry.LineID
		}
		lines, err := r.CycleCounts.LockLines(ctx, id, ids)
		if err != nil {
			return err
		}
		if len(lines) != len(ids) {
			return invalid("one or more lines do not belong to this cycle count")
		}
		tolerance, ok := new(big.Rat).SetString(header.ToleranceQty)
		if !ok {
			return fmt.Errorf("invalid stored tolerance")
		}
		for _, line := range lines {
			if line.DecisionCode != "OPEN" && !(line.DecisionCode == "COUNTED" && line.RequiresRecount) {
				return invalid("one or more lines cannot be counted again")
			}
			entryRequest := entries[line.ID]
			counted, countedNumber, err := number(entryRequest.CountedQty, false)
			if err != nil {
				return err
			}
			balance, err := loadCountBalance(ctx, r, ir, header, line, false)
			if err != nil {
				return err
			}
			if err = validateCountStorageLocation(ctx, ir, balance.WarehouseID, balance.LocationID, "cycle count", header.CountTypeCode == "GRAND"); err != nil {
				return err
			}
			if balance.SerialControlled && !countedNumber.IsInt() {
				return invalid("serialized counts must be a whole number of base units")
			}
			if line.FoundStock && balance.SerialControlled && countedNumber.Cmp(big.NewRat(5000, 1)) > 0 {
				return invalid("serialized found stock cannot exceed 5000 base units")
			}
			system, ok := new(big.Rat).SetString(line.SystemQty)
			if !ok {
				return fmt.Errorf("invalid stored system quantity")
			}
			variance := new(big.Rat).Sub(countedNumber, system)
			attempt := line.CountAttempts + 1
			line.RequiresRecount = attempt == 1 && new(big.Rat).Abs(variance).Cmp(tolerance) > 0
			entry := model.CycleCountEntry{ID: fmt.Sprintf("%s-A%02d", line.ID, attempt), CycleCountLineID: line.ID, AttemptNo: attempt, SystemQty: line.SystemQty, SnapshotVersionNo: line.SnapshotVersionNo, CountedQty: counted, VarianceQty: variance.FloatString(6), Notes: entryRequest.Notes, CountedBy: actor}
			if err = r.CycleCounts.Record(ctx, line, entry, actor); err != nil {
				return err
			}
		}
		return updateCycleCountHeader(ctx, r, header, actor, nil)
	})
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	return s.GetCycleCount(ctx, id, actor)
}
func (s *Service) ApproveCycleCount(ctx context.Context, id string, q dto.CycleCountDecisionRequest, actor string) (dto.CycleCountResponse, error) {
	ids, err := normalizeLineIDs(q.LineIDs)
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		header, err := r.CycleCounts.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := r.CycleCounts.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "REVIEW" && row.StatusCode != "PARTIALLY_POSTED" {
			return invalid("cycle count must be ready for review")
		}
		if header.VersionNo != q.ExpectedVersion {
			return inventoryrepo.ErrConflict
		}
		if header.CreatedBy == actor {
			return invalid("the cycle count creator cannot approve it")
		}
		if err = authorize(ctx, r, actor, scopeBalance(header.OwnerID, header.WarehouseID)); err != nil {
			return err
		}
		lines, err := r.CycleCounts.LockLines(ctx, id, ids)
		if err != nil {
			return err
		}
		if len(lines) != len(ids) {
			return invalid("one or more lines do not belong to this cycle count")
		}
		for _, line := range lines {
			if line.DecisionCode != "COUNTED" || line.RequiresRecount || line.CountedQty == nil || line.VarianceQty == nil {
				return invalid("one or more lines are not ready for approval")
			}
			if countLineSelfReview(header, line, actor) {
				return invalid("found stock must be reviewed by a different account from its recorder and counter")
			}
			balance, err := loadCountBalance(ctx, r, ir, header, line, true)
			if err != nil {
				return err
			}
			variance, ok := new(big.Rat).SetString(*line.VarianceQty)
			if !ok {
				return fmt.Errorf("invalid stored variance")
			}
			if variance.Sign() == 0 {
				if err = r.CycleCounts.Decide(ctx, line.ID, "NO_VARIANCE", actor, "Count matched system quantity", nil, nil); err != nil {
					return err
				}
				continue
			}
			if balance.SerialControlled {
				return invalid("serial-controlled count variances require serial reconciliation")
			}
			reason := "COUNT_VARIANCE"
			note := "Cycle count " + header.ID + ", line " + fmt.Sprint(line.LineNo)
			reasonID, err := resolveReason(ctx, r, dto.CommandBase{ReasonCode: &reason, Notes: &note}, true)
			if err != nil {
				return err
			}
			posting := inventorydto.PostingRequest{OperationKey: "stock.cycle-count." + header.ID + "." + line.ID, MovementTypeCode: "COUNT_CORRECTION", OwnerID: header.OwnerID, WarehouseID: header.WarehouseID, BusinessDate: header.BusinessDate.Format("2006-01-02"), ItemID: line.ItemID, LotID: line.LotID, HandlingUnitID: line.HandlingUnitID, Quantity: new(big.Rat).Abs(variance).FloatString(6), SourceDocumentID: header.ID, SourceLineID: &line.ID, ReasonCodeID: reasonID, Notes: &note}
			expected := line.SnapshotVersionNo
			if variance.Sign() > 0 {
				posting.To = dimension(balance)
				if expected > 0 {
					posting.ExpectedDestinationVersion = &expected
				}
			} else {
				posting.From = dimension(balance)
				posting.ExpectedSourceVersion = &expected
			}
			posted, err := inventory.NewService(ir).PostMovement(countPostingContext(ctx, header), posting, actor)
			if err != nil {
				return err
			}
			resultID := balance.ID
			if posted.FromBalance != nil {
				resultID = posted.FromBalance.ID
			}
			if posted.ToBalance != nil {
				resultID = posted.ToBalance.ID
			}
			if err = r.CycleCounts.Decide(ctx, line.ID, "POSTED", actor, "Approved count variance", &posted.Movement.ID, &resultID); err != nil {
				return err
			}
		}
		return updateCycleCountHeader(ctx, r, header, actor, nil)
	})
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	return s.GetCycleCount(ctx, id, actor)
}

func (s *Service) ReconcileCycleCountSerials(ctx context.Context, id string, q dto.ReconcileCycleCountSerialsRequest, actor string) (dto.CycleCountResponse, error) {
	id = strings.TrimSpace(id)
	q.LineID = strings.TrimSpace(q.LineID)
	if id == "" || len(id) > 140 || q.LineID == "" || len(q.LineID) > 180 {
		return dto.CycleCountResponse{}, invalid("invalid cycle count or line identifier")
	}
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		header, err := r.CycleCounts.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := r.CycleCounts.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "REVIEW" && row.StatusCode != "PARTIALLY_POSTED" {
			return invalid("cycle count must be ready for review")
		}
		if header.VersionNo != q.ExpectedVersion {
			return inventoryrepo.ErrConflict
		}
		if header.CreatedBy == actor {
			return invalid("the cycle count creator cannot reconcile serials")
		}
		if err = authorize(ctx, r, actor, scopeBalance(header.OwnerID, header.WarehouseID)); err != nil {
			return err
		}
		lines, err := r.CycleCounts.LockLines(ctx, id, []string{q.LineID})
		if err != nil {
			return err
		}
		if len(lines) != 1 {
			return invalid("cycle count line does not belong to this document")
		}
		line := lines[0]
		if line.DecisionCode != "COUNTED" || line.RequiresRecount || line.CountedQty == nil || line.VarianceQty == nil {
			return invalid("cycle count line is not ready for serial reconciliation")
		}
		if countLineSelfReview(header, line, actor) {
			return invalid("found stock must be reconciled by a different account from its recorder and counter")
		}
		balance, err := loadCountBalance(ctx, r, ir, header, line, true)
		if err != nil {
			return err
		}
		if !balance.SerialControlled {
			return invalid("cycle count line is not serial-controlled")
		}
		direction, units, err := serialVarianceUnits(*line.VarianceQty)
		if err != nil {
			return err
		}
		if direction < 0 && (len(q.MissingSerialIDs) != units || len(q.FoundSerialNumbers) != 0) {
			return invalid(fmt.Sprintf("select exactly %d missing serials", units))
		}
		if direction > 0 && (len(q.FoundSerialNumbers) != units || len(q.MissingSerialIDs) != 0) {
			return invalid(fmt.Sprintf("enter exactly %d found serial numbers", units))
		}

		reason := "COUNT_VARIANCE"
		note := "Cycle count " + header.ID + ", line " + fmt.Sprint(line.LineNo) + ": serial reconciliation"
		reasonID, err := resolveReason(ctx, r, dto.CommandBase{ReasonCode: &reason, Notes: &note}, true)
		if err != nil {
			return err
		}
		expectedVersion := line.SnapshotVersionNo
		resultBalanceID := balance.ID
		lastMovementID := ""
		seen := make(map[string]struct{}, units)
		for index := 0; index < units; index++ {
			serialID := ""
			directionCode := "FOUND"
			if direction < 0 {
				directionCode = "MISSING"
				serialID = strings.TrimSpace(q.MissingSerialIDs[index])
				if serialID == "" || len(serialID) > 160 {
					return invalid("missing_serial_ids must contain valid serial identifiers")
				}
				if _, duplicate := seen[serialID]; duplicate {
					return invalid("missing_serial_ids must be unique")
				}
				seen[serialID] = struct{}{}
				state, stateErr := ir.SerialState.Get(ctx, serialID)
				if stateErr != nil {
					if errors.Is(stateErr, inventoryrepo.ErrNotFound) {
						return invalid("a selected missing serial is not currently in inventory")
					}
					return stateErr
				}
				if state.OwnerID != header.OwnerID || state.ItemID != line.ItemID || state.BalanceID != balance.ID {
					return invalid("a selected missing serial is not in the counted balance")
				}
			} else {
				serialNo, normalizeErr := normalizeFoundSerialNumber(q.FoundSerialNumbers[index])
				if normalizeErr != nil {
					return normalizeErr
				}
				key := strings.ToUpper(serialNo)
				if _, duplicate := seen[key]; duplicate {
					return invalid("found_serial_numbers must be unique")
				}
				seen[key] = struct{}{}
				serials, total, listErr := ir.Serial.List(ctx, inventoryrepo.Filter{OwnerID: header.OwnerID, ItemID: line.ItemID, Number: serialNo, Page: 1, PageSize: 2})
				if listErr != nil {
					return listErr
				}
				if total > 1 {
					return fmt.Errorf("multiple serial identities found for owner, item and serial number")
				}
				if total == 1 {
					serialID = serials[0].ID
					if _, stateErr := ir.SerialState.Get(ctx, serialID); stateErr == nil {
						return invalid("found serial " + serialNo + " is already in inventory")
					} else if !errors.Is(stateErr, inventoryrepo.ErrNotFound) {
						return stateErr
					}
				} else {
					created, createErr := inventory.NewService(ir).CreateSerial(ctx, inventorydto.CreateSerialRequest{OwnerID: header.OwnerID, ItemID: line.ItemID, SerialNo: serialNo}, actor)
					if createErr != nil {
						return createErr
					}
					serialID = created.ID
				}
			}

			posting := inventorydto.PostingRequest{OperationKey: fmt.Sprintf("stock.count-serial.%s.%d.%d", header.ID, line.LineNo, index+1), MovementTypeCode: "COUNT_CORRECTION", OwnerID: header.OwnerID, WarehouseID: header.WarehouseID, BusinessDate: header.BusinessDate.Format("2006-01-02"), ItemID: line.ItemID, LotID: line.LotID, HandlingUnitID: line.HandlingUnitID, Quantity: "1", SerialIDs: []string{serialID}, SourceDocumentID: header.ID, SourceLineID: &line.ID, ReasonCodeID: reasonID, Notes: &note}
			if direction < 0 {
				posting.From = dimension(balance)
				posting.ExpectedSourceVersion = &expectedVersion
			} else {
				posting.To = dimension(balance)
				if expectedVersion > 0 {
					posting.ExpectedDestinationVersion = &expectedVersion
				}
			}
			posted, postErr := inventory.NewService(ir).PostMovement(countPostingContext(ctx, header), posting, actor)
			if postErr != nil {
				return postErr
			}
			lastMovementID = posted.Movement.ID
			if posted.FromBalance != nil {
				expectedVersion = posted.FromBalance.VersionNo
				resultBalanceID = posted.FromBalance.ID
			}
			if posted.ToBalance != nil {
				expectedVersion = posted.ToBalance.VersionNo
				resultBalanceID = posted.ToBalance.ID
			}
			reconciliation := model.CycleCountSerialReconciliation{ID: fmt.Sprintf("%s-SR%04d", line.ID, index+1), CycleCountLineID: line.ID, SerialID: serialID, DirectionCode: directionCode, InventoryMovementID: posted.Movement.ID, CreatedBy: actor}
			if err = r.CycleCounts.CreateSerialReconciliation(ctx, &reconciliation); err != nil {
				return err
			}
		}
		if err = r.CycleCounts.Decide(ctx, line.ID, "POSTED", actor, "Serialized count variance reconciled", &lastMovementID, &resultBalanceID); err != nil {
			return err
		}
		return updateCycleCountHeader(ctx, r, header, actor, nil)
	})
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	return s.GetCycleCount(ctx, id, actor)
}

func (s *Service) RejectCycleCount(ctx context.Context, id string, q dto.RejectCycleCountRequest, actor string) (dto.CycleCountResponse, error) {
	reason := strings.TrimSpace(q.Reason)
	if reason == "" {
		return dto.CycleCountResponse{}, invalid("reason is required")
	}
	ids, err := normalizeLineIDs(q.LineIDs)
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, _ *inventoryrepo.Repositories) error {
		header, err := r.CycleCounts.Lock(ctx, id)
		if err != nil {
			return err
		}
		if header.VersionNo != q.ExpectedVersion {
			return inventoryrepo.ErrConflict
		}
		if header.CreatedBy == actor {
			return invalid("the cycle count creator cannot reject it")
		}
		if err = authorize(ctx, r, actor, scopeBalance(header.OwnerID, header.WarehouseID)); err != nil {
			return err
		}
		lines, err := r.CycleCounts.LockLines(ctx, id, ids)
		if err != nil {
			return err
		}
		if len(lines) != len(ids) {
			return invalid("invalid selected lines")
		}
		for _, line := range lines {
			if line.DecisionCode != "COUNTED" || line.RequiresRecount {
				return invalid("one or more lines are not ready for review")
			}
			if countLineSelfReview(header, line, actor) {
				return invalid("found stock must be reviewed by a different account from its recorder and counter")
			}
			if err = r.CycleCounts.Decide(ctx, line.ID, "REJECTED", actor, reason, nil, nil); err != nil {
				return err
			}
		}
		return updateCycleCountHeader(ctx, r, header, actor, nil)
	})
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	return s.GetCycleCount(ctx, id, actor)
}
func (s *Service) CancelCycleCount(ctx context.Context, id string, q dto.CancelCycleCountRequest, actor string) (dto.CycleCountResponse, error) {
	reason := strings.TrimSpace(q.Reason)
	if reason == "" {
		return dto.CycleCountResponse{}, invalid("reason is required")
	}
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, _ *inventoryrepo.Repositories) error {
		header, err := r.CycleCounts.Lock(ctx, id)
		if err != nil {
			return err
		}
		if header.VersionNo != q.ExpectedVersion {
			return inventoryrepo.ErrConflict
		}
		if header.CreatedBy != actor {
			return ErrForbidden
		}
		if err = authorize(ctx, r, actor, scopeBalance(header.OwnerID, header.WarehouseID)); err != nil {
			return err
		}
		_, _, _, finalBefore, err := r.CycleCounts.Progress(ctx, id)
		if err != nil {
			return err
		}
		if err = r.CycleCounts.CancelRemaining(ctx, id, actor, reason); err != nil {
			return err
		}
		statusCode := "CANCELLED"
		if finalBefore > 0 {
			statusCode = "POSTED"
		}
		status, err := r.CycleCounts.Status(ctx, header.DocumentTypeID, statusCode)
		if err != nil {
			return invalid("STOCK_COUNT " + statusCode + " status is not configured")
		}
		now := time.Now()
		if err = r.CycleCounts.UpdateHeader(ctx, id, header.VersionNo, map[string]any{"status_id": status.ID, "cancelled_at": now, "cancelled_by": actor, "cancellation_reason": reason, "completed_at": now, "updated_by": actor}); err != nil {
			return err
		}
		if header.CountTypeCode == "GRAND" {
			return r.Balances.ReleaseCountFreeze(ctx, header.ID)
		}
		return nil
	})
	if err != nil {
		return dto.CycleCountResponse{}, err
	}
	return s.GetCycleCount(ctx, id, actor)
}
