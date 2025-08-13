package test_items

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
	CreateTestItemEvent = "testitems.create"
	UpdateTestItemEvent = "testitems.update"
	DeleteTestItemEvent = "testitems.delete"
)

type TestItemsService struct {
	DB      *gorm.DB
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
	Logger  logger.Logger
}

func NewTestItemsService(db *gorm.DB, emitter *emitter.Emitter, storage *storage.ActiveStorage, logger logger.Logger) *TestItemsService {
	return &TestItemsService{
		DB:      db,
		Emitter: emitter,
		Storage: storage,
		Logger:  logger,
	}
}

// applySorting applies sorting to the query based on the sort and order parameters
func (s *TestItemsService) applySorting(query *gorm.DB, sortBy *string, sortOrder *string) {
	// Valid sortable fields for TestItem
	validSortFields := map[string]string{
		"id":         "id",
		"created_at": "created_at",
		"updated_at": "updated_at",
		"name":       "name",
		"email":      "email",
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

func (s *TestItemService) Create(req *models.CreateTestItemRequest) (*models.TestItem, error) {
	item := &models.TestItem{
		Name:  req.Name,
		Email: req.Email,
	}

	if err := s.DB.Create(item).Error; err != nil {
		s.Logger.Error("failed to create testitem", logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create testitem: %w", err)
	}

	// Emit create event
	s.Emitter.Emit(CreateTestItemEvent, item)

	return s.GetById(item.Id)
}

func (s *TestItemService) Update(id uint, req *models.UpdateTestItemRequest) (*models.TestItem, error) {
	item := &models.TestItem{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find testitem for update",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to find testitem: %w", err)
	}

	// Build updates map
	updates := make(map[string]any)
	// For string and other fields
	if req.Name != "" {
		updates["name"] = req.Name
	}
	// For string and other fields
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if err := s.DB.Model(item).Updates(updates).Error; err != nil {
		s.Logger.Error("failed to update testitem",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to update testitem: %w", err)
	}

	result, err := s.GetById(item.Id)
	if err != nil {
		s.Logger.Error("failed to get updated testitem",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to get updated testitem: %w", err)
	}

	// Emit update event
	s.Emitter.Emit(UpdateTestItemEvent, result)

	return result, nil
}

func (s *TestItemService) Delete(id uint) error {
	item := &models.TestItem{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find testitem for deletion",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return fmt.Errorf("failed to find testitem: %w", err)
	}

	// Delete file attachments if any

	if err := s.DB.Delete(item).Error; err != nil {
		s.Logger.Error("failed to delete testitem",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return fmt.Errorf("failed to delete testitem: %w", err)
	}

	// Emit delete event
	s.Emitter.Emit(DeleteTestItemEvent, item)

	return nil
}

func (s *TestItemService) GetById(id uint) (*models.TestItem, error) {
	item := &models.TestItem{}

	query := item.Preload(s.DB)

	if err := query.First(item, id).Error; err != nil {
		s.Logger.Error("failed to get testitem",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to get testitem: %w", err)
	}

	return item, nil
}

func (s *TestItemService) GetAll(page *int, limit *int, sortBy *string, sortOrder *string) (*types.PaginatedResponse, error) {
	var items []*models.TestItem
	var total int64

	query := s.DB.Model(&models.TestItem{})
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
		s.Logger.Error("failed to count testitems",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to count testitems: %w", err)
	}

	// Apply pagination if provided
	if page != nil && limit != nil {
		offset := (*page - 1) * *limit
		query = query.Offset(offset).Limit(*limit)
	}

	// Apply sorting
	s.applySorting(query, sortBy, sortOrder)

	// Preload relationships
	query = (&models.TestItem{}).Preload(query)

	// Execute query
	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("failed to get testitems",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get testitems: %w", err)
	}

	// Convert to response type
	responses := make([]*models.TestItemListResponse, len(items))
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
