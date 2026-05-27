package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"backend/internal/blockchain"
	"backend/internal/config"
	"backend/internal/handler"
	"backend/internal/repository"
	"backend/internal/router"
	"backend/internal/scheduler"
	"backend/internal/service"
	"backend/pkg/db"
	"backend/pkg/logger"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("load config failed: %v\n", err)
		os.Exit(1)
	}

	// 2. 初始化日志
	if err := logger.Init(cfg.Logger.Level); err != nil {
		fmt.Printf("init logger failed: %v\n", err)
		os.Exit(1)
	}
	defer logger.Log.Sync()

	logger.Log.Info("server starting...")

	// 3. 初始化数据库
	if err := db.Init(cfg.Database.DSN()); err != nil {
		//logger.Log.Warn("database init failed, continuing without db", zap.Error(err))
		logger.Log.Fatal("database init failed", zap.Error(err))
	} else {
		//if err := db.DB.AutoMigrate(&model.User{}); err != nil {
		//logger.Log.Warn("auto migrate failed", zap.Error(err))
		//} else {
		logger.Log.Info("database connected")
		//}
	}

	// 4. 初始化区块链客户端（如果没有配置 RPC URL，可以跳过）
	var bcClient *blockchain.Client
	if cfg.Blockchain.RPCURL != "" {
		bcClient, err = blockchain.NewClient(cfg.Blockchain.RPCURL, cfg.Blockchain.ChainID)
		if err != nil {
			logger.Log.Warn("blockchain client init failed, continuing without it", zap.Error(err))
		} else {
			defer bcClient.Close()
			logger.Log.Info("blockchain client connected")
		}
	}

	// 5. 初始化定时任务
	var sched *scheduler.Scheduler
	if bcClient != nil {
		sched = scheduler.New(bcClient)
		sched.Start()
		defer sched.Stop()
	}

	// 6. 组装三层：repo -> service -> handler
	userRepo := repository.NewUserRepository()
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// 7. 注册路由并启动 HTTP 服务
	r := router.Setup(userHandler)

	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	// 优雅退出
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		logger.Log.Info("http server listening", zap.String("port", port))
		if err := r.Run(":" + port); err != nil {
			logger.Log.Fatal("http server error", zap.Error(err))
		}
	}()

	<-quit
	logger.Log.Info("server shutting down...")
}
