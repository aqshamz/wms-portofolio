package stockcontrol

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	inventorydto "wms-api/dto/inventory"
	dto "wms-api/dto/stock_control"
	inventoryrepo "wms-api/repository/inventory"
	repository "wms-api/repository/stock_control"
	inventory "wms-api/services/inventory"
)

var ErrInvalidInput = errors.New("invalid stock-control command")
var nonnegative = regexp.MustCompile(`^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$`)
var reasonPattern = regexp.MustCompile(`^[A-Z0-9][A-Z0-9_-]*$`)

type Service struct{ repositories *repository.Repositories }

func NewService(r *repository.Repositories) *Service { return &Service{repositories: r} }
func invalid(message string) error                   { return fmt.Errorf("%w: %s", ErrInvalidInput, message) }
func number(value string, positive bool) (string, *big.Rat, error) {
	value = strings.TrimSpace(value)
	if !nonnegative.MatchString(value) {
		return "", nil, invalid("quantity must fit numeric(20,6)")
	}
	n, ok := new(big.Rat).SetString(value)
	if !ok || (positive && n.Sign() <= 0) {
		return "", nil, invalid("quantity must be greater than zero")
	}
	return n.FloatString(6), n, nil
}
func base(q dto.CommandBase) inventorydto.PostingRequest {
	return inventorydto.PostingRequest{OperationKey: q.OperationKey, BusinessDate: q.BusinessDate, SourceDocumentID: q.SourceDocumentID, SourceLineID: q.SourceLineID, Notes: q.Notes}
}
func resolveReason(ctx context.Context, r *repository.Repositories, q dto.CommandBase, required bool) (*string, error) {
	if q.ReasonCode == nil {
		if required {
			return nil, invalid("reason_code is required")
		}
		return nil, nil
	}
	code := strings.ToUpper(strings.TrimSpace(*q.ReasonCode))
	if !reasonPattern.MatchString(code) {
		return nil, invalid("invalid reason_code")
	}
	reason, err := r.Reasons.ByCode(ctx, "INVENTORY", code)
	if err != nil {
		return nil, invalid("reason_code does not exist")
	}
	if !reason.IsActive {
		return nil, invalid("reason_code is inactive")
	}
	if reason.RequiresNote && (q.Notes == nil || strings.TrimSpace(*q.Notes) == "") {
		return nil, invalid("notes are required for reason_code " + code)
	}
	return &reason.ID, nil
}

