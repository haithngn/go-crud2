package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"time"
)

type Post struct {
	ID int `json:"id" gorm:"primary key; auto_increment;"`
	Title string `json:"title" gorm:"size 255;"`
	Content string `json:"content" gorm:"size 255;"`
	CreatedAt time.Time `json:"created_at" gorm:"CURRENT_TIMESTAMP;"`
	UpdatedAt time.Time `json:"updated_at" gorm:"CURRENT_TIMESTAMP;"`
}

const API_KEY = "API_KEY"

var db *gorm.DB

func main() {
	var err error
	db, err = seeding()

	if err != nil {
		log.Fatal("Database cannot be established.")
	}

	router := gin.Default()
	router.Use(logging)

	v1 := router.Group("/v1")
	{
		v1.POST("/post", ensureAuthorized, createPost)
		v1.GET("/post/:id", readPost)
		v1.PUT("/post", ensureAuthorized, updatePost)
		v1.DELETE("/post/:id", ensureAuthorized, deletePost)
	}

	log.Fatal(router.Run(":8080"))
}

func seeding() (*gorm.DB, error) {
	// https://gorm.io/docs/connecting_to_the_database.html
	// Postgresql
	dsn := "user=developer password=golangdev dbname=news port=5432 sslmode=disable TimeZone=Asia/Ho_Chi_Minh"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	err = database.AutoMigrate(&Post{})
	if err != nil {
		return nil, err
	}

	return database, nil
}

type CreatePostParams struct {
	Title string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func createPost(context *gin.Context) {
	var params CreatePostParams
	err := context.ShouldBindJSON(&params)

	if err != nil {
		context.AbortWithStatus(http.StatusBadRequest)
		return
	}
	post := Post{
		Title:  params.Title,
		Content: params.Content,
	}
	err = db.Create(&post).Error

	if err != nil {
		context.JSON(http.StatusUnprocessableEntity, gin.H{"error": err})
		return
	}

	context.JSON(http.StatusOK, gin.H{"post": post})
	return
}

func readPost(context *gin.Context) {
	id := context.Param("id")

	post := Post{}

	err := db.First(&post,id).Error

	if err != nil {
		context.JSON(http.StatusOK, gin.H{"message": err})
		return
	}
	context.JSON(http.StatusOK, gin.H{"post": post})
}

type UpdatePostParams struct {
	ID int `json:"id,string,int" binding:"required"`
	Title string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func updatePost(context *gin.Context) {
	var params UpdatePostParams
	err := context.ShouldBindJSON(&params)
	if err != nil {
		_ = context.AbortWithError(http.StatusBadRequest, err)
		return
	}

	post := Post{
		ID:        params.ID,
		Title:     params.Title,
		Content:   params.Content,
		UpdatedAt: time.Time{},
	}

	err = db.Model(&post).Updates(&Post{
		Title:     params.Title,
		Content:   params.Content,
		UpdatedAt: time.Time{},
	}).Error

	if err != nil {
		_ = context.AbortWithError(http.StatusInsufficientStorage, err)
		return
	}

	context.JSON(http.StatusOK, gin.H{"post": post})
	return
}

func deletePost(context *gin.Context) {
	id := context.Param("id")

	var post Post

	err := db.First(&post, id).Error
	if err != nil {
		context.JSON(http.StatusOK, gin.H{"message": "This post does not exist."})
		return
	}

	err = db.Delete(&Post{}, id).Error

	if err != nil {
		context.AbortWithStatus(http.StatusUnprocessableEntity)
		return
	}

	context.JSON(http.StatusOK, gin.H{"message": "The post has been removed."})
}

type RQHeader struct {
	Authorization string
}

func ensureAuthorized(context *gin.Context) {
	var header RQHeader
	err := context.ShouldBindHeader(&header)
	if err != nil {
		context.AbortWithStatus(http.StatusBadRequest)
		return
	}

	if header.Authorization != API_KEY {
		context.AbortWithStatus(http.StatusNotAcceptable)
		return
	}
	context.Next()
}

func logging(context *gin.Context) {
	context.Next()
	log.Println(context.Request.Method, context.Request.URL)
}
