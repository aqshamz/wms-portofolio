package inventory

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"
	"time"
	dto "wms-api/dto/inventory"
	model "wms-api/models/inventory"
	repository "wms-api/repository/inventory"
)

var decimalPattern = regexp.MustCompile(`^(0|[1-9][0-9]{0,13})(\.[0-9]{1,6})?$`)
var operationKeyPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]*$`)

func quantity(value string) (string, *big.Rat, error) {
	value = strings.TrimSpace(value)
	if !decimalPattern.MatchString(value) {
		return "", nil, invalid("quantity must be positive numeric(20,6)")
	}
	n, ok := new(big.Rat).SetString(value)
	if !ok || n.Sign() <= 0 {
		return "", nil, invalid("quantity must be greater than zero")
	}
	return n.FloatString(6), n, nil
}
func balanceKey(owner, warehouse, item string, lot, hu *string, d *dto.BalanceDimension) repository.BalanceIdentity {
	return repository.BalanceIdentity{OwnerID: owner, WarehouseID: warehouse, LocationID: d.LocationID, ItemID: item, LotID: lot, HandlingUnitID: hu, InventoryStatusID: d.InventoryStatusID}
}
func balanceModel(id string, k repository.BalanceIdentity, uom string) *model.InventoryBalance {
	return &model.InventoryBalance{ID: id, OwnerID: k.OwnerID, WarehouseID: k.WarehouseID, LocationID: k.LocationID, ItemID: k.ItemID, LotID: k.LotID, HandlingUnitID: k.HandlingUnitID, InventoryStatusID: k.InventoryStatusID, OnHandQty: "0.000000", ReservedQty: "0.000000", UOMID: uom, VersionNo: 1}
}
func trimPointer(value **string, max int, label string) error {
	if *value == nil {
		return nil
	}
	v, err := normalizeNumber(**value, max)
	if err != nil {
		return invalid("invalid " + label)
	}
	*value = &v
	return nil
}
func normalizePosting(q *dto.PostingRequest) (string, *big.Rat, time.Time, error) {
	q.OperationKey = strings.TrimSpace(q.OperationKey)
	q.MovementTypeCode = strings.ToUpper(strings.TrimSpace(q.MovementTypeCode))
	q.OwnerID = strings.ToLower(strings.TrimSpace(q.OwnerID))
	q.WarehouseID = strings.ToLower(strings.TrimSpace(q.WarehouseID))
	q.ItemID = strings.ToLower(strings.TrimSpace(q.ItemID))
	if !uuid(q.OwnerID) || !uuid(q.WarehouseID) || !uuid(q.ItemID) {
		return "", nil, time.Time{}, invalid("owner_id, warehouse_id and item_id must be UUIDs")
	}
	if len(q.OperationKey) > 160 || !operationKeyPattern.MatchString(q.OperationKey) {
		return "", nil, time.Time{}, invalid("operation_key must use letters, digits, dot, underscore or hyphen")
	}
	if _, err := normalizeNumber(q.MovementTypeCode, 40); err != nil {
		return "", nil, time.Time{}, invalid("invalid movement_type_code")
	}
	if !businessCodePattern.MatchString(q.MovementTypeCode) {
		return "", nil, time.Time{}, invalid("invalid movement_type_code")
	}
	source, err := normalizeNumber(q.SourceDocumentID, 140)
	if err != nil {
		return "", nil, time.Time{}, invalid("invalid source_document_id")
	}
	q.SourceDocumentID = source
	for pointer, rule := range map[**string]struct {
		max   int
		label string
	}{&q.SourceLineID: {160, "source_line_id"}, &q.LotID: {120, "lot_id"}, &q.HandlingUnitID: {120, "handling_unit_id"}} {
		if err := trimPointer(pointer, rule.max, rule.label); err != nil {
			return "", nil, time.Time{}, err
		}
		if *pointer != nil && !identityID(**pointer, rule.max) {
			return "", nil, time.Time{}, invalid("invalid " + rule.label)
		}
	}
	if q.ReasonCodeID != nil && !uuid(*q.ReasonCodeID) {
		return "", nil, time.Time{}, invalid("reason_code_id must be a UUID")
	}
	if q.ReasonCodeID != nil {
		value := strings.ToLower(*q.ReasonCodeID)
		q.ReasonCodeID = &value
	}
	if q.Notes != nil {
		v := strings.TrimSpace(*q.Notes)
		if len(v) > 4000 {
			return "", nil, time.Time{}, invalid("notes is too long")
		}
		q.Notes = &v
	}
	if (q.From == nil) == (q.To == nil) {
		if q.From == nil {
			return "", nil, time.Time{}, invalid("from or to dimension is required")
		}
		if q.From.LocationID == q.To.LocationID && q.From.InventoryStatusID == q.To.InventoryStatusID {
			return "", nil, time.Time{}, invalid("posting must change location or inventory status")
		}
	}
	for _, d := range []*dto.BalanceDimension{q.From, q.To} {
		if d != nil {
			d.LocationID = strings.ToLower(strings.TrimSpace(d.LocationID))
			d.InventoryStatusID = strings.ToLower(strings.TrimSpace(d.InventoryStatusID))
			if !uuid(d.LocationID) || !uuid(d.InventoryStatusID) {
				return "", nil, time.Time{}, invalid("location and inventory status must be UUIDs")
			}
		}
	}
	businessDate, err := time.Parse("2006-01-02", q.BusinessDate)
	if err != nil || businessDate.Year() < 1 {
		return "", nil, time.Time{}, invalid("business_date must be YYYY-MM-DD")
	}
	qty, n, err := quantity(q.Quantity)
	if err != nil {
		return "", nil, time.Time{}, err
	}
	q.Quantity = qty
	for i, id := range q.SerialIDs {
		if !identityID(id, 160) {
			return "", nil, time.Time{}, invalid("invalid serial_id")
		}
		q.SerialIDs[i] = id
	}
	return qty, n, businessDate, nil
}
func postingFingerprint(q dto.PostingRequest) string {
	raw, _ := json.Marshal(q)
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
func validateDimension(ctx context.Context, r *repository.Repositories, warehouseID string, d *dto.BalanceDimension, target bool) error {
	if d == nil {
		return nil
	}
	location, err := r.Location.GetShared(ctx, d.LocationID)
	if err != nil {
		return reference(err, "location")
	}
	if location.WarehouseID != warehouseID {
		return invalid("location does not belong to warehouse")
	}
	if target && (!location.IsActive || location.IsLocked) {
		return invalid("target location is inactive or locked")
	}
	status, err := r.Catalog.InventoryStatus.GetShared(ctx, d.InventoryStatusID)
	if err != nil {
		return reference(err, "inventory status")
	}
	if target && !status.IsActive {
		return invalid("target inventory status is inactive")
	}
	return nil
}
func rat(value string) (*big.Rat, error) {
	v, ok := new(big.Rat).SetString(value)
	if !ok {
		return nil, fmt.Errorf("invalid stored quantity")
	}
	return v, nil
}
func (s *Service) PostMovement(ctx context.Context, q dto.PostingRequest, actor string) (dto.PostingResult, error) {
	qty, number, businessDate, err := normalizePosting(&q)
	if err != nil {
		return dto.PostingResult{}, err
	}
	if !uuid(actor) {
		return dto.PostingResult{}, invalid("authenticated actor must be a UUID")
	}
	fingerprint := postingFingerprint(q)
	if prior, e := s.repositories.Movement.GetByOperationKey(ctx, q.OperationKey); e == nil {
		if prior.OperationFingerprint == nil || *prior.OperationFingerprint != fingerprint {
			return dto.PostingResult{}, invalid("operation_key was already used for different posting data")
		}
		return dto.PostingResult{Movement: mapMovement(prior), IdempotentReplay: true}, nil
	} else if !errors.Is(e, repository.ErrNotFound) {
		return dto.PostingResult{}, e
	}
	movementID, err := newID("MOV")
	if err != nil {
		return dto.PostingResult{}, err
	}
	var fromResult, toResult *dto.BalanceResponse
	var fromBalanceID, toBalanceID string
	movement := model.InventoryMovement{ID: movementID, OwnerID: q.OwnerID, WarehouseID: q.WarehouseID, BusinessDate: businessDate, ItemID: q.ItemID, LotID: q.LotID, HandlingUnitID: q.HandlingUnitID, Quantity: qty, SourceDocumentID: q.SourceDocumentID, SourceLineID: q.SourceLineID, ReasonCodeID: q.ReasonCodeID, Notes: q.Notes, OperationKey: &q.OperationKey, OperationFingerprint: &fingerprint, CreatedBy: actor}
	if q.From != nil {
		movement.FromLocationID = &q.From.LocationID
		movement.FromStatusID = &q.From.InventoryStatusID
	}
	if q.To != nil {
		movement.ToLocationID = &q.To.LocationID
		movement.ToStatusID = &q.To.InventoryStatusID
	}
	err = s.repositories.Transaction(ctx, func(r *repository.Repositories) error {
		if err := r.Balance.LockStockKey(ctx, q.OwnerID, q.WarehouseID, q.ItemID); err != nil {
			return err
		}
		item, err := requireItem(ctx, r, q.OwnerID, q.ItemID)
		if err != nil {
			return err
		}
		warehouse, err := r.Warehouse.GetShared(ctx, q.WarehouseID)
		if err != nil {
			return reference(err, "warehouse")
		}
		if !warehouse.IsActive {
			return invalid("warehouse is inactive")
		}
		assignment, err := r.WarehouseOwner.GetShared(ctx, q.WarehouseID, q.OwnerID)
		if err != nil {
			return reference(err, "warehouse-owner assignment")
		}
		if !assignment.IsActive {
			return invalid("warehouse-owner assignment is inactive")
		}
		kind, err := r.MovementType.ByCodeShared(ctx, q.MovementTypeCode)
		if err != nil {
			return reference(err, "movement type")
		}
		if !kind.IsActive {
			return invalid("movement type is inactive")
		}
		movement.MovementTypeID = kind.ID
		movement.UOMID = item.BaseUOMID
		if err := validateDimension(ctx, r, q.WarehouseID, q.From, false); err != nil {
			return err
		}
		if err := validateDimension(ctx, r, q.WarehouseID, q.To, true); err != nil {
			return err
		}
		if item.LotControlled != (q.LotID != nil) {
			return invalid("lot presence must match item lot control")
		}
		if q.LotID != nil {
			lot, err := r.Lot.GetShared(ctx, *q.LotID)
			if err != nil {
				return reference(err, "lot")
			}
			if lot.OwnerID != q.OwnerID || lot.ItemID != q.ItemID {
				return invalid("lot does not match owner and item")
			}
		}
		if q.HandlingUnitID != nil {
			hu, err := r.HandlingUnit.GetShared(ctx, *q.HandlingUnitID)
			if err != nil {
				return reference(err, "handling unit")
			}
			if hu.OwnerID != q.OwnerID || hu.WarehouseID != q.WarehouseID {
				return invalid("handling unit does not match owner and warehouse")
			}
			relevant := q.To
			if relevant == nil {
				relevant = q.From
			}
			if hu.CurrentLocationID == nil || *hu.CurrentLocationID != relevant.LocationID {
				return invalid("handling unit current location does not match posting")
			}
			if q.From != nil && q.To != nil && q.From.LocationID != q.To.LocationID {
				return invalid("generic posting cannot relocate a handling unit")
			}
			if hu.IsClosed && q.From == nil {
				return invalid("cannot receive into a closed handling unit")
			}
		}
		var fromBalance, toBalance *model.InventoryBalance
		if q.From != nil {
			value, err := r.Balance.FindLocked(ctx, balanceKey(q.OwnerID, q.WarehouseID, q.ItemID, q.LotID, q.HandlingUnitID, q.From))
			if err != nil {
				return invalid("source balance does not exist")
			}
			if q.ExpectedSourceVersion != nil && value.VersionNo != *q.ExpectedSourceVersion {
				return invalid("source balance version changed; reload before retrying")
			}
			onHand, _ := rat(value.OnHandQty)
			reserved, _ := rat(value.ReservedQty)
			remaining := new(big.Rat).Sub(onHand, number)
			if remaining.Sign() < 0 || remaining.Cmp(reserved) < 0 {
				return invalid("insufficient unreserved source quantity")
			}
			if err := r.Balance.SetOnHand(ctx, value.ID, remaining.FloatString(6)); err != nil {
				return err
			}
			value.OnHandQty = remaining.FloatString(6)
			value.VersionNo++
			fromBalance = &value
			fromBalanceID = value.ID
		}
		if q.To != nil {
			key := balanceKey(q.OwnerID, q.WarehouseID, q.ItemID, q.LotID, q.HandlingUnitID, q.To)
			value, err := r.Balance.FindLocked(ctx, key)
			if errors.Is(err, repository.ErrNotFound) {
				id, e := newID("BAL")
				if e != nil {
					return e
				}
				value = *balanceModel(id, key, item.BaseUOMID)
				if e = r.Balance.Create(ctx, &value); e != nil {
					return e
				}
			} else if err != nil {
				return err
			}
			if q.ExpectedDestinationVersion != nil && value.VersionNo != *q.ExpectedDestinationVersion {
				return invalid("destination balance version changed; reload before retrying")
			}
			onHand, _ := rat(value.OnHandQty)
			onHand.Add(onHand, number)
			if err := r.Balance.SetOnHand(ctx, value.ID, onHand.FloatString(6)); err != nil {
				return err
			}
			value.OnHandQty = onHand.FloatString(6)
			value.VersionNo++
			toBalance = &value
			toBalanceID = value.ID
		}
		if item.SerialControlled {
			if number.Cmp(big.NewRat(1, 1)) != 0 || len(q.SerialIDs) != 1 {
				return invalid("serialized postings require quantity 1 and exactly one serial_id")
			}
			serial, err := r.Serial.GetShared(ctx, q.SerialIDs[0])
			if err != nil {
				return reference(err, "serial")
			}
			if serial.OwnerID != q.OwnerID || serial.ItemID != q.ItemID {
				return invalid("serial does not match owner and item")
			}
			movement.SerialID = &serial.ID
			state, err := r.SerialState.GetLocked(ctx, serial.ID)
			if q.From == nil {
				if err == nil {
					return invalid("serial is already in inventory")
				}
				if !errors.Is(err, repository.ErrNotFound) {
					return err
				}
				if toBalance == nil {
					return invalid("serial destination is missing")
				}
				if err := r.SerialState.Create(ctx, &model.SerialInventory{SerialID: serial.ID, BalanceID: toBalance.ID, OwnerID: q.OwnerID, ItemID: q.ItemID}); err != nil {
					return err
				}
			} else {
				if err != nil {
					return invalid("serial is not currently in inventory")
				}
				if fromBalance == nil || state.BalanceID != fromBalance.ID {
					return invalid("serial is not in source balance")
				}
				if toBalance == nil {
					if err := r.SerialState.Delete(ctx, serial.ID); err != nil {
						return err
					}
				} else if err := r.SerialState.Move(ctx, serial.ID, toBalance.ID); err != nil {
					return err
				}
			}
		} else if len(q.SerialIDs) != 0 {
			return invalid("serial_ids require a serial-controlled item")
		}
		if err := r.Movement.Create(ctx, &movement); err != nil {
			return err
		}
		return nil
	})
	if errors.Is(err, repository.ErrConflict) {
		if prior, e := s.repositories.Movement.GetByOperationKey(ctx, q.OperationKey); e == nil && prior.OperationFingerprint != nil && *prior.OperationFingerprint == fingerprint {
			return dto.PostingResult{Movement: mapMovement(prior), IdempotentReplay: true}, nil
		}
	}
	if err != nil {
		return dto.PostingResult{}, err
	}
	row, err := s.repositories.Movement.Get(ctx, movement.ID)
	if err != nil {
		return dto.PostingResult{}, err
	}
	if fromBalanceID != "" {
		if b, e := s.repositories.Balance.Get(ctx, fromBalanceID); e == nil {
			mapped := mapBalance(b)
			fromResult = &mapped
		}
	}
	if toBalanceID != "" {
		if b, e := s.repositories.Balance.Get(ctx, toBalanceID); e == nil {
			mapped := mapBalance(b)
			toResult = &mapped
		}
	}
	return dto.PostingResult{Movement: mapMovement(row), FromBalance: fromResult, ToBalance: toResult}, nil
}
