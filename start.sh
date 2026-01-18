#!/bin/bash

echo "===================================="
echo "地图位置共享应用 - 启动脚本"
echo "===================================="
echo ""

echo "[1/3] 检查Redis服务..."
if ! redis-cli ping > /dev/null 2>&1; then
    echo "Redis服务未运行，请先启动Redis服务"
    echo "可以运行: redis-server"
    exit 1
fi
echo "Redis服务运行正常"
echo ""

echo "[2/3] 启动后端服务..."
cd backend
go run main.go &
BACKEND_PID=$!
cd ..
sleep 3
echo "后端服务已启动 (http://localhost:8080)"
echo ""

echo "[3/3] 启动前端服务..."
cd frontend
npm run dev &
FRONTEND_PID=$!
cd ..
sleep 3
echo "前端服务已启动 (http://localhost:3000)"
echo ""

echo "===================================="
echo "所有服务已启动！"
echo "前端地址: http://localhost:3000"
echo "后端地址: http://localhost:8080"
echo "===================================="
echo ""
echo "按 Ctrl+C 停止所有服务"

trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT TERM

wait
