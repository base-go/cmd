package posts

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	 
)

type PostController struct {
	PostService *PostService
}

func NewPostController(service *PostService) *PostController {
	return &PostController{
		PostService: service,
	}
}

func (c *PostController) Routes(router *gin.RouterGroup) {
	router.GET("/posts", c.List)
	router.GET("/posts/:id", c.Get)
	router.POST("/posts", c.Create)
	router.PUT("/posts/:id", c.Update)
	router.DELETE("/posts/:id", c.Delete)
}

// CreatePost godoc
// @Summary Create a new Post
// @Description Create a new Post with the input payload
// @Tags App/Post
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param posts body CreatePostRequest true "Create Post"
// @Success 201 {object} PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [post]
func (c *PostController) Create(ctx *gin.Context) {
	var req CreatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.PostService.Create(&req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, item.ToResponse())
}

// GetPost godoc
// @Summary Get a Post
// @Description Get a Post by its ID
// @Tags App/Post
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} PostResponse
// @Failure 404 {object} ErrorResponse
// @Router /posts/{id} [get]
func (c *PostController) Get(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	item, err := c.PostService.GetByID(uint(id))
	if err != nil {
		ctx.JSON(http.StatusNotFound, ErrorResponse{Error: "Item not found"})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// ListPost godoc
// @Summary List Post
// @Description Get a list of all Post
// @Tags App/Post
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Success 200 {array} PostResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts [get]
func (c *PostController) List(ctx *gin.Context) {
	items, err := c.PostService.GetAll()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to fetch items: " + err.Error()})
		return
	}

	responses := make([]*PostResponse, len(items))
	for i, item := range items {
		responses[i] = item.ToResponse()
	}

	ctx.JSON(http.StatusOK, responses)
}

// UpdatePost godoc
// @Summary Update a Post
// @Description Update a Post by its ID
// @Tags App/Post
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Param posts body UpdatePostRequest true "Update Post"
// @Success 200 {object} PostResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id} [put]
func (c *PostController) Update(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	var req UpdatePostRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	item, err := c.PostService.Update(uint(id), &req)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to update item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, item.ToResponse())
}

// DeletePost godoc
// @Summary Delete a Post
// @Description Delete a Post by its ID
// @Tags App/Post
// @Security ApiKeyAuth
// @Accept json
// @Produce json
// @Param id path int true "ID"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /posts/{id} [delete]
func (c *PostController) Delete(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid ID format"})
		return
	}

	if err := c.PostService.Delete(uint(id)); err != nil {
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to delete item: " + err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, SuccessResponse{Message: "Item deleted successfully"})
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}