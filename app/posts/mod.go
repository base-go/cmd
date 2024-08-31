package posts

import (
	"base/core/module"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PostModule struct {
	module.DefaultModule
	DB         *gorm.DB
	Controller *PostController
	Service    *PostService
}

func NewPostModule(db *gorm.DB, router *gin.RouterGroup) module.Module {
	service := NewPostService(db)
	controller := NewPostController(service)

	postsModule := &PostModule{
		DB:         db,
		Controller: controller,
		Service:    service,
	}

	postsModule.Routes(router)
	postsModule.Migrate()

	return postsModule
}

func (m *PostModule) Routes(router *gin.RouterGroup) {
	m.Controller.Routes(router)
}

func (m *PostModule) Migrate() error {
	return m.DB.AutoMigrate(&Post{})
}

func (m *PostModule) GetModels() []interface{} {
    return []interface{}{&Post{}}
}

