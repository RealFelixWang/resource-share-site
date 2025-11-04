# 多阶段构建 Dockerfile for Resource Share Site
FROM golang:1.25.3-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# 运行阶段
FROM alpine:latest

# 安装 ca-certificates
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .
COPY --from=builder /app/web ./web
COPY --from=builder /app/config ./config

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV GIN_MODE=release

# 启动应用
CMD ["./main"]
