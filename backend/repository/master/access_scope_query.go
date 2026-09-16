package master

import (
	"context"

	"wms-api/requestscope"

	"gorm.io/gorm"
)

func scopedPrincipal(ctx context.Context) (requestscope.Principal, bool) {
	principal, exists := requestscope.FromContext(ctx)
	return principal, exists && !principal.Unrestricted && principal.AccountID != ""
}

func applyOwnerAccess(ctx context.Context, query *gorm.DB, column string) *gorm.DB {
	principal, scoped := scopedPrincipal(ctx)
	if !scoped {
		return query
	}
	return query.Where(
		column+" IN (?)",
		query.Session(&gorm.Session{NewDB: true}).Table("account_owner_access").
			Select("owner_id").Where("account_id = ?", principal.AccountID),
	)
}

func applyWarehouseAccess(ctx context.Context, query *gorm.DB, column string) *gorm.DB {
	principal, scoped := scopedPrincipal(ctx)
	if !scoped {
		return query
	}
	return query.Where(
		column+" IN (?)",
		query.Session(&gorm.Session{NewDB: true}).Table("account_warehouse_access").
			Select("warehouse_id").Where("account_id = ?", principal.AccountID),
	)
}

func applyOptionalOwnerWarehouseAccess(
	ctx context.Context,
	query *gorm.DB,
	ownerColumn, warehouseColumn string,
) *gorm.DB {
	principal, scoped := scopedPrincipal(ctx)
	if !scoped {
		return query
	}
	ownerAccess := query.Session(&gorm.Session{NewDB: true}).Table("account_owner_access").
		Select("owner_id").Where("account_id = ?", principal.AccountID)
	warehouseAccess := query.Session(&gorm.Session{NewDB: true}).Table("account_warehouse_access").
		Select("warehouse_id").Where("account_id = ?", principal.AccountID)
	return query.
		Where("("+ownerColumn+" IS NULL OR "+ownerColumn+" IN (?))", ownerAccess).
		Where("("+warehouseColumn+" IS NULL OR "+warehouseColumn+" IN (?))", warehouseAccess)
}

func applyItemOwnerAccess(ctx context.Context, query *gorm.DB, itemColumn string) *gorm.DB {
	principal, scoped := scopedPrincipal(ctx)
	if !scoped {
		return query
	}
	return query.Where(
		itemColumn+" IN (?)",
		query.Session(&gorm.Session{NewDB: true}).Table("item").Select("item_id").
			Where("owner_id IN (?)",
				query.Session(&gorm.Session{NewDB: true}).Table("account_owner_access").
					Select("owner_id").Where("account_id = ?", principal.AccountID)),
	)
}
