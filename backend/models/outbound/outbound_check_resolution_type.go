package outbound

type OutboundCheckResolutionType struct {
	ID                      string  `gorm:"column:outbound_check_resolution_type_id;type:uuid;default:gen_random_uuid();primaryKey"`
	Code                    string  `gorm:"column:code;size:40;not null;uniqueIndex"`
	Name                    string  `gorm:"column:name;size:120;not null"`
	Description             *string `gorm:"column:description;type:text"`
	CountsAsStockCorrection bool    `gorm:"column:counts_as_stock_correction;not null;default:false"`
	CountsAsReplacement     bool    `gorm:"column:counts_as_replacement;not null;default:false"`
	CountsAsShortAcceptance bool    `gorm:"column:counts_as_short_acceptance;not null;default:false"`
	RequiresApproval        bool    `gorm:"column:requires_approval;not null;default:false"`
	IsActive                bool    `gorm:"column:is_active;not null;default:true"`
}

func (OutboundCheckResolutionType) TableName() string { return "outbound_check_resolution_type" }
