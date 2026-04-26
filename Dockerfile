FROM golang:1.25 AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app

FROM alpine:latest

WORKDIR /root/

RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/app .

COPY config.yaml .

COPY --from=builder /app/views ./views

RUN mkdir -p /var/log/app

EXPOSE 9090

CMD ["./app"]
