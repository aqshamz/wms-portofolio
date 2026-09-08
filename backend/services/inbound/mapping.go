package inbound

import (
	dto "wms-api/dto/inbound"
	model "wms-api/models/inbound"
	repository "wms-api/repository/inbound"
)

func dateText(value interface{ Format(string) string }) string { return value.Format("2006-01-02") }
func optionalDateText(value interface{ Format(string) string }) *string {
	text := value.Format("2006-01-02")
	return &text
}
func mapPurchaseOrder(row repository.PurchaseOrderRow) dto.PurchaseOrderResponse {
	return dto.PurchaseOrderResponse{ID: row.ID, OwnerID: row.OwnerID, OwnerCode: row.OwnerCode, VendorID: row.VendorID, VendorCode: row.VendorCode, VendorName: row.VendorName, WarehouseID: row.WarehouseID, WarehouseCode: row.WarehouseCode, BusinessDate: dateText(row.BusinessDate), PurchaseOrderNo: row.PurchaseOrderNo, OrderedAt: row.OrderedAt, ExpectedArrivalAt: row.ExpectedArrivalAt, StatusCode: row.StatusCode, Notes: row.Notes, VersionNo: row.VersionNo, CreatedAt: row.CreatedAt}
}
func mapPurchaseOrderLine(row repository.PurchaseOrderLineRow) dto.PurchaseOrderLineResponse {
	var expiry *string
	if row.ExpectedExpiryDate != nil {
		expiry = optionalDateText(*row.ExpectedExpiryDate)
	}
	return dto.PurchaseOrderLineResponse{ID: row.ID, LineNo: row.LineNo, ItemID: row.ItemID, ItemCode: row.ItemCode, ItemName: row.ItemName, OrderedQty: row.OrderedQty, ScheduledQty: row.ScheduledQty, CompletedReceiptQty: row.CompletedReceiptQty, OverReceiptTolerancePct: row.OverReceiptTolerancePct, UnderReceiptTolerancePct: row.UnderReceiptTolerancePct, UOMID: row.UOMID, UOMCode: row.UOMCode, VendorItemCode: row.VendorItemCode, ExpectedLotNo: row.ExpectedLotNo, ExpectedExpiryDate: expiry, Notes: row.Notes}
}
func mapInboundOrder(row repository.InboundOrderRow) dto.InboundOrderResponse {
	return dto.InboundOrderResponse{ID: row.ID, OwnerID: row.OwnerID, OwnerCode: row.OwnerCode, VendorID: row.VendorID, VendorCode: row.VendorCode, VendorName: row.VendorName, WarehouseID: row.WarehouseID, WarehouseCode: row.WarehouseCode, BusinessDate: dateText(row.BusinessDate), ExpectedArrivalAt: row.ExpectedArrivalAt, ExternalReference: row.ExternalReference, SupplierReference: row.SupplierReference, StatusCode: row.StatusCode, Notes: row.Notes, VersionNo: row.VersionNo, CreatedAt: row.CreatedAt}
}
func mapInboundOrderLine(row repository.InboundOrderLineRow) dto.InboundOrderLineResponse {
	var expiry *string
	if row.ExpectedExpiryDate != nil {
		expiry = optionalDateText(*row.ExpectedExpiryDate)
	}
	return dto.InboundOrderLineResponse{ID: row.ID, PurchaseOrderLineID: row.PurchaseOrderLineID, LineNo: row.LineNo, ItemID: row.ItemID, ItemCode: row.ItemCode, ItemName: row.ItemName, ExpectedQty: row.ExpectedQty, CompletedReceiptQty: row.CompletedReceiptQty, UOMID: row.UOMID, UOMCode: row.UOMCode, ExpectedLotNo: row.ExpectedLotNo, ExpectedExpiryDate: expiry, CustomerLineReference: row.CustomerLineReference, Notes: row.Notes}
}
func mapReceipt(row repository.ReceiptRow) dto.ReceiptResponse {
	return dto.ReceiptResponse{ID: row.ID, InboundID: row.InboundID, OwnerID: row.OwnerID, OwnerCode: row.OwnerCode, WarehouseID: row.WarehouseID, WarehouseCode: row.WarehouseCode, BusinessDate: dateText(row.BusinessDate), ReceivedAt: row.ReceivedAt, DockLocationID: row.DockLocationID, VehicleNumber: row.VehicleNumber, SealNumber: row.SealNumber, DeliveryNoteNo: row.DeliveryNoteNo, StatusCode: row.StatusCode, Notes: row.Notes, VersionNo: row.VersionNo, CreatedAt: row.CreatedAt}
}
func mapReceiptLine(row repository.ReceiptLineRow) dto.ReceiptLineResponse {
	return dto.ReceiptLineResponse{ID: row.ID, InboundLineID: row.InboundLineID, LineNo: row.LineNo, ItemID: row.ItemID, ItemCode: row.ItemCode, ItemName: row.ItemName, ReceivedQty: row.ReceivedQty, RejectedQty: row.RejectedQty, ExceptionNotes: row.ExceptionNotes, ExceptionTypeCode: row.ExceptionTypeCode, AcceptedQty: row.AcceptedQty, BatchedQty: row.BatchedQty, UOMID: row.UOMID, UOMCode: row.UOMCode, Batches: make([]dto.ReceiptBatchResponse, 0)}
}
func mapReceiptBatch(row repository.ReceiptInventoryRow) dto.ReceiptBatchResponse {
	return dto.ReceiptBatchResponse{ID: row.ID, ItemID: row.ItemID, SourceQty: row.SourceQty, SourceUOMID: row.SourceUOMID, SourceUOMCode: row.SourceUOMCode, BaseQty: row.BaseQty, BaseUOMID: row.BaseUOMID, BaseUOMCode: row.BaseUOMCode, LotID: row.LotID, LotNumber: row.LotNumber, HandlingUnitID: row.HandlingUnitID, HandlingUnitBarcode: row.HandlingUnitBarcode, SerialID: row.SerialID, SerialNo: row.SerialNo, ReceivedLocationID: row.ReceivedLocationID, ReceivedLocationCode: row.ReceivedLocationCode, InitialInventoryStatusID: row.InitialInventoryStatusID, InitialInventoryStatusCode: row.InitialInventoryStatusCode, InitialBalanceID: row.InitialBalanceID}
}

