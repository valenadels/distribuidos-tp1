FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/gateway/gateway.go ./main.go
COPY pkg/ ./pkg/
COPY internal/gateway ./internal/gateway/
COPY internal/health_check/health_check_service.go ./internal/health_check/health_check_service.go

RUN CGO_ENABLED=0 GOOS=linux go build -o /main

ENTRYPOINT ["/main"]