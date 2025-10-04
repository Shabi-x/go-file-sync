# 使用多阶段构建
# 第一阶段：构建阶段
FROM golang:1.25-alpine AS builder

# 安装Node.js和npm，用于构建前端
RUN apk add --no-cache nodejs npm

# 设置工作目录
WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建前端
WORKDIR /app/server/frontend
RUN npm install --legacy-peer-deps
RUN npm run build

# 返回到根目录
WORKDIR /app

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 第二阶段：运行阶段
FROM alpine:latest

# 安装ca-certificates用于HTTPS请求，以及必要的浏览器依赖
RUN apk --no-cache add ca-certificates tzdata chromium

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制可执行文件
COPY --from=builder /app/main .

# 复制前端静态文件（如果有）
COPY --from=builder /app/server/frontend/dist ./static

# 暴露端口
EXPOSE 27149

# 设置环境变量
ENV DISPLAY=:99

# 运行应用
CMD ["./main"]