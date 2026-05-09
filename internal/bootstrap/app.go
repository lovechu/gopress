package bootstrap

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/yourorg/gopress/internal/config"
	"github.com/yourorg/gopress/internal/content"
	"github.com/yourorg/gopress/internal/media"
	"github.com/yourorg/gopress/internal/taxonomy"
	"github.com/yourorg/gopress/internal/theme"
	"github.com/yourorg/gopress/internal/user"
	"github.com/yourorg/gopress/pkg/cache"
	"github.com/yourorg/gopress/pkg/database"
	"github.com/yourorg/gopress/pkg/jwt"
	"github.com/yourorg/gopress/pkg/performance"
	"go.uber.org/zap"
)

type App struct {
	Router *gin.Engine
	Config *config.Config
	Perf   *performance.Monitor
	Cache  cache.Cache
}

func NewApp(cfg *config.Config) (*App, error) {
	// 初始化日志
	log, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	// 初始化性能监控
	perfMonitor := performance.NewMonitor(log)

	// 初始化 MySQL
	db, err := database.NewMySQL(database.Database{
		Driver:         cfg.Database.Driver,
		Host:           cfg.Database.Host,
		Port:           cfg.Database.Port,
		Name:           cfg.Database.Name,
		User:           cfg.Database.User,
		Password:       cfg.Database.Password,
		Charset:        cfg.Database.Charset,
		MaxOpenConns:   cfg.Database.MaxOpenConns,
		MaxIdleConns:   cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
		LogLevel:       cfg.Database.LogLevel,
	})
	if err != nil {
		return nil, err
	}

	// 自动迁移（开发环境，生产用 migration 脚本）
	if err := db.AutoMigrate(
		&user.User{},
		&taxonomy.Term{},
		&content.Post{},
		&content.PostTerm{},
		&content.PostRevision{},
		&media.Media{},
		&theme.Theme{},
	); err != nil {
		return nil, err
	}

	// 初始化 Redis 缓存
	var cacheClient cache.Cache
	redisClient, redisErr := database.NewRedis(database.Redis{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: 5,
		MaxRetries:   3,
		Prefix:       cfg.Redis.Prefix,
	})

	if redisErr != nil {
		log.Warn("Redis connection failed, caching disabled", zap.Error(redisErr))
	} else {
		cacheClient = cache.NewRedisCache(redisClient, cfg.Redis.Prefix)
		log.Info("Redis cache initialized successfully")
	}

	// 初始化 Storage
	var storage media.Storage
	storageCfg := cfg.Media.Storage
	if storageCfg == "minio" {
		minioCfg := media.MinIOConfig{
			Endpoint:  cfg.Media.MinIO.Endpoint,
			AccessKey: cfg.Media.MinIO.AccessKey,
			SecretKey: cfg.Media.MinIO.SecretKey,
			Bucket:   cfg.Media.MinIO.Bucket,
			UseSSL:   cfg.Media.MinIO.UseSSL,
			Region:   cfg.Media.MinIO.Region,
			BaseURL:  cfg.Media.MinIO.BaseURL,
		}
		minioStorage, err := media.NewMinIOStorage(minioCfg)
		if err != nil {
			return nil, fmt.Errorf("init minio: %w", err)
		}
		storage = minioStorage
	} else {
		// 默认使用本地存储
		storage = media.NewLocalStorage(cfg.Media.Local.BasePath, cfg.Media.Local.BaseURL)
	}

	// 初始化 User 模块
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, jwt.Config{
		AccessSecret:   cfg.JWT.AccessSecret,
		RefreshSecret:  cfg.JWT.RefreshSecret,
		AccessExpire:   cfg.JWT.AccessExpire,
		RefreshExpire:  cfg.JWT.RefreshExpire,
	}, log)
	userHandler := user.NewHandler(userSvc)

	// 初始化 Taxonomy 模块
	termRepo := taxonomy.NewRepository(db)
	termSvc := taxonomy.NewService(termRepo)
	termHandler := taxonomy.NewHandler(termSvc)

	// 初始化 Content 模块
	postRepo := content.NewRepository(db)
	postSvc := content.NewService(postRepo, termRepo, userRepo)
	postHandler := content.NewHandler(postSvc)

	// 初始化 Media 模块
	mediaRepo := media.NewRepository(db)
	mediaCfg := media.Config{
		MaxFileSize:   cfg.Media.MaxFileSize,
		AllowedTypes:  cfg.Media.AllowedTypes,
	}
	mediaSvc := media.NewService(mediaRepo, storage, mediaCfg, storage.GetBaseURL())
	mediaHandler := media.NewHandler(mediaSvc)

	// 初始化 Theme 模块
	themeRepo := theme.NewRepository(db)
	shortcodeRegistry := theme.NewShortcodeRegistry()
	theme.RegisterDefaultShortcodes(shortcodeRegistry)
	themeEngine := theme.NewEngine(cfg.App.BasePath+"/themes", shortcodeRegistry, log)

	// 预热模板
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := themeEngine.Prewarm(ctx); err != nil {
			log.Warn("template prewarming failed", zap.Error(err))
		}
	}()

	// 设置默认主题
	if err := themeEngine.SetActiveTheme("default"); err != nil {
		fmt.Printf("warning: failed to set default theme: %v\n", err)
	}
	themeSvc := theme.NewService(themeEngine, themeRepo)
	themeHandler := theme.NewHandler(
		themeSvc,
		postSvc,
		userRepo,
		termRepo,
		cfg.App.BasePath+"/themes",
	)

	// 启动性能指标定期报告
	go perfMonitor.StartPeriodicReport(context.Background(), 60*time.Second)

	// 注册路由
	router := NewRouter(cfg, RouterDeps{
		UserHandler:  userHandler,
		TermHandler:  termHandler,
		PostHandler:  postHandler,
		MediaHandler: mediaHandler,
		ThemeHandler: themeHandler,
		Config:       cfg,
		Log:          log,
		PerfMonitor:  perfMonitor,
	})

	return &App{Router: router, Config: cfg, Perf: perfMonitor, Cache: cacheClient}, nil
}

// LoadConfig 加载配置文件
func LoadConfig(path string) (*config.Config, error) {
	return config.Load(path)
}
