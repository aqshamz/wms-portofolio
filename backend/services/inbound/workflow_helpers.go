package inbound

import (
	"context"
	"strings"

	mastermodel "wms-api/models/master"
	repository "wms-api/repository/inbound"
	masterrepository "wms-api/repository/master"
)

func (s *Service) inspectionResult(ctx context.Context, code string) (mastermodel.InspectionResult, error) {
	active := true
	rows, _, err := s.repositories.Master.Catalog.InspectionResult.List(ctx, masterrepository.CatalogFilter{Search: code, Active: &active, Page: 1, PageSize: 100})
	if err != nil {
		return mastermodel.InspectionResult{}, repository.Error(err)
	}
	for _, row := range rows {
		if row.Code == code {
			return row, nil
		}
	}
	return mastermodel.InspectionResult{}, state("inspection result " + code + " is not configured")
}

func (s *Service) taskType(ctx context.Context, code string) (mastermodel.TaskType, error) {
	value, err := s.repositories.Master.TaskType.ByCode(ctx, code)
	if err != nil || !value.IsActive {
		return value, state("task type " + code + " is not configured")
	}
	return value, nil
}

func (s *Service) taskStatus(ctx context.Context, code string) (mastermodel.TaskStatus, error) {
	value, err := s.repositories.Master.TaskStatus.ByCode(ctx, code)
	if err != nil || !value.IsActive {
		return value, state("task status " + code + " is not configured")
	}
	return value, nil
}

func (s *Service) initialTaskStatus(ctx context.Context) (mastermodel.TaskStatus, error) {
	active := true
	rows, _, err := s.repositories.Master.TaskStatus.List(ctx, masterrepository.OperationalFilter{Active: &active})
	if err != nil {
		return mastermodel.TaskStatus{}, repository.Error(err)
	}
	var result *mastermodel.TaskStatus
	for index := range rows {
		if rows[index].IsInitial {
			if result != nil {
				return mastermodel.TaskStatus{}, state("multiple initial task statuses are configured")
			}
			result = &rows[index]
		}
	}
	if result == nil {
		return mastermodel.TaskStatus{}, state("initial task status is not configured")
	}
	return *result, nil
}

func (s *Service) taskTransitionTarget(ctx context.Context, fromID, code string) (mastermodel.TaskStatus, error) {
	target, err := s.taskStatus(ctx, code)
	if err != nil {
		return target, err
	}
	active := true
	rows, _, err := s.repositories.Master.TaskStatusTransition.List(ctx, masterrepository.OperationalFilter{Active: &active})
	if err != nil {
		return target, repository.Error(err)
	}
	for _, row := range rows {
		if row.FromStatusID == fromID && row.ToStatusID == target.ID {
			return target, nil
		}
	}
	return target, state("task workflow transition to " + code + " is not configured")
}

func (s *Service) normalPriority(ctx context.Context) (mastermodel.TaskPriority, error) {
	value, err := s.repositories.Master.TaskPriority.ByCode(ctx, "NORMAL")
	if err != nil || !value.IsActive {
		return value, state("NORMAL task priority is not configured")
	}
	return value, nil
}

func scopeMatch(value *string, expected string) bool { return value == nil || *value == expected }

func (s *Service) validatePutawayTarget(ctx context.Context, ownerID, warehouseID string, item mastermodel.Item, targetID string) (mastermodel.WarehouseLocation, error) {
	targetID = strings.ToLower(strings.TrimSpace(targetID))
	if !inboundUUID(targetID) {
		return mastermodel.WarehouseLocation{}, invalid("target_location_id must be a UUID")
	}
	target, err := s.repositories.Inventory.Location.GetShared(ctx, targetID)
	if err != nil || target.WarehouseID != warehouseID || !target.IsActive || target.IsLocked {
		return target, invalid("putaway target location is unavailable")
	}
	locationType, err := s.repositories.Master.LocationType.GetShared(ctx, target.LocationTypeID)
	if err != nil || !locationType.IsActive || !locationType.AllowsStorage {
		return target, invalid("putaway target must use an active storage location type")
	}
	active := true
	strategies, _, err := s.repositories.Master.PutawayStrategy.List(ctx, masterrepository.OperationalFilter{Active: &active})
	if err != nil {
		return target, repository.Error(err)
	}
	bestScore := -1
	matched := false
	for _, strategy := range strategies {
		if !scopeMatch(strategy.OwnerID, ownerID) || !scopeMatch(strategy.WarehouseID, warehouseID) {
			continue
		}
		score := 0
		if strategy.OwnerID != nil {
			score += 2
		}
		if strategy.WarehouseID != nil {
			score++
		}
		if score < bestScore {
			continue
		}
		if score > bestScore {
			bestScore, matched = score, false
		}
		rules, _, ruleErr := s.repositories.Master.PutawayStrategyRule.List(ctx, masterrepository.OperationalFilter{ParentID: strategy.ID, Active: &active})
		if ruleErr != nil {
			return target, repository.Error(ruleErr)
		}
		for _, rule := range rules {
			if rule.CategoryID != nil && (item.CategoryID == nil || *rule.CategoryID != *item.CategoryID) {
				continue
			}
			if rule.LocationTypeID != nil && *rule.LocationTypeID != target.LocationTypeID {
				continue
			}
			if rule.ZoneID != nil && *rule.ZoneID != target.ZoneID {
				continue
			}
			matched = true
			break
		}
	}
	if bestScore < 0 {
		return target, state("no active putaway strategy is configured for the owner and warehouse")
	}
	if !matched {
		return target, invalid("target location does not match the active putaway strategy")
	}
	return target, nil
}
