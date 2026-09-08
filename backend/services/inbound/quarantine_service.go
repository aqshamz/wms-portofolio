package inbound

import (
	"context"
	"errors"
	"math/big"
	"strings"

	dto "wms-api/dto/inbound"
	inventorydto "wms-api/dto/inventory"
	model "wms-api/models/inbound"
	repository "wms-api/repository/inbound"
	inventoryservice "wms-api/services/inventory"
)

func (s *Service) createQuarantineCase(ctx context.Context, inspection model.QualityInspection, batch repository.ReceiptInventoryContext, balanceID, quantity string, notes *string, actor string) error {
	kind, err := s.documentType(ctx, "QUARANTINE_CASE")
	if err != nil {
		return err
	}
	status, err := s.initialStatus(ctx, kind.ID)
	if err != nil {
		return err
	}
	id, err := s.generateID(ctx, kind.ID, batch.BusinessDate, batch.VendorID, batch.WarehouseID)
	if err != nil {
		return err
	}
	var parentCaseID *string
	if inspection.ParentInspectionID != nil {
		if parent, parentErr := s.repositories.QuarantineCase.GetByInspection(ctx, *inspection.ParentInspectionID); parentErr == nil {
			parentID := parent.ID
			parentCaseID = &parentID
		} else if !errors.Is(parentErr, repository.ErrNotFound) {
			return parentErr
		}
	}
	return s.repositories.QuarantineCase.Create(ctx, &model.QuarantineCase{ID: id, ParentQuarantineCaseID: parentCaseID, DocumentTypeID: kind.ID, StatusID: status.ID, ReceiptInventoryID: batch.ID, InspectionID: inspection.ID, QuarantineBalanceID: balanceID, OwnerID: batch.OwnerID, WarehouseID: batch.WarehouseID, QuarantineQty: quantity, UOMID: batch.BaseUOMID, Notes: notes, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1})
}

func (s *Service) ListQuarantineDispositionTypes(ctx context.Context, active *bool) ([]dto.QuarantineDispositionTypeResponse, error) {
	rows, err := s.repositories.QuarantineType.List(ctx, active)
	items := make([]dto.QuarantineDispositionTypeResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapDispositionType(row))
	}
	return items, err
}

