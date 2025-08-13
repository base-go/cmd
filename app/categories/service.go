package categories

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
	CreateCategoryEvent = "categories.create"
	UpdateCategoryEvent = "categories.update"
	DeleteCategoryEvent = "categories.delete"
)

type CategoriesService struct {
	DB      *gorm.DB
	Emitter *emitter.Emitter
	Storage *storage.ActiveStorage
	Logger  logger.Logger
}

func NewCategoriesService(db *gorm.DB, emitter *emitter.Emitter, storage *storage.ActiveStorage, logger logger.Logger) *CategoriesService {
	return &CategoriesService{
		DB:      db,
		Emitter: emitter,
		Storage: storage,
		Logger:  logger,
	}
}

// applySorting applies sorting to the query based on the sort and order parameters
func (s *CategoriesService) applySorting(query *gorm.DB, sortBy *string, sortOrder *string) {
	// Valid sortable fields for Category
	validSortFields := map[string]string{
		"id":          "id",
		"created_at":  "created_at",
		"updated_at":  "updated_at",
		"name":        "name",
		"description": "description",
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

func (s *CategoryService) Create(req *models.CreateCategoryRequest) (*models.Category, error) {
	item := &models.Category{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.DB.Create(item).Error; err != nil {
		s.Logger.Error("failed to create category", logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	// Emit create event
	s.Emitter.Emit(CreateCategoryEvent, item)

	return s.GetById(item.Id)
}

func (s *CategoryService) Update(id uint, req *models.UpdateCategoryRequest) (*models.Category, error) {
	item := &models.Category{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find category for update",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to find category: %w", err)
	}

	// Build updates map
	updates := make(map[string]any)
	// For string and other fields
	if req.Name != "" {
		updates["name"] = req.Name
	}
	// For string and other fields
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if err := s.DB.Model(item).Updates(updates).Error; err != nil {
		s.Logger.Error("failed to update category",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	result, err := s.GetById(item.Id)
	if err != nil {
		s.Logger.Error("failed to get updated category",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to get updated category: %w", err)
	}

	// Emit update event
	s.Emitter.Emit(UpdateCategoryEvent, result)

	return result, nil
}

func (s *CategoryService) Delete(id uint) error {
	item := &models.Category{}
	if err := s.DB.First(item, id).Error; err != nil {
		s.Logger.Error("failed to find category for deletion",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return fmt.Errorf("failed to find category: %w", err)
	}

	// Delete file attachments if any

	if err := s.DB.Delete(item).Error; err != nil {
		s.Logger.Error("failed to delete category",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return fmt.Errorf("failed to delete category: %w", err)
	}

	// Emit delete event
	s.Emitter.Emit(DeleteCategoryEvent, item)

	return nil
}

func (s *CategoryService) GetById(id uint) (*models.Category, error) {
	item := &models.Category{}

	query := item.Preload(s.DB)

	if err := query.First(item, id).Error; err != nil {
		s.Logger.Error("failed to get category",
			logger.String("error", err.Error()),
			logger.Int("id", int(id)))
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return item, nil
}

func (s *CategoryService) GetAll(page *int, limit *int, sortBy *string, sortOrder *string) (*types.PaginatedResponse, error) {
	var items []*models.Category
	var total int64

	query := s.DB.Model(&models.Category{})
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
		s.Logger.Error("failed to count categories",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to count categories: %w", err)
	}

	// Apply pagination if provided
	if page != nil && limit != nil {
		offset := (*page - 1) * *limit
		query = query.Offset(offset).Limit(*limit)
	}

	// Apply sorting
	s.applySorting(query, sortBy, sortOrder)

	// Preload relationships
	query = (&models.Category{}).Preload(query)

	// Execute query
	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("failed to get categories",
			logger.String("error", err.Error()))
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	// Convert to response type
	responses := make([]*models.CategoryListResponse, len(items))
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

// GetAllForSelect gets all items for select box/dropdown options (simplified response)
func (s *CategoryService) GetAllForSelect() ([]*models.Category, error) {
	var items []*models.Category

	query := s.DB.Model(&models.Category{})

	// Only select the necessary fields for select options
	query = query.Select("id, name")

	// Order by name/title for better UX
	query = query.Order("name ASC")

	if err := query.Find(&items).Error; err != nil {
		s.Logger.Error("Failed to fetch items for select", "error", err)
		return nil, err
	}

	return items, nil
}
