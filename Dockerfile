# 多阶段构建 - 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的包
RUN apk add --no-cache git ca-certificates tzdata

# 复制 go mod 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o nofx main.go

# 运行阶段
FROM alpine:latest

# 安装必要的包（包括wget用于健康检查）
RUN apk --no-cache add ca-certificates tzdata wget

# 创建非root用户
RUN addgroup -g 1001 -S nofx && \
    adduser -S nofx -u 1001

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/nofx .

# 复制配置文件（如果存在）
COPY --from=builder /app/config ./config/

# 设置文件权限
RUN chown -R nofx:nofx /app
USER nofx

# 暴露端口（Railway会动态分配）
EXPOSE $PORT

# 健康检查（使用动态端口）
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:${PORT:-8080}/health || exit 1

# 启动命令
CMD ["./nofx"]