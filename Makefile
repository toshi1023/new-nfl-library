GOBIN ?= $(shell go env GOPATH)/bin

# 開発ツールのパス。PATH に通っている別のものを使いたい場合は `make lint GOLANGCI_LINT=golangci-lint` のように上書きする。
AIR           ?= $(GOBIN)/air
SQLC          ?= $(GOBIN)/sqlc
OAPI_CODEGEN  ?= $(GOBIN)/oapi-codegen
MIGRATE       ?= $(GOBIN)/migrate
GOLANGCI_LINT ?= $(GOBIN)/golangci-lint

DATABASE_URL ?= postgres://app:app@localhost:5432/app?sslmode=disable
MIGRATIONS_DIR := db/migrations

.PHONY: help up down logs run dev build test test-integration lint fmt sqlc gen \
        migrate-up migrate-down migrate-new tidy tools scrape-rosters

help: ## ターゲット一覧を表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

up: ## docker compose で API + Postgres を起動（ホットリロード）
	docker compose up -d --build

down: ## docker compose を停止
	docker compose down

logs: ## API コンテナのログを追う
	docker compose logs -f api

run: ## ローカルで API を起動（.env を読み込む）
	@set -a; [ -f .env ] && . ./.env; set +a; go run ./cmd/api

dev: ## ローカルで air によるホットリロード起動（.env を読み込む）
	@set -a; [ -f .env ] && . ./.env; set +a; $(AIR) -c .air.toml

build: ## バイナリをビルド
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/api ./cmd/api
	CGO_ENABLED=0 go build -ldflags="-s -w" -o bin/scraperosters ./cmd/scraperosters

scrape-rosters: ## ロスターをスクレイプして outputs/csv/ に CSV 出力 (make scrape-rosters season=2024 [teams=buf,nwe] [out=x.csv])
	@test -n "$(season)" || (echo "usage: make scrape-rosters season=2024 [teams=buf,nwe] [out=x.csv]" && exit 1)
	go run ./cmd/scraperosters -season $(season) $(if $(teams),-teams $(teams)) $(if $(out),-out $(out))

test: ## ユニットテスト
	go test -race -cover ./...

test-integration: ## 統合テスト（docker compose の Postgres が必要）
	DATABASE_URL=$(DATABASE_URL) go test -race -tags integration ./...

lint: ## golangci-lint
	$(GOLANGCI_LINT) run ./...

fmt: ## golangci-lint によるフォーマット
	$(GOLANGCI_LINT) fmt ./...

sqlc: ## sqlc でコード生成
	$(SQLC) generate

gen: ## oapi-codegen で OpenAPI から型を生成
	$(OAPI_CODEGEN) -config api/oapi-codegen.yaml api/openapi.yaml

migrate-up: ## マイグレーション適用
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down: ## マイグレーションを 1 つ戻す
	$(MIGRATE) -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-new: ## マイグレーションファイル作成 (make migrate-new name=xxx)
	@test -n "$(name)" || (echo "usage: make migrate-new name=xxx" && exit 1)
	$(MIGRATE) create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

tidy: ## go mod tidy
	go mod tidy

tools: ## 開発ツールを $(GOBIN) にインストール
	GOBIN=$(GOBIN) go install github.com/air-verse/air@latest
	GOBIN=$(GOBIN) go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	GOBIN=$(GOBIN) go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
	GOBIN=$(GOBIN) go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	GOBIN=$(GOBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
