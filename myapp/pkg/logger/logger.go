package logger

import (
	"log/slog"
	"os"
)

// 使い回すためのグローバルロガーを保持する
var appLogger *slog.Logger

// InitLogger はアプリ全体で使うロガーを初期化します。
// 「JSON形式で出力するか」「ログレベルをどうするか」などをまとめて設定。
func InitLogger(isJSON bool, level slog.Level) {
	var handler slog.Handler

	if isJSON {
		// JSON形式で出力するハンドラ
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: level, // ログレベル
		})
	} else {
		// テキスト形式で出力するハンドラ
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})
	}

	// グローバルで使うロガーを作成
	appLogger = slog.New(handler)
}

// GetLogger は初期化済みのロガーを返します。
func GetLogger() *slog.Logger {
	// まだ初期化されていない場合はデフォルトを作る
	if appLogger == nil {
		InitLogger(false, slog.LevelInfo)
	}
	return appLogger
}
