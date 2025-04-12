package handler

import (
	"context"
	"errors"
	"fmt"
	"myapp/internal/instrumentation"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// JSON は Gin の c.JSON をラップし、OTel のスパンを計測する。
func JSON(c *gin.Context, ctx context.Context, code int, obj interface{}) {
	_, span := instrumentation.TracerRenderer.Start(ctx, "gin.renderer.json", oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	span.SetAttributes(
		semconv.HTTPResponseStatusCodeKey.Int(code),
		attribute.String("response.type", "json"),
		attribute.String("response.content_type", "application/json"),
	)
	defer span.End()

	// 例：動作確認のため少しスリープ
	time.Sleep(time.Millisecond)

	// HTTPステータスコードが200以外の場合は、エラー内容をスパンに記録
	if code >= 400 {
		var errMsg string
		if errObj, ok := obj.(gin.H); ok {
			if errStr, exists := errObj["error"]; exists {
				errMsg = fmt.Sprintf("%v", errStr)
			}
		}

		if errMsg == "" {
			errMsg = fmt.Sprintf("HTTP %d error", code)
		}

		err := errors.New(errMsg)
		instrumentation.RecordError(span, err)
	}

	// Gin 標準の JSON レスポンス
	c.JSON(code, obj)
}
