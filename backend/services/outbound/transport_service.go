package outbound

import (
	"context"
	"strings"

	dto "wms-api/dto/outbound"
	model "wms-api/models/outbound"
	repository "wms-api/repository/outbound"
)

func mapCarrier(v model.Carrier) dto.CarrierResponse {
	return dto.CarrierResponse{ID: v.ID, BusinessPartnerID: v.BusinessPartnerID, Code: v.Code, Name: v.Name, IsActive: v.IsActive}
}

func mapCarrierService(v repository.CarrierServiceRow) dto.CarrierServiceResponse {
	return dto.CarrierServiceResponse{ID: v.ID, CarrierID: v.CarrierID, CarrierCode: v.CarrierCode, Code: v.Code, Name: v.Name, IsActive: v.IsActive}
}

func mapCarrierDriver(v repository.CarrierDriverRow) dto.CarrierDriverResponse {
	return dto.CarrierDriverResponse{ID: v.ID, CarrierID: v.CarrierID, CarrierCode: v.CarrierCode, AccountID: v.AccountID, Code: v.Code, Name: v.Name, PhoneNumber: v.PhoneNumber, LicenseNumber: v.LicenseNumber, IsActive: v.IsActive, CreatedAt: v.CreatedAt}
}

func mapShipmentDriver(v repository.ShipmentDriverRow) dto.ShipmentDriverResponse {
	return dto.ShipmentDriverResponse{DriverID: v.DriverID, DriverCode: v.DriverCode, DriverName: v.DriverName, PhoneNumber: v.PhoneNumber, LicenseNumber: v.LicenseNumber, IsPrimary: v.IsPrimary, AssignedAt: v.AssignedAt, AssignedBy: v.AssignedBy}
}

func transportCode(value, label string) (string, error) {
	value, err := clean(value, 40, label)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(value), nil
}

func (s *Service) CreateCarrier(ctx context.Context, q dto.CreateCarrierRequest) (dto.CarrierResponse, error) {
	code, err := transportCode(q.Code, "code")
	if err != nil {
		return dto.CarrierResponse{}, err
	}
	name, err := clean(q.Name, 150, "name")
	if err != nil {
		return dto.CarrierResponse{}, err
	}
	if q.BusinessPartnerID != nil && !validUUID(*q.BusinessPartnerID) {
		return dto.CarrierResponse{}, invalid("invalid business_partner_id")
	}
	v := model.Carrier{BusinessPartnerID: q.BusinessPartnerID, Code: code, Name: name, IsActive: true}
	if err := s.repositories.Carrier.Create(ctx, &v); err != nil {
		return dto.CarrierResponse{}, err
	}
	return mapCarrier(v), nil
}

func (s *Service) ListCarriers(ctx context.Context, active *bool, search string) ([]dto.CarrierResponse, error) {
	rows, err := s.repositories.Carrier.List(ctx, active, strings.TrimSpace(search))
	if err != nil {
		return nil, err
	}
	out := make([]dto.CarrierResponse, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapCarrier(v))
	}
	return out, nil
}

func (s *Service) GetCarrier(ctx context.Context, id string) (dto.CarrierResponse, error) {
	if !validUUID(id) {
		return dto.CarrierResponse{}, invalid("invalid carrier_id")
	}
	v, err := s.repositories.Carrier.Get(ctx, id)
	if err != nil {
		return dto.CarrierResponse{}, err
	}
	return mapCarrier(v), nil
}

func (s *Service) UpdateCarrier(ctx context.Context, id string, q dto.UpdateCarrierRequest) (dto.CarrierResponse, error) {
	if !validUUID(id) {
		return dto.CarrierResponse{}, invalid("invalid carrier_id")
	}
	code, err := transportCode(q.Code, "code")
	if err != nil {
		return dto.CarrierResponse{}, err
	}
	name, err := clean(q.Name, 150, "name")
	if err != nil {
		return dto.CarrierResponse{}, err
	}
	if q.BusinessPartnerID != nil && !validUUID(*q.BusinessPartnerID) {
		return dto.CarrierResponse{}, invalid("invalid business_partner_id")
	}
	values := map[string]interface{}{"business_partner_id": q.BusinessPartnerID, "code": code, "name": name}
	if q.IsActive != nil {
		values["is_active"] = *q.IsActive
	}
	err = s.repositories.Carrier.Update(ctx, id, values)
	if err != nil {
		return dto.CarrierResponse{}, err
	}
	return s.GetCarrier(ctx, id)
}

