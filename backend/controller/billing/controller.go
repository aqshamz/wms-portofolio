package billing

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	authdto "wms-api/dto/authentication"
	dto "wms-api/dto/billing"
	"wms-api/middleware"
	repository "wms-api/repository/billing"
	service "wms-api/services/billing"
	"wms-api/utils"
)

type Controller struct{ service *service.Service }

func NewController(s *service.Service) *Controller { return &Controller{service: s} }
func actor(c *gin.Context) string {
	v, _ := c.Get(middleware.ContextUserKey)
	u, _ := v.(authdto.UserResponse)
	return u.AccountID
}
func fail(c *gin.Context, e error) {
	switch {
	case errors.Is(e, service.ErrInvalidInput), errors.Is(e, repository.ErrConstraint):
		utils.Failure(c, http.StatusBadRequest, e.Error(), nil)
	case errors.Is(e, repository.ErrNotFound):
		utils.Failure(c, http.StatusNotFound, e.Error(), nil)
	case errors.Is(e, service.ErrInvalidState), errors.Is(e, repository.ErrConcurrentWrite), errors.Is(e, repository.ErrConflict):
		utils.Failure(c, http.StatusConflict, e.Error(), nil)
	case errors.Is(e, service.ErrForbidden):
		utils.Failure(c, http.StatusForbidden, e.Error(), nil)
	default:
		_ = c.Error(e)
		utils.Failure(c, http.StatusInternalServerError, "unable to process billing request", nil)
	}
}
func listFilter(c *gin.Context, dates bool) (repository.ListFilter, bool) {
	allowed := map[string]bool{"owner_id": true, "warehouse_id": true, "status_code": true, "search": true, "page": true, "page_size": true}
	if dates {
		allowed["date_from"] = true
		allowed["date_until"] = true
	}
	for k, v := range c.Request.URL.Query() {
		if !allowed[k] || len(v) != 1 {
			utils.Failure(c, http.StatusBadRequest, "unknown or repeated query parameter: "+k, nil)
			return repository.ListFilter{}, false
		}
	}
	p, z, e := utils.ParsePagination(c.Query("page"), c.Query("page_size"))
	if e != nil {
		utils.Failure(c, http.StatusBadRequest, e.Error(), nil)
		return repository.ListFilter{}, false
	}
	return repository.ListFilter{OwnerID: strings.TrimSpace(c.Query("owner_id")), WarehouseID: strings.TrimSpace(c.Query("warehouse_id")), StatusCode: strings.ToUpper(strings.TrimSpace(c.Query("status_code"))), Search: strings.TrimSpace(c.Query("search")), DateFrom: strings.TrimSpace(c.Query("date_from")), DateUntil: strings.TrimSpace(c.Query("date_until")), Page: p, PageSize: z}, true
}
func bind[T any](c *gin.Context) (T, bool) { var q T; return q, utils.BindStrictJSON(c, &q) }

