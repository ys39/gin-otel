package instrumentation

import (
	"context"
	"fmt"
	"os"
	"runtime"

	"myapp/internal/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	TracerAPI      trace.Tracer
	TracerDB       trace.Tracer
	TracerCache    trace.Tracer
	TracerRenderer trace.Tracer
)

// RecordError はエラーをスパンに記録するヘルパー関数です。
// エラーの重大度を設定し、スタックトレース情報も追加します。
func RecordError(span trace.Span, err error) {
	if err == nil {
		return
	}

	span.RecordError(err, trace.WithStackTrace(true))
	span.SetStatus(codes.Error, err.Error())

	_, file, line, ok := runtime.Caller(1)
	if ok {
		span.SetAttributes(
			attribute.String("error.file", file),
			attribute.Int("error.line", line),
		)
	}
}

// InitTracerProvider は OpenTelemetry のトレーサーを初期化し、TracerProvider を返します。
func InitTracerProvider(cfg *config.Config) (*sdktrace.TracerProvider, error) {

	// ここで config を使ってスコープ名を確定し、グローバル変数に割り当て
	TracerAPI = otel.Tracer(cfg.OtelScopeAPIName)
	TracerDB = otel.Tracer(cfg.OtelScopeDBName)
	TracerCache = otel.Tracer(cfg.OtelScopeCacheName)
	TracerRenderer = otel.Tracer(cfg.OtelScopeRendererName)

	// OTLP(gRPC) クライアントを作成
	client := otlptracegrpc.NewClient(
		otlptracegrpc.WithEndpoint(cfg.OtelHost+":"+cfg.OtelPort),
		otlptracegrpc.WithInsecure(), // ローカルなので認証なし
	)

	// 例: OTLP HTTP エクスポーターを利用する
	exporter, err := otlptrace.New(context.Background(), client)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}

	// リソース属性を詳細に設定
	hostname, _ := os.Hostname()
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(cfg.OtelServiceName),
		semconv.ServiceVersionKey.String(cfg.OtelServiceVersion),
		semconv.ServiceInstanceIDKey.String(hostname),
		semconv.DeploymentEnvironmentKey.String("development"),
		semconv.TelemetrySDKNameKey.String("opentelemetry"),
		semconv.TelemetrySDKLanguageKey.String("go"),
		semconv.TelemetrySDKVersionKey.String("1.24.0"),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return tp, nil
}
