package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"subscription_service/internal/domain"
	"subscription_service/internal/service"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents error message
type ErrorResponse struct {
	Error string `json:"error"`
}

type Handler struct {
	svc *service.SubscriptionService
}

func NewHandler(svc *service.SubscriptionService) *Handler {
	return &Handler{svc: svc}
}

type CreateSubscriptionRequest struct {
	ServiceName string  `json:"service_name" binding:"required"`
	Price       int64   `json:"price" binding:"required"`
	UserID      string  `json:"user_id" binding:"required"`
	StartDate   string  `json:"start_date" binding:"required"` // MM-YYYY
	EndDate     *string `json:"end_date"`                      // MM-YYYY | null
}

// Create godoc
// @Summary Create subscription
// @Description Create a new subscription for a user
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param request body CreateSubscriptionRequest true "Subscription payload"
// @Success 201 {object} domain.SubscriptionDTO
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [post]
func (h *Handler) Create(c *gin.Context) {
	var req CreateSubscriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	svcReq := service.CreateSubscriptionRequest{
		ServiceName: req.ServiceName,
		Price:       req.Price,
		UserID:      req.UserID,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
	}

	id, err := h.svc.Create(c.Request.Context(), svcReq)
	if err != nil {
		writeError(c, err)
		return
	}

	created, _ := h.svc.GetByID(c.Request.Context(), id)
	if created != nil {
		c.JSON(http.StatusCreated, domain.ToDTO(*created))
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetByID godoc
// @Summary Get subscription by ID
// @Description Get subscription details by its ID
// @Tags subscriptions
// @Produce json
// @Param id path int true "Subscription ID"
// @Success 200 {object} domain.SubscriptionDTO
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [get]
func (h *Handler) GetByID(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	sub, err := h.svc.GetByID(c.Request.Context(), id)
	if sub == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.ToDTO(*sub))
}

// Update godoc
// @Summary Update subscription
// @Description Partially update subscription fields
// @Tags subscriptions
// @Accept json
// @Produce json
// @Param id path int true "Subscription ID"
// @Param request body object true "Partial update payload"
// @Success 200 {object} domain.SubscriptionDTO
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	var raw map[string]any
	if err := c.ShouldBindJSON(&raw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var req service.UpdateSubscriptionRequest

	if v, ok := raw["service_name"]; ok {
		if s, ok := v.(string); ok {
			req.ServiceName = &s
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "service_name must be string"})
			return
		}
	}
	if v, ok := raw["price"]; ok {
		// gin/json decodes numbers as float64 when using map[string]any
		f, ok := v.(float64)
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price must be number"})
			return
		}
		p := int64(f)
		if f != float64(int64(f)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price must be integer"})
			return
		}
		req.Price = &p
	}
	if v, ok := raw["user_id"]; ok {
		if s, ok := v.(string); ok {
			req.UserID = &s
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user_id must be string"})
			return
		}
	}
	if v, ok := raw["start_date"]; ok {
		if s, ok := v.(string); ok {
			req.StartDate = &s
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be string"})
			return
		}
	}

	// end_date: 3-state
	if _, exists := raw["end_date"]; exists {
		if raw["end_date"] == nil {
			req.EndDate = service.EndDateSetNull()
		} else {
			s, ok := raw["end_date"].(string)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "end_date must be string or null"})
				return
			}
			req.EndDate = service.EndDateSetValue(s)
		}
	} else {
		req.EndDate = service.EndDateNotProvided()
	}

	updated, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.ToDTO(*updated))
}

// Delete godoc
// @Summary Delete subscription
// @Description Delete subscription by ID
// @Tags subscriptions
// @Param id path int true "Subscription ID"
// @Success 204 "No Content"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c, "id")
	if !ok {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		writeError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// List godoc
// @Summary List subscriptions
// @Description Get list of subscriptions with optional filters
// @Tags subscriptions
// @Produce json
// @Param user_id query string false "User ID (UUID)"
// @Param service_name query string false "Service name"
// @Param from query string false "Start month (MM-YYYY)"
// @Param to query string false "End month (MM-YYYY)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {array} domain.SubscriptionDTO
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions [get]
func (h *Handler) List(c *gin.Context) {
	f := domain.ListFilter{
		Limit:  50,
		Offset: 0,
	}

	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		f.UserID = &v
	}
	if v := strings.TrimSpace(c.Query("service_name")); v != "" {
		f.ServiceName = &v
	}

	if v := strings.TrimSpace(c.Query("from")); v != "" {
		t, err := domain.ParseMonthYear(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' (expected MM-YYYY)"})
			return
		}
		f.From = &t
	}
	if v := strings.TrimSpace(c.Query("to")); v != "" {
		t, err := domain.ParseMonthYear(v)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' (expected MM-YYYY)"})
			return
		}
		f.To = &t
	}

	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'limit'"})
			return
		}
		f.Limit = n
	}
	if v := strings.TrimSpace(c.Query("offset")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'offset'"})
			return
		}
		f.Offset = n
	}

	items, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		writeError(c, err)
		return
	}

	out := make([]domain.SubscriptionDTO, 0, len(items))
	for _, s := range items {
		out = append(out, domain.ToDTO(s))
	}
	c.JSON(http.StatusOK, out)
}

// Total godoc
// @Summary Calculate total subscription cost
// @Description Calculate total cost of subscriptions for a given period
// @Tags subscriptions
// @Produce json
// @Param from query string true "Start month (MM-YYYY)"
// @Param to query string true "End month (MM-YYYY)"
// @Param user_id query string false "User ID (UUID)"
// @Param service_name query string false "Service name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/v1/subscriptions/total [get]
func (h *Handler) Total(c *gin.Context) {
	fromStr := strings.TrimSpace(c.Query("from"))
	toStr := strings.TrimSpace(c.Query("to"))
	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "'from' and 'to' are required (MM-YYYY)"})
		return
	}

	from, err := domain.ParseMonthYear(fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'from' (expected MM-YYYY)"})
		return
	}
	to, err := domain.ParseMonthYear(toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'to' (expected MM-YYYY)"})
		return
	}

	var userID *string
	if v := strings.TrimSpace(c.Query("user_id")); v != "" {
		userID = &v
	}
	var serviceName *string
	if v := strings.TrimSpace(c.Query("service_name")); v != "" {
		serviceName = &v
	}

	total, err := h.svc.TotalCost(c.Request.Context(), domain.TotalFilter{
		UserID:      userID,
		ServiceName: serviceName,
		From:        from,
		To:          to,
	})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":    total,
		"currency": "RUB",
		"from":     fromStr,
		"to":       toStr,
	})
}

// ---------- Helpers ----------

func parseIDParam(c *gin.Context, name string) (int64, bool) {
	raw := c.Param(name)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidInput),
		errors.Is(err, service.ErrInvalidDateRange):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
