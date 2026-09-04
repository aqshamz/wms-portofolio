package master

import (
	"strconv"
	"time"
	dto "wms-api/dto/master"
	model "wms-api/models/master"
)

func operationalDate(value time.Time) string { return value.Format("2006-01-02") }
func operationalOptionalDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	date := operationalDate(*value)
	return &date
}
func mapAppModule(value model.AppModule) dto.AppModuleResponse {
	return dto.AppModuleResponse{
		ID:           value.ID,
		Code:         value.Code,
		Name:         value.Name,
		DisplayOrder: value.DisplayOrder,
		IsActive:     value.IsActive,
	}
}

func mapAppPermission(value model.AppPermission) dto.AppPermissionResponse {
	return dto.AppPermissionResponse{
		ID:          value.ID,
		Code:        value.Code,
		Name:        value.Name,
		ModuleCode:  value.ModuleCode,
		Description: value.Description,
		IsActive:    value.IsActive,
		CreatedAt:   value.CreatedAt,
	}
}

func mapDocumentType(value model.DocumentType) dto.DocumentTypeResponse {
	return dto.DocumentTypeResponse{
		ID:          value.ID,
		Code:        value.Code,
		Name:        value.Name,
		ModuleCode:  value.ModuleCode,
		Description: value.Description,
		IsActive:    value.IsActive,
	}
}

func mapDocumentStatus(value model.DocumentStatus) dto.DocumentStatusResponse {
	return dto.DocumentStatusResponse{
		ID:             value.ID,
		DocumentTypeID: value.DocumentTypeID,
		Code:           value.Code,
		Name:           value.Name,
		Description:    value.Description,
		IsInitial:      value.IsInitial,
		IsFinal:        value.IsFinal,
		IsCancelled:    value.IsCancelled,
		DisplayOrder:   value.DisplayOrder,
		IsActive:       value.IsActive,
	}
}

func mapDocumentStatusTransition(value model.DocumentStatusTransition) dto.DocumentStatusTransitionResponse {
	return dto.DocumentStatusTransitionResponse{
		ID:                   value.ID,
		DocumentTypeID:       value.DocumentTypeID,
		FromStatusID:         value.FromStatusID,
		ToStatusID:           value.ToStatusID,
		RequiredPermissionID: value.RequiredPermissionID,
		IsActive:             value.IsActive,
	}
}

func mapDocumentNumberRule(value model.DocumentNumberRule) dto.DocumentNumberRuleResponse {
	return dto.DocumentNumberRuleResponse{
		ID:                   value.ID,
		DocumentTypeID:       value.DocumentTypeID,
		Prefix:               value.Prefix,
		Separator:            value.Separator,
		SequenceLength:       value.SequenceLength,
		IncludePartnerCode:   value.IncludePartnerCode,
		IncludeWarehouseCode: value.IncludeWarehouseCode,
		IsActive:             value.IsActive,
		EffectiveFrom:        operationalDate(value.EffectiveFrom),
		EffectiveUntil:       operationalOptionalDate(value.EffectiveUntil),
		CreatedAt:            value.CreatedAt,
		CreatedBy:            value.CreatedBy,
	}
}

func mapDocumentDailyCounter(value model.DocumentDailyCounter) dto.DocumentDailyCounterResponse {
	return dto.DocumentDailyCounterResponse{
		DocumentTypeID: value.DocumentTypeID,
		BusinessDate:   operationalDate(value.BusinessDate),
		LastNumber:     strconv.FormatInt(value.LastNumber, 10),
		UpdatedAt:      value.UpdatedAt,
	}
}

func mapTaskType(value model.TaskType) dto.TaskTypeResponse {
	return dto.TaskTypeResponse{
		ID:          value.ID,
		Code:        value.Code,
		Name:        value.Name,
		Description: value.Description,
		IsActive:    value.IsActive,
	}
}

func mapTaskStatus(value model.TaskStatus) dto.TaskStatusResponse {
	return dto.TaskStatusResponse{
		ID:          value.ID,
		Code:        value.Code,
		Name:        value.Name,
		IsInitial:   value.IsInitial,
		IsFinal:     value.IsFinal,
		IsCancelled: value.IsCancelled,
		IsActive:    value.IsActive,
	}
}

func mapTaskStatusTransition(value model.TaskStatusTransition) dto.TaskStatusTransitionResponse {
	return dto.TaskStatusTransitionResponse{
		ID:                   value.ID,
		FromStatusID:         value.FromStatusID,
		ToStatusID:           value.ToStatusID,
		RequiredPermissionID: value.RequiredPermissionID,
		IsActive:             value.IsActive,
	}
}

func mapTaskPriority(value model.TaskPriority) dto.TaskPriorityResponse {
	return dto.TaskPriorityResponse{
		ID:            value.ID,
		Code:          value.Code,
		Name:          value.Name,
		PriorityValue: value.PriorityValue,
		IsActive:      value.IsActive,
	}
}

func mapPickingSortMethod(value model.PickingSortMethod) dto.PickingSortMethodResponse {
	return dto.PickingSortMethodResponse{
		ID:          value.ID,
		Code:        value.Code,
		Name:        value.Name,
		Description: value.Description,
		IsActive:    value.IsActive,
	}
}

func mapPickingStrategy(value model.PickingStrategy) dto.PickingStrategyResponse {
	return dto.PickingStrategyResponse{
		ID:          value.ID,
		OwnerID:     value.OwnerID,
		WarehouseID: value.WarehouseID,
		Code:        value.Code,
		Name:        value.Name,
		Description: value.Description,
		IsActive:    value.IsActive,
	}
}

func mapPickingStrategyRule(value model.PickingStrategyRule) dto.PickingStrategyRuleResponse {
	return dto.PickingStrategyRuleResponse{
		ID:                  value.ID,
		PickingStrategyID:   value.PickingStrategyID,
		SequenceNo:          value.SequenceNo,
		InventoryStatusID:   value.InventoryStatusID,
		ZoneID:              value.ZoneID,
		PickingSortMethodID: value.PickingSortMethodID,
		IsActive:            value.IsActive,
	}
}

func mapPutawayStrategy(value model.PutawayStrategy) dto.PutawayStrategyResponse {
	return dto.PutawayStrategyResponse{
		ID:          value.ID,
		OwnerID:     value.OwnerID,
		WarehouseID: value.WarehouseID,
		Code:        value.Code,
		Name:        value.Name,
		Description: value.Description,
		IsActive:    value.IsActive,
	}
}

func mapPutawayStrategyRule(value model.PutawayStrategyRule) dto.PutawayStrategyRuleResponse {
	return dto.PutawayStrategyRuleResponse{
		ID:                  value.ID,
		PutawayStrategyID:   value.PutawayStrategyID,
		SequenceNo:          value.SequenceNo,
		CategoryID:          value.CategoryID,
		LocationTypeID:      value.LocationTypeID,
		ZoneID:              value.ZoneID,
		MinimumEmptyPercent: value.MinimumEmptyPercent,
		IsActive:            value.IsActive,
	}
}
