export const connectWebSocket = (sessionId, userId, userLocation, onLocationUpdate) => {
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsHost = window.location.host

  const wsUrl = `${wsProtocol}//${wsHost}/ws/${sessionId}`

  let ws = null
  let reconnectTimer = null
  let heartbeatTimer = null
  let isManualClose = false

  const connect = () => {
    try {
      console.log('尝试连接WebSocket:', wsUrl)
      ws = new WebSocket(wsUrl)

      ws.onopen = () => {
        console.log('WebSocket连接已建立')

        const initialMessage = {
          type: 'location_update',
          data: {
            userId,
            latitude: userLocation.latitude,
            longitude: userLocation.longitude,
            heading: userLocation.heading || 0,
            timestamp: Date.now()
          }
        }
        ws.send(JSON.stringify(initialMessage))

        heartbeatTimer = setInterval(() => {
          if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({
              type: 'ping',
              timestamp: Date.now()
            }))
          }
        }, 30000)
      }

      ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data)
          console.log('收到WebSocket消息:', message.type)

          if (message.type === 'all_locations') {
            const otherUsers = message.data.filter(loc => loc.userId !== userId)
            console.log('其他用户位置:', otherUsers)
            onLocationUpdate(otherUsers)
          } else if (message.type === 'pong') {
            console.log('收到心跳响应')
          }
        } catch (error) {
          console.error('解析WebSocket消息失败:', error)
        }
      }

      ws.onerror = (error) => {
        console.error('WebSocket错误:', error)
      }

      ws.onclose = (event) => {
        console.log('WebSocket连接已关闭:', event.code, event.reason, 'isManualClose:', isManualClose)

        if (heartbeatTimer) {
          clearInterval(heartbeatTimer)
          heartbeatTimer = null
        }

        if (!isManualClose && event.code !== 1000) {
          console.log('5秒后尝试重新连接...')
          reconnectTimer = setTimeout(() => {
            connect()
          }, 5000)
        }
      }
    } catch (error) {
      console.error('创建WebSocket连接失败:', error)
    }
  }

  connect()

  return {
    close: () => {
      console.log('手动关闭WebSocket连接')
      isManualClose = true
      if (ws) {
        ws.close(1000, 'User closed')
      }
      if (reconnectTimer) {
        clearTimeout(reconnectTimer)
      }
      if (heartbeatTimer) {
        clearInterval(heartbeatTimer)
      }
    },
    sendLocation: (latitude, longitude, heading) => {
      if (ws && ws.readyState === WebSocket.OPEN) {
        const message = {
          type: 'location_update',
          data: {
            userId,
            latitude,
            longitude,
            heading: heading || 0,
            timestamp: Date.now()
          }
        }
        console.log('发送位置更新:', message)
        ws.send(JSON.stringify(message))
      } else {
        console.warn('WebSocket未连接，无法发送位置')
      }
    }
  }
}