func (cn *Controller) CreateContract(c *gin.Context) {
	q, ok := bind[dto.CreateContractRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.CreateContract(c.Request.Context(), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusCreated, "billing contract created", v)
}
func (cn *Controller) ListContracts(c *gin.Context) {
	f, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, e := cn.service.ListContracts(c.Request.Context(), f)
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billing contracts retrieved", v)
}
func (cn *Controller) GetContract(c *gin.Context) {
	v, e := cn.service.GetContract(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billing contract retrieved", v)
}
func (cn *Controller) UpdateContract(c *gin.Context) {
	q, ok := bind[dto.UpdateContractRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.UpdateContract(c.Request.Context(), c.Param("id"), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billing contract updated", v)
}
func (cn *Controller) contractTransition(c *gin.Context, to, msg string) {
	q, ok := bind[dto.TransitionRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.TransitionContract(c.Request.Context(), c.Param("id"), to, q.ExpectedVersion, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, msg, v)
}
func (cn *Controller) ActivateContract(c *gin.Context) {
	cn.contractTransition(c, "ACTIVE", "billing contract activated")
}
func (cn *Controller) SuspendContract(c *gin.Context) {
	cn.contractTransition(c, "SUSPENDED", "billing contract suspended")
}
func (cn *Controller) ExpireContract(c *gin.Context) {
	cn.contractTransition(c, "EXPIRED", "billing contract expired")
}
func (cn *Controller) ResumeContract(c *gin.Context) {
	cn.contractTransition(c, "ACTIVE", "billing contract resumed")
}
func (cn *Controller) TerminateContract(c *gin.Context) {
	cn.contractTransition(c, "TERMINATED", "billing contract terminated")
}

func (cn *Controller) CreateRateCard(c *gin.Context) {
	q, ok := bind[dto.CreateRateCardRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.CreateRateCard(c.Request.Context(), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusCreated, "rate card created", v)
}
func (cn *Controller) ListRateCards(c *gin.Context) {
	f, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, e := cn.service.ListRateCards(c.Request.Context(), f)
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "rate cards retrieved", v)
}
func (cn *Controller) GetRateCard(c *gin.Context) {
	v, e := cn.service.GetRateCard(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "rate card retrieved", v)
}
func (cn *Controller) UpdateRateCard(c *gin.Context) {
	q, ok := bind[dto.UpdateRateCardRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.UpdateRateCard(c.Request.Context(), c.Param("id"), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "rate card updated", v)
}
func (cn *Controller) AddRateCardLine(c *gin.Context) {
	q, ok := bind[dto.AddRateCardLineRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.AddRateCardLine(c.Request.Context(), c.Param("id"), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusCreated, "rate card line added", v)
}
func (cn *Controller) UpdateRateCardLine(c *gin.Context) {
	q, ok := bind[dto.UpdateRateCardLineRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.UpdateRateCardLine(c.Request.Context(), c.Param("id"), c.Param("line_id"), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "rate card line updated", v)
}
func (cn *Controller) DeleteRateCardLine(c *gin.Context) {
	q, ok := bind[dto.DeleteRateCardLineRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.DeleteRateCardLine(c.Request.Context(), c.Param("id"), c.Param("line_id"), q.ExpectedVersion, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "rate card line deleted", v)
}
func (cn *Controller) rateTransition(c *gin.Context, to, msg string) {
	q, ok := bind[dto.TransitionRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.TransitionRateCard(c.Request.Context(), c.Param("id"), to, q.ExpectedVersion, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, msg, v)
}
func (cn *Controller) ApproveRateCard(c *gin.Context) {
	cn.rateTransition(c, "APPROVED", "rate card approved")
}
func (cn *Controller) ActivateRateCard(c *gin.Context) {
	cn.rateTransition(c, "ACTIVE", "rate card activated")
}
func (cn *Controller) ExpireRateCard(c *gin.Context) {
	cn.rateTransition(c, "EXPIRED", "rate card expired")
}
func (cn *Controller) CancelRateCard(c *gin.Context) {
	cn.rateTransition(c, "CANCELLED", "rate card cancelled")
}

func (cn *Controller) CollectEvents(c *gin.Context) {
	q, ok := bind[dto.CollectEventsRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.CollectEvents(c.Request.Context(), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billable events collected", v)
}
func (cn *Controller) SnapshotStorage(c *gin.Context) {
	q, ok := bind[dto.StorageSnapshotRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.SnapshotStorage(c.Request.Context(), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "storage snapshot events collected", v)
}
func (cn *Controller) CreateManualEvent(c *gin.Context) {
	q, ok := bind[dto.CreateManualEventRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.CreateManualEvent(c.Request.Context(), c.Param("id"), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusCreated, "manual billable event created", v)
}
func (cn *Controller) ListEvents(c *gin.Context) {
	f, ok := listFilter(c, true)
	if !ok {
		return
	}
	v, e := cn.service.ListEvents(c.Request.Context(), f)
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billable events retrieved", v)
}
func (cn *Controller) GetEvent(c *gin.Context) {
	v, e := cn.service.GetEvent(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billable event retrieved", v)
}
func (cn *Controller) ExcludeEvent(c *gin.Context) {
	q, ok := bind[dto.ExcludeEventRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.ExcludeEvent(c.Request.Context(), c.Param("id"), q.Reason, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billable event excluded", v)
}

func (cn *Controller) CreateRun(c *gin.Context) {
	q, ok := bind[dto.CreateBillingRunRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.CreateRun(c.Request.Context(), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusCreated, "billing run created", v)
}
func (cn *Controller) ListRuns(c *gin.Context) {
	f, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, e := cn.service.ListRuns(c.Request.Context(), f)
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billing runs retrieved", v)
}
func (cn *Controller) GetRun(c *gin.Context) {
	v, e := cn.service.GetRun(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billing run retrieved", v)
}
func (cn *Controller) CalculateRun(c *gin.Context) {
	q, ok := bind[dto.TransitionRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.CalculateRun(c.Request.Context(), c.Param("id"), q.ExpectedVersion, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "billing run calculated", v)
}
func (cn *Controller) runTransition(c *gin.Context, to, msg string) {
	q, ok := bind[dto.TransitionRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.TransitionRun(c.Request.Context(), c.Param("id"), to, q.ExpectedVersion, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, msg, v)
}
func (cn *Controller) ReviewRun(c *gin.Context) {
	cn.runTransition(c, "REVIEWED", "billing run reviewed")
}
func (cn *Controller) ReopenRun(c *gin.Context) { cn.runTransition(c, "DRAFT", "billing run reopened") }
func (cn *Controller) CancelRun(c *gin.Context) {
	cn.runTransition(c, "CANCELLED", "billing run cancelled")
}
func (cn *Controller) CreateInvoice(c *gin.Context) {
	q, ok := bind[dto.CreateInvoiceRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.CreateInvoice(c.Request.Context(), c.Param("id"), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusCreated, "invoice created", v)
}
func (cn *Controller) ListInvoices(c *gin.Context) {
	f, ok := listFilter(c, false)
	if !ok {
		return
	}
	v, e := cn.service.ListInvoices(c.Request.Context(), f)
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "invoices retrieved", v)
}
func (cn *Controller) GetInvoice(c *gin.Context) {
	v, e := cn.service.GetInvoice(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "invoice retrieved", v)
}
func (cn *Controller) UpdateInvoice(c *gin.Context) {
	q, ok := bind[dto.UpdateInvoiceRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.UpdateInvoice(c.Request.Context(), c.Param("id"), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "invoice updated", v)
}
func (cn *Controller) invoiceTransition(c *gin.Context, to, msg string) {
	q, ok := bind[dto.TransitionRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.TransitionInvoice(c.Request.Context(), c.Param("id"), to, q.ExpectedVersion, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, msg, v)
}
func (cn *Controller) ReviewInvoice(c *gin.Context) {
	cn.invoiceTransition(c, "REVIEWED", "invoice reviewed")
}
func (cn *Controller) ReopenInvoice(c *gin.Context) {
	cn.invoiceTransition(c, "DRAFT", "invoice reopened")
}
func (cn *Controller) IssueInvoice(c *gin.Context) {
	cn.invoiceTransition(c, "ISSUED", "invoice issued")
}
func (cn *Controller) VoidInvoice(c *gin.Context) { cn.invoiceTransition(c, "VOID", "invoice voided") }
func (cn *Controller) IssueCreditNote(c *gin.Context) {
	q, ok := bind[dto.CreditNoteRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.IssueCreditNote(c.Request.Context(), c.Param("id"), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusCreated, "credit note issued", v)
}
func (cn *Controller) ListCreditNotes(c *gin.Context) {
	v, e := cn.service.ListCreditNotes(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "credit notes retrieved", v)
}
func (cn *Controller) RecordPayment(c *gin.Context) {
	q, ok := bind[dto.PaymentRequest](c)
	if !ok {
		return
	}
	v, e := cn.service.RecordPayment(c.Request.Context(), c.Param("id"), q, actor(c))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusCreated, "payment recorded and allocated", v)
}
func (cn *Controller) ListPayments(c *gin.Context) {
	v, e := cn.service.ListPayments(c.Request.Context(), c.Param("id"))
	if e != nil {
		fail(c, e)
		return
	}
	utils.Success(c, http.StatusOK, "payments retrieved", v)
}
