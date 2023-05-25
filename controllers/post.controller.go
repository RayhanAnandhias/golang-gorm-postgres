package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/RayhanAnandhias/golang-gorm-postgres/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// The type PostController contains a pointer to a gorm.DB object.
// @property DB - DB is a property of the PostController struct that holds a pointer to a gorm.DB
// object. This is likely used to interact with a database in the context of the PostController.
type PostController struct {
	DB *gorm.DB
}

// The function returns a new instance of the PostController struct with a given DB object.
func NewPostController(DB *gorm.DB) PostController {
	return PostController{DB}
}

// This function is creating a new post by parsing the request body for a JSON payload containing the
// post's title, content, and image. It then creates a new Post object with the parsed data, along with
// the current user's ID and the current time as the creation and update timestamps. The function then
// attempts to create the new post in the database using the gorm.DB object stored in the
// PostController struct. If the creation is successful, the function returns a JSON response with a
// status of 201 (created) and the newly created post data. If there is an error, the function returns
// a JSON response with an appropriate error status code and message.
func (pc *PostController) CreatePost(ctx *gin.Context) {
	currentUser := ctx.MustGet("currentUser").(models.User)
	var payload *models.CreatePostRequest

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, err.Error())
		return
	}

	now := time.Now()
	newPost := models.Post{
		Title:     payload.Title,
		Content:   payload.Content,
		Image:     payload.Image,
		User:      currentUser.ID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	result := pc.DB.Create(&newPost)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key") {
			ctx.JSON(http.StatusConflict, gin.H{"status": "fail", "message": "Post with that title already exists"})
			return
		}
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": result.Error.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"status": "success", "data": newPost})
}

// The `UpdatePost` function is a method of the `PostController` struct that updates an existing post
// in the database. It first retrieves the post ID from the request parameters and the current user
// from the context. It then parses the request body for a JSON payload containing the updated post's
// title, content, and image. If there is an error in parsing the payload, the function returns a JSON
// response with an appropriate error status code and message.
func (pc *PostController) UpdatePost(ctx *gin.Context) {
	postId := ctx.Param("postId")
	currentUser := ctx.MustGet("currentUser").(models.User)

	var payload *models.UpdatePost
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "fail", "message": err.Error()})
		return
	}
	var updatedPost models.Post
	result := pc.DB.First(&updatedPost, "id = ?", postId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No post with that title exists"})
		return
	}
	now := time.Now()
	postToUpdate := models.Post{
		Title:     payload.Title,
		Content:   payload.Content,
		Image:     payload.Image,
		User:      currentUser.ID,
		CreatedAt: updatedPost.CreatedAt,
		UpdatedAt: now,
	}

	pc.DB.Model(&updatedPost).Updates(postToUpdate)

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": updatedPost})
}

// The `FindPostById` function is a method of the `PostController` struct that retrieves a single post
// from the database based on the post ID provided in the request parameters. It first retrieves the
// post ID from the request parameters and then uses the gorm.DB object stored in the PostController
// struct to query the database for a post with that ID. If the post is found, the function returns a
// JSON response with a status of 200 (OK) and the post data. If the post is not found, the function
// returns a JSON response with a status of 404 (Not Found) and an appropriate error message.
func (pc *PostController) FindPostById(ctx *gin.Context) {
	postId := ctx.Param("postId")

	var post models.Post
	result := pc.DB.First(&post, "id = ?", postId)
	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No post with that title exists"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "data": post})
}

// The `FindPosts` function is a method of the `PostController` struct that retrieves a list of posts
// from the database based on the provided query parameters. It first retrieves the `page` and `limit`
// query parameters from the request, and then converts them to integers using the `strconv.Atoi`
// function. It then calculates the `offset` value based on the `page` and `limit` values.
func (pc *PostController) FindPosts(ctx *gin.Context) {
	var page = ctx.DefaultQuery("page", "1")
	var limit = ctx.DefaultQuery("limit", "10")

	intPage, _ := strconv.Atoi(page)
	intLimit, _ := strconv.Atoi(limit)
	offset := (intPage - 1) * intLimit

	var posts []models.Post
	results := pc.DB.Limit(intLimit).Offset(offset).Find(&posts)
	if results.Error != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"status": "error", "message": results.Error})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success", "results": len(posts), "data": posts})
}

// The `DeletePost` function is a method of the `PostController` struct that deletes a post from the
// database based on the post ID provided in the request parameters. It first retrieves the post ID
// from the request parameters and then uses the gorm.DB object stored in the PostController struct to
// delete the post with that ID from the database. If the post is not found, the function returns a
// JSON response with a status of 404 (Not Found) and an appropriate error message. If the deletion is
// successful, the function returns a JSON response with a status of 204 (No Content).
func (pc *PostController) DeletePost(ctx *gin.Context) {
	postId := ctx.Param("postId")

	result := pc.DB.Delete(&models.Post{}, "id = ?", postId)

	if result.Error != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"status": "fail", "message": "No post with that title exists"})
		return
	}

	ctx.JSON(http.StatusNoContent, gin.H{"status": "Success", "message": "Succesfully delete a record"})
}