func mapQualityInspection(row repository.QualityInspectionRow) dto.QualityInspectionResponse {
	return dto.QualityInspectionResponse{ID: row.ID, ReceiptInventoryID: row.ReceiptInventoryID, ParentInspectionID: row.ParentInspectionID, SourceBalanceID: row.SourceBalanceID, ReceiptID: row.ReceiptID, OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, ItemID: row.ItemID, ItemCode: row.ItemCode, LotNumber: row.LotNumber, LocationCode: row.LocationCode, QualityStatusCode: row.QualityStatusCode, InspectionResultCode: row.InspectionResultCode, InspectedQty: row.InspectedQty, PassedQty: row.PassedQty, FailedQty: row.FailedQty, InspectedAt: row.InspectedAt, InspectedBy: row.InspectedBy, CancelledAt: row.CancelledAt, CancelledBy: row.CancelledBy, CancellationReason: row.CancellationReason, Notes: row.Notes, VersionNo: row.VersionNo, CreatedAt: row.CreatedAt}
}

func mapPutawayTask(row repository.PutawayTaskRow) dto.PutawayTaskResponse {
	return dto.PutawayTaskResponse{ID: row.ID, InspectionID: row.InspectionID, ReceiptInventoryID: row.ReceiptInventoryID, SourceBalanceID: row.SourceBalanceID, OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, ItemID: row.ItemID, ItemCode: row.ItemCode, LotID: row.LotID, LotNumber: row.LotNumber, HandlingUnitID: row.HandlingUnitID, SourceLocationID: row.SourceLocationID, SourceLocationCode: row.SourceLocationCode, TargetLocationID: row.TargetLocationID, TargetLocationCode: row.TargetLocationCode, PlannedQty: row.PlannedQty, CompletedQty: row.CompletedQty, UOMID: row.UOMID, TaskStatusCode: row.TaskStatusCode, TaskPriorityCode: row.TaskPriorityCode, AssignedTo: row.AssignedTo, StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, InventoryMovementID: row.InventoryMovementID, ResultingBalanceID: row.ResultingBalanceID, ReversalMovementID: row.ReversalMovementID, ReversedAt: row.ReversedAt, ReversedBy: row.ReversedBy, ReversalReason: row.ReversalReason, ReplacementInspectionID: row.ReplacementInspectionID, VersionNo: row.VersionNo, CreatedAt: row.CreatedAt}
}

