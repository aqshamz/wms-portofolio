package outbound

import (
	dto "wms-api/dto/outbound"
	repository "wms-api/repository/outbound"
)

func mapCheckLine(v repository.CheckLineRow) dto.CheckLineResponse {
	return dto.CheckLineResponse{ID: v.ID, StagingLineID: v.StagingLineID, ItemCode: v.ItemCode, LotNumber: v.LotNumber, LineNo: v.LineNo, ExpectedQty: v.ExpectedQty, CheckedQty: v.CheckedQty, ExceptionQty: v.ExceptionQty, ResultCode: v.ResultCode, Notes: v.Notes, CheckedAt: v.CheckedAt}
}
func mapCheck(v repository.CheckRow) dto.CheckResponse {
	return dto.CheckResponse{ID: v.ID, ParentCheckID: v.ParentCheckID, StagingID: v.StagingID, OutboundID: v.OutboundID, ClientDeliveryOrderNo: v.ClientDeliveryOrderNo, OwnerID: v.OwnerID, WarehouseID: v.WarehouseID, StatusCode: v.StatusCode, CheckedAt: v.CheckedAt, Notes: v.Notes, LineCount: v.LineCount}
}
func mapException(v repository.CheckExceptionRow) dto.CheckExceptionResponse {
	return dto.CheckExceptionResponse{ID: v.ID, CheckLineID: v.OutboundCheckLineID, ResultCode: v.ResultCode, StatusCode: v.StatusCode, ExceptionQty: v.ExceptionQty, ResolvedQty: v.ResolvedQty, OpenedAt: v.OpenedAt, ResolvedAt: v.ResolvedAt, Notes: v.Notes}
}
func mapPackingLine(v repository.PackingLineRow) dto.PackingLineResponse {
	return dto.PackingLineResponse{ID: v.ID, OutboundCheckLineID: v.OutboundCheckLineID, PickTaskID: v.PickTaskID, ItemCode: v.ItemCode, LotNumber: v.LotNumber, SourceBalanceID: v.SourceBalanceID, SourceBalanceVersion: v.SourceBalanceVersion, PackingBalanceID: v.PackingBalanceID, HandlingUnitID: v.HandlingUnitID, PackedQty: v.PackedQty, UOMCode: v.UOMCode, MovementID: v.MovementID}
}
func mapPacking(v repository.PackingRow) dto.PackingResponse {
	return dto.PackingResponse{ID: v.ID, OutboundID: v.OutboundID, ClientDeliveryOrderNo: v.ClientDeliveryOrderNo, OwnerID: v.OwnerID, WarehouseID: v.WarehouseID, PackingLocationID: v.PackingLocationID, PackingLocationCode: v.PackingLocationCode, StatusCode: v.StatusCode, PackedAt: v.PackedAt, LineCount: v.LineCount}
}
func mapShipmentLine(v repository.ShipmentLineRow) dto.ShipmentLineResponse {
	return dto.ShipmentLineResponse{ID: v.ID, PackingLineID: v.PackingLineID, OutboundID: v.OutboundID, ItemCode: v.ItemCode, LotNumber: v.LotNumber, SourceBalanceID: v.SourceBalanceID, SourceBalanceVersion: v.SourceBalanceVersion, ShippedQty: v.ShippedQty, UOMCode: v.UOMCode, MovementID: v.MovementID}
}
func mapShipment(v repository.ShipmentRow) dto.ShipmentResponse {
	return dto.ShipmentResponse{ID: v.ID, OwnerID: v.OwnerID, WarehouseID: v.WarehouseID, CarrierServiceID: v.CarrierServiceID, CarrierCode: v.CarrierCode, CarrierServiceCode: v.CarrierServiceCode, BusinessDate: v.BusinessDate.Format("2006-01-02"), StatusCode: v.StatusCode, RouteReference: v.RouteReference, TrackingNumber: v.TrackingNumber, VehicleNumber: v.VehicleNumber, SealNumber: v.SealNumber, ShippedAt: v.ShippedAt, Notes: v.Notes, OrderCount: v.OrderCount, LineCount: v.LineCount}
}
func mapDeliveryLine(v repository.DeliveryLineRow) dto.DeliveryLineResponse {
	return dto.DeliveryLineResponse{ID: v.ID, ShipmentLineID: v.ShipmentLineID, ItemCode: v.ItemCode, LotNumber: v.LotNumber, PlannedQty: v.PlannedQty, DeliveredQty: v.DeliveredQty, ReturnedQty: v.ReturnedQty, UOMCode: v.UOMCode}
}
func mapDeliveryEvent(v repository.DeliveryEventRow) dto.DeliveryEventResponse {
	return dto.DeliveryEventResponse{ID: v.ID, EventTypeCode: v.EventTypeCode, FailureReasonCode: v.FailureReasonCode, EventAt: v.EventAt, RecipientName: v.RecipientName, ProofReference: v.ProofReference, ProofURI: v.ProofURI, Notes: v.Notes, Latitude: v.Latitude, Longitude: v.Longitude}
}
func mapDelivery(v repository.DeliveryRow) dto.DeliveryResponse {
	return dto.DeliveryResponse{ID: v.ID, ShipmentID: v.ShipmentID, OutboundID: v.OutboundID, ClientDeliveryOrderNo: v.ClientDeliveryOrderNo, OwnerID: v.OwnerID, WarehouseID: v.WarehouseID, BusinessDate: v.BusinessDate.Format("2006-01-02"), StatusCode: v.StatusCode, PlannedDeliveryAt: v.PlannedDeliveryAt, ArrivedAt: v.ArrivedAt, DeliveredAt: v.DeliveredAt, RecipientName: v.RecipientName, RecipientReference: v.RecipientReference, ProofReference: v.ProofReference, ProofURI: v.ProofURI, Notes: v.Notes}
}
