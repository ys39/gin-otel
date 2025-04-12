package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"myapp/internal/config"
	"myapp/internal/handler"
	"myapp/internal/instrumentation"
	"myapp/internal/repository"
	"myapp/internal/service"
	"myapp/pkg/logger"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. 設定読み込み
	cfg := config.Load()

	// 2. ロガーを初期化（log/slogを使う例）
	//    第1引数: JSON 形式で出力するかどうか (true なら JSON)
	//    第2引数: ログレベル (Debug, Info, Warn, Errorなど)
	logger.InitLogger(false, slog.LevelDebug)
	appLogger := logger.GetLogger()

	// 2. OTel 初期化
	tp, err := instrumentation.InitTracerProvider(cfg)
	if err != nil {
		log.Fatalf("failed to initialize tracer provider: %v", err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Printf("Error shutting down tracer provider: %v", err)
		}
	}()

	// 3. Gin のセットアップ
	r := gin.Default()

	// 4. リポジトリ / サービス初期化
	// userRepo := repository.NewUserRepository(cfg) // 例: 実際のDBへ接続する実装
	userRepo := repository.NewMockUserRepository()  // モック実装
	userService := service.NewUserService(userRepo) // DI

	// 5. Handler 登録
	handler.NewUserHandler(r, userService)

	// 6. サーバーを起動 (cfg.AppPort を使用)
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	appLogger.Info("Starting server", "port", cfg.AppPort)
	if err := r.Run(addr); err != nil {
		appLogger.Error("Server run failed", "error", err)
	}
}
