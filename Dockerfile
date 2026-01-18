# 使用Go 1.21作为基础镜像构建后端
FROM golang:1.21-alpine AS backend-builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum文件
COPY backend/go.mod backend/go.sum ./

# 下载依赖
RUN go mod download

# 复制后端源代码
COPY backend/ ./

# 构建后端应用
RUN go build -o main main.go

# 使用Node 18作为基础镜像构建前端
FROM node:18-alpine AS frontend-builder

# 设置工作目录
WORKDIR /app

# 复制前端package.json和package-lock.json
COPY frontend/package.json frontend/package-lock.json ./

# 安装依赖
RUN npm install

# 复制前端源代码
COPY frontend/ ./

# 构建前端应用
RUN npm run build

# 使用Alpine作为基础镜像，运行后端应用
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /app

# 从后端构建阶段复制编译好的二进制文件
COPY --from=backend-builder /app/main .

# 从前端构建阶段复制构建好的前端文件到dist目录
COPY --from=frontend-builder /app/dist ./dist

# 暴露端口
EXPOSE 8080

# 设置环境变量
ENV PORT=8080

# 运行后端应用
CMD ["./main"]
