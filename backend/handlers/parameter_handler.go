package handlers

import (
	"net/http"
	"strconv"

	"lms-backend/usecases"

	"github.com/gin-gonic/gin"
)

type ParameterHandler struct {
	usecase usecases.ParameterUseCase
}

func NewParameterHandler(u usecases.ParameterUseCase) *ParameterHandler {
	return &ParameterHandler{usecase: u}
}

func (h *ParameterHandler) GetAll(c *gin.Context) {
	params, err := h.usecase.GetAllParameters()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": params})
}

func (h *ParameterHandler) Save(c *gin.Context) {
	var req usecases.SaveParameterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	param, err := h.usecase.SaveParameter(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": param, "message": "Parameter saved successfully"})
}

func (h *ParameterHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := h.usecase.DeleteParameter(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Parameter deleted successfully"})
}
