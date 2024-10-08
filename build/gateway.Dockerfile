FROM golang:latest

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

# Update path to desired entrypoint
COPY cmd/gateway/gateway.go ./main.go
COPY pkg/ ./pkg/
COPY internal/ ./internal/

ENTRYPOINT ["go", "run", "main.go"]