func (s *Service) CreateCarrierService(ctx context.Context, carrierID string, q dto.CreateCarrierServiceRequest) (dto.CarrierServiceResponse, error) {
	if !validUUID(carrierID) {
		return dto.CarrierServiceResponse{}, invalid("invalid carrier_id")
	}
	carrier, err := s.repositories.Carrier.Get(ctx, carrierID)
	if err != nil {
		return dto.CarrierServiceResponse{}, err
	}
	if !carrier.IsActive {
		return dto.CarrierServiceResponse{}, state("carrier is inactive")
	}
	code, err := transportCode(q.Code, "code")
	if err != nil {
		return dto.CarrierServiceResponse{}, err
	}
	name, err := clean(q.Name, 100, "name")
	if err != nil {
		return dto.CarrierServiceResponse{}, err
	}
	v := model.CarrierService{CarrierID: carrierID, Code: code, Name: name, IsActive: true}
	if err := s.repositories.CarrierService.Create(ctx, &v); err != nil {
		return dto.CarrierServiceResponse{}, err
	}
	return s.GetCarrierService(ctx, carrierID, v.ID)
}

func (s *Service) ListCarrierServices(ctx context.Context, carrierID string, active *bool) ([]dto.CarrierServiceResponse, error) {
	if !validUUID(carrierID) {
		return nil, invalid("invalid carrier_id")
	}
	if _, err := s.repositories.Carrier.Get(ctx, carrierID); err != nil {
		return nil, err
	}
	rows, err := s.repositories.CarrierService.List(ctx, carrierID, active)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CarrierServiceResponse, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapCarrierService(v))
	}
	return out, nil
}

func (s *Service) GetCarrierService(ctx context.Context, carrierID, id string) (dto.CarrierServiceResponse, error) {
	if !validUUID(carrierID) || !validUUID(id) {
		return dto.CarrierServiceResponse{}, invalid("invalid carrier service identity")
	}
	v, err := s.repositories.CarrierService.Get(ctx, id)
	if err != nil {
		return dto.CarrierServiceResponse{}, err
	}
	if v.CarrierID != carrierID {
		return dto.CarrierServiceResponse{}, repository.ErrNotFound
	}
	return mapCarrierService(v), nil
}

func (s *Service) UpdateCarrierService(ctx context.Context, carrierID, id string, q dto.UpdateCarrierServiceRequest) (dto.CarrierServiceResponse, error) {
	if !validUUID(carrierID) || !validUUID(id) {
		return dto.CarrierServiceResponse{}, invalid("invalid carrier service identity")
	}
	code, err := transportCode(q.Code, "code")
	if err != nil {
		return dto.CarrierServiceResponse{}, err
	}
	name, err := clean(q.Name, 100, "name")
	if err != nil {
		return dto.CarrierServiceResponse{}, err
	}
	values := map[string]interface{}{"code": code, "name": name}
	if q.IsActive != nil {
		values["is_active"] = *q.IsActive
	}
	if err := s.repositories.CarrierService.Update(ctx, id, carrierID, values); err != nil {
		return dto.CarrierServiceResponse{}, err
	}
	return s.GetCarrierService(ctx, carrierID, id)
}

func (s *Service) CreateCarrierDriver(ctx context.Context, carrierID string, q dto.CreateCarrierDriverRequest, actor string) (dto.CarrierDriverResponse, error) {
	if !validUUID(carrierID) || !validUUID(actor) {
		return dto.CarrierDriverResponse{}, invalid("invalid carrier driver identity")
	}
	carrier, err := s.repositories.Carrier.Get(ctx, carrierID)
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	if !carrier.IsActive {
		return dto.CarrierDriverResponse{}, state("carrier is inactive")
	}
	if q.AccountID != nil && !validUUID(*q.AccountID) {
		return dto.CarrierDriverResponse{}, invalid("invalid account_id")
	}
	code, err := transportCode(q.Code, "code")
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	name, err := clean(q.Name, 150, "name")
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	phone, err := optional(q.PhoneNumber, 50, "phone_number")
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	license, err := optional(q.LicenseNumber, 80, "license_number")
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	v := model.CarrierDriver{CarrierID: carrierID, AccountID: q.AccountID, Code: code, Name: name, PhoneNumber: phone, LicenseNumber: license, IsActive: true, CreatedBy: &actor}
	if err := s.repositories.CarrierDriver.Create(ctx, &v); err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	return s.GetCarrierDriver(ctx, carrierID, v.ID)
}

func (s *Service) ListCarrierDrivers(ctx context.Context, carrierID string, active *bool) ([]dto.CarrierDriverResponse, error) {
	if !validUUID(carrierID) {
		return nil, invalid("invalid carrier_id")
	}
	if _, err := s.repositories.Carrier.Get(ctx, carrierID); err != nil {
		return nil, err
	}
	rows, err := s.repositories.CarrierDriver.List(ctx, carrierID, active)
	if err != nil {
		return nil, err
	}
	out := make([]dto.CarrierDriverResponse, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapCarrierDriver(v))
	}
	return out, nil
}

