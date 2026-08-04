package handlers

import (
	"net/http"
	"strconv"

	"lms-backend/usecases"

	"github.com/gin-gonic/gin"
)

type ApplicationHandler struct {
	usecase usecases.ApplicationUseCase
}

func NewApplicationHandler(usecase usecases.ApplicationUseCase) *ApplicationHandler {
	return &ApplicationHandler{usecase}
}

func (h *ApplicationHandler) GetAll(c *gin.Context) {
	period := c.Query("period")
	status := c.Query("status")
	apps, err := h.usecase.GetApplicationsByPeriodAndStatus(period, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": apps})
}

func (h *ApplicationHandler) Simulate(c *gin.Context) {
	var req usecases.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	sim, err := h.usecase.SimulateApplication(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": sim})
}

func (h *ApplicationHandler) Submit(c *gin.Context) {
	var req usecases.SubmitApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	app, err := h.usecase.SubmitApplication(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": app, "message": "Application submitted successfully"})
}

type ApproveRequest struct {
	Action      string  `json:"action"`
	ApprovedAmount float64 `json:"approved_amount"`
	Notes       string  `json:"notes"`
	UpdatedUser string  `json:"updated_user"`
}

func (h *ApplicationHandler) Approve(c *gin.Context) {
	idStr := c.Param("id")
	applicationNo, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	var req ApproveRequest
	_ = c.ShouldBindJSON(&req)

	if req.Notes == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Catatan approval wajib diisi!"})
		return
	}

	action := req.Action
	if action == "" {
		action = "APPROVED"
	}

	if err := h.usecase.ProcessApproval(applicationNo, action, req.Notes, req.UpdatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Approval status updated successfully"})
}

func (h *ApplicationHandler) GetTrackings(c *gin.Context) {
	idStr := c.Param("id")
	appNo, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	trackings, err := h.usecase.GetLoanTrackings(appNo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": trackings})
}

func (h *ApplicationHandler) Disburse(c *gin.Context) {
	idStr := c.Param("id")
	appNo, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid application ID format"})
		return
	}

	var req usecases.DisburseRequest
	_ = c.ShouldBindJSON(&req)

	if err := h.usecase.DisburseApplication(appNo, req, req.UpdatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pencairan dana pinjaman berhasil diproses dan rekening pinjaman aktif!"})
}
