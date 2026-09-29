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

	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/stock_control"
	model "wms-api/models/stock_control"
	inventoryrepo "wms-api/repository/inventory"
	repository "wms-api/repository/stock_control"
	inventory "wms-api/services/inventory"
)

var adjustmentReasonCodes = map[string]bool{"DAMAGE": true, "EXPIRY": true, "MANUAL_ADJUSTMENT": true, "TRANSFER_VARIANCE": true}

func adjustmentID() (string, error) {
	data := make([]byte, 12)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return "IADJ-" + strings.ToUpper(hex.EncodeToString(data)), nil
}

func mapAdjustment(row repository.InventoryAdjustmentRow) dto.AdjustmentResponse {
	return dto.AdjustmentResponse{
		ID: row.ID, StatusCode: row.StatusCode, OwnerID: row.OwnerID, OwnerCode: row.OwnerCode, OwnerName: row.OwnerName,
		WarehouseID: row.WarehouseID, WarehouseCode: row.WarehouseCode, WarehouseName: row.WarehouseName,
		BusinessDate: row.BusinessDate.Format("2006-01-02"), Direction: row.Direction, ReasonCode: row.ReasonCode, ReasonName: row.ReasonName, Notes: row.Notes,
		TotalLines: row.TotalLines, PendingLines: row.PendingLines, PostedLines: row.PostedLines, RejectedLines: row.RejectedLines, CancelledLines: row.CancelledLines,
		Lines: []dto.AdjustmentLineResponse{}, CompletedAt: row.CompletedAt, CancelledAt: row.CancelledAt, CancelledBy: row.CancelledBy,
		CancelledByDisplayName: row.CancelledByDisplayName, CancellationReason: row.CancellationReason,
		CreatedAt: row.CreatedAt, CreatedBy: row.CreatedBy, CreatedByUsername: row.CreatedByUsername, CreatedByDisplayName: row.CreatedByDisplayName, VersionNo: row.VersionNo,
	}
}
func mapAdjustmentLine(row repository.InventoryAdjustmentLineRow) dto.AdjustmentLineResponse {
	return dto.AdjustmentLineResponse{
		ID: row.ID, LineNo: row.LineNo, DecisionCode: row.DecisionCode, BalanceID: row.BalanceID,
		PlannedBalanceVersionNo: row.PlannedBalanceVersionNo, CurrentBalanceVersionNo: row.CurrentBalanceVersionNo,
		ItemID: row.ItemID, ItemCode: row.ItemCode, ItemName: row.ItemName, LotID: row.LotID, LotNumber: row.LotNumber,
		SerialID: row.SerialID, SerialNumber: row.SerialNumber, HandlingUnitID: row.HandlingUnitID, HandlingUnitBarcode: row.HandlingUnitBarcode,
		LocationID: row.LocationID, LocationCode: row.LocationCode, InventoryStatusID: row.InventoryStatusID, InventoryStatusCode: row.InventoryStatusCode,
		Quantity: row.Quantity, UOMID: row.UOMID, UOMCode: row.UOMCode, ApprovedAt: row.ApprovedAt, ApprovedBy: row.ApprovedBy,
		ApprovedByDisplayName: row.ApprovedByDisplayName, RejectedAt: row.RejectedAt, RejectedBy: row.RejectedBy,
		RejectedByDisplayName: row.RejectedByDisplayName, RejectionReason: row.RejectionReason, CancelledAt: row.CancelledAt,
		CancelledBy: row.CancelledBy, CancellationReason: row.CancellationReason, InventoryMovementID: row.InventoryMovementID,
		ResultingBalanceID: row.ResultingBalanceID, VersionNo: row.VersionNo,
	}
}

