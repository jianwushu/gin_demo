package bootstrap

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	docs "gin-demo/docs"
	appconfig "gin-demo/internal/config"
	"gin-demo/internal/handler"
	"gin-demo/internal/middleware"
	"gin-demo/internal/model"
	"gin-demo/internal/repository"
	"gin-demo/internal/service"
	pkglogger "gin-demo/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	_ "modernc.org/sqlite"
)

type App struct {
	Config *appconfig.Config
	Logger *log.Logger
	DB     *gorm.DB
	Router *gin.Engine
	Server *http.Server
}

func NewApp(configPath string) (*App, error) {
	cfg, err := appconfig.Load(configPath)
	if err != nil {
		return nil, err
	}

	logger, err := pkglogger.New(pkglogger.Config{
		Dir:           cfg.Log.Dir,
		Filename:      cfg.Log.Filename,
		RotateBy:      cfg.Log.RotateBy,
		RetentionDays: cfg.Log.RetentionDays,
	})
	if err != nil {
		return nil, err
	}

	db, err := newDatabase(cfg)
	if err != nil {
		return nil, err
	}

	router := gin.New()
	router.Use(gin.Recovery(), middleware.Audit(logger))
	registerRoutes(router, db, cfg)

	server := &http.Server{
		Addr:         cfg.Server.Address(),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	return &App{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Router: router,
		Server: server,
	}, nil
}

func newDatabase(cfg *appconfig.Config) (*gorm.DB, error) {
	if cfg.Database.Driver != "sqlite" {
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	if err := os.MkdirAll(filepath.Dir(cfg.Database.DSN), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Dialector{
		DriverName: "sqlite",
		DSN:        cfg.Database.DSN,
	}, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if cfg.Database.AutoMigrate {
		if err := db.AutoMigrate(&model.User{}); err != nil {
			return nil, fmt.Errorf("auto migrate database: %w", err)
		}
	}

	return db, nil
}

func registerRoutes(router *gin.Engine, db *gorm.DB, cfg *appconfig.Config) {
	router.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if cfg.Swagger.Enabled && cfg.App.Env == "dev" {
		docs.SwaggerInfo.BasePath = "/"
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	api := router.Group("/api/v1")
	userHandler.RegisterRoutes(api)
}
