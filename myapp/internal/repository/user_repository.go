package repository

import (
	"context"
	"myapp/internal/config"
	"myapp/internal/domain"
	"myapp/internal/instrumentation"
	"strconv"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type UserRepository interface {
	FindByID(ctx context.Context, id int) (*domain.User, error)
	FindDetailByID(ctx context.Context, id int) (*domain.UserDetail, error)
}

// 実際の DB アクセスを行う実装例
type userRepository struct {
	Users       map[int]domain.User
	UserDetails map[int]domain.UserDetail
}

func NewUserRepository(cfg *config.Config) UserRepository {
	return &userRepository{}
}

func (r *userRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	_, span := instrumentation.TracerDB.Start(ctx, "DB.GetUser", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		semconv.DBSystemKey.String("mock"),
		semconv.DBStatementKey.String("SELECT * FROM users WHERE id=?"),
		semconv.DBOperationKey.String("SELECT"),
		semconv.DBNameKey.String("users_db"),
		attribute.String("db.table", "users"),
		attribute.String("db.user.id", strconv.Itoa(id)),
	)
	defer span.End()

	// TODO ここで実際のDBからユーザー情報を取得する処理を実装

	// return
	return &domain.User{}, nil
}

func (r *userRepository) FindDetailByID(ctx context.Context, id int) (*domain.UserDetail, error) {
	_, span := instrumentation.TracerDB.Start(ctx, "DB.GetUserDetail", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		semconv.DBSystemKey.String("mock"),
		semconv.DBStatementKey.String("SELECT * FROM user_details WHERE id=?"),
		semconv.DBOperationKey.String("SELECT"),
		semconv.DBNameKey.String("users_db"),
		attribute.String("db.table", "user_details"),
		attribute.String("db.user.id", strconv.Itoa(id)),
	)
	defer span.End()

	// TODO ここで実際のDBからユーザー詳細情報を取得する処理を実装

	return &domain.UserDetail{}, nil
}
