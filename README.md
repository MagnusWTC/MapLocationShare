# 地图位置共享应用

一个基于React、Go和WebSocket的实时地图位置共享应用。

## 功能特性

- 实时位置共享
- 地图自动缩放和定位
- 支持分享链接邀请他人查看位置
- 位置精度优化
- 前后端统一部署在一个端口
- 基于内存的24小时滑动过期存储

## 技术栈

- **前端**: React + Vite
- **后端**: Go + Gin + WebSocket
- **地图**: 高德地图API
- **存储**: 内存存储（24小时滑动过期）
- **部署**: Docker容器化

## 快速开始

### 本地开发

#### 前端开发

```bash
cd frontend
npm install
npm run dev
```

#### 后端开发

```bash
cd backend
go run main.go
```

### Docker部署

#### 构建镜像

```bash
docker build -t map-location-share .
```

#### 运行容器

```bash
docker run -p 8080:8080 map-location-share
```

应用将在 http://localhost:8080 启动

## 环境变量

### 前端

创建 `.env` 文件在 `frontend` 目录下：

```
VITE_AMAP_KEY=your_amap_key
```

### 后端

创建 `.env` 文件在 `backend` 目录下：

```
PORT=8080
```

## 项目结构

```
.
├── backend/          # 后端代码
│   ├── handlers/     # HTTP处理函数
│   ├── models/       # 数据模型
│   ├── storage/      # 存储实现
│   ├── websocket/    # WebSocket处理
│   └── main.go       # 入口文件
├── frontend/         # 前端代码
│   ├── src/
│   │   ├── components/  # React组件
│   │   ├── services/    # API和WebSocket服务
│   │   └── App.jsx      # 主应用组件
│   └── index.html       # HTML入口
├── Dockerfile        # Docker构建文件
└── README.md         # 项目说明
```

## 核心功能实现

### 位置共享

1. 用户通过浏览器获取地理位置
2. 位置信息通过WebSocket实时发送到后端
3. 后端广播位置信息给同一房间的所有用户
4. 其他用户在地图上实时看到位置更新

### 存储机制

- 使用内存存储，数据自动24小时滑动过期
- 定期清理过期数据，优化内存使用

### 地图功能

- 自动定位到当前用户位置
- 支持手动缩放和拖动，不会自动复位
- 显示所有共享用户的位置标记

## 许可证

MIT
