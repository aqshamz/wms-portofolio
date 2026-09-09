package outbound

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"
	dto "wms-api/dto/outbound"
	model "wms-api/models/outbound"
	repository "wms-api/repository/outbound"
)

func (s *Service) GetCheck(ctx context.Context, id string) (dto.CheckResponse, error) {
	if !validID(id, 120) {
		return dto.CheckResponse{}, invalid("invalid outbound_check_id")
	}
	v, err := s.repositories.Check.Get(ctx, id)
	if err != nil {
		return dto.CheckResponse{}, err
	}
	out := mapCheck(v)
	rows, err := s.repositories.CheckLine.List(ctx, id)
	if err != nil {
		return dto.CheckResponse{}, err
	}
	out.Lines = make([]dto.CheckLineResponse, 0, len(rows))
	for _, row := range rows {
		out.Lines = append(out.Lines, mapCheckLine(row))
	}
	return out, nil
}
func (s *Service) ListChecks(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.CheckResponse], error) {
	if !validUUID(f.OwnerID) {
		return dto.PageResponse[dto.CheckResponse]{}, invalid("owner_id is required and must be a UUID")
	}
	rows, total, err := s.repositories.Check.List(ctx, f)
	if err != nil {
		return dto.PageResponse[dto.CheckResponse]{}, err
	}
	items := make([]dto.CheckResponse, 0, len(rows))
	for _, v := range rows {
		items = append(items, mapCheck(v))
	}
	return page(items, f.Page, f.PageSize, total), nil
}
func (s *Service) CreateCheck(ctx context.Context, q dto.CreateCheckRequest, actor string) (dto.CheckResponse, error) {
	if !validID(q.StagingID, 120) || !validUUID(actor) {
		return dto.CheckResponse{}, invalid("invalid check request")
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.CheckResponse{}, err
	}
	var id string
	err = s.transaction(ctx, func(local *Service) error {
		staging, err := local.repositories.Staging.Get(ctx, q.StagingID)
		if err != nil {
			return err
		}
		if staging.StatusCode != "COMPLETED" {
			return state("staging must be COMPLETED before checking")
		}
		order, err := local.repositories.Order.Lock(ctx, staging.OutboundID)
		if err != nil {
			return err
		}
		orderRow, err := local.repositories.Order.Get(ctx, order.ID)
		if err != nil {
			return err
		}
		if q.ParentCheckID == nil {
			if orderRow.StatusCode != "STAGED" {
				return state("only a STAGED order can start its first check")
			}
		} else {
			parentID, err := clean(*q.ParentCheckID, 120, "parent_check_id")
			if err != nil {
				return err
			}
			parent, err := local.repositories.Check.Get(ctx, parentID)
			if err != nil {
				return err
			}
			if parent.StagingID != staging.ID || parent.StatusCode != "FAILED" {
				return state("parent check must be a FAILED check for the same staging document")
			}
			unresolved, err := local.repositories.Check.HasUnresolvedParent(ctx, parentID)
			if err != nil {
				return err
			}
			if unresolved {
				return state("resolve every parent-check exception before rechecking")
			}
			q.ParentCheckID = &parentID
			if orderRow.StatusCode != "CHECK_FAILED" {
				return state("only a CHECK_FAILED order can be rechecked")
			}
		}
		open, err := local.repositories.Check.HasOpenForStaging(ctx, staging.ID)
		if err != nil {
			return err
		}
		if open {
			return state("staging already has an OPEN check")
		}
		kind, initial, err := local.document(ctx, "OUTBOUND_CHECK")
		if err != nil {
			return err
		}
		id, err = local.generateID(ctx, "OUTBOUND_CHECK", order.BusinessDate, nil, &order.WarehouseID)
		if err != nil {
			return err
		}
		header := model.OutboundCheck{ID: id, ParentCheckID: q.ParentCheckID, DocumentTypeID: kind.ID, StatusID: initial.ID, StagingID: staging.ID, Notes: notes, CreatedBy: actor}
		if err := local.repositories.Check.Create(ctx, &header); err != nil {
			return err
		}
		stagingLines, err := local.repositories.StagingLine.List(ctx, staging.ID)
		if err != nil {
			return err
		}
		lines := make([]model.OutboundCheckLine, 0, len(stagingLines))
		for i, v := range stagingLines {
			expected := new(big.Rat).Sub(mustRat(v.StagedQty), mustRat(v.RemovedQty))
			if expected.Sign() <= 0 {
				continue
			}
			lineID := fmt.Sprintf("%s-L%04d", id, i+1)
			lines = append(lines, model.OutboundCheckLine{ID: lineID, OutboundCheckID: id, StagingLineID: v.ID, LineNo: i + 1, ExpectedQty: decimal(expected), ExceptionQty: "0.000000", UOMID: v.UOMID, CreatedBy: actor})
		}
		if len(lines) == 0 {
			return state("staging has no checkable quantity")
		}
		if err := local.repositories.CheckLine.CreateBatch(ctx, lines); err != nil {
			return err
		}
		next, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, "CHECKING")
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, next.ID, actor, nil)
	})
	if err != nil {
		return dto.CheckResponse{}, err
	}
	return s.GetCheck(ctx, id)
}
func (s *Service) RecordCheckLine(ctx context.Context, checkID, lineID string, q dto.RecordCheckLineRequest, actor string) (dto.CheckResponse, error) {
	if !validID(checkID, 120) || !validID(lineID, 160) || !validUUID(actor) {
		return dto.CheckResponse{}, invalid("invalid check-line request")
	}
	qty, checked, err := quantity(q.CheckedQty, "checked_qty", true)
	if err != nil {
		return dto.CheckResponse{}, err
	}
	code := strings.ToUpper(strings.TrimSpace(q.ResultCode))
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.CheckResponse{}, err
	}
	err = s.transaction(ctx, func(local *Service) error {
		header, err := local.repositories.Check.Lock(ctx, checkID)
		if err != nil {
			return err
		}
		row, err := local.repositories.Check.Get(ctx, checkID)
		if err != nil {
			return err
		}
		if row.StatusCode != "OPEN" {
			return state("only an OPEN check can be recorded")
		}
		line, err := local.repositories.CheckLine.Lock(ctx, lineID)
		if err != nil {
			return err
		}
		if line.OutboundCheckID != header.ID {
			return invalid("check line does not belong to check")
		}
		result, err := local.repositories.CheckResult.ByCode(ctx, code)
		if err != nil {
			return invalid("unknown or inactive result_code")
		}
		if result.RequiresNote && (notes == nil || *notes == "") {
			return invalid("notes are required for this result")
		}
		expected := mustRat(line.ExpectedQty)
		exception := new(big.Rat)
		switch code {
		case "PASS":
			if checked.Cmp(expected) != 0 {
				return invalid("PASS requires checked_qty equal to expected_qty")
			}
		case "SHORT":
			if checked.Cmp(expected) >= 0 {
				return invalid("SHORT requires checked_qty below expected_qty")
			}
			exception.Sub(expected, checked)
		case "OVER":
			if checked.Cmp(expected) <= 0 {
				return invalid("OVER requires checked_qty above expected_qty")
			}
			exception.Sub(checked, expected)
		case "WRONG_ITEM", "DAMAGED":
			if checked.Sign() <= 0 || checked.Cmp(expected) > 0 {
				return invalid(code + " requires checked_qty between zero and expected_qty")
			}
			exception.Set(checked)
		default:
			return invalid("unsupported result_code")
		}
		return local.repositories.CheckLine.Record(ctx, lineID, result.ID, qty, decimal(exception), notes, actor)
	})
	if err != nil {
		return dto.CheckResponse{}, err
	}
	return s.GetCheck(ctx, checkID)
}
func (s *Service) CompleteCheck(ctx context.Context, id, actor string) (dto.CheckResponse, error) {
	if !validID(id, 120) || !validUUID(actor) {
		return dto.CheckResponse{}, invalid("invalid check completion")
	}
	err := s.transaction(ctx, func(local *Service) error {
		check, err := local.repositories.Check.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.Check.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode != "OPEN" {
			return state("only an OPEN check can be completed")
		}
		n, err := local.repositories.CheckLine.Unrecorded(ctx, id)
		if err != nil {
			return err
		}
		if n > 0 {
			return state("record every check line before completion")
		}
		failures, err := local.repositories.CheckLine.FailureCount(ctx, id)
		if err != nil {
			return err
		}
		outcome := "PASSED"
		orderOutcome := "CHECKED"
		if failures > 0 {
			outcome = "FAILED"
			orderOutcome = "CHECK_FAILED"
		}
		status, err := local.transition(ctx, check.DocumentTypeID, check.StatusID, outcome)
		if err != nil {
			return err
		}
		if err := local.repositories.Check.Complete(ctx, id, status.ID, actor); err != nil {
			return err
		}
		lines, err := local.repositories.CheckLine.List(ctx, id)
		if err != nil {
			return err
		}
		if failures > 0 {
			open, err := local.repositories.ExceptionStatus.ByCode(ctx, "OPEN")
			if err != nil {
				return err
			}
			for _, line := range lines {
				if line.ResultCode == "PASS" || mustRat(line.ExceptionQty).Sign() == 0 {
					continue
				}
				exception := model.OutboundCheckException{ID: line.ID + "-EX", OutboundCheckLineID: line.ID, StatusID: open.ID, ExceptionQty: line.ExceptionQty, Notes: line.Notes, CreatedBy: actor}
				if err := local.repositories.CheckException.Create(ctx, &exception); err != nil {
					return err
				}
			}
		} else {
			for _, line := range lines {
				if err := local.repositories.Line.AddChecked(ctx, line.OutboundLineID, *line.CheckedQty); err != nil {
					return err
				}
			}
		}
		order, err := local.repositories.Order.Lock(ctx, row.OutboundID)
		if err != nil {
			return err
		}
		next, err := local.transition(ctx, order.DocumentTypeID, order.StatusID, orderOutcome)
		if err != nil {
			return err
		}
		return local.repositories.Order.SetStatusLocked(ctx, order.ID, next.ID, actor, nil)
	})
	if err != nil {
		return dto.CheckResponse{}, err
	}
	return s.GetCheck(ctx, id)
}
func (s *Service) ListCheckExceptions(ctx context.Context, checkID string) ([]dto.CheckExceptionResponse, error) {
	if !validID(checkID, 120) {
		return nil, invalid("invalid outbound_check_id")
	}
	rows, err := s.repositories.CheckException.ListByCheck(ctx, checkID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CheckExceptionResponse, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapException(v))
	}
	return out, nil
}
func (s *Service) ResolveCheckException(ctx context.Context, id string, q dto.ResolveCheckExceptionRequest, actor string) (dto.CheckExceptionResponse, error) {
	if !validID(id, 170) || !validUUID(actor) {
		return dto.CheckExceptionResponse{}, invalid("invalid exception resolution")
	}
	qty, n, err := quantity(q.ResolvedQty, "resolved_qty", false)
	if err != nil {
		return dto.CheckExceptionResponse{}, err
	}
	notes, err := optional(q.Notes, 4000, "notes")
	if err != nil {
		return dto.CheckExceptionResponse{}, err
	}
	code := strings.ToUpper(strings.TrimSpace(q.ResolutionTypeCode))
	err = s.transaction(ctx, func(local *Service) error {
		exception, err := local.repositories.CheckException.Lock(ctx, id)
		if err != nil {
			return err
		}
		row, err := local.repositories.CheckException.Get(ctx, id)
		if err != nil {
			return err
		}
		if row.StatusCode == "RESOLVED" {
			return state("check exception is already resolved")
		}
		kind, err := local.repositories.ResolutionType.ByCode(ctx, code)
		if err != nil {
			return invalid("unknown or inactive resolution_type_code")
		}
		if kind.RequiresApproval && !q.Approved {
			return invalid("this resolution requires approved=true")
		}
		totals, err := local.repositories.CheckResolution.Totals(ctx, id)
		if err != nil {
			return err
		}
		exceptionQty := mustRat(exception.ExceptionQty)
		corrected := mustRat(totals.CorrectedQty)
		accepted := mustRat(totals.AcceptedQty)
		replaced := mustRat(totals.ReplacementQty)
		fulfilled := new(big.Rat).Add(accepted, replaced)
		if kind.CountsAsStockCorrection {
			if q.MovementID == nil {
				return invalid("movement_id is required for STOCK_CORRECTION")
			}
			movement, err := local.repositories.Inventory.Movement.Get(ctx, *q.MovementID)
			if err != nil {
				return invalid("movement_id does not exist")
			}
			if movement.OwnerID != row.OwnerID || movement.WarehouseID != row.WarehouseID || movement.ItemID != row.ItemID || mustRat(movement.Quantity).Cmp(n) != 0 {
				return invalid("movement must correct the same warehouse, item, and exact quantity")
			}
			source, err := local.repositories.Inventory.Balance.Get(ctx, row.StagingBalanceID)
			if err != nil {
				return err
			}
			if row.ResultCode == "OVER" {
				if movement.ToLocationID == nil || *movement.ToLocationID != source.LocationID {
					return invalid("OVER correction movement must enter the staging location")
				}
			} else if movement.FromLocationID == nil || *movement.FromLocationID != source.LocationID {
				return invalid("correction movement must leave the staging location")
			}
			if new(big.Rat).Add(corrected, n).Cmp(exceptionQty) > 0 {
				return invalid("stock-correction quantity exceeds exception quantity")
			}
		}
		if kind.CountsAsReplacement {
			if q.ReservationID == nil {
				return invalid("reservation_id is required for REPLACEMENT")
			}
			if row.ResultCode == "OVER" {
				return invalid("OVER exceptions do not accept replacement stock")
			}
			eligible, err := local.repositories.Reservation.ReplacementEligible(ctx, *q.ReservationID, row.OutboundLineID, row.StagingID)
			if err != nil {
				return err
			}
			if !eligible {
				return invalid("replacement reservation must be consumed and staged for this outbound line")
			}
			newFulfilled := new(big.Rat).Add(fulfilled, n)
			if newFulfilled.Cmp(exceptionQty) > 0 || corrected.Cmp(newFulfilled) < 0 {
				return invalid("replacement quantity exceeds corrected exception quantity")
			}
		}
		if kind.CountsAsShortAcceptance {
			if row.ResultCode != "SHORT" {
				return invalid("ACCEPT_SHORT applies only to SHORT exceptions")
			}
			newFulfilled := new(big.Rat).Add(fulfilled, n)
			if newFulfilled.Cmp(exceptionQty) > 0 || corrected.Cmp(newFulfilled) < 0 {
				return invalid("accepted-short quantity exceeds corrected exception quantity")
			}
		}
		resolutionID, err := randomID(id + "-R")
		if err != nil {
			return err
		}
		resolution := model.OutboundCheckResolution{ID: resolutionID, OutboundCheckExceptionID: id, OutboundCheckResolutionTypeID: kind.ID, ResolvedQty: qty, Notes: notes, CreatedBy: actor}
		if kind.RequiresApproval {
			now := time.Now()
			resolution.ApprovedAt = &now
			resolution.ApprovedBy = &actor
		}
		if err := local.repositories.CheckResolution.Create(ctx, &resolution); err != nil {
			return err
		}
		if q.MovementID != nil {
			if err := local.repositories.ResolutionMovement.Create(ctx, &model.OutboundCheckResolutionMovement{OutboundCheckResolutionID: resolutionID, MovementID: *q.MovementID}); err != nil {
				return err
			}
		}
		if q.ReservationID != nil {
			if err := local.repositories.ResolutionReservation.Create(ctx, &model.OutboundCheckResolutionReservation{OutboundCheckResolutionID: resolutionID, ReservationID: *q.ReservationID}); err != nil {
				return err
			}
		}
		if kind.CountsAsStockCorrection && row.ResultCode != "OVER" {
			if err := local.repositories.StagingLine.AddRemoved(ctx, row.StagingLineID, qty); err != nil {
				return err
			}
			if err := local.repositories.Line.AddRejected(ctx, row.OutboundLineID, qty); err != nil {
				return err
			}
		}
		if kind.CountsAsShortAcceptance {
			if err := local.repositories.Line.AcceptShort(ctx, row.OutboundLineID, qty); err != nil {
				return err
			}
		}
		totals, err = local.repositories.CheckResolution.Totals(ctx, id)
		if err != nil {
			return err
		}
		corrected = mustRat(totals.CorrectedQty)
		accepted = mustRat(totals.AcceptedQty)
		replaced = mustRat(totals.ReplacementQty)
		eligible := corrected.Cmp(exceptionQty) >= 0
		switch row.ResultCode {
		case "SHORT":
			eligible = eligible && new(big.Rat).Add(accepted, replaced).Cmp(exceptionQty) >= 0
		case "DAMAGED", "WRONG_ITEM":
			eligible = eligible && replaced.Cmp(exceptionQty) >= 0
		case "OVER":
		default:
			eligible = false
		}
		if eligible {
			resolved, err := local.repositories.ExceptionStatus.ByCode(ctx, "RESOLVED")
			if err != nil {
				return err
			}
			return local.repositories.CheckException.Resolve(ctx, id, resolved.ID, actor)
		}
		return nil
	})
	if err != nil {
		return dto.CheckExceptionResponse{}, err
	}
	row, err := s.repositories.CheckException.Get(ctx, id)
	if err != nil {
		return dto.CheckExceptionResponse{}, err
	}
	return mapException(row), nil
}
