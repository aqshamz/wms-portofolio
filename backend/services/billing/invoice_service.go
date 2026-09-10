package billing

import (
	"context"
	"math/big"
	"strings"
	"time"

	dto "wms-api/dto/billing"
	model "wms-api/models/billing"
	repository "wms-api/repository/billing"
)

func (s *Service) CreateInvoice(ctx context.Context, runID string, q dto.CreateInvoiceRequest, actor string) (out dto.InvoiceResponse, err error) {
	issue, e := date(q.IssueDate, "issue_date")
	if e != nil {
		return out, e
	}
	err = s.transaction(ctx, func(local *Service) error {
		run, e := local.repositories.Run.Lock(ctx, runID)
		if e != nil {
			return e
		}
		if run.VersionNo != q.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		reviewed, e := local.status(ctx, run.DocumentTypeID, "REVIEWED")
		if e != nil || run.StatusID != reviewed.ID {
			return state("billing run must be REVIEWED")
		}
		contract, e := local.repositories.Contract.Get(ctx, run.BillingContractID)
		if e != nil {
			return e
		}
		due := issue.AddDate(0, 0, contract.PaymentTermDays)
		if q.DueDate != nil {
			due, e = date(*q.DueDate, "due_date")
			if e != nil {
				return e
			}
			if due.Before(issue) {
				return invalid("due_date cannot precede issue_date")
			}
		}
		kind, e := local.documentType(ctx, "INVOICE")
		if e != nil {
			return e
		}
		status, e := local.initialStatus(ctx, kind.ID)
		if e != nil {
			return e
		}
		id, e := local.number(ctx, "INVOICE", issue)
		if e != nil {
			return e
		}
		invoice := model.Invoice{ID: id, BillingRunID: run.ID, DocumentTypeID: kind.ID, StatusID: status.ID, OwnerID: run.OwnerID, WarehouseID: run.WarehouseID, IssueDate: issue, DueDate: due, CurrencyCode: run.CurrencyCode, SubtotalAmount: run.SubtotalAmount, TaxAmount: run.TaxAmount, CreditAmount: "0.000000", TotalAmount: run.TotalAmount, PaidAmount: "0.000000", Notes: q.Notes, VersionNo: 1, CreatedBy: actor}
		if e = local.repositories.Invoice.Create(ctx, &invoice); e != nil {
			return e
		}
		charges, e := local.repositories.Charge.List(ctx, run.ID)
		if e != nil {
			return e
		}
		if len(charges) == 0 {
			return state("reviewed billing run has no charges")
		}
		for i, c := range charges {
			line := model.InvoiceLine{InvoiceID: id, BillingChargeID: c.ID, LineNo: i + 1, ServiceCode: c.ServiceCode, Description: c.Description, Quantity: c.Quantity, UnitRate: c.UnitRate, SubtotalAmount: c.SubtotalAmount, TaxAmount: c.TaxAmount, TotalAmount: c.TotalAmount}
			if e = local.repositories.InvoiceLine.Create(ctx, &line); e != nil {
				return e
			}
		}
		target, e := local.transition(ctx, run.DocumentTypeID, run.StatusID, "INVOICED")
		if e != nil {
			return e
		}
		if e = local.repositories.Run.SetStatus(ctx, run.ID, target.ID, actor, run.VersionNo); e != nil {
			return e
		}
		out, e = local.GetInvoice(ctx, id)
		return e
	})
	return out, err
}

