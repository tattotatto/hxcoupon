# --- Stage 1: Build Go backend ---
FROM golang:1.22-alpine AS go-builder

ENV GOPROXY https://goproxy.cn,direct

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -mod=mod -o /server ./cmd/server

# --- Stage 2: Build React frontend ---
FROM node:20-alpine AS web-builder

WORKDIR /web

COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

# --- Stage 3: Final image ---
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

COPY --from=go-builder /server /server
COPY --from=web-builder /web/dist /static
COPY config/config.yaml /config/config.yaml

EXPOSE 8080

CMD ["/server", "--config", "/config/config.yaml"]