func (s *Service) CreateQuarantineDisposition(ctx context.Context, caseID string, request dto.CreateQuarantineDispositionRequest, actor string) (dto.QuarantineCaseResponse, error) {
	if !inboundID(caseID, 140) || !inboundUUID(actor) || request.ExpectedCaseVersion < 1 || request.ExpectedBalanceVersion < 1 {
		return dto.QuarantineCaseResponse{}, invalid("invalid quarantine disposition")
	}
	quantity, quantityNumber, err := inboundQuantity(request.DispositionQty, "disposition_qty", false)
	if err != nil {
		return dto.QuarantineCaseResponse{}, err
	}
	if _, err = inboundDate(request.BusinessDate, "business_date"); err != nil {
		return dto.QuarantineCaseResponse{}, err
	}
	decidedAt, err := inboundTimestamp(request.DecidedAt, "decided_at")
	if err != nil {
		return dto.QuarantineCaseResponse{}, err
	}
	decisionReference, err := optional(request.ClientDecisionReference, 120, "client_decision_reference")
	if err != nil {
		return dto.QuarantineCaseResponse{}, err
	}
	decisionNotes, err := optional(request.DecisionNotes, 4000, "decision_notes")
	if err != nil {
		return dto.QuarantineCaseResponse{}, err
	}
	workInstructions, err := optional(request.WorkInstructions, 4000, "work_instructions")
	if err != nil {
		return dto.QuarantineCaseResponse{}, err
	}
	typeCode := strings.ToUpper(strings.TrimSpace(request.DispositionTypeCode))
	if !inboundID(typeCode, 40) {
		return dto.QuarantineCaseResponse{}, invalid("invalid disposition_type_code")
	}
	err = s.transaction(ctx, func(local *Service) error {
		caseModel, err := local.repositories.QuarantineCase.Lock(ctx, caseID)
		if err != nil {
			return err
		}
		if caseModel.VersionNo != request.ExpectedCaseVersion {
			return repository.ErrConcurrentWrite
		}
		caseRow, err := local.repositories.QuarantineCase.Get(ctx, caseID)
		if err != nil {
			return err
		}
		if caseRow.StatusCode != "OPEN" && caseRow.StatusCode != "PARTIALLY_DECIDED" {
			return state("quarantine case is not open for disposition")
		}
		total, err := storedRatio(caseRow.QuarantineQty)
		if err != nil {
			return err
		}
		disposed, err := storedRatio(caseRow.DisposedQty)
		if err != nil {
			return err
		}
		if new(big.Rat).Add(disposed, quantityNumber).Cmp(total) > 0 {
			return invalid("disposition_qty exceeds the undecided quarantine quantity")
		}
		kind, err := local.repositories.QuarantineType.ByCode(ctx, typeCode)
		if err != nil || !kind.IsActive {
			return invalid("disposition type is unavailable")
		}
		batch, err := local.repositories.ReceiptInventory.GetContext(ctx, caseModel.ReceiptInventoryID)
		if err != nil {
			return err
		}
		balance, err := local.repositories.Inventory.Balance.Get(ctx, caseModel.QuarantineBalanceID)
		if err != nil || balance.InventoryStatusCode != "QUARANTINE" {
			return state("quarantine source balance is unavailable")
		}
		if balance.VersionNo != request.ExpectedBalanceVersion {
			return repository.ErrConcurrentWrite
		}
		available, err := storedRatio(balance.AvailableQty)
		if err != nil || available.Cmp(quantityNumber) < 0 {
			return invalid("insufficient unreserved quarantine quantity")
		}
		if batch.HandlingUnitID != nil && available.Cmp(quantityNumber) != 0 {
			return invalid("handling-unit disposition must process its entire source balance")
		}
		item, err := local.repositories.Master.Catalog.Item.Get(ctx, batch.ItemID)
		if err != nil || !item.IsActive {
			return invalid("quarantine item is unavailable")
		}
		var targetLocationID *string
		movementTypeCode := ""
		var target *inventorydto.BalanceDimension
		if kind.RequiresReinspection {
			if request.TargetLocationID != nil {
				return invalid("target_location_id is not allowed for REWORK")
			}
			if workInstructions == nil {
				return invalid("work_instructions is required for REWORK")
			}
			pending, err := local.inventoryStatus(ctx, "QC_PENDING")
			if err != nil {
				return err
			}
			target = &inventorydto.BalanceDimension{LocationID: balance.LocationID, InventoryStatusID: pending.ID}
			movementTypeCode = "STATUS_CHANGE"
		} else if kind.ReleasesToAvailable {
			if request.TargetLocationID == nil {
				return invalid("target_location_id is required for ACCEPT")
			}
			location, err := local.validatePutawayTarget(ctx, caseModel.OwnerID, caseModel.WarehouseID, item, *request.TargetLocationID)
			if err != nil {
				return err
			}
			availableStatus, err := local.inventoryStatus(ctx, "AVAILABLE")
			if err != nil {
				return err
			}
			targetLocationID = &location.ID
			target = &inventorydto.BalanceDimension{LocationID: location.ID, InventoryStatusID: availableStatus.ID}
			movementTypeCode = "PUTAWAY"
		} else if kind.RemovesInventory {
			if request.TargetLocationID != nil {
				return invalid("target_location_id is not allowed for inventory removal")
			}
			if kind.RemovalMovementTypeID == nil {
				return state("disposition removal movement type is not configured")
			}
			movementType, err := local.repositories.Inventory.MovementType.Get(ctx, *kind.RemovalMovementTypeID)
			if err != nil || !movementType.IsActive {
				return state("disposition removal movement type is unavailable")
			}
			movementTypeCode = movementType.Code
		} else {
			return state("disposition type has no supported inventory action")
		}
		documentType, err := local.documentType(ctx, "QUARANTINE_DISPOSITION")
		if err != nil {
			return err
		}
		initial, err := local.initialStatus(ctx, documentType.ID)
		if err != nil {
			return err
		}
		dispositionID, err := local.generateID(ctx, documentType.ID, request.BusinessDate, batch.VendorID, batch.WarehouseID)
		if err != nil {
			return err
		}
		disposition := model.QuarantineDisposition{ID: dispositionID, QuarantineCaseID: caseID, DocumentTypeID: documentType.ID, StatusID: initial.ID, QuarantineDispositionTypeID: kind.ID, DispositionQty: quantity, UOMID: caseModel.UOMID, ClientDecisionReference: decisionReference, DecisionNotes: decisionNotes, DecidedAt: decidedAt, DecidedBy: actor, TargetLocationID: targetLocationID, CreatedBy: actor}
		if err := local.repositories.Disposition.Create(ctx, &disposition); err != nil {
			return err
		}
		serialIDs := make([]string, 0, 1)
		if batch.SerialID != nil {
			if quantityNumber.Cmp(big.NewRat(1, 1)) != 0 {
				return invalid("serialized quarantine disposition must process exactly one unit")
			}
			serialIDs = append(serialIDs, *batch.SerialID)
		}
		posted, err := inventoryservice.NewService(local.repositories.Inventory).PostMovement(ctx, inventorydto.PostingRequest{OperationKey: "inbound.quarantine." + dispositionID, MovementTypeCode: movementTypeCode, OwnerID: caseModel.OwnerID, WarehouseID: caseModel.WarehouseID, BusinessDate: request.BusinessDate, ItemID: batch.ItemID, LotID: batch.LotID, HandlingUnitID: batch.HandlingUnitID, From: &inventorydto.BalanceDimension{LocationID: balance.LocationID, InventoryStatusID: balance.InventoryStatusID}, To: target, Quantity: quantity, SerialIDs: serialIDs, ExpectedSourceVersion: &request.ExpectedBalanceVersion, SourceDocumentID: dispositionID, SourceLineID: &caseID, Notes: decisionNotes, RelocateHandlingUnit: kind.ReleasesToAvailable}, actor)
		if err != nil {
			return err
		}
		processed, err := local.transitionTarget(ctx, documentType.ID, initial.ID, "PROCESSED")
		if err != nil {
			return err
		}
		var resultingBalanceID *string
		if posted.ToBalance != nil {
			resultingBalanceID = &posted.ToBalance.ID
		}
		if err := local.repositories.Disposition.Process(ctx, dispositionID, processed.ID, posted.Movement.ID, resultingBalanceID); err != nil {
			return err
		}
		if kind.RequiresReinspection {
			if posted.ToBalance == nil {
				return state("rework disposition returned no QC balance")
			}
			taskID, err := newInboundID("RWK")
			if err != nil {
				return err
			}
			taskType, err := local.taskType(ctx, "REWORK")
			if err != nil {
				return err
			}
			taskStatus, err := local.initialTaskStatus(ctx)
			if err != nil {
				return err
			}
			priority, err := local.normalPriority(ctx)
			if err != nil {
				return err
			}
			if err := local.repositories.ReworkTask.Create(ctx, &model.ReworkTask{ID: taskID, QuarantineDispositionID: dispositionID, TaskTypeID: taskType.ID, TaskStatusID: taskStatus.ID, TaskPriorityID: priority.ID, SourceBalanceID: posted.ToBalance.ID, PlannedQty: quantity, CompletedQty: "0.000000", UOMID: caseModel.UOMID, WorkInstructions: workInstructions, CreatedBy: actor, UpdatedBy: &actor, VersionNo: 1}); err != nil {
				return err
			}
		}
		updatedCase, err := local.repositories.QuarantineCase.Get(ctx, caseID)
		if err != nil {
			return err
		}
		processedQty, err := storedRatio(updatedCase.DisposedQty)
		if err != nil {
			return err
		}
		targetCode := "PARTIALLY_DECIDED"
		closeCase := false
		if processedQty.Cmp(total) == 0 {
			targetCode, closeCase = "CLOSED", true
		}
		if updatedCase.StatusCode != targetCode {
			targetStatus, err := local.transitionTarget(ctx, caseModel.DocumentTypeID, caseModel.StatusID, targetCode)
			if err != nil {
				return err
			}
			if err := local.repositories.QuarantineCase.SetStatus(ctx, caseID, targetStatus.ID, actor, closeCase, caseModel.VersionNo); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return dto.QuarantineCaseResponse{}, err
	}
	return s.GetQuarantineCase(ctx, caseID)
}
