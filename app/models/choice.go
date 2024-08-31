package models

import (
	"gorm.io/gorm"
	"time"
)

type Choice struct {
	gorm.Model
	Content string `json:"content" gorm:"column:content"`
	DialogID uint `json:"dialogId" gorm:"column:dialog_id"`
	Dialog *Dialog `json:"dialog,omitempty" gorm:"foreignKey:DialogID"`
	Next_sceneID uint `json:"next_sceneId" gorm:"column:next_scene_id"`
	Next_scene *Scene `json:"next_scene,omitempty" gorm:"foreignKey:Next_sceneID"`
}

type CreateChoiceRequest struct {
	Content string `json:"content"`
	DialogID uint `json:"dialogId"`
	Next_sceneID uint `json:"next_sceneId"`
}

type UpdateChoiceRequest struct {
	Content *string `json:"content,omitempty"`
	DialogID *uint `json:"dialogId,omitempty"`
	Next_sceneID *uint `json:"next_sceneId,omitempty"`
}

type ChoiceResponse struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Content string `json:"content"`
	DialogID uint `json:"dialogId"`
	Dialog *DialogResponse `json:"dialog,omitempty"`
	Next_sceneID uint `json:"next_sceneId"`
	Next_scene *SceneResponse `json:"next_scene,omitempty"`
}

func (Choice) TableName() string {
	return "choices"
}

func (item *Choice) ToResponse() *ChoiceResponse {
	response := &ChoiceResponse{
		ID:        item.ID,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
		Content: item.Content,
		DialogID: item.DialogID,
		Next_sceneID: item.Next_sceneID,
	}
	if item.Dialog != nil {
		response.Dialog = item.Dialog.ToResponse()
	}
	if item.Next_scene != nil {
		response.Next_scene = item.Next_scene.ToResponse()
	}

	return response
}

func ToResponseSlice(items []*Choice) []*ChoiceResponse {
	responses := make([]*ChoiceResponse, len(items))
	for i, item := range items {
		responses[i] = item.ToResponse()
	}
	return responses
}