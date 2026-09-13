ARG GO_VERSION=1.26
ARG AIR_VERSION=v1.67.4

# ---- dev: ホットリロード用（docker-compose の target: dev） ----
FROM golang:${GO_VERSION}-alpine AS dev
ARG AIR_VERSION
WORKDIR /app
RUN apk add --no-cache git \
 && go install github.com/air-verse/air@${AIR_VERSION}
COPY go.mod go.sum* ./
RUN go mod download
CMD ["air", "-c", ".air.toml"]

# ---- build: 本番バイナリ作成 ----
FROM golang:${GO_VERSION}-alpine AS build
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/api ./cmd/api

# ---- prod: 実行イメージ ----
FROM gcr.io/distroless/static-debian12:nonroot AS prod
WORKDIR /
COPY --from=build /out/api /api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
