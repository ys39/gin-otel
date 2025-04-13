# Gin + OpenTelemetry サンプルプロジェクト

このプロジェクトは、Go言語のWebフレームワーク「Gin」とOpenTelemetryを連携させた分散トレーシングの実装サンプルです。クリーンアーキテクチャに近い構造で、各レイヤー（ハンドラー、サービス、リポジトリ）でのトレーシングの実装方法を示しています。

## 技術スタック

- Go
- [Gin](https://github.com/gin-gonic/gin) - Webフレームワーク
- [OpenTelemetry](https://opentelemetry.io/) - 分散トレーシングライブラリ
- [Jaeger](https://www.jaegertracing.io/) - トレーシングバックエンド
- [log/slog](https://pkg.go.dev/log/slog) - 構造化ロギング

## プロジェクト構造

```
gin-otel/
├── myapp/                  # アプリケーション本体
│   ├── cmd/                # エントリーポイント
│   │   └── server/         # サーバー起動コード
│   ├── internal/           # 内部パッケージ
│   │   ├── config/         # 設定
│   │   ├── domain/         # ドメインモデル
│   │   ├── handler/        # HTTPハンドラー
│   │   ├── instrumentation/# トレーシング関連
│   │   ├── repository/     # データアクセス層
│   │   └── service/        # ビジネスロジック層
│   └── pkg/                # 外部公開可能なパッケージ
│       └── logger/         # ロギング
└── tracer/                 # トレーシングバックエンド関連
    └── docker.sh           # Jaeger起動スクリプト
```

## セットアップ

### 前提条件

- Go 1.21以上
- Docker

### Jaegerの起動

トレーシングデータを収集・可視化するためのJaegerを起動します：

```bash
cd tracer
chmod +x docker.sh
./docker.sh
```

これにより、以下のポートでJaegerが起動します：
- 16686: Jaeger UI
- 4317: OTLP gRPC
- 4318: OTLP HTTP
- 5778: Agent configs
- 9411: Zipkin collector

### アプリケーションの実行

```bash
cd myapp
go run cmd/server/main.go
```

デフォルトでは、アプリケーションは8080ポートで起動します。

## 使用方法

### APIエンドポイント

- `GET /users/:id` - 指定されたIDのユーザー情報を取得

例：
```bash
curl http://localhost:8080/users/1
```

レスポンス例：
```json
{
  "user": {
    "id": "sample123",
    "name": "otelgin tester"
  },
  "user_detail": {
    "id": "sample123",
    "age": 30,
    "mail": "test1@example.com"
  }
}
```

## トレーシングの確認

1. Jaeger UIにアクセス: http://localhost:16686
2. Service: "myapp-api" を選択
3. Find Traces ボタンをクリック
4. トレース一覧から詳細を確認したいトレースをクリック

## 実装の解説

### トレーシングの初期化 (instrumentation/trace.go)

```go
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
		semconv.DeploymentEnvironmentKey.String(cfg.OtelDeploymentEnv),
		semconv.TelemetrySDKNameKey.String(cfg.OtelSDKName),
		semconv.TelemetrySDKLanguageKey.String(cfg.OtelSDKLanguage),
		semconv.TelemetrySDKVersionKey.String(cfg.OtelSDKVersion),
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
```

### ハンドラーでのトレーシング (handler/user_handler.go)

```go
func (h *UserHandler) GetUser(c *gin.Context) {
	// OpenTelemetry のスパンを開始
	ctx, span := instrumentation.TracerAPI.Start(c.Request.Context(), "GetUser", oteltrace.WithSpanKind(oteltrace.SpanKindServer))
	defer span.End()
    
	// スキーム情報を正確に取得
	scheme := c.Request.Header.Get("X-Forwarded-Proto")
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	// HTTPリクエスト関連の属性を設定
	span.SetAttributes(
		semconv.HTTPMethodKey.String(c.Request.Method),
		semconv.URLFullKey.String(scheme+"://"+c.Request.Host+c.Request.URL.Path),
		semconv.URLPathKey.String(c.Request.URL.Path),
		semconv.HTTPUserAgentKey.String(c.Request.UserAgent()),
		semconv.HTTPRequestContentLengthKey.Int64(c.Request.ContentLength),
		semconv.URLSchemeKey.String(scheme),
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
		semconv.HTTPRouteKey.String("/users/:id"),
	)

    // サービス呼び出し
    user, err := h.userService.GetUserByID(ctx, id)
    // ...
}
```

### リポジトリでのトレーシング (repository/user_repository_mock.go)

```go
func (m *mockUserRepository) FindByID(ctx context.Context, id int) (*domain.User, error) {
    // DBアクセス用のスパンを開始
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
    
    // DBアクセス処理
    // ...
    
    // エラー発生時はスパンにエラーを記録
	if !ok {
		err := errors.New("user not found")
		instrumentation.RecordError(span, err)
		return nil, err
	}
    
    return &user, nil
}
```

### レンダリングでのトレーシング (handler/render.go)

```go
func JSON(c *gin.Context, ctx context.Context, code int, obj interface{}) {
    // レンダリング用のスパンを開始
	_, span := instrumentation.TracerRenderer.Start(ctx, "gin.renderer.json", oteltrace.WithSpanKind(oteltrace.SpanKindInternal))
	span.SetAttributes(
		semconv.HTTPResponseStatusCodeKey.Int(code),
		attribute.String("response.type", "json"),
		attribute.String("response.content_type", "application/json"),
	)
	defer span.End()
    
    // エラーレスポンスの場合はスパンに記録
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

    
    // レスポンス送信
    c.JSON(code, obj)
}
```

## 環境変数

アプリケーションは以下の環境変数で設定可能です：

| 環境変数 | デフォルト値 | 説明 |
|----------|--------------|------|
| SERVICE_NAME | myapp | アプリケーション名 |
| SERVICE_VERSION | 0.0.1 | アプリケーションバージョン |
| APP_PORT | 8080 | アプリケーションポート |
| OTEL_HOST | localhost | OpenTelemetryコレクターのホスト |
| OTEL_PORT | 4317 | OpenTelemetryコレクターのポート(gRPC) |
| OTEL_SERVICE_NAME | myapp-api | トレーシングに使用するサービス名 |
| OTEL_SERVICE_VERSION | 0.0.1 | トレーシングに使用するサービスバージョン |

## ライセンス

MIT