func (s *Service) GetCarrierDriver(ctx context.Context, carrierID, id string) (dto.CarrierDriverResponse, error) {
	if !validUUID(carrierID) || !validUUID(id) {
		return dto.CarrierDriverResponse{}, invalid("invalid carrier driver identity")
	}
	v, err := s.repositories.CarrierDriver.Get(ctx, id)
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	if v.CarrierID != carrierID {
		return dto.CarrierDriverResponse{}, repository.ErrNotFound
	}
	return mapCarrierDriver(v), nil
}

func (s *Service) UpdateCarrierDriver(ctx context.Context, carrierID, id string, q dto.UpdateCarrierDriverRequest) (dto.CarrierDriverResponse, error) {
	if !validUUID(carrierID) || !validUUID(id) {
		return dto.CarrierDriverResponse{}, invalid("invalid carrier driver identity")
	}
	if q.AccountID != nil && !validUUID(*q.AccountID) {
		return dto.CarrierDriverResponse{}, invalid("invalid account_id")
	}
	code, err := transportCode(q.Code, "code")
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	name, err := clean(q.Name, 150, "name")
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	phone, err := optional(q.PhoneNumber, 50, "phone_number")
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	license, err := optional(q.LicenseNumber, 80, "license_number")
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	values := map[string]interface{}{"account_id": q.AccountID, "code": code, "name": name, "phone_number": phone, "license_number": license}
	if q.IsActive != nil {
		values["is_active"] = *q.IsActive
	}
	err = s.repositories.CarrierDriver.Update(ctx, id, carrierID, values)
	if err != nil {
		return dto.CarrierDriverResponse{}, err
	}
	return s.GetCarrierDriver(ctx, carrierID, id)
}

func (s *Service) AssignShipmentDriver(ctx context.Context, shipmentID string, q dto.AssignShipmentDriverRequest, actor string) ([]dto.ShipmentDriverResponse, error) {
	if !validID(shipmentID, 120) || !validUUID(q.DriverID) || !validUUID(actor) {
		return nil, invalid("invalid shipment driver assignment")
	}
	err := s.transaction(ctx, func(local *Service) error {
		shipment, err := local.repositories.Shipment.Get(ctx, shipmentID)
		if err != nil {
			return err
		}
		if shipment.StatusCode != "PLANNED" {
			return state("drivers can only be assigned to a PLANNED shipment")
		}
		if shipment.CarrierServiceID == nil {
			return state("shipment carrier_service_id is required before assigning a driver")
		}
		service, err := local.repositories.CarrierService.Get(ctx, *shipment.CarrierServiceID)
		if err != nil {
			return err
		}
		driver, err := local.repositories.CarrierDriver.Get(ctx, q.DriverID)
		if err != nil {
			return err
		}
		if !service.IsActive || !driver.IsActive || service.CarrierID != driver.CarrierID {
			return state("active driver must belong to the shipment carrier")
		}
		return local.repositories.ShipmentDriver.Assign(ctx, &model.ShipmentDriver{ShipmentID: shipmentID, DriverID: q.DriverID, IsPrimary: q.IsPrimary, AssignedBy: actor})
	})
	if err != nil {
		return nil, err
	}
	return s.ListShipmentDrivers(ctx, shipmentID)
}

func (s *Service) ListShipmentDrivers(ctx context.Context, shipmentID string) ([]dto.ShipmentDriverResponse, error) {
	if !validID(shipmentID, 120) {
		return nil, invalid("invalid shipment_id")
	}
	if _, err := s.repositories.Shipment.Get(ctx, shipmentID); err != nil {
		return nil, err
	}
	rows, err := s.repositories.ShipmentDriver.List(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.ShipmentDriverResponse, 0, len(rows))
	for _, v := range rows {
		out = append(out, mapShipmentDriver(v))
	}
	return out, nil
}

func (s *Service) RemoveShipmentDriver(ctx context.Context, shipmentID, driverID string) ([]dto.ShipmentDriverResponse, error) {
	if !validID(shipmentID, 120) || !validUUID(driverID) {
		return nil, invalid("invalid shipment driver identity")
	}
	shipment, err := s.repositories.Shipment.Get(ctx, shipmentID)
	if err != nil {
		return nil, err
	}
	if shipment.StatusCode != "PLANNED" {
		return nil, state("drivers can only be removed from a PLANNED shipment")
	}
	if err := s.repositories.ShipmentDriver.Remove(ctx, shipmentID, driverID); err != nil {
		return nil, err
	}
	return s.ListShipmentDrivers(ctx, shipmentID)
}
