FROM golang:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o o2wa .

FROM ubuntu:latest

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/o2wa /app/o2wa
COPY --from=builder /app/config.example.yaml /app/config.example.yaml

EXPOSE 8080

CMD ["/app/o2wa"]