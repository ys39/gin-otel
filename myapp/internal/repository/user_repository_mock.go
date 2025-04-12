package repository

import (
	"context"
	"errors"
	"fmt"
	"myapp/internal/domain"
	"myapp/internal/instrumentation"
	"strconv"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type mockUserRepository struct {
	Users       map[int]domain.User
	UserDetails map[int]domain.UserDetail
}

func NewMockUserRepository() UserRepository {

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
		Users:       mockUsers,
		UserDetails: mockUserDetails,
	}
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
	_, span := instrumentation.TracerDB.Start(ctx, "MockDB.GetUser", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", "mock"),
		attribute.String("user.id", strconv.Itoa(id)),
		attribute.Key("sql.query").String(fmt.Sprintf("SELECT * FROM users WHERE id='%s'", strconv.Itoa(id))),
	)
	defer span.End()

	// 適当にスリープを入れてDBからの読み取りをシミュレート
	time.Sleep(time.Millisecond)

	user, ok := m.Users[id]
	if !ok {
		span.RecordError(errors.New("user not found"))
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (m *mockUserRepository) FindDetailByID(ctx context.Context, id int) (*domain.UserDetail, error) {
	_, span := instrumentation.TracerDB.Start(ctx, "MockDB.GetUserDetail", trace.WithSpanKind(trace.SpanKindClient))
	span.SetAttributes(
		attribute.String("db.system", "mock"),
		attribute.String("user.id", strconv.Itoa(id)),
		attribute.Key("sql.query").String(fmt.Sprintf("SELECT * FROM user_details WHERE id='%s'", strconv.Itoa(id))),
	)
	defer span.End()

	// 適当にスリープを入れてDBからの読み取りをシミュレート
	time.Sleep(time.Millisecond)

	userDetail, ok := m.UserDetails[id]
	if !ok {
		span.RecordError(errors.New("user detail not found"))
		return nil, errors.New("user detail not found")
	}
	return &userDetail, nil
}
