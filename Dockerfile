# multi-stage build: builder -> runtime
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates gcc musl-dev
WORKDIR /src

# copy mod first for caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG CGO_ENABLED=1
ENV CGO_ENABLED=${CGO_ENABLED}
RUN go build -ldflags="-s -w" -o /out/aiservice ./cmd/server

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/aiservice /usr/local/bin/aiservice
EXPOSE 8080
ENV SERVER_PORT=8080
# non-root user
RUN addgroup -S app && adduser -S -G app app
USER app
ENTRYPOINT ["/usr/local/bin/aiservice"]