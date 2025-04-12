package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"myapp/internal/instrumentation"
	"myapp/internal/service"
)

// UserHandler はユーザー関連の処理を行う Handler をまとめた構造体です。
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler はユーザー用のエンドポイントを Gin のルーターに紐付けます。
func NewUserHandler(r *gin.Engine, us service.UserService) {
	handler := &UserHandler{
		userService: us,
	}
	r.GET("/users/:id", handler.GetUser)
}

// GetUser は /users/:id に対して、指定されたユーザーを取得して返却するハンドラです。
func (h *UserHandler) GetUser(c *gin.Context) {

	// OpenTelemetry のスパンを開始
	ctx, span := instrumentation.TracerAPI.Start(c.Request.Context(), "GetUser", oteltrace.WithSpanKind(oteltrace.SpanKindServer))
	defer span.End()

	// HTTPリクエスト関連の属性を設定
	span.SetAttributes(
		semconv.HTTPMethodKey.String(c.Request.Method),
		semconv.HTTPURLKey.String(c.Request.URL.String()),
		semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
		semconv.HTTPRequestContentLengthKey.Int64(c.Request.ContentLength),
		attribute.String("http.scheme", c.Request.URL.Scheme),
		attribute.String("http.host", c.Request.Host),
		attribute.String("http.target", c.Request.URL.Path),
	)

	// パスパラメータから ID を取得
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		instrumentation.RecordError(span, err)
		JSON(c, ctx, http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// スパンにユーザー ID をセット
	span.SetAttributes(
		attribute.String("user.id", strconv.Itoa(id)),
		attribute.String("endpoint.path", "/users/:id"),
	)

	// Service レイヤーを通じてユーザーを取得
	user, err := h.userService.GetUserByID(ctx, id)
	if err != nil {
		instrumentation.RecordError(span, err)
		JSON(c, ctx, http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// ユーザー詳細情報を取得
	userDetail, err := h.userService.GetUserDetailByID(ctx, id)
	if err != nil {
		instrumentation.RecordError(span, err)
		JSON(c, ctx, http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	result := gin.H{
		"user":        user,
		"user_detail": userDetail,
	}

	// 正常時は 200 OK とともにユーザー情報を JSON で返す
	JSON(c, ctx, http.StatusOK, result)
}
