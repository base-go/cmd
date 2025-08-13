package models

import (
	"time"

	"gorm.io/gorm"
)

// Category represents a category entity
type Category struct {
	Id          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Name        string         `json:"name" gorm:"size:255"`
	Description string         `json:"description" gorm:"type:text"`
}

// TableName returns the table name for the Category model
func (m *Category) TableName() string {
	return "categories"
}

// GetId returns the Id of the model
func (m *Category) GetId() uint {
	return m.Id
}

// GetModelName returns the model name
func (m *Category) GetModelName() string {
	return "category"
}

// CategoryResponse represents the API response for Category
type CategoryResponse struct {
	Id          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
}

// CategorySelectOption represents a simplified response for select boxes and dropdowns
type CategorySelectOption struct {
	Id   uint   `json:"id"`
	Name string `json:"name"` // From Name field (string)
}

// CreateCategoryRequest represents the request payload for creating a Category
type CreateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateCategoryRequest represents the request payload for updating a Category
type UpdateCategoryRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// ToResponse converts the model to an API response
func (m *Category) ToResponse() *CategoryResponse {
	if m == nil {
		return nil
	}
	return &CategoryResponse{
		Id:          m.Id,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
		Name:        m.Name,
		Description: m.Description,
	}
}

// ToSelectOption converts the model to a select option for dropdowns
func (m *Category) ToSelectOption() *CategorySelectOption {
	if m == nil {
		return nil
	}

	var displayName string
	displayName = m.Name

	return &CategorySelectOption{
		Id:   m.Id,
		Name: displayName,
	}
}

// Preload preloads all the model's relationships
func (m *Category) Preload(db *gorm.DB) *gorm.DB {
	query := db
	return query
}