func mapDispositionType(row model.QuarantineDispositionType) dto.QuarantineDispositionTypeResponse {
	return dto.QuarantineDispositionTypeResponse{ID: row.ID, Code: row.Code, Name: row.Name, Description: row.Description, ReleasesToAvailable: row.ReleasesToAvailable, RequiresReinspection: row.RequiresReinspection, RemovesInventory: row.RemovesInventory, IsActive: row.IsActive}
}

func mapDisposition(row repository.QuarantineDispositionRow) dto.QuarantineDispositionResponse {
	return dto.QuarantineDispositionResponse{ID: row.ID, QuarantineCaseID: row.QuarantineCaseID, DispositionTypeCode: row.DispositionTypeCode, StatusCode: row.StatusCode, DispositionQty: row.DispositionQty, UOMID: row.UOMID, ClientDecisionReference: row.ClientDecisionReference, DecisionNotes: row.DecisionNotes, DecidedAt: row.DecidedAt, DecidedBy: row.DecidedBy, ProcessedAt: row.ProcessedAt, InventoryMovementID: row.InventoryMovementID, ResultingBalanceID: row.ResultingBalanceID, TargetLocationID: row.TargetLocationID, CreatedAt: row.CreatedAt}
}

func mapQuarantineCase(row repository.QuarantineCaseRow) dto.QuarantineCaseResponse {
	return dto.QuarantineCaseResponse{ID: row.ID, ParentQuarantineCaseID: row.ParentQuarantineCaseID, InspectionID: row.InspectionID, ReceiptInventoryID: row.ReceiptInventoryID, QuarantineBalanceID: row.QuarantineBalanceID, OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, ItemID: row.ItemID, ItemCode: row.ItemCode, LotNumber: row.LotNumber, LocationCode: row.LocationCode, StatusCode: row.StatusCode, QuarantineQty: row.QuarantineQty, DisposedQty: row.DisposedQty, UOMID: row.UOMID, OpenedAt: row.OpenedAt, ClosedAt: row.ClosedAt, Notes: row.Notes, VersionNo: row.VersionNo, Dispositions: make([]dto.QuarantineDispositionResponse, 0)}
}

func mapInboundException(row model.InboundException) dto.InboundExceptionResponse {
	return dto.InboundExceptionResponse{ID: row.ID, OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, SourceDocumentID: row.SourceDocumentID, SourceLineID: row.SourceLineID, ExceptionTypeCode: row.ExceptionTypeCode, ExpectedQty: row.ExpectedQty, ActualQty: row.ActualQty, VarianceQty: row.VarianceQty, Notes: row.Notes, CreatedAt: row.CreatedAt, CreatedBy: row.CreatedBy}
}

func mapReworkTask(row repository.ReworkTaskRow) dto.ReworkTaskResponse {
	return dto.ReworkTaskResponse{ID: row.ID, QuarantineDispositionID: row.QuarantineDispositionID, QuarantineCaseID: row.QuarantineCaseID, SourceBalanceID: row.SourceBalanceID, OwnerID: row.OwnerID, WarehouseID: row.WarehouseID, ItemID: row.ItemID, ItemCode: row.ItemCode, TaskStatusCode: row.TaskStatusCode, TaskPriorityCode: row.TaskPriorityCode, PlannedQty: row.PlannedQty, CompletedQty: row.CompletedQty, UOMID: row.UOMID, AssignedTo: row.AssignedTo, WorkInstructions: row.WorkInstructions, ResultNotes: row.ResultNotes, StartedAt: row.StartedAt, CompletedAt: row.CompletedAt, ReinspectionID: row.ReinspectionID, VersionNo: row.VersionNo, CreatedAt: row.CreatedAt}
}
