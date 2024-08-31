package choices

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"base/app/models"
)

type ChoiceService struct {
	DB *gorm.DB
}

func NewChoiceService(db *gorm.DB) *ChoiceService {
	return &ChoiceService{
		DB: db,
	}
}

func (s *ChoiceService) Create(req *models.CreateChoiceRequest) (*models.Choice, error) {
	item := models.Choice{
		Content: req.Content,
		DialogID: req.DialogID,
		Next_sceneID: req.Next_sceneID,
	}

	if err := s.DB.Create(&item).Error; err != nil {
		return nil, err
	}

	return s.GetByID(item.ID)
}

func (s *ChoiceService) GetByID(id uint) (*models.Choice, error) {
	var item models.Choice
	if err := s.DB.Preload(clause.Associations).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *ChoiceService) GetAll() ([]*models.Choice, error) {
	var items []*models.Choice
	if err := s.DB.Preload(clause.Associations).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *ChoiceService) Update(id uint, req *models.UpdateChoiceRequest) (*models.Choice, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Content != nil {
		item.Content = *req.Content
	}
	if req.DialogID != nil {
		item.DialogID = *req.DialogID
	}
	if req.Next_sceneID != nil {
		item.Next_sceneID = *req.Next_sceneID
	}

	if err := s.DB.Save(item).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

func (s *ChoiceService) Delete(id uint) error {
	result := s.DB.Delete(&models.Choice{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	return nil
}

func (s *ChoiceService) GetAssociated(id uint, associationName string) (interface{}, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}

	var associated interface{}
	if err := s.DB.Model(item).Association(associationName).Find(&associated); err != nil {
		return nil, err
	}

	return associated, nil
}