func (s *Service) GetAdjustment(ctx context.Context, id, actor string) (dto.AdjustmentResponse, error) {
	id = strings.TrimSpace(id)
	if id == "" || len(id) > 140 {
		return dto.AdjustmentResponse{}, invalid("invalid inventory_adjustment_id")
	}
	row, err := s.repositories.Adjustments.Get(ctx, id)
	if err != nil {
		return dto.AdjustmentResponse{}, err
	}
	if err = authorize(ctx, s.repositories, actor, scopeBalance(row.OwnerID, row.WarehouseID)); err != nil {
		return dto.AdjustmentResponse{}, err
	}
	lines, err := s.repositories.Adjustments.Lines(ctx, id)
	if err != nil {
		return dto.AdjustmentResponse{}, err
	}
	response := mapAdjustment(row)
	for _, line := range lines {
		response.Lines = append(response.Lines, mapAdjustmentLine(line))
	}
	return response, nil
}
func (s *Service) ListAdjustments(ctx context.Context, filter repository.AdjustmentFilter, actor string) (dto.PageResponse[dto.AdjustmentResponse], error) {
	filter.OwnerID = strings.ToLower(strings.TrimSpace(filter.OwnerID))
	filter.WarehouseID = strings.ToLower(strings.TrimSpace(filter.WarehouseID))
	filter.StatusCode = strings.ToUpper(strings.TrimSpace(filter.StatusCode))
	filter.Search = strings.TrimSpace(filter.Search)
	if filter.OwnerID == "" || filter.WarehouseID == "" || filter.Page < 1 || filter.PageSize < 1 || filter.PageSize > 100 || len(filter.Search) > 160 {
		return dto.PageResponse[dto.AdjustmentResponse]{}, invalid("invalid adjustment filters")
	}
	allowed, err := s.repositories.Scope.Allowed(ctx, actor, filter.OwnerID, filter.WarehouseID)
	if err != nil {
		return dto.PageResponse[dto.AdjustmentResponse]{}, err
	}
	if !allowed {
		return dto.PageResponse[dto.AdjustmentResponse]{}, ErrForbidden
	}
	rows, total, err := s.repositories.Adjustments.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.AdjustmentResponse]{}, err
	}
	items := make([]dto.AdjustmentResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapAdjustment(row))
	}
	return replenishmentPage(items, filter.Page, filter.PageSize, total), nil
}

func validateAdjustmentLine(ctx context.Context, r *repository.Repositories, ir *inventoryrepo.Repositories, request dto.CreateAdjustmentLineRequest, direction string) (model.InventoryAdjustmentLine, error) {
	quantity, quantityNumber, err := number(request.Quantity, true)
	if err != nil {
		return model.InventoryAdjustmentLine{}, err
	}
	balance, err := r.Balances.Get(ctx, strings.TrimSpace(request.BalanceID))
	if err != nil {
		return model.InventoryAdjustmentLine{}, err
	}
	if err := validateStorageLocation(ctx, ir, balance.WarehouseID, balance.LocationID, "inventory adjustment"); err != nil {
		return model.InventoryAdjustmentLine{}, err
	}
	if balance.VersionNo != request.ExpectedVersion {
		return model.InventoryAdjustmentLine{}, invalid("a selected balance changed; reload before requesting the adjustment")
	}
	if balance.SerialControlled {
		if request.SerialID == nil || quantityNumber.Cmp(big.NewRat(1, 1)) != 0 {
			return model.InventoryAdjustmentLine{}, invalid("serialized adjustment lines require one serial and quantity 1")
		}
		serial, serialErr := ir.Serial.GetShared(ctx, *request.SerialID)
		if serialErr != nil || serial.OwnerID != balance.OwnerID || serial.ItemID != balance.ItemID {
			return model.InventoryAdjustmentLine{}, invalid("serial does not match the balance owner and item")
		}
		state, stateErr := ir.SerialState.Get(ctx, serial.ID)
		if direction == "DECREASE" && (stateErr != nil || state.BalanceID != balance.ID) {
			return model.InventoryAdjustmentLine{}, invalid("serial is not in the selected balance")
		}
		if direction == "INCREASE" && stateErr == nil {
			return model.InventoryAdjustmentLine{}, invalid("serial is already in inventory")
		}
		if direction == "INCREASE" && !errors.Is(stateErr, inventoryrepo.ErrNotFound) {
			return model.InventoryAdjustmentLine{}, stateErr
		}
	} else if request.SerialID != nil {
		return model.InventoryAdjustmentLine{}, invalid("serial_id is only allowed for serial-controlled stock")
	}
	if direction == "DECREASE" {
		available, ok := new(big.Rat).SetString(balance.AvailableQty)
		if !ok || available.Cmp(quantityNumber) < 0 {
			return model.InventoryAdjustmentLine{}, invalid("decrease exceeds available unreserved quantity")
		}
	}
	return model.InventoryAdjustmentLine{BalanceID: balance.ID, PlannedBalanceVersionNo: balance.VersionNo, ItemID: balance.ItemID, LotID: balance.LotID, SerialID: request.SerialID, HandlingUnitID: balance.HandlingUnitID, LocationID: balance.LocationID, InventoryStatusID: balance.InventoryStatusID, UOMID: balance.UOMID, Quantity: quantity, DecisionCode: "PENDING"}, nil
}

