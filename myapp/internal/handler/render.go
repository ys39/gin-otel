package handler

import (
	"context"
	"fmt"
	"myapp/internal/instrumentation"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// JSON は Gin の c.JSON をラップし、OTel のスパンを計測する。
func JSON(c *gin.Context, ctx context.Context, code int, obj interface{}) {
	_, span := instrumentation.TracerRenderer.Start(ctx, "gin.renderer.json", oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	span.SetAttributes(
		attribute.String("response.type", "json"),
		attribute.String("response.code", fmt.Sprintf("%d", code)),
	)
	defer span.End()

	// 例：動作確認のため少しスリープ
	time.Sleep(time.Millisecond)

	// HTTPステータスコードが200以外の場合は、エラー内容をスパンにセット
	if code != 200 {
		span.SetAttributes(
			attribute.String("error", fmt.Sprintf("%v", obj)),
		)
	}

	// Gin 標準の JSON レスポンス
	c.JSON(code, obj)
}
