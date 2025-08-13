package test_items

import (
	"net/http"
	"strconv"

	"base/app/models"
	"base/core/router"
	"base/core/storage"
)

type TestItemController struct {
	Service *TestItemService
	Storage *storage.ActiveStorage
}

func NewTestItemController(service *TestItemService, storage *storage.ActiveStorage) *TestItemController {
	return &TestItemController{
		Service: service,
		Storage: storage,
	}
}

func (c *TestItemController) Routes(router *router.RouterGroup) {
	// Main CRUD endpoints
	router.GET("/test-items", c.List)        // Paginated list
	router.GET("/test-items/all", c.ListAll) // Unpaginated list
	router.GET("/test-items/:id", c.Get)
	router.POST("/test-items", c.Create)
	router.PUT("/test-items/:id", c.Update)
	router.DELETE("/test-items/:id", c.Delete)
}

// CreateTestItem godoc
// @Summary Create a new TestItem
// @Description Create a new TestItem with the input payload
// @Tags App/TestItem
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param test-items body models.CreateTestItemRequest true "Create TestItem request"
// @Success 201 {object} models.TestItemResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/test-items [post]
func (c *TestItemController) Create(ctx *router.Context) error {
	var req models.CreateTestItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item: " + err.Error()})
	}

	return ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetTestItem godoc
// @Summary Get a TestItem
// @Description Get a TestItem by its id
// @Tags App/TestItem
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "TestItem id"
// @Success 200 {object} models.TestItemResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/test-items/{id} [get]
func (c *TestItemController) Get(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
	}

	item, err := c.Service.GetById(uint(id))
	if err != nil {
		return ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
	}

	return ctx.JSON(http.StatusOK, item.ToResponse())
}

// ListTestItems godoc
// @Summary List test-items
// @Description Get a list of test-items
// @Tags App/TestItem
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort query string false "Sort field (id, created_at, updated_at,name,email,)"
// @Param order query string false "Sort order (asc, desc)"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/test-items [get]
func (c *TestItemController) List(ctx *router.Context) error {
	var page, limit *int
	var sortBy, sortOrder *string

	// Parse page parameter
	if pageStr := ctx.Query("page"); pageStr != "" {
		if pageNum, err := strconv.Atoi(pageStr); err == nil && pageNum > 0 {
			page = &pageNum
		} else {
			return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid page number"})
		}
	}

	// Parse limit parameter
	if limitStr := ctx.Query("limit"); limitStr != "" {
		if limitNum, err := strconv.Atoi(limitStr); err == nil && limitNum > 0 {
			limit = &limitNum
		} else {
			return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid limit number"})
		}
	}

	// Parse sort parameters
	if sortStr := ctx.Query("sort"); sortStr != "" {
		sortBy = &sortStr
	}

	if orderStr := ctx.Query("order"); orderStr != "" {
		if orderStr == "asc" || orderStr == "desc" {
			sortOrder = &orderStr
		} else {
			return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid sort order. Use 'asc' or 'desc'"})
		}
	}

	paginatedResponse, err := c.Service.GetAll(page, limit, sortBy, sortOrder)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, paginatedResponse)
}

// ListAllTestItems godoc
// @Summary List all test-items without pagination
// @Description Get a list of all test-items without pagination
// @Tags App/TestItem
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} types.PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/test-items/all [get]
func (c *TestItemController) ListAll(ctx *router.Context) error {
	paginatedResponse, err := c.Service.GetAll(nil, nil, nil, nil)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch all items: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, paginatedResponse)
}

// UpdateTestItem godoc
// @Summary Update a TestItem
// @Description Update a TestItem by its id
// @Tags App/TestItem
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "TestItem id"
// @Param test-items body models.UpdateTestItemRequest true "Update TestItem request"
// @Success 200 {object} models.TestItemResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/test-items/{id} [put]
func (c *TestItemController) Update(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
	}

	var req models.UpdateTestItemRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Update(uint(id), &req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, item.ToResponse())
}

// DeleteTestItem godoc
// @Summary Delete a TestItem
// @Description Delete a TestItem by its id
// @Tags App/TestItem
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "TestItem id"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/test-items/{id} [delete]
func (c *TestItemController) Delete(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
	}

	if err := c.Service.Delete(uint(id)); err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete item: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, SuccessResponse{Message: "Item deleted successfully"})
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}
