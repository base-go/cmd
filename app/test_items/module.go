package test_items

import (
	"base/app/models"
	"base/core/module"

	"gorm.io/gorm"
)

func init() {
	// Self-register the module for auto-discovery
	module.RegisterAppModule("test_items", func(deps module.Dependencies) module.Module {
		mod := &Module{
			DB: deps.DB,
		}

		// Setup service and controller
		service := NewTestItemsService(deps.DB, deps.Emitter, deps.Storage, deps.Logger)
		controller := NewTestItemsController(service, deps.Storage)

		// Register routes using kebab-case
		group := deps.Router.Group("/test-items")
		controller.Routes(group)

		return mod
	})
}

type Module struct {
	module.DefaultModule
	DB *gorm.DB
}

func NewTestItemModule(db *gorm.DB) *Module {
	return &Module{
		DB: db,
	}
}

func (m *Module) Init() error {
	return nil
}

func (m *Module) Migrate() error {
	return m.DB.AutoMigrate(&models.TestItem{})
}

func (m *Module) GetModels() []any {
	return []any{
		&models.TestItem{},
	}
}
