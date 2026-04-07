package main

import (
	"Forum/internal/config"
	"Forum/internal/domain"
	"Forum/internal/handler"
	"Forum/internal/metrics"
	"Forum/internal/middleware"
	"Forum/internal/pkg/cache"
	"Forum/internal/repository"
	"Forum/internal/search"
	"Forum/internal/service"
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/opensearch-project/opensearch-go/v4/opensearchapi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()

	wd, _ := os.Getwd()
	log.Println("working dir:", wd)

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

	// elastic search
	es, err := search.NewESClient(cfg)
	if err != nil {
		log.Fatal("Error initializing elasticsearch:", err)
	}

	// create elastic search index
	createEsIndex(es)

	// for migrate data
	runBackfill(ctx, es, db)

	fmt.Println("Connected to database successfully")

	initDBMigrationAndIndex(db)

	// init layers
	userRepo := repository.NewUserRepo(db)
	postRepo := repository.NewPostRepo(db)
	commentRepo := repository.NewCommentRepo(db)
	likeRepo := repository.NewLikeRepo(db)
	likeCache := cache.NewLikeCache(rdb)
	followRepo := repository.NewFollowRepo(db)
	feedRepo := repository.NewFeedRepo(db)
	searchRepo := repository.NewSearchRepo(es)
	notificationRepo := repository.NewNotificationRepo(db)

	postIndexer := search.NewPostIndexer(es)

	notificationService := service.NewNotificationService(notificationRepo)
	userService := service.NewUserService(userRepo, cfg.JWTKey)
	postService := service.NewPostService(postRepo, userRepo, postIndexer)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo, notificationService)
	likeService := service.NewLikeService(likeRepo, postRepo, commentRepo, likeCache)
	followService := service.NewFollowService(followRepo)
	feedService := service.NewFeedService(feedRepo)
	searchService := service.NewSearchService(searchRepo)

	userHandler := handler.NewUserHandler(userService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)
	likeHandler := handler.NewLikeHandler(likeService)
	followHandler := handler.NewFollowHandler(followService)
	feedHandler := handler.NewFeedHandler(feedService)
	searchHandler := handler.NewSearchHandler(searchService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// rate limiter middleware
	rateLimiterRepo := repository.NewRedisRateLimiter(rdb)
	rateLimiterService := service.NewRateLimitService(rateLimiterRepo)

	// set router
	r := gin.Default()

	// prometheus
	metrics.MustRegister()
	r.Use(middleware.PrometheusMiddleware())
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// public routes, don't need login
	r.POST("/api/v1/register", userHandler.Register)
	r.POST("/api/v1/login", userHandler.Login)
	r.GET("/api/v1/posts", postHandler.ListPosts)
	r.GET("/api/v1/posts/:id", postHandler.GetPostByID)
	r.GET("/api/v1/posts/:id/comments", commentHandler.ListComments)
	r.GET("/api/v1/users/:id/following", followHandler.ListFollowing)
	r.GET("/api/v1/users/:id/followers", followHandler.ListFollowers)
	r.GET("/api/v1/search/posts", searchHandler.SearchPosts)

	// protected routes
	auth := r.Group("api/v1")
	auth.Use(middleware.AuthMiddleware(cfg.JWTKey))
	{
		auth.GET("/profile", userHandler.GetProfile)

		// post routes
		auth.POST("/posts", middleware.RateLimitMiddleware(rateLimiterService, domain.RuleCreatePost), postHandler.CreatePost) // use rate limiter
		auth.PUT("/posts/:id", postHandler.UpdatePost)
		auth.DELETE("/posts/:id", postHandler.DeletePost)

		// comment routes
		auth.POST("/posts/:id/comments", middleware.RateLimitMiddleware(rateLimiterService, domain.RuleCreateComment), commentHandler.CreateComment) // user rate limiter
		auth.DELETE("/comments/:id", commentHandler.DeleteComment)

		// like routes
		auth.POST("/likes/toggle", likeHandler.Toggle)
		auth.GET("/likes/:type/:id", likeHandler.GetStatus)

		// follow routes
		auth.POST("/users/:id/follow/toggle", followHandler.Toggle)
		auth.GET("/users/:id/follow/status", followHandler.GetStatus)

		// feed routes
		auth.GET("/feed", feedHandler.ListHomeFeed)

		// notification routes
		auth.GET("/notification", notificationHandler.List)
		auth.GET("/notification/unread_count", notificationHandler.UnreadCount)
		auth.POST("/notification/:id/read", notificationHandler.MarkRead)
		auth.POST("/notification/read_all", notificationHandler.MarkAllRead)
	}

	// launch server
	log.Printf("Server started on port %s", cfg.ServerPort)
	r.Run(":" + cfg.ServerPort)
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
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

	addr := rdb.Options().Addr
	db := rdb.Options().DB
	log.Printf("redis addr=%s db=%d", addr, db)
	return rdb, nil
}

func initDBMigrationAndIndex(db *gorm.DB) {
	// auto migrate models
	_ = db.AutoMigrate(&domain.User{}, &domain.Post{}, &domain.Comment{}, &domain.Like{}, &domain.Follow{}, domain.Notification{})

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

	// create search full text search tsvector and index

	db.Exec(`
		ALTER TABLE posts
		ADD COLUMN IF NOT EXISTS search_tsv tsvector
		GENERATED ALWAYS AS (
		  to_tsvector('simple', coalesce(title,'') || ' ' || coalesce(content,''))
		) STORED;
	`)

	db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_posts_search_tsv
		ON posts USING GIN (search_tsv);
	`)

	db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_notif_user_created
		ON notifications (user_id, created_at DESC)
		WHERE deleted_at IS NULL;
	`)
}

func createEsIndex(es *opensearchapi.Client) {
	if err := search.EnsurePostsIndex(es); err != nil {
		log.Fatal(err)
	}
	log.Println("create index succeeded")
}

func runBackfill(ctx context.Context, es *opensearchapi.Client, db *gorm.DB) {
	bi, err := search.NewPostBulkIndexer(es)
	if err != nil {
		log.Fatal(err)
	}

	runner := &search.BackfillRunner{
		Db: db,
		Bi: bi,
	}

	if err := runner.Run(ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("backfill done")
}