func (s *Service) ListReasons(ctx context.Context, active *bool) ([]dto.ReasonCodeResponse, error) {
	rows, err := s.repositories.Reasons.List(ctx, "INVENTORY", active)
	items := make([]dto.ReasonCodeResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.ReasonCodeResponse{ID: row.ID, Code: row.Code, Name: row.Name, Description: row.Description, RequiresNote: row.RequiresNote, IsActive: row.IsActive})
	}
	return items, err
}
func dimension(b inventoryrepo.BalanceRow) *inventorydto.BalanceDimension {
	return &inventorydto.BalanceDimension{LocationID: b.LocationID, InventoryStatusID: b.InventoryStatusID}
}
func trace(q *inventorydto.PostingRequest, b inventoryrepo.BalanceRow) {
	q.OwnerID = b.OwnerID
	q.WarehouseID = b.WarehouseID
	q.ItemID = b.ItemID
	q.LotID = b.LotID
	q.HandlingUnitID = b.HandlingUnitID
}
func response(operation string, results ...inventorydto.PostingResult) dto.CommandResponse {
	r := dto.CommandResponse{Operation: operation, Movements: make([]inventorydto.MovementResponse, 0, len(results)), IdempotentReplay: true}
	for _, v := range results {
		r.Movements = append(r.Movements, v.Movement)
		if v.FromBalance != nil {
			r.SourceBalance = v.FromBalance
		}
		if v.ToBalance != nil {
			r.DestinationBalance = v.ToBalance
		}
		r.IdempotentReplay = r.IdempotentReplay && v.IdempotentReplay
	}
	return r
}
func (s *Service) InternalMove(ctx context.Context, q dto.InternalMoveRequest, actor string) (dto.CommandResponse, error) {
	var result inventorydto.PostingResult
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		source, err := r.Balances.Get(ctx, q.SourceBalanceID)
		if err != nil {
			return err
		}
		p := base(q.CommandBase)
		reason, err := resolveReason(ctx, r, q.CommandBase, false)
		if err != nil {
			return err
		}
		p.ReasonCodeID = reason
		trace(&p, source)
		p.MovementTypeCode = "INTERNAL_MOVE"
		p.From = dimension(source)
		p.To = &inventorydto.BalanceDimension{LocationID: q.TargetLocationID, InventoryStatusID: source.InventoryStatusID}
		p.Quantity = q.Quantity
		p.SerialIDs = q.SerialIDs
		p.ExpectedSourceVersion = &q.ExpectedVersion
		result, err = inventory.NewService(ir).PostMovement(ctx, p, actor)
		return err
	})
	return response("INTERNAL_MOVE", result), err
}
func (s *Service) StatusChange(ctx context.Context, q dto.StatusChangeRequest, actor string) (dto.CommandResponse, error) {
	var result inventorydto.PostingResult
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		source, err := r.Balances.Get(ctx, q.SourceBalanceID)
		if err != nil {
			return err
		}
		p := base(q.CommandBase)
		reason, err := resolveReason(ctx, r, q.CommandBase, true)
		if err != nil {
			return err
		}
		p.ReasonCodeID = reason
		trace(&p, source)
		p.MovementTypeCode = "STATUS_CHANGE"
		p.From = dimension(source)
		p.To = &inventorydto.BalanceDimension{LocationID: source.LocationID, InventoryStatusID: q.TargetInventoryStatusID}
		p.Quantity = q.Quantity
		p.SerialIDs = q.SerialIDs
		p.ExpectedSourceVersion = &q.ExpectedVersion
		result, err = inventory.NewService(ir).PostMovement(ctx, p, actor)
		return err
	})
	return response("STATUS_CHANGE", result), err
}
func (s *Service) Adjustment(ctx context.Context, q dto.AdjustmentRequest, actor string) (dto.CommandResponse, error) {
	var result inventorydto.PostingResult
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		balance, err := r.Balances.Get(ctx, q.BalanceID)
		if err != nil {
			return err
		}
		p := base(q.CommandBase)
		reason, err := resolveReason(ctx, r, q.CommandBase, true)
		if err != nil {
			return err
		}
		p.ReasonCodeID = reason
		trace(&p, balance)
		p.MovementTypeCode = "ADJUSTMENT"
		p.Quantity = q.Quantity
		p.SerialIDs = q.SerialIDs
		if q.Direction == "INCREASE" {
			p.To = dimension(balance)
			p.ExpectedDestinationVersion = &q.ExpectedVersion
		} else if q.Direction == "DECREASE" {
			p.From = dimension(balance)
			p.ExpectedSourceVersion = &q.ExpectedVersion
		} else {
			return invalid("direction must be INCREASE or DECREASE")
		}
		result, err = inventory.NewService(ir).PostMovement(ctx, p, actor)
		return err
	})
	return response("ADJUSTMENT_"+q.Direction, result), err
}
func (s *Service) ReconcileCount(ctx context.Context, q dto.StockCountReconcileRequest, actor string) (dto.CommandResponse, error) {
	counted, countedNumber, err := number(q.CountedQty, false)
	if err != nil {
		return dto.CommandResponse{}, err
	}
	q.CountedQty = counted
	var result inventorydto.PostingResult
	noVariance := false
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		balance, err := r.Balances.Get(ctx, q.BalanceID)
		if err != nil {
			return err
		}
		if err := ir.Balance.LockStockKey(ctx, balance.OwnerID, balance.WarehouseID, balance.ItemID); err != nil {
			return err
		}
		locked, err := ir.Balance.FindLocked(ctx, inventoryrepo.BalanceIdentity{OwnerID: balance.OwnerID, WarehouseID: balance.WarehouseID, LocationID: balance.LocationID, ItemID: balance.ItemID, LotID: balance.LotID, HandlingUnitID: balance.HandlingUnitID, InventoryStatusID: balance.InventoryStatusID})
		if err != nil || locked.ID != balance.ID {
			return invalid("balance identity changed; recount required")
		}
		balance.OnHandQty = locked.OnHandQty
		balance.VersionNo = locked.VersionNo
		system, ok := new(big.Rat).SetString(balance.OnHandQty)
		if !ok {
			return fmt.Errorf("invalid stored balance quantity")
		}
		difference := new(big.Rat).Sub(countedNumber, system)
		if difference.Sign() == 0 {
			if balance.VersionNo != q.ExpectedVersion {
				return invalid("balance version changed; recount required")
			}
			noVariance = true
			return nil
		}
		p := base(q.CommandBase)
		reason, err := resolveReason(ctx, r, q.CommandBase, true)
		if err != nil {
			return err
		}
		p.ReasonCodeID = reason
		trace(&p, balance)
		p.MovementTypeCode = "COUNT_CORRECTION"
		p.Quantity = new(big.Rat).Abs(difference).FloatString(6)
		p.SerialIDs = q.SerialIDs
		if difference.Sign() > 0 {
			p.To = dimension(balance)
			p.ExpectedDestinationVersion = &q.ExpectedVersion
		} else {
			p.From = dimension(balance)
			p.ExpectedSourceVersion = &q.ExpectedVersion
		}
		result, err = inventory.NewService(ir).PostMovement(ctx, p, actor)
		return err
	})
	if noVariance {
		return dto.CommandResponse{Operation: "COUNT_RECONCILIATION", Movements: []inventorydto.MovementResponse{}, NoVariance: true}, err
	}
	return response("COUNT_RECONCILIATION", result), err
}
func (s *Service) WarehouseTransfer(ctx context.Context, q dto.WarehouseTransferRequest, actor string) (dto.CommandResponse, error) {
	if len(q.OperationKey) > 150 {
		return dto.CommandResponse{}, invalid("operation_key is too long for transfer suffixes")
	}
	var out, in inventorydto.PostingResult
	err := s.repositories.Transaction(ctx, func(r *repository.Repositories, ir *inventoryrepo.Repositories) error {
		source, err := r.Balances.Get(ctx, q.SourceBalanceID)
		if err != nil {
			return err
		}
		if source.WarehouseID == q.TargetWarehouseID {
			return invalid("source and target warehouses must differ")
		}
		if source.HandlingUnitID != nil {
			return invalid("direct warehouse transfer of a handling unit is not supported")
		}
		common := base(q.CommandBase)
		reason, err := resolveReason(ctx, r, q.CommandBase, false)
		if err != nil {
			return err
		}
		common.ReasonCodeID = reason
		trace(&common, source)
		common.Quantity = q.Quantity
		common.SerialIDs = q.SerialIDs
		common.ExpectedSourceVersion = &q.ExpectedVersion
		common.OperationKey = q.OperationKey + ":out"
		common.MovementTypeCode = "TRANSFER_OUT"
		common.From = dimension(source)
		out, err = inventory.NewService(ir).PostMovement(ctx, common, actor)
		if err != nil {
			return err
		}
		incoming := base(q.CommandBase)
		incoming.ReasonCodeID = reason
		incoming.OperationKey = q.OperationKey + ":in"
		incoming.MovementTypeCode = "TRANSFER_IN"
		incoming.OwnerID = source.OwnerID
		incoming.WarehouseID = q.TargetWarehouseID
		incoming.ItemID = source.ItemID
		incoming.LotID = source.LotID
		incoming.Quantity = q.Quantity
		incoming.SerialIDs = q.SerialIDs
		incoming.To = &inventorydto.BalanceDimension{LocationID: q.TargetLocationID, InventoryStatusID: q.TargetInventoryStatusID}
		in, err = inventory.NewService(ir).PostMovement(ctx, incoming, actor)
		return err
	})
	return response("WAREHOUSE_TRANSFER", out, in), err
}
