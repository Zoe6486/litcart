package main

import (
	"context"
	"fmt"
	"litcart/controller"
	"litcart/dao/mysql"
	"litcart/dao/redis"
	"litcart/logger"
	"litcart/pkg/snowflake"
	"litcart/router"
	"litcart/setting"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

func main() {

	//configPath := "./conf/config.yaml" // 本地 fallback，生产无文件
	configPath := "./conf/dev.yaml"
	if len(os.Args) >= 2 {
		configPath = os.Args[1]
	}

	if err := setting.Init(configPath); err != nil {
		fmt.Printf("Fatal: config init failed: %v\n", err)
		os.Exit(1)
	}

	if err := logger.Init(setting.Conf.LogConfig, setting.Conf.Mode); err != nil {
		fmt.Printf("Fatal: logger init failed: %v\n", err)
		os.Exit(1)
	}
	// defer zap.L().Sync()
	defer func() {
		_ = zap.L().Sync()
	}()

	// 加关键配置校验（防止 nil/空密码导致崩溃）
	if setting.Conf.MySQLConfig.Password == "" {
		zap.L().Warn("Missing required env: MYSQL_PASSWORD (or equivalent)")
	}
	if setting.Conf.RedisConfig.Password == "" && setting.Conf.RedisConfig.Host != "localhost" { // 根据需要
		zap.L().Warn("Missing Redis password for production-like env")
	}
	if setting.Conf.RedisConfig.Host == "" || setting.Conf.RedisConfig.Host == "localhost" {
		zap.L().Warn("Redis host is empty or localhost, may not connect in production")
	}
	if setting.Conf.Port == 0 {
		zap.L().Warn("Port not set, check config or env")
	}

	zap.L().Info("Config loaded (env-first)",
		zap.String("source", "defaults + env"),
		zap.String("mode", setting.Conf.Mode),
		// zap.String("log_level", setting.Conf.LogConfig.Level),
		zap.String("log_level", setting.Conf.Level),
	)

	//
	// 初始化 snowflake（你已经硬编码了时间和 machine id，也可从 Conf 读）
	if err := snowflake.Init("2026-01-01", 1); err != nil {
		zap.L().Fatal("snowflake init failed", zap.Error(err))
	}

	// 初始化自定义 validator
	if err := controller.InitValidator(); err != nil {
		zap.L().Fatal("validator init failed", zap.Error(err))
	}

	// 初始化 MySQL
	db, err := mysql.Init(setting.Conf.MySQLConfig)
	if err != nil {
		zap.L().Fatal("mysql init failed", zap.Error(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			zap.L().Error("mysql db close failed", zap.Error(err))
		}
	}()

	// 初始化 Redis
	if err := redis.Init(setting.Conf.RedisConfig); err != nil {
		zap.L().Fatal("redis init failed", zap.Error(err))
	}
	defer redis.Close()

	// 注册路由
	r := router.Setup(db)

	// // 启动 HTTP 服务 + 优雅关闭
	// srv := &http.Server{
	// 	Addr:    fmt.Sprintf(":%d", setting.Conf.Port),
	// 	Handler: r,
	// }

	// 强制使用 Railway 的动态端口（最关键）
	// Railway 不使用你 Dockerfile 里 EXPOSE 的 8084，它会注入环境变量 PORT（随机，如 12345），你的应用必须监听 ${PORT}。
	// 在 srv 定义前加
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084" // 本地 fallback
	}
	// 新增诊断日志：方便排查 Railway 是否正确注入了 PORT
	zap.L().Info("准备监听端口",
		zap.String("resolved_port", port),
		zap.String("env_PORT_value", os.Getenv("PORT")),
		zap.Bool("using_fallback", port == "8084"),
	)
	srv := &http.Server{
		Addr:    ":" + port, // 改成 ":" + port
		Handler: r,
	}

	go func() {
		zap.L().Info("Starting server", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.L().Fatal("listen failed", zap.Error(err))
		}
	}()

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zap.L().Info("Shutdown Server ...")

	// 优雅关闭（5秒超时）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		zap.L().Fatal("Server Shutdown failed", zap.Error(err))
	}

	zap.L().Info("Server exiting")
}
