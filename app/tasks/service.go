package tasks

import (
	"fmt"
	"math"

	"base/app/models"
	"base/core/emitter"
	"base/core/logger"
	"base/core/storage"
	"base/core/types"

	"gorm.io/gorm"
)

const (
	CreateTaskEvent = "tasks.create"
	UpdateTaskEvent = "tasks.update"
	DeleteTaskEvent = "tasks.delete"
)

type TasksService struct {
	DB      *gorm.DB
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
	Logger  logger.Logger
}

func NewTasksService(db *gorm.DB, emitter *emitter.Emitter, storage *storage.ActiveStorage, logger logger.Logger) *TasksService {
	return &TasksService{
		DB:      db,
		Emitter: emitter,
		Storage: storage,
		Logger:  logger,
	}
}

// applySorting applies sorting to the query based on the sort and order parameters
func (s *TasksService) applySorting(query *gorm.DB, sortBy *string, sortOrder *string) {
	// Valid sortable fields for Task
	validSortFields := map[string]string{
		"id":          "id",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
		"title":       "title",
		"description": "description",
		"status":      "status",
	}

	// Default sorting - if sort_order exists, always use it for custom ordering
	defaultSortBy := "id"
	defaultSortOrder := "desc"

	// Determine sort field
	sortField := defaultSortBy
	if sortBy != nil && *sortBy != "" {
		if field, exists := validSortFields[*sortBy]; exists {
			sortField = field
		}
	}

	// Determine sort direction (order parameter)
	sortDirection := defaultSortOrder
	if sortOrder != nil && (*sortOrder == "asc" || *sortOrder == "desc") {
		sortDirection = *sortOrder
	}

	// Apply sorting
	query.Order(sortField + " " + sortDirection)
}

func (s *TaskService) Create(req *models.CreateTaskRequest) (*models.Task, error) {
	item := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := s.DB.Create(item).Error; err != nil {
		s.Logger.Error("failed to create task", logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	// Emit create event
	s.Emitter.Emit(CreateTaskEvent, item)

	return s.GetById(item.Id)
}

func (s *TaskService) Update(id uint, req *models.UpdateTaskRequest) (*models.Task, error) {
	item := &models.Task{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find task for update",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to find task: %w", err)
	}

	// Build updates map
	updates := make(map[string]any)
	// For string and other fields
	if req.Title != "" {
		updates["title"] = req.Title
	}
	// For string and other fields
	if req.Description != "" {
		updates["description"] = req.Description
	}
	// For string and other fields
	if req.Status != "" {
		updates["status"] = req.Status
	}

	if err := s.DB.Model(item).Updates(updates).Error; err != nil {
		s.Logger.Error("failed to update task",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	result, err := s.GetById(item.Id)
	if err != nil {
		s.Logger.Error("failed to get updated task",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to get updated task: %w", err)
	}

	// Emit update event
	s.Emitter.Emit(UpdateTaskEvent, result)

	return result, nil
}

func (s *TaskService) Delete(id uint) error {
	item := &models.Task{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find task for deletion",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return fmt.Errorf("failed to find task: %w", err)
	}

	// Delete file attachments if any

	if err := s.DB.Delete(item).Error; err != nil {
		s.Logger.Error("failed to delete task",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Emit delete event
	s.Emitter.Emit(DeleteTaskEvent, item)

	return nil
}

func (s *TaskService) GetById(id uint) (*models.Task, error) {
	item := &models.Task{}

	query := item.Preload(s.DB)

	if err := query.First(item, id).Error; err != nil {
		s.Logger.Error("failed to get task",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	return item, nil
}

func (s *TaskService) GetAll(page *int, limit *int, sortBy *string, sortOrder *string) (*types.PaginatedResponse, error) {
	var items []*models.Task
	var total int64

	query := s.DB.Model(&models.Task{})
	// Set default values if nil
	defaultPage := 1
	defaultLimit := 10
	if page == nil {
		page = &defaultPage
	}
	if limit == nil {
		limit = &defaultLimit
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		s.Logger.Error("failed to count tasks",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to count tasks: %w", err)
	}

	// Apply pagination if provided
	if page != nil && limit != nil {
		offset := (*page - 1) * *limit
		query = query.Offset(offset).Limit(*limit)
	}

	// Apply sorting
	s.applySorting(query, sortBy, sortOrder)

	// Preload relationships
	query = (&models.Task{}).Preload(query)

	// Execute query
	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("failed to get tasks",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get tasks: %w", err)
	}

	// Convert to response type
	responses := make([]*models.TaskListResponse, len(items))
	for i, item := range items {
		responses[i] = item.ToListResponse()
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(total) / float64(*limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &types.PaginatedResponse{
		Data: responses,
		Pagination: types.Pagination{
			Total:      int(total),
			Page:       *page,
			PageSize:   *limit,
			TotalPages: totalPages,
		},
	}, nil
}