func (s *Service) GetInvoice(ctx context.Context, id string) (dto.InvoiceResponse, error) {
	v, e := s.repositories.Invoice.Get(ctx, id)
	if e != nil {
		return dto.InvoiceResponse{}, e
	}
	out := mapInvoice(v)
	lines, e := s.repositories.InvoiceLine.List(ctx, id)
	if e != nil {
		return out, e
	}
	out.Lines = make([]dto.InvoiceLineResponse, 0, len(lines))
	for _, x := range lines {
		out.Lines = append(out.Lines, mapInvoiceLine(x))
	}
	return out, nil
}
func (s *Service) ListInvoices(ctx context.Context, f repository.ListFilter) (dto.PageResponse[dto.InvoiceResponse], error) {
	v, n, e := s.repositories.Invoice.List(ctx, f)
	if e != nil {
		return dto.PageResponse[dto.InvoiceResponse]{}, e
	}
	x := make([]dto.InvoiceResponse, 0, len(v))
	for _, r := range v {
		x = append(x, mapInvoice(r))
	}
	return page(x, f.Page, f.PageSize, n), nil
}
func (s *Service) UpdateInvoice(ctx context.Context, id string, q dto.UpdateInvoiceRequest, actor string) (dto.InvoiceResponse, error) {
	issue, e := date(q.IssueDate, "issue_date")
	if e != nil {
		return dto.InvoiceResponse{}, e
	}
	due, e := date(q.DueDate, "due_date")
	if e != nil {
		return dto.InvoiceResponse{}, e
	}
	if due.Before(issue) {
		return dto.InvoiceResponse{}, invalid("due_date cannot precede issue_date")
	}
	if e = s.repositories.Invoice.UpdateDraft(ctx, id, actor, q.ExpectedVersion, map[string]interface{}{"issue_date": issue, "due_date": due, "notes": q.Notes}); e != nil {
		return dto.InvoiceResponse{}, e
	}
	return s.GetInvoice(ctx, id)
}
func (s *Service) TransitionInvoice(ctx context.Context, id, to string, version int64, actor string) (out dto.InvoiceResponse, err error) {
	err = s.transaction(ctx, func(local *Service) error {
		v, e := local.repositories.Invoice.Lock(ctx, id)
		if e != nil {
			return e
		}
		if v.VersionNo != version {
			return repository.ErrConcurrentWrite
		}
		target, e := local.transition(ctx, v.DocumentTypeID, v.StatusID, to)
		if e != nil {
			return e
		}
		if to == "VOID" && ((v.PaidAmount != "0" && v.PaidAmount != "0.000000") || (v.CreditAmount != "0" && v.CreditAmount != "0.000000")) {
			return state("an invoice with payments or issued credits cannot be voided")
		}
		if e = local.repositories.Invoice.SetStatus(ctx, id, target.ID, actor, version); e != nil {
			return e
		}
		out, e = local.GetInvoice(ctx, id)
		return e
	})
	return out, err
}

