package choices

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"base/app/models"
)

type ChoiceController struct {
	ChoiceService *ChoiceService
}

func NewChoiceController(service *ChoiceService) *ChoiceController {
	return &ChoiceController{
		ChoiceService: service,
	}
}

func (c *ChoiceController) Routes(router *gin.RouterGroup) {
	router.GET("/choices", c.List)
	router.GET("/choices/:id", c.Get)
	router.POST("/choices", c.Create)
	router.PUT("/choices/:id", c.Update)
	router.DELETE("/choices/:id", c.Delete)
}

// CreateChoice godoc
// @Summary Create a new Choice
// @Description Create a new Choice with the input payload
// @Tags App/Choice
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param choices body models.CreateChoiceRequest true "Create Choice"
// @Success 201 {object} models.ChoiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /choices [post]
func (c *ChoiceController) Create(ctx *gin.Context) {
	var req models.CreateChoiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.ChoiceService.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetChoice godoc
// @Summary Get a Choice
// @Description Get a Choice by its ID
// @Tags App/Choice
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} models.ChoiceResponse
// @Failure 404 {object} ErrorResponse
// @Router /choices/{id} [get]
func (c *ChoiceController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	item, err := c.ChoiceService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// ListChoice godoc
// @Summary List Choice
// @Description Get a list of all Choice
// @Tags App/Choice
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} models.ChoiceResponse
// @Failure 500 {object} ErrorResponse
// @Router /choices [get]
func (c *ChoiceController) List(ctx *gin.Context) {
	items, err := c.ChoiceService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items: " + err.Error()})
		return
	}

	responses := models.ToResponseSlice(items)
	ctx.JSON(http.StatusOK, responses)
}

// UpdateChoice godoc
// @Summary Update a Choice
// @Description Update a Choice by its ID
// @Tags App/Choice
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param choices body models.UpdateChoiceRequest true "Update Choice"
// @Success 200 {object} models.ChoiceResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /choices/{id} [put]
func (c *ChoiceController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	var req models.UpdateChoiceRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.ChoiceService.Update(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// DeleteChoice godoc
// @Summary Delete a Choice
// @Description Delete a Choice by its ID
// @Tags App/Choice
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /choices/{id} [delete]
func (c *ChoiceController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	if err := c.ChoiceService.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Item deleted successfully"})
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}