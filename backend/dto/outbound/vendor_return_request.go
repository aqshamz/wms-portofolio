package outbound

type CompleteVendorReturnRequest struct {
	ExpectedVersion        int64  `json:"expected_version" binding:"required,min=1"`
	ExpectedBalanceVersion int64  `json:"expected_balance_version" binding:"required,min=1"`
	CompletedAt            string `json:"completed_at" binding:"required"`
}

type CancelVendorReturnRequest struct {
	ExpectedVersion int64  `json:"expected_version" binding:"required,min=1"`
	Reason          string `json:"reason" binding:"required,max=4000"`
}
