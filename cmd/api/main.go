package main

import (
	"Forum/internal/config"
	"Forum/internal/domain"
	"Forum/internal/handler"
	"Forum/internal/middleware"
	"Forum/internal/pkg/cache"
	"Forum/internal/repository"
	"Forum/internal/service"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Error loading config:", err)
	}

	// init and ping DB
	db, err := initDB(cfg)
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}

	// init and ping Redis
	rdb, err := initRedis(cfg)
	if err != nil {
		log.Fatal("Error initializing redis:", err)
	}

	fmt.Println("Connected to database successfully")

	initDBMigrationAndIndex(db)

	// init layers
	userRepo := repository.NewUserRepo(db)
	postRepo := repository.NewPostRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	likeRepo := repository.NewLikeRepo(db)
	likeCache := cache.NewLikeCache(rdb)

	userService := service.NewUserService(userRepo, cfg.JWTKey)
	postService := service.NewPostService(postRepo, userRepo)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo)
	likeService := service.NewLikeService(likeRepo, likeCache)

	userHandler := handler.NewUserHandler(userService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	likeHandler := handler.NewLikeHandler(likeService)

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

		// like routes
		auth.POST("/likes/toggle", likeHandler.Toggle)
		auth.GET("/likes/:type/:id", likeHandler.GetStatus)
	}

	// launch server
	log.Printf("Server started on port %s", cfg.ServerPort)
	r.Run(":" + cfg.ServerPort)
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Error getting sql.DB: ", err)
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(25)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Fatal("Error pinging database: ", err)
		return nil, err
	}

	return db, nil
}

func initRedis(cfg *config.Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.RedisAddr,
		Password:     cfg.RedisPassword,
		DB:           0,
		PoolSize:     20,
		MinIdleConns: 5,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := rdb.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	info, err := rdb.Info(ctx, "server").Result()
	log.Println("redis INFO server err =", err)
	log.Println(info)

	addr := rdb.Options().Addr
	db := rdb.Options().DB
	log.Printf("redis addr=%s db=%d", addr, db)
	return rdb, nil
}

func initDBMigrationAndIndex(db *gorm.DB) {
	// auto migrate models
	_ = db.AutoMigrate(&domain.User{}, &domain.Post{}, &domain.Comment{}, &domain.Like{}, &domain.Follow{})

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

	// follow table index
	db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_follow_follower_followee
		ON follows (follower_id, followee_id)
		WHERE deleted_at IS NULL;
	`)

	db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_follow_follower_created_at
		ON follows (follower_id, created_at DESC)
		WHERE deleted_at IS NULL;
	`)

	db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_follow_followee_created_at
		ON follows (followee_id, created_at DESC)
		WHERE deleted_at IS NULL;
	`)
}
