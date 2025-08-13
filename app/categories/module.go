package categories

import (
	"base/app/models"
	"base/core/module"

	"gorm.io/gorm"
)

func init() {
	// Self-register the module for auto-discovery
	module.RegisterAppModule("categories", func(deps module.Dependencies) module.Module {
		mod := &Module{
			DB: deps.DB,
		}

		// Setup service and controller
		service := NewCategoriesService(deps.DB, deps.Emitter, deps.Storage, deps.Logger)
		controller := NewCategoriesController(service, deps.Storage)

		// Register routes
		controller.Routes(deps.Router)

		return mod
	})
}

type Module struct {
	module.DefaultModule
	DB *gorm.DB
}

func NewCategoryModule(db *gorm.DB) *Module {
	return &Module{
		DB: db,
	}
}

func (m *Module) Init() error {
	return nil
}

func (m *Module) Migrate() error {
	return m.DB.AutoMigrate(&models.Category{})
}

func (m *Module) GetModels() []any {
	return []any{
		&models.Category{},
	}
}
