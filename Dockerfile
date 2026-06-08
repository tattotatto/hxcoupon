FROM golang:1.22-alpine AS builder

ENV GOPROXY https://goproxy.cn,direct

WORKDIR /app

COPY go.mod ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /server ./cmd/server

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

COPY --from=builder /server /server
COPY config/config.yaml /config/config.yaml

EXPOSE 8080

CMD ["/server", "--config", "/config/config.yaml"]