func (s *Service) CreateAdjustment(ctx context.Context, request dto.CreateAdjustmentRequest, actor string) (dto.AdjustmentResponse, error) {
	request.Direction = strings.ToUpper(strings.TrimSpace(request.Direction))
	request.ReasonCode = strings.ToUpper(strings.TrimSpace(request.ReasonCode))
	if request.Direction != "INCREASE" && request.Direction != "DECREASE" {
		return dto.AdjustmentResponse{}, invalid("direction must be INCREASE or DECREASE")
	}
	if !adjustmentReasonCodes[request.ReasonCode] {
		return dto.AdjustmentResponse{}, invalid("reason_code is not valid for inventory adjustments")
	}
	if len(request.Lines) == 0 || len(request.Lines) > 100 {
		return dto.AdjustmentResponse{}, invalid("an adjustment requires between 1 and 100 lines")
	}
	businessDate, err := time.Parse("2006-01-02", request.BusinessDate)
	if err != nil {
		return dto.AdjustmentResponse{}, invalid("business_date must be YYYY-MM-DD")
	}
	if request.Notes != nil {
		notes := strings.TrimSpace(*request.Notes)
		request.Notes = &notes
	}
	id, err := adjustmentID()
	if err != nil {
		return dto.AdjustmentResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		documentType, err := r.Adjustments.DocumentType(ctx)
		if err != nil {
			return invalid("INVENTORY_ADJUSTMENT document type is not configured")
		}
		status, err := r.Adjustments.Status(ctx, documentType.ID, "DRAFT")
		if err != nil {
			return invalid("INVENTORY_ADJUSTMENT DRAFT status is not configured")
		}
		reasonCode := request.ReasonCode
		reasonID, err := resolveReason(ctx, r, dto.CommandBase{ReasonCode: &reasonCode, Notes: request.Notes}, true)
		if err != nil {
			return err
		}
		lines := make([]model.InventoryAdjustmentLine, 0, len(request.Lines))
		seen := map[string]bool{}
		var ownerID, warehouseID string
		for index, lineRequest := range request.Lines {
			line, err := validateAdjustmentLine(ctx, r, ir, lineRequest, request.Direction)
			if err != nil {
				return fmt.Errorf("line %d: %w", index+1, err)
			}
			balance, err := r.Balances.Get(ctx, line.BalanceID)
			if err != nil {
				return err
			}
			if index == 0 {
				ownerID, warehouseID = balance.OwnerID, balance.WarehouseID
				if err = authorize(ctx, r, actor, balance); err != nil {
					return err
				}
			} else if balance.OwnerID != ownerID || balance.WarehouseID != warehouseID {
				return invalid("all adjustment lines must belong to the same owner and warehouse")
			}
			key := line.BalanceID
			if seen[key] {
				return invalid("each inventory balance can appear only once in an adjustment document")
			}
			seen[key] = true
			line.ID = fmt.Sprintf("%s-L%04d", id, index+1)
			line.AdjustmentID = id
			line.LineNo = index + 1
			line.CreatedBy = actor
			line.UpdatedBy = &actor
			line.VersionNo = 1
			lines = append(lines, line)
		}
		header := &model.InventoryAdjustment{ID: id, DocumentTypeID: documentType.ID, StatusID: status.ID, OwnerID: ownerID, WarehouseID: warehouseID, BusinessDate: businessDate, Direction: request.Direction, ReasonCodeID: *reasonID, Notes: request.Notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}
		return r.Adjustments.Create(ctx, header, lines)
	})
	if err != nil {
		return dto.AdjustmentResponse{}, err
	}
	return s.GetAdjustment(ctx, id, actor)
}