func (s *Service) IssueCreditNote(ctx context.Context, invoiceID string, q dto.CreditNoteRequest, actor string) (out dto.CreditNoteResponse, err error) {
	business, e := date(q.BusinessDate, "business_date")
	if e != nil {
		return out, e
	}
	creditText, credit, e := amount(q.Amount, "amount", false)
	if e != nil {
		return out, e
	}
	reason := strings.TrimSpace(q.Reason)
	err = s.transaction(ctx, func(local *Service) error {
		invoice, e := local.repositories.Invoice.Lock(ctx, invoiceID)
		if e != nil {
			return e
		}
		if invoice.VersionNo != q.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if business.Before(invoice.IssueDate) {
			return invalid("credit note business_date cannot precede invoice issue_date")
		}
		issued, e := local.status(ctx, invoice.DocumentTypeID, "ISSUED")
		if e != nil {
			return e
		}
		partial, e := local.status(ctx, invoice.DocumentTypeID, "PARTIALLY_PAID")
		if e != nil {
			return e
		}
		if invoice.StatusID != issued.ID && invoice.StatusID != partial.ID {
			return state("credit notes require an ISSUED or PARTIALLY_PAID invoice")
		}
		total, _ := new(big.Rat).SetString(invoice.TotalAmount)
		paid, _ := new(big.Rat).SetString(invoice.PaidAmount)
		outstanding := new(big.Rat).Sub(total, paid)
		if credit.Cmp(outstanding) > 0 {
			return invalid("credit amount exceeds invoice outstanding amount")
		}
		kind, e := local.documentType(ctx, "CREDIT_NOTE")
		if e != nil {
			return e
		}
		draft, e := local.initialStatus(ctx, kind.ID)
		if e != nil {
			return e
		}
		issuedCredit, e := local.transition(ctx, kind.ID, draft.ID, "ISSUED")
		if e != nil {
			return e
		}
		id, e := local.number(ctx, "CREDIT_NOTE", business)
		if e != nil {
			return e
		}
		now := time.Now()
		note := model.CreditNote{ID: id, InvoiceID: invoiceID, DocumentTypeID: kind.ID, StatusID: issuedCredit.ID, BusinessDate: business, Amount: creditText, Reason: reason, CreatedBy: actor, IssuedAt: &now, IssuedBy: &actor}
		if e = local.repositories.CreditNote.Create(ctx, &note); e != nil {
			return e
		}
		oldCredit, _ := new(big.Rat).SetString(invoice.CreditAmount)
		newCredit := new(big.Rat).Add(oldCredit, credit)
		newTotal := new(big.Rat).Sub(total, credit)
		targetStatus := invoice.StatusID
		if newTotal.Cmp(paid) == 0 {
			target, e := local.transition(ctx, invoice.DocumentTypeID, invoice.StatusID, "SETTLED")
			if e != nil {
				return e
			}
			targetStatus = target.ID
		}
		if e = local.repositories.Invoice.SetFinancialsAndStatus(ctx, invoiceID, targetStatus, actor, invoice.VersionNo, newCredit.FloatString(6), newTotal.FloatString(6), invoice.PaidAmount); e != nil {
			return e
		}
		rows, e := local.repositories.CreditNote.List(ctx, invoiceID)
		if e != nil {
			return e
		}
		for _, v := range rows {
			if v.ID == id {
				out = mapCredit(v)
				return nil
			}
		}
		return repository.ErrNotFound
	})
	return out, err
}
func (s *Service) ListCreditNotes(ctx context.Context, invoiceID string) ([]dto.CreditNoteResponse, error) {
	v, e := s.repositories.CreditNote.List(ctx, invoiceID)
	if e != nil {
		return nil, e
	}
	x := make([]dto.CreditNoteResponse, 0, len(v))
	for _, r := range v {
		x = append(x, mapCredit(r))
	}
	return x, nil
}
func (s *Service) RecordPayment(ctx context.Context, invoiceID string, q dto.PaymentRequest, actor string) (out dto.PaymentResponse, err error) {
	business, e := date(q.BusinessDate, "business_date")
	if e != nil {
		return out, e
	}
	paidText, paid, e := amount(q.Amount, "amount", false)
	if e != nil {
		return out, e
	}
	reference := strings.TrimSpace(q.Reference)
	err = s.transaction(ctx, func(local *Service) error {
		invoice, e := local.repositories.Invoice.Lock(ctx, invoiceID)
		if e != nil {
			return e
		}
		if invoice.VersionNo != q.ExpectedVersion {
			return repository.ErrConcurrentWrite
		}
		if business.Before(invoice.IssueDate) {
			return invalid("payment business_date cannot precede invoice issue_date")
		}
		issued, e := local.status(ctx, invoice.DocumentTypeID, "ISSUED")
		if e != nil {
			return e
		}
		partial, e := local.status(ctx, invoice.DocumentTypeID, "PARTIALLY_PAID")
		if e != nil {
			return e
		}
		if invoice.StatusID != issued.ID && invoice.StatusID != partial.ID {
			return state("payments require an ISSUED or PARTIALLY_PAID invoice")
		}
		total, _ := new(big.Rat).SetString(invoice.TotalAmount)
		oldPaid, _ := new(big.Rat).SetString(invoice.PaidAmount)
		outstanding := new(big.Rat).Sub(total, oldPaid)
		if paid.Cmp(outstanding) > 0 {
			return invalid("payment amount exceeds invoice outstanding amount")
		}
		newPaid := new(big.Rat).Add(oldPaid, paid)
		targetID := invoice.StatusID
		if newPaid.Cmp(total) == 0 {
			target, e := local.transition(ctx, invoice.DocumentTypeID, invoice.StatusID, "PAID")
			if e != nil {
				return e
			}
			targetID = target.ID
		} else if invoice.StatusID == issued.ID {
			target, e := local.transition(ctx, invoice.DocumentTypeID, invoice.StatusID, "PARTIALLY_PAID")
			if e != nil {
				return e
			}
			targetID = target.ID
		}
		kind, e := local.documentType(ctx, "PAYMENT")
		if e != nil {
			return e
		}
		initial, e := local.initialStatus(ctx, kind.ID)
		if e != nil {
			return e
		}
		allocated, e := local.transition(ctx, kind.ID, initial.ID, "ALLOCATED")
		if e != nil {
			return e
		}
		id, e := local.number(ctx, "PAYMENT", business)
		if e != nil {
			return e
		}
		payment := model.Payment{ID: id, InvoiceID: invoiceID, DocumentTypeID: kind.ID, StatusID: allocated.ID, BusinessDate: business, Amount: paidText, Reference: reference, Notes: q.Notes, CreatedBy: actor}
		if e = local.repositories.Payment.Create(ctx, &payment); e != nil {
			return e
		}
		if e = local.repositories.Invoice.SetFinancialsAndStatus(ctx, invoiceID, targetID, actor, invoice.VersionNo, invoice.CreditAmount, invoice.TotalAmount, newPaid.FloatString(6)); e != nil {
			return e
		}
		rows, e := local.repositories.Payment.List(ctx, invoiceID)
		if e != nil {
			return e
		}
		for _, v := range rows {
			if v.ID == id {
				out = mapPayment(v)
				return nil
			}
		}
		return repository.ErrNotFound
	})
	return out, err
}
func (s *Service) ListPayments(ctx context.Context, invoiceID string) ([]dto.PaymentResponse, error) {
	v, e := s.repositories.Payment.List(ctx, invoiceID)
	if e != nil {
		return nil, e
	}
	x := make([]dto.PaymentResponse, 0, len(v))
	for _, r := range v {
		x = append(x, mapPayment(r))
	}
	return x, nil
}
