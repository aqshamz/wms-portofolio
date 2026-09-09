package outbound

import (
	dto "wms-api/dto/outbound"
	repository "wms-api/repository/outbound"
)

func mapOrder(v repository.OutboundOrderRow, lines []repository.OutboundOrderLineRow) dto.OutboundOrderResponse {
	r := dto.OutboundOrderResponse{ID: v.ID, OwnerID: v.OwnerID, OwnerCode: v.OwnerCode, CustomerID: v.CustomerID, CustomerCode: v.CustomerCode, CustomerName: v.CustomerName, ShipToPartnerID: v.ShipToPartnerID, ShipToCode: v.ShipToCode, WarehouseID: v.WarehouseID, WarehouseCode: v.WarehouseCode, BusinessDate: v.BusinessDate.Format("2006-01-02"), RequestedShipAt: v.RequestedShipAt, ClientDeliveryOrderNo: v.ClientDeliveryOrderNo, CustomerOrderNo: v.CustomerOrderNo, ExternalReference: v.ExternalReference, ShipToName: v.ShipToName, ShipToAddress1: v.ShipToAddress1, ShipToAddress2: v.ShipToAddress2, ShipToCity: v.ShipToCity, ShipToProvince: v.ShipToProvince, ShipToPostalCode: v.ShipToPostalCode, ShipToCountryCode: v.ShipToCountryCode, StatusCode: v.StatusCode, Notes: v.Notes, VersionNo: v.VersionNo, CreatedAt: v.CreatedAt}
	for _, x := range lines {
		r.Lines = append(r.Lines, dto.OutboundOrderLineResponse{ID: x.ID, LineNo: x.LineNo, ItemID: x.ItemID, ItemCode: x.ItemCode, ItemName: x.ItemName, OrderedQty: x.OrderedQty, AllocatedQty: x.AllocatedQty, PickedQty: x.PickedQty, CheckedQty: x.CheckedQty, RejectedQty: x.RejectedQty, ShortAcceptedQty: x.ShortAcceptedQty, PackedQty: x.PackedQty, ShippedQty: x.ShippedQty, DeliveredQty: x.DeliveredQty, UOMID: x.UOMID, UOMCode: x.UOMCode, RequestedLotNo: x.RequestedLotNo, CustomerLineReference: x.CustomerLineReference, Notes: x.Notes})
	}
	return r
}
func mapReservation(v repository.ReservationRow) dto.ReservationResponse {
	return dto.ReservationResponse{ID: v.ID, OutboundID: v.OutboundID, OutboundLineID: v.OutboundLineID, BalanceID: v.BalanceID, ItemCode: v.ItemCode, LocationCode: v.LocationCode, LotNumber: v.LotNumber, ReservedQty: v.ReservedQty, PickedQty: v.PickedQty, UOMID: v.UOMID, StatusCode: v.StatusCode, ReservedAt: v.ReservedAt}
}
func mapWave(v repository.OutboundWaveRow, orders []repository.WaveOrderRow) dto.WaveResponse {
	r := dto.WaveResponse{ID: v.ID, OwnerID: v.OwnerID, OwnerCode: v.OwnerCode, WarehouseID: v.WarehouseID, WarehouseCode: v.WarehouseCode, WaveTypeCode: v.WaveTypeCode, PickingStrategyID: v.PickingStrategyID, BusinessDate: v.BusinessDate.Format("2006-01-02"), PlannedReleaseAt: v.PlannedReleaseAt, ReleasedAt: v.ReleasedAt, CompletedAt: v.CompletedAt, StatusCode: v.StatusCode, Notes: v.Notes, VersionNo: v.VersionNo, OrderCount: v.OrderCount, PickTaskCount: v.PickTaskCount}
	for _, x := range orders {
		r.Orders = append(r.Orders, dto.WaveOrderResponse{OutboundID: x.OutboundID, ClientDeliveryOrderNo: x.ClientDeliveryOrderNo, StatusCode: x.StatusCode})
	}
	return r
}
func mapPick(v repository.PickTaskRow) dto.PickTaskResponse {
	return dto.PickTaskResponse{ID: v.ID, WaveID: v.WaveID, ReservationID: v.ReservationID, OutboundID: v.OutboundID, ClientDeliveryOrderNo: v.ClientDeliveryOrderNo, OutboundLineID: v.OutboundLineID, ItemCode: v.ItemCode, SourceLocationID: v.SourceLocationID, SourceLocationCode: v.SourceLocationCode, TargetLocationID: v.TargetLocationID, TargetLocationCode: v.TargetLocationCode, LotNumber: v.LotNumber, PlannedQty: v.PlannedQty, PickedQty: v.PickedQty, ShortQty: v.ShortQty, UOMID: v.UOMID, StatusCode: v.TaskStatusCode, PriorityCode: v.PriorityCode, AssignedTo: v.AssignedTo, BalanceID: v.BalanceID, BalanceVersion: v.BalanceVersion, StartedAt: v.StartedAt, CompletedAt: v.CompletedAt}
}
func mapStaging(v repository.StagingRow, lines []repository.StagingLineRow) dto.StagingResponse {
	r := dto.StagingResponse{ID: v.ID, OutboundID: v.OutboundID, ClientDeliveryOrderNo: v.ClientDeliveryOrderNo, WaveID: v.WaveID, WarehouseID: v.WarehouseID, WarehouseCode: v.WarehouseCode, StagingLocationID: v.StagingLocationID, StagingLocationCode: v.StagingLocationCode, StatusCode: v.StatusCode, StagedAt: v.StagedAt, Notes: v.Notes, LineCount: v.LineCount}
	for _, x := range lines {
		r.Lines = append(r.Lines, dto.StagingLineResponse{ID: x.ID, PickExecutionID: x.PickExecutionID, PickTaskID: x.PickTaskID, StagingBalanceID: x.StagingBalanceID, ItemCode: x.ItemCode, StagedQty: x.StagedQty, RemovedQty: x.RemovedQty, UOMCode: x.UOMCode})
	}
	return r
}