func normalizeLineIDs(ids []string) ([]string, error) {
	seen := map[string]bool{}
	result := make([]string, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			return nil, invalid("line_ids must contain unique non-empty values")
		}
		seen[id] = true
		result = append(result, id)
	}
	return result, nil
}
func nextAdjustmentStatus(pending, posted int64) string {
	if pending > 0 && posted > 0 {
		return "PARTIALLY_POSTED"
	}
	if pending > 0 {
		return "DRAFT"
	}
	if posted > 0 {
		return "POSTED"
	}
	return "CANCELLED"
}
func completeHeader(ctx context.Context, r *repository.Repositories, adjustment model.InventoryAdjustment, actor string, extra map[string]any) error {
	pending, posted, err := r.Adjustments.Counts(ctx, adjustment.ID)
	if err != nil {
		return err
	}
	code := nextAdjustmentStatus(pending, posted)
	status, err := r.Adjustments.Status(ctx, adjustment.DocumentTypeID, code)
	if err != nil {
		return invalid("INVENTORY_ADJUSTMENT " + code + " status is not configured")
	}
	values := map[string]any{"status_id": status.ID, "updated_by": actor}
	for key, value := range extra {
		values[key] = value
	}
	if pending == 0 {
		values["completed_at"] = time.Now()
	}
	return r.Adjustments.Update(ctx, adjustment.ID, adjustment.VersionNo, values)
}

func (s *Service) ApproveAdjustment(ctx context.Context, id string, request dto.AdjustmentLineSelectionRequest, actor string) (dto.AdjustmentResponse, error) {
	lineIDs, err := normalizeLineIDs(request.LineIDs)
	if err != nil {
		return dto.AdjustmentResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		adjustment, err := r.Adjustments.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := r.Adjustments.Get(ctx, id)
		if err != nil {
			return err
		}
		if err = authorize(ctx, r, actor, scopeBalance(adjustment.OwnerID, adjustment.WarehouseID)); err != nil {
			return err
		}
		if adjustment.VersionNo != request.ExpectedVersion {
			return inventoryrepo.ErrConflict
		}
		if row.StatusCode != "DRAFT" && row.StatusCode != "PARTIALLY_POSTED" {
			return invalid("only pending adjustment lines can be approved")
		}
		if adjustment.CreatedBy == actor {
			return invalid("the requester cannot approve their own adjustment")
		}
		lines, err := r.Adjustments.LockLines(ctx, id, lineIDs)
		if err != nil {
			return err
		}
		if len(lines) != len(lineIDs) {
			return invalid("one or more selected lines do not belong to this adjustment")
		}
		for _, line := range lines {
			if line.DecisionCode != "PENDING" {
				return invalid("one or more selected lines are no longer pending")
			}
			balance, err := r.Balances.Get(ctx, line.BalanceID)
			if err != nil {
				return err
			}
			if err := validateStorageLocation(ctx, ir, balance.WarehouseID, balance.LocationID, "inventory adjustment"); err != nil {
				return err
			}
			if balance.VersionNo != line.PlannedBalanceVersionNo {
				return invalid("a selected balance changed after the request; reject it and create a new line")
			}
		}
		for _, line := range lines {
			balance, err := r.Balances.Get(ctx, line.BalanceID)
			if err != nil {
				return err
			}
			posting := inventorydto.PostingRequest{OperationKey: "stock.adjustment." + adjustment.ID + "." + line.ID, MovementTypeCode: "ADJUSTMENT", OwnerID: adjustment.OwnerID, WarehouseID: adjustment.WarehouseID, BusinessDate: adjustment.BusinessDate.Format("2006-01-02"), ItemID: line.ItemID, LotID: line.LotID, HandlingUnitID: line.HandlingUnitID, Quantity: line.Quantity, SourceDocumentID: adjustment.ID, ReasonCodeID: &adjustment.ReasonCodeID, Notes: adjustment.Notes}
			if line.SerialID != nil {
				posting.SerialIDs = []string{*line.SerialID}
			}
			expected := balance.VersionNo
			if adjustment.Direction == "INCREASE" {
				posting.To = dimension(balance)
				posting.ExpectedDestinationVersion = &expected
			} else {
				posting.From = dimension(balance)
				posting.ExpectedSourceVersion = &expected
			}
			posted, err := inventory.NewService(ir).PostMovement(ctx, posting, actor)
			if err != nil {
				return err
			}
			resulting := line.BalanceID
			if posted.FromBalance != nil {
				resulting = posted.FromBalance.ID
			}
			if posted.ToBalance != nil {
				resulting = posted.ToBalance.ID
			}
			if err = r.Adjustments.PostLine(ctx, line.ID, posted.Movement.ID, resulting, actor); err != nil {
				return err
			}
		}
		return completeHeader(ctx, r, adjustment, actor, nil)
	})
	if err != nil {
		return dto.AdjustmentResponse{}, err
	}
	return s.GetAdjustment(ctx, id, actor)
}

