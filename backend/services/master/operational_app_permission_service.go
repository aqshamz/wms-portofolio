package master

import (
	"context"

	dto "wms-api/dto/master"
)

func (s *OperationalService) GetAppPermission(ctx context.Context, id string) (response dto.AppPermissionResponse, err error) {
	if validateID(id) != nil {
		return response, ErrInvalidInput
	}

	value, err := s.repositories.AppPermission.Get(ctx, id)
	if err != nil {
		return response, catalogError(err)
	}

	return mapAppPermission(value), nil
}
func (s *OperationalService) ListAppPermission(ctx context.Context, request dto.OperationalListRequest) (dto.PageResponse[dto.AppPermissionResponse], error) {
	filter, err := operationalFilter(request, "")
	if err != nil {
		return dto.PageResponse[dto.AppPermissionResponse]{}, err
	}

	rows, total, err := s.repositories.AppPermission.List(ctx, filter)
	if err != nil {
		return dto.PageResponse[dto.AppPermissionResponse]{}, catalogError(err)
	}
	items := make([]dto.AppPermissionResponse, 0, len(rows))
	for _, row := range rows {
		items = append(items, mapAppPermission(row))
	}
	return pageResponse(items, request.Page, request.PageSize, total), nil
}
