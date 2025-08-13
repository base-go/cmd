package models

import (
	"time"

	"gorm.io/gorm"
)

// Task represents a task entity
type Task struct {
	Id          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
	Title       string         `json:"title" gorm:"size:255"`
	Description string         `json:"description" gorm:"type:text"`
	Status      string         `json:"status" gorm:"size:255"`
}

// TableName returns the table name for the Task model
func (m *Task) TableName() string {
	return "tasks"
}

// GetId returns the Id of the model
func (m *Task) GetId() uint {
	return m.Id
}

// GetModelName returns the model name
func (m *Task) GetModelName() string {
	return "task"
}

// TaskResponse represents the API response for Task
type TaskResponse struct {
	Id          uint           `json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at,omitempty"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
}

// CreateTaskRequest represents the request payload for creating a Task
type CreateTaskRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// UpdateTaskRequest represents the request payload for updating a Task
type UpdateTaskRequest struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
}

// ToResponse converts the model to an API response
func (m *Task) ToResponse() *TaskResponse {
	if m == nil {
		return nil
	}
	return &TaskResponse{
		Id:          m.Id,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
		Title:       m.Title,
		Description: m.Description,
		Status:      m.Status,
	}
}

// Preload preloads all the model's relationships
func (m *Task) Preload(db *gorm.DB) *gorm.DB {
	query := db
	return query
}
