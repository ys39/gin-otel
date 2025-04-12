#!/bin/bash

# すでにコンテナが起動中なら止めて削除
docker stop jaeger 2>/dev/null && docker rm jaeger 2>/dev/null

# コンテナを起動
docker run --rm --name jaeger \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  -p 5778:5778 \
  -p 9411:9411 \
  jaegertracing/jaeger:2.4.0
