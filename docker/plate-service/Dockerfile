FROM golang:1.25-alpine AS build

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.22

# Сертификаты — для HTTPS к площадкам, tzdata — чтобы TZ в .env влиял на крон
RUN apk add --no-cache ca-certificates tzdata && adduser -D -H app

WORKDIR /app
COPY --from=build /out/ ./
COPY migrations ./migrations

USER app
CMD ["/app/api"]
