package tasks

import (
	"net/http"
	"strconv"

	"base/app/models"
	"base/core/router"
	"base/core/storage"
)

type TaskController struct {
	Service *TaskService
	Storage *storage.ActiveStorage
}

func NewTaskController(service *TaskService, storage *storage.ActiveStorage) *TaskController {
	return &TaskController{
		Service: service,
		Storage: storage,
	}
}

func (c *TaskController) Routes(router *router.Router) {
	// Main CRUD endpoints
	router.GET("/tasks", c.List)        // Paginated list
	router.GET("/tasks/all", c.ListAll) // Unpaginated list
	router.GET("/tasks/:id", c.Get)
	router.POST("/tasks", c.Create)
	router.PUT("/tasks/:id", c.Update)
	router.DELETE("/tasks/:id", c.Delete)
}

// CreateTask godoc
// @Summary Create a new Task
// @Description Create a new Task with the input payload
// @Tags App/Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param tasks body models.CreateTaskRequest true "Create Task request"
// @Success 201 {object} models.TaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tasks [post]
func (c *TaskController) Create(ctx *router.Context) error {
	var req models.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Create(&req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item: " + err.Error()})
	}

	return ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetTask godoc
// @Summary Get a Task
// @Description Get a Task by its id
// @Tags App/Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Task id"
// @Success 200 {object} models.TaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Router /api/tasks/{id} [get]
func (c *TaskController) Get(ctx *router.Context) error {
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

// ListTasks godoc
// @Summary List tasks
// @Description Get a list of tasks
// @Tags App/Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param limit query int false "Number of items per page"
// @Param sort query string false "Sort field (id, created_at, updated_at,title,description,status,)"
// @Param order query string false "Sort order (asc, desc)"
// @Success 200 {object} types.PaginatedResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tasks [get]
func (c *TaskController) List(ctx *router.Context) error {
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

// ListAllTasks godoc
// @Summary List all tasks without pagination
// @Description Get a list of all tasks without pagination
// @Tags App/Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} types.PaginatedResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tasks/all [get]
func (c *TaskController) ListAll(ctx *router.Context) error {
	paginatedResponse, err := c.Service.GetAll(nil, nil, nil, nil)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch all items: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, paginatedResponse)
}

// UpdateTask godoc
// @Summary Update a Task
// @Description Update a Task by its id
// @Tags App/Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Task id"
// @Param tasks body models.UpdateTaskRequest true "Update Task request"
// @Success 200 {object} models.TaskResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tasks/{id} [put]
func (c *TaskController) Update(ctx *router.Context) error {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid id format"})
	}

	var req models.UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	}

	item, err := c.Service.Update(uint(id), &req)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item: " + err.Error()})
	}

	return ctx.JSON(http.StatusOK, item.ToResponse())
}

// DeleteTask godoc
// @Summary Delete a Task
// @Description Delete a Task by its id
// @Tags App/Task
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Task id"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tasks/{id} [delete]
func (c *TaskController) Delete(ctx *router.Context) error {
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
