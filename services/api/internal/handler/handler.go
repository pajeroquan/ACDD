package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/local-life/partner/services/api/internal/middleware"
	"github.com/local-life/partner/services/api/internal/pkg/response"
	"github.com/local-life/partner/services/api/internal/service"
)

type Handler struct {
	Svc *service.Services
}

func New(svc *service.Services) *Handler { return &Handler{Svc: svc} }

func (h *Handler) Health(c *gin.Context) { response.OK(c, gin.H{"status": "up"}) }

func (h *Handler) AdminLogin(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	user, token, err := h.Svc.AdminLogin(req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token, "user": user})
}

func (h *Handler) ListUnions(c *gin.Context) {
	list, err := h.Svc.ListUnions()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *Handler) CreateUnion(c *gin.Context) {
	var in service.UnionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	u, err := h.Svc.CreateUnion(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	aid := middleware.AdminID(c)
	h.Svc.Audit(&aid, "union.create", "union", strconv.FormatUint(u.ID, 10), nil)
	response.OK(c, u)
}

func (h *Handler) UpdateUnion(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.UnionInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	u, err := h.Svc.UpdateUnion(id, in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	aid := middleware.AdminID(c)
	h.Svc.Audit(&aid, "union.update", "union", strconv.FormatUint(id, 10), map[string]any{"commission_rate": in.CommissionRate})
	response.OK(c, u)
}

func (h *Handler) ListPartners(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.Svc.ListPartners(c.Query("city"), c.Query("status"), page, size)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

func (h *Handler) GetPartner(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	p, err := h.Svc.GetPartner(id)
	if err != nil {
		response.NotFound(c, "partner not found")
		return
	}
	response.OK(c, p)
}

func (h *Handler) CreatePartner(c *gin.Context) {
	var in service.PartnerInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.Svc.CreatePartner(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	aid := middleware.AdminID(c)
	h.Svc.Audit(&aid, "partner.create", "partner", strconv.FormatUint(p.ID, 10), nil)
	response.OK(c, p)
}

func (h *Handler) UpdatePartner(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.PartnerInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	p, err := h.Svc.UpdatePartner(id, in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	aid := middleware.AdminID(c)
	h.Svc.Audit(&aid, "partner.update", "partner", strconv.FormatUint(id, 10), nil)
	response.OK(c, p)
}

func (h *Handler) MatchChat(c *gin.Context) {
	var req struct {
		MatchRequestID *uint64 `json:"match_request_id"`
		Text           string  `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	res, err := h.Svc.MatchChat(c.Request.Context(), middleware.AdminID(c), req.MatchRequestID, req.Text)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *Handler) MatchConfirm(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var in service.ConfirmMatchInput
	_ = c.ShouldBindJSON(&in)
	res, err := h.Svc.ConfirmMatch(middleware.AdminID(c), id, in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, res)
}

func (h *Handler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	list, total, err := h.Svc.ListOrders(c.Query("status"), page, size)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"list": list, "total": total})
}

func (h *Handler) GetOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	o, err := h.Svc.GetOrder(id)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}
	response.OK(c, o)
}

func (h *Handler) RefundOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	o, err := h.Svc.RefundOrder(id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	aid := middleware.AdminID(c)
	h.Svc.Audit(&aid, "order.refund", "order", strconv.FormatUint(id, 10), nil)
	response.OK(c, o)
}

func (h *Handler) CommissionReport(c *gin.Context) {
	list, err := h.Svc.CommissionReport()
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *Handler) ListNotifications(c *gin.Context) {
	var pid uint64
	if v := c.Query("partner_id"); v != "" {
		pid, _ = strconv.ParseUint(v, 10, 64)
	}
	list, err := h.Svc.ListNotifications(pid)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, list)
}

func (h *Handler) WxLogin(c *gin.Context) {
	var req struct {
		Code     string `json:"code"`
		Nickname string `json:"nickname"`
	}
	_ = c.ShouldBindJSON(&req)
	user, token, err := h.Svc.WxLogin(req.Code, req.Nickname)
	if err != nil {
		response.ServerError(c, err.Error())
		return
	}
	response.OK(c, gin.H{"token": token, "user": user})
}

func (h *Handler) BrowseSession(c *gin.Context) {
	cards, session, err := h.Svc.GetBrowseSession(c.Param("sid"), middleware.UserID(c))
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"session": session, "cards": cards})
}

func (h *Handler) Quote(c *gin.Context) {
	var in service.QuoteInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	q, err := h.Svc.Quote(in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, q)
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var in service.CreateOrderInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	o, err := h.Svc.CreateOrder(middleware.UserID(c), in)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, o)
}

func (h *Handler) PayOrder(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	params, err := h.Svc.CreatePayment(middleware.UserID(c), id)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, params)
}

func (h *Handler) MockPayNotify(c *gin.Context) {
	var req struct {
		OutTradeNo string `json:"out_trade_no" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if err := h.Svc.MockPayNotify(req.OutTradeNo); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.OK(c, gin.H{"ok": true})
}

func (h *Handler) PayNotify(c *gin.Context) {
	var req struct {
		OutTradeNo    string `json:"out_trade_no"`
		TransactionID string `json:"transaction_id"`
	}
	_ = c.ShouldBindJSON(&req)
	if req.OutTradeNo == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": "missing out_trade_no"})
		return
	}
	if req.TransactionID == "" {
		req.TransactionID = "wx_" + req.OutTradeNo
	}
	if err := h.Svc.HandlePaySuccess(req.OutTradeNo, req.TransactionID, map[string]any{"source": "notify"}); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "FAIL", "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": "SUCCESS", "message": "成功"})
}
