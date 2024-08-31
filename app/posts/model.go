package posts

import (
	"gorm.io/gorm"
	"time"
)

type Post struct {
	gorm.Model
	Title string `json:"title" gorm:"column:title"`
	Body string `json:"body" gorm:"column:body"`
	Published time `json:"published" gorm:"column:published"`
}

type CreatePostRequest struct {
	Title string `json:"title"`
	Body string `json:"body"`
	Published time `json:"published"`
}

type UpdatePostRequest struct {
	Title *string `json:"title,omitempty"`
	Body *string `json:"body,omitempty"`
	Published *time `json:"published,omitempty"`
}

type PostResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Title string `json:"title"`
	Body string `json:"body"`
	Published time `json:"published"`
}

func (Post) TableName() string {
	return "posts"
}

func (item *Post) ToResponse() *PostResponse {
	response := &PostResponse{
		ID:        item.ID,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		Title: item.Title,
		Body: item.Body,
		Published: item.Published,
	}

	return response
}

func ToResponseSlice(items []*Post) []*PostResponse {
	responses := make([]*PostResponse, len(items))
	for i, item := range items {
		responses[i] = item.ToResponse()
	}
	return responses
}