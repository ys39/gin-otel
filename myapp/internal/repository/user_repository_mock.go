package repository

import (
	"context"
	"errors"
	"myapp/internal/config"
	"myapp/internal/domain"
	"myapp/internal/instrumentation"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

type mockUserRepository struct {
	cfg         *config.Config
	Users       map[int]domain.User
	UserDetails map[int]domain.UserDetail
}

func NewMockUserRepository(cfg *config.Config) UserRepository {

	mockUsers := map[int]domain.User{
		1: {ID: "sample123", Name: "otelgin tester"},
		2: {ID: "sample456", Name: "another user"},
		3: {ID: "sample789", Name: "test user"},
	}
	mockUserDetails := map[int]domain.UserDetail{
		1: {ID: "sample123", Age: 30, Mail: "test1@example.com"},
		2: {ID: "sample456", Age: 31, Mail: "test2@example.com"},
		3: {ID: "sample789", Age: 32, Mail: "test3@example.com"},
	}

	return &mockUserRepository{
		cfg:         cfg,
		Users:       mockUsers,
		UserDetails: mockUserDetails,
	}
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	_, span := instrumentation.TracerDB.Start(ctx, "MockDB.GetUser", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		semconv.DBSystemKey.String(m.cfg.DBMockSystem),
		semconv.DBStatementKey.String("SELECT * FROM users WHERE id=?"),
		semconv.DBOperationKey.String("SELECT"),
		semconv.DBNameKey.String(m.cfg.DBName),
		semconv.DBSQLTableKey.String("users"),
		attribute.String("db.users.id", strconv.Itoa(id)),
	)
	defer span.End()

	// 適当にスリープを入れてDBからの読み取りをシミュレート
	time.Sleep(time.Millisecond)

	user, ok := m.Users[id]
	if !ok {
		err := errors.New("user not found")
		instrumentation.RecordError(span, err)
		return nil, err
	}
	return &user, nil
}

func (m *mockUserRepository) FindDetailByID(ctx context.Context, id int) (*domain.UserDetail, error) {
	_, span := instrumentation.TracerDB.Start(ctx, "MockDB.GetUserDetail", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		semconv.DBSystemKey.String(m.cfg.DBMockSystem),
		semconv.DBStatementKey.String("SELECT * FROM user_details WHERE id=?"),
		semconv.DBOperationKey.String("SELECT"),
		semconv.DBNameKey.String(m.cfg.DBName),
		semconv.DBSQLTableKey.String("user_details"),
		attribute.String("db.user.id", strconv.Itoa(id)),
	)
	defer span.End()

	// 適当にスリープを入れてDBからの読み取りをシミュレート
	time.Sleep(time.Millisecond)

	userDetail, ok := m.UserDetails[id]
	if !ok {
		err := errors.New("user detail not found")
		instrumentation.RecordError(span, err)
		return nil, err
	}
	return &userDetail, nil
}
