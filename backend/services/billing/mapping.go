package billing

import (
	"math/big"
	dto "wms-api/dto/billing"
	model "wms-api/models/billing"
	repository "wms-api/repository/billing"
)

func dateText(v interface{ Format(string) string }) string { return v.Format("2006-01-02") }
func optDate(v interface{ Format(string) string }) *string { x := dateText(v); return &x }
func mapContract(v repository.ContractRow) dto.ContractResponse {
	var until *string
	if v.EffectiveUntil != nil {
		until = optDate(*v.EffectiveUntil)
	}
	return dto.ContractResponse{ID: v.ID, OwnerID: v.OwnerID, OwnerCode: v.OwnerCode, WarehouseID: v.WarehouseID, WarehouseCode: v.WarehouseCode, CurrencyCode: v.CurrencyCode, BillingCycle: v.BillingCycle, StatusCode: v.StatusCode, PaymentTermDays: v.PaymentTermDays, EffectiveFrom: dateText(v.EffectiveFrom), EffectiveUntil: until, Notes: v.Notes, VersionNo: v.VersionNo, CreatedAt: v.CreatedAt}
}
func mapRateLine(v repository.RateCardLineRow) dto.RateCardLineResponse {
	return dto.RateCardLineResponse{ID: v.ID, ServiceCode: v.ServiceCode, Description: v.Description, SourceKind: v.SourceKind, MovementTypeID: v.MovementTypeID, MovementTypeCode: v.MovementTypeCode, BillingBasis: v.BillingBasis, UnitRate: v.UnitRate, MinimumCharge: v.MinimumCharge, TaxPercent: v.TaxPercent, IsActive: v.IsActive}
}
func mapRateCard(v repository.RateCardRow) dto.RateCardResponse {
	var until *string
	if v.EffectiveUntil != nil {
		until = optDate(*v.EffectiveUntil)
	}
	return dto.RateCardResponse{ID: v.ID, BillingContractID: v.BillingContractID, OwnerID: v.OwnerID, OwnerCode: v.OwnerCode, WarehouseID: v.WarehouseID, WarehouseCode: v.WarehouseCode, Name: v.Name, CurrencyCode: v.CurrencyCode, StatusCode: v.StatusCode, EffectiveFrom: dateText(v.EffectiveFrom), EffectiveUntil: until, VersionNo: v.VersionNo, CreatedAt: v.CreatedAt}
}
func mapEvent(v repository.EventRow) dto.EventResponse {
	return dto.EventResponse{ID: v.ID, OwnerID: v.OwnerID, WarehouseID: v.WarehouseID, RateCardLineID: v.RateCardLineID, ServiceCode: v.ServiceCode, MovementTypeCode: v.MovementTypeCode, InventoryMovementID: v.InventoryMovementID, EventKey: v.EventKey, BusinessDate: dateText(v.BusinessDate), SourceDocumentID: v.SourceDocumentID, SourceLineID: v.SourceLineID, Quantity: v.Quantity, UOMID: v.UOMID, UOMCode: v.UOMCode, StatusCode: v.StatusCode, Notes: v.Notes, CreatedAt: v.CreatedAt}
}
func mapCharge(v model.BillingCharge) dto.ChargeResponse {
	return dto.ChargeResponse{ID: v.ID, BillableEventID: v.BillableEventID, ServiceCode: v.ServiceCode, Description: v.Description, SourceDocumentID: v.SourceDocumentID, Quantity: v.Quantity, UnitRate: v.UnitRate, SubtotalAmount: v.SubtotalAmount, TaxPercent: v.TaxPercent, TaxAmount: v.TaxAmount, TotalAmount: v.TotalAmount}
}
func mapRun(v repository.BillingRunRow) dto.BillingRunResponse {
	return dto.BillingRunResponse{ID: v.ID, BillingContractID: v.BillingContractID, OwnerID: v.OwnerID, OwnerCode: v.OwnerCode, WarehouseID: v.WarehouseID, WarehouseCode: v.WarehouseCode, PeriodFrom: dateText(v.PeriodFrom), PeriodUntil: dateText(v.PeriodUntil), BusinessDate: dateText(v.BusinessDate), CurrencyCode: v.CurrencyCode, StatusCode: v.StatusCode, SubtotalAmount: v.SubtotalAmount, TaxAmount: v.TaxAmount, TotalAmount: v.TotalAmount, Notes: v.Notes, VersionNo: v.VersionNo, CreatedAt: v.CreatedAt}
}
func mapInvoiceLine(v model.InvoiceLine) dto.InvoiceLineResponse {
	return dto.InvoiceLineResponse{ID: v.ID, LineNo: v.LineNo, BillingChargeID: v.BillingChargeID, ServiceCode: v.ServiceCode, Description: v.Description, Quantity: v.Quantity, UnitRate: v.UnitRate, SubtotalAmount: v.SubtotalAmount, TaxAmount: v.TaxAmount, TotalAmount: v.TotalAmount}
}
func mapInvoice(v repository.InvoiceRow) dto.InvoiceResponse {
	total, _ := new(big.Rat).SetString(v.TotalAmount)
	paid, _ := new(big.Rat).SetString(v.PaidAmount)
	out := new(big.Rat).Sub(total, paid).FloatString(6)
	return dto.InvoiceResponse{ID: v.ID, BillingRunID: v.BillingRunID, OwnerID: v.OwnerID, OwnerCode: v.OwnerCode, WarehouseID: v.WarehouseID, WarehouseCode: v.WarehouseCode, IssueDate: dateText(v.IssueDate), DueDate: dateText(v.DueDate), CurrencyCode: v.CurrencyCode, StatusCode: v.StatusCode, SubtotalAmount: v.SubtotalAmount, TaxAmount: v.TaxAmount, CreditAmount: v.CreditAmount, TotalAmount: v.TotalAmount, PaidAmount: v.PaidAmount, OutstandingAmount: out, Notes: v.Notes, VersionNo: v.VersionNo, CreatedAt: v.CreatedAt}
}
func mapCredit(v repository.CreditNoteRow) dto.CreditNoteResponse {
	return dto.CreditNoteResponse{ID: v.ID, InvoiceID: v.InvoiceID, BusinessDate: dateText(v.BusinessDate), Amount: v.Amount, Reason: v.Reason, StatusCode: v.StatusCode, CreatedAt: v.CreatedAt, IssuedAt: v.IssuedAt}
}
func mapPayment(v repository.PaymentRow) dto.PaymentResponse {
	return dto.PaymentResponse{ID: v.ID, InvoiceID: v.InvoiceID, BusinessDate: dateText(v.BusinessDate), Amount: v.Amount, Reference: v.Reference, Notes: v.Notes, StatusCode: v.StatusCode, CreatedAt: v.CreatedAt}
}