func (s *Service) RejectAdjustment(ctx context.Context, id string, request dto.RejectAdjustmentLinesRequest, actor string) (dto.AdjustmentResponse, error) {
	reason := strings.TrimSpace(request.Reason)
	if reason == "" {
		return dto.AdjustmentResponse{}, invalid("reason is required")
	}
	lineIDs, err := normalizeLineIDs(request.LineIDs)
	if err != nil {
		return dto.AdjustmentResponse{}, err
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, _ *inventoryrepo.Repositories) error {
		adjustment, err := r.Adjustments.Lock(ctx, id)
		if err != nil {
			return err
		}
		if err = authorize(ctx, r, actor, scopeBalance(adjustment.OwnerID, adjustment.WarehouseID)); err != nil {
			return err
		}
		if adjustment.VersionNo != request.ExpectedVersion {
			return inventoryrepo.ErrConflict
		}
		if adjustment.CreatedBy == actor {
			return invalid("the requester cannot reject their own adjustment lines")
		}
		lines, err := r.Adjustments.LockLines(ctx, id, lineIDs)
		if err != nil {
			return err
		}
		if len(lines) != len(lineIDs) {
			return invalid("one or more selected lines do not belong to this adjustment")
		}
		for _, line := range lines {
			if line.DecisionCode != "PENDING" {
				return invalid("one or more selected lines are no longer pending")
			}
		}
		if err = r.Adjustments.DecideLines(ctx, id, lineIDs, "REJECTED", actor, reason); err != nil {
			return err
		}
		return completeHeader(ctx, r, adjustment, actor, nil)
	})
	if err != nil {
		return dto.AdjustmentResponse{}, err
	}
	return s.GetAdjustment(ctx, id, actor)
}

func (s *Service) CancelAdjustment(ctx context.Context, id string, request dto.CancelAdjustmentRequest, actor string) (dto.AdjustmentResponse, error) {
	reason := strings.TrimSpace(request.Reason)
	if reason == "" {
		return dto.AdjustmentResponse{}, invalid("reason is required")
	}
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, _ *inventoryrepo.Repositories) error {
		adjustment, err := r.Adjustments.Lock(ctx, id)
		if err != nil {
			return err
		}
		if err = authorize(ctx, r, actor, scopeBalance(adjustment.OwnerID, adjustment.WarehouseID)); err != nil {
			return err
		}
		if adjustment.VersionNo != request.ExpectedVersion {
			return inventoryrepo.ErrConflict
		}
		if adjustment.CreatedBy != actor {
			return ErrForbidden
		}
		ids, err := r.Adjustments.PendingLineIDs(ctx, id)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			return invalid("this adjustment has no pending lines")
		}
		if err = r.Adjustments.DecideLines(ctx, id, ids, "CANCELLED", actor, reason); err != nil {
			return err
		}
		now := time.Now()
		return completeHeader(ctx, r, adjustment, actor, map[string]any{"cancelled_at": now, "cancelled_by": actor, "cancellation_reason": reason})
	})
	if err != nil {
		return dto.AdjustmentResponse{}, err
	}
	return s.GetAdjustment(ctx, id, actor)
}
