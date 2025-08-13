package categories

import (
	"net/http"
	"strconv"

	"base/app/models"
	"base/core/router"
	"base/core/storage"
)

type CategoryController struct {
	Service *CategoryService
	Storage *storage.ActiveStorage
}

func NewCategoryController(service *CategoryService, storage *storage.ActiveStorage) *CategoryController {
	return &CategoryController{
		Service: service,
		Storage: storage,
	}
}

func (c *CategoryController) Routes(router *router.Router) {
	// Main CRUD endpoints
	router.GET("/categories", c.List)        // Paginated list
	router.GET("/categories/all", c.ListAll) // Unpaginated list
	router.GET("/categories/:id", c.Get)
	router.POST("/categories", c.Create)
	router.PUT("/categories/:id", c.Update)
	router.DELETE("/categories/:id", c.Delete)
}

// CreateCategory godoc
// @Summary Create a new Category
// @Description Create a new Category with the input payload
// @Tags App/Category
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param categories body models.CreateCategoryRequest true "Create Category request"
// @Success 201 {object} models.CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/categories [post]
func (c *CategoryController) Create(ctx *router.Context) error {
	var req models.CreateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item: " + err.Error()})
	}

	return ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetCategory godoc
// @Summary Get a Category
// @Description Get a Category by its id
// @Tags App/Category
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Category id"
// @Success 200 {object} models.CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/categories/{id} [get]
func (c *CategoryController) Get(ctx *router.Context) error {
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

// ListCategories godoc
// @Summary List categories
// @Description Get a list of categories
// @Tags App/Category
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort query string false "Sort field (id, created_at, updated_at,name,description,)"
// @Param order query string false "Sort order (asc, desc)"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/categories [get]
func (c *CategoryController) List(ctx *router.Context) error {
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

// ListAllCategories godoc
// @Summary List all categories for select options
// @Description Get a simplified list of all categories with id and name only (for dropdowns/select boxes)
// @Tags App/Category
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {array} models.CategorySelectOption
// @Failure 500 {object} ErrorResponse
// @Router /api/categories/all [get]
func (c *CategoryController) ListAll(ctx *router.Context) error {
	items, err := c.Service.GetAllForSelect()
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch select options: " + err.Error()})
	}

	// Convert to select options
	var selectOptions []*models.CategorySelectOption
	for _, item := range items {
		selectOptions = append(selectOptions, item.ToSelectOption())
	}

	return ctx.JSON(http.StatusOK, selectOptions)
}

// UpdateCategory godoc
// @Summary Update a Category
// @Description Update a Category by its id
// @Tags App/Category
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Category id"
// @Param categories body models.UpdateCategoryRequest true "Update Category request"
// @Success 200 {object} models.CategoryResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/categories/{id} [put]
func (c *CategoryController) Update(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
	}

	var req models.UpdateCategoryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Update(uint(id), &req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, item.ToResponse())
}

// DeleteCategory godoc
// @Summary Delete a Category
// @Description Delete a Category by its id
// @Tags App/Category
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Category id"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/categories/{id} [delete]
func (c *CategoryController) Delete(ctx *router.Context) error {
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
