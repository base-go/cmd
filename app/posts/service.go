package posts

import (
	"errors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

)

type PostService struct {
	DB *gorm.DB
}

func NewPostService(db *gorm.DB) *PostService {
	return &PostService{
		DB: db,
	}
}

func (s *PostService) Create(req *CreatePostRequest) (*Post, error) {
	item := Post{
		Title: req.Title,
		Body: req.Body,
		Published: req.Published,
	}

	if err := s.DB.Create(&item).Error; err != nil {
		return nil, err
	}

	return s.GetByID(item.ID)
}

func (s *PostService) GetByID(id uint) (*Post, error) {
	var item Post
	if err := s.DB.Preload(clause.Associations).First(&item, id).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *PostService) GetAll() ([]Post, error) {
	var items []Post
	if err := s.DB.Preload(clause.Associations).Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *PostService) Update(id uint, req *UpdatePostRequest) (*Post, error) {
	item, err := s.GetByID(id)
	if err != nil {
		return nil, err
	}
	if req.Title != nil {
		item.Title = *req.Title
	}
	if req.Body != nil {
		item.Body = *req.Body
	}
	if req.Published != nil {
		item.Published = *req.Published
	}

	if err := s.DB.Save(item).Error; err != nil {
		return nil, err
	}

	return s.GetByID(id)
}

func (s *PostService) Delete(id uint) error {
	result := s.DB.Delete(&Post{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("item not found")
	}
	return nil
}

func (s *PostService) GetAssociated(id uint, associationName string) (interface{}, error) {
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