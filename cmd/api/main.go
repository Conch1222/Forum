package main

import (
	"Forum/internal/config"
	"Forum/internal/domain"
	"Forum/internal/handler"
	"Forum/internal/middleware"
	"Forum/internal/repository"
	"Forum/internal/service"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	fmt.Println("Connected to database successfully")

	// auto migrate models
	db.AutoMigrate(&domain.User{}, &domain.Post{}, &domain.Comment{}, &domain.Like{}, &domain.Like{})

	// like table index
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_like_user_target 
        ON likes(user_id, target_type, target_id) 
        WHERE deleted_at IS NULL
	`)

	db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_like_target 
        ON likes(target_id, target_type) 
        WHERE deleted_at IS NULL
    `)

	// init layers
	userRepo := repository.NewUserRepo(db)
	postRepo := repository.NewPostRepo(db)
	commentRepo := repository.NewCommentRepo(db)

	userService := service.NewUserService(userRepo, cfg.JWTKey)
	postService := service.NewPostService(postRepo, userRepo)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo)

	userHandler := handler.NewUserHandler(userService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)

	// set router
	r := gin.Default()

	// public routes
	r.POST("/api/v1/register", userHandler.Register)
	r.POST("/api/v1/login", userHandler.Login)
	r.GET("/api/v1/posts", postHandler.ListPosts)                    // don't need login
	r.GET("/api/v1/posts/:id", postHandler.GetPostByID)              // don't need login
	r.GET("/api/v1/posts/:id/comments", commentHandler.ListComments) // don't need login

	// protected routes
	auth := r.Group("api/v1")
	auth.Use(middleware.AuthMiddleware(cfg.JWTKey))
	{
		auth.GET("/profile", userHandler.GetProfile)

		// post routes
		auth.POST("/posts", postHandler.CreatePost)
		auth.PUT("/posts/:id", postHandler.UpdatePost)
		auth.DELETE("/posts/:id", postHandler.DeletePost)

		// comment routes
		auth.POST("/posts/:id/comments", commentHandler.CreateComment)
		auth.DELETE("/comments/:id", commentHandler.DeleteComment)
	}

	// launch server
	log.Printf("Server started on port %s", cfg.ServerPort)
	r.Run(":" + cfg.ServerPort)
}
