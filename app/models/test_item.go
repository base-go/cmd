package models

import (
	"time"

	"gorm.io/gorm"
)

// TestItem represents a testItem entity
type TestItem struct {
	Id        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Name      string         `json:"name" gorm:"size:255"`
	Email     string         `json:"email" gorm:"size:255;index"`
}

// TableName returns the table name for the TestItem model
func (m *TestItem) TableName() string {
	return "test_items"
}

// GetId returns the Id of the model
func (m *TestItem) GetId() uint {
	return m.Id
}

// GetModelName returns the model name
func (m *TestItem) GetModelName() string {
	return "test_item"
}

// TestItemResponse represents the API response for TestItem
type TestItemResponse struct {
	Id        uint           `json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
}

// CreateTestItemRequest represents the request payload for creating a TestItem
type CreateTestItemRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// UpdateTestItemRequest represents the request payload for updating a TestItem
type UpdateTestItemRequest struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}

// ToResponse converts the model to an API response
func (m *TestItem) ToResponse() *TestItemResponse {
	if m == nil {
		return nil
	}
	return &TestItemResponse{
		Id:        m.Id,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: m.DeletedAt,
		Name:      m.Name,
		Email:     m.Email,
	}
}

// Preload preloads all the model's relationships
func (m *TestItem) Preload(db *gorm.DB) *gorm.DB {
	query := db
	return query
}
