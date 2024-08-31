package choices

import (
	"base/core/module"
	"base/app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ChoiceModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *ChoiceController
	Service    *ChoiceService
}

func NewChoiceModule(db *gorm.DB, router *gin.RouterGroup) module.Module {
	service := NewChoiceService(db)
	controller := NewChoiceController(service)

	choicesModule := &ChoiceModule{
		DB:         db,
		Controller: controller,
		Service:    service,
	}

	choicesModule.Routes(router)
	choicesModule.Migrate()

	return choicesModule
}

func (m *ChoiceModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *ChoiceModule) Migrate() error {
	return m.DB.AutoMigrate(&models.Choice{})
}

func (m *ChoiceModule) GetModels() []interface{} {
    return []interface{}{&models.Choice{}}
}