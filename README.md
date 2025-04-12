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
    // OTLPエクスポーターの設定
    client := otlptracegrpc.NewClient(
        otlptracegrpc.WithEndpoint(cfg.OtelHost+":"+cfg.OtelPort),
        otlptracegrpc.WithInsecure(),
    )
    
    exporter, err := otlptrace.New(context.Background(), client)
    if err != nil {
        return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
    }
    
    // リソース属性の設定
    resource := resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String(cfg.OtelServiceName),
        semconv.ServiceVersionKey.String(cfg.OtelServiceVersion),
    )
    
    // TracerProviderの作成
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
    // スパンの開始
    ctx, span := instrumentation.TracerAPI.Start(c.Request.Context(), "GetUser", oteltrace.WithSpanKind(oteltrace.SpanKindServer))
    defer span.End()
    
    // パラメータの取得と属性の設定
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        JSON(c, ctx, http.StatusBadRequest, gin.H{"error": "invalid user id"})
        return
    }
    
    span.SetAttributes(
        attribute.String("user.id", strconv.Itoa(id)),
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
        attribute.String("db.system", "mock"),
        attribute.String("user.id", strconv.Itoa(id)),
        attribute.Key("sql.query").String(fmt.Sprintf("SELECT * FROM users WHERE id='%s'", strconv.Itoa(id))),
    )
    defer span.End()
    
    // DBアクセス処理
    // ...
    
    // エラー発生時はスパンにエラーを記録
    if !ok {
        span.RecordError(errors.New("user not found"))
        return nil, errors.New("user not found")
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
        attribute.String("response.type", "json"),
        attribute.String("response.code", fmt.Sprintf("%d", code)),
    )
    defer span.End()
    
    // エラーレスポンスの場合はスパンに記録
    if code != 200 {
        span.SetAttributes(
            attribute.String("error", fmt.Sprintf("%v", obj)),
        )
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
