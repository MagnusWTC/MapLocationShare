import { useEffect, useRef, useState } from 'react'
import { connectWebSocket } from '../services/websocket'
import './Map.css'

function Map({ userLocation, otherLocations, userId, sessionId, onLocationUpdate }) {
  const mapRef = useRef(null)
  const mapInstanceRef = useRef(null)
  const markersRef = useRef([])
  const wsRef = useRef(null)
  const [mapLoaded, setMapLoaded] = useState(false)
  const [mapInitialized, setMapInitialized] = useState(false)
  const [viewInitialized, setViewInitialized] = useState(false)
  const [userInteracted, setUserInteracted] = useState(false)

  useEffect(() => {
    if (!window.AMap) {
      console.error('高德地图API未加载')
      return
    }

    const map = new AMap.Map(mapRef.current, {
      zoom: 15,
      center: [116.397428, 39.90923],
      mapStyle: 'amap://styles/normal',
      viewMode: '2D'
    })

    mapInstanceRef.current = map

    // 添加地图交互事件监听器
    const handleUserInteraction = () => {
      setUserInteracted(true)
    }

    // 监听用户缩放、拖拽、旋转等交互事件
    map.on('zoomstart', handleUserInteraction)
    map.on('dragstart', handleUserInteraction)
    map.on('rotate', handleUserInteraction)

    map.on('complete', () => {
      setMapLoaded(true)
    })

    return () => {
      if (mapInstanceRef.current) {
        // 移除事件监听器
        map.off('zoomstart', handleUserInteraction)
        map.off('dragstart', handleUserInteraction)
        map.off('rotate', handleUserInteraction)
        mapInstanceRef.current.destroy()
      }
    }
  }, [])

  useEffect(() => {
    if (!mapLoaded || !userLocation || mapInitialized) return

    if (mapInstanceRef.current) {
      // 用户第一次进入地图时，将当前位置居中显示
      mapInstanceRef.current.setCenter([userLocation.longitude, userLocation.latitude])
      mapInstanceRef.current.setZoom(15)
      setMapInitialized(true)
      // 不在这里设置viewInitialized，让标记处理逻辑来决定是否需要调整视图
    }
  }, [userLocation, mapLoaded, mapInitialized])

  useEffect(() => {
    if (!mapLoaded) return

    markersRef.current.forEach(marker => marker.setMap(null))
    markersRef.current = []

    const allLocations = [...otherLocations]
    if (userLocation) {
      allLocations.push(userLocation)
    }

    allLocations.forEach(location => {
      const isCurrentUser = location.userId === userId
      const marker = new AMap.Marker({
        position: [location.longitude, location.latitude],
        title: isCurrentUser ? '我的位置' : '其他用户',
        icon: isCurrentUser ? createCustomIcon('#667eea', true) : createCustomIcon('#e53e3e', false),
        animation: 'AMAP_ANIMATION_DROP'
      })

      const infoWindow = new AMap.InfoWindow({
        content: `
          <div style="padding: 10px; min-width: 150px;">
            <h3 style="margin: 0 0 8px 0; font-size: 16px; color: #333;">
              ${isCurrentUser ? '我的位置' : '其他用户'}
            </h3>
            <p style="margin: 0; font-size: 14px; color: #666;">
              纬度: ${location.latitude.toFixed(6)}<br/>
              经度: ${location.longitude.toFixed(6)}
            </p>
          </div>
        `,
        offset: new AMap.Pixel(0, -30)
      })

      marker.on('click', () => {
        infoWindow.open(mapInstanceRef.current, marker.getPosition())
      })

      marker.setMap(mapInstanceRef.current)
      markersRef.current.push(marker)
    })

    if (markersRef.current.length > 0 && !userInteracted && !viewInitialized) {
      // 对于分享链接进入的用户，始终优先显示当前用户位置
      if (userLocation) {
        mapInstanceRef.current.setCenter([userLocation.longitude, userLocation.latitude])
        mapInstanceRef.current.setZoom(15)
      } else if (markersRef.current.length === 1) {
        // 只有当没有当前用户位置且只有一个标记时，才使用该标记的位置
        const marker = markersRef.current[0]
        mapInstanceRef.current.setCenter(marker.getPosition())
        mapInstanceRef.current.setZoom(15)
      } else {
        // 当有多个其他位置时，先显示当前用户位置，不自动调整边界
        // 这样可以确保分享链接进入的用户能看到自己的位置
      }
      setViewInitialized(true)
    }
  }, [userLocation, otherLocations, userId, mapLoaded, viewInitialized])

  useEffect(() => {
    if (!sessionId || !userId) return

    const ws = connectWebSocket(sessionId, userId, userLocation, onLocationUpdate)
    wsRef.current = ws

    return () => {
      if (ws) {
        ws.close()
        wsRef.current = null
      }
    }
  }, [sessionId, userId, onLocationUpdate])

  useEffect(() => {
    if (!sessionId || !userId || !userLocation || !wsRef.current) return

    if (wsRef.current.sendLocation) {
      wsRef.current.sendLocation(userLocation.latitude, userLocation.longitude, userLocation.heading)
    }
  }, [userLocation, sessionId, userId])

  const createCustomIcon = (color, isCurrentUser) => {
    return new AMap.Icon({
      size: new AMap.Size(32, 32),
      image: `data:image/svg+xml;charset=utf-8,${encodeURIComponent(`
        <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">
          <circle cx="16" cy="16" r="${isCurrentUser ? 14 : 12}" fill="${color}" stroke="white" stroke-width="3"/>
          ${isCurrentUser ? '<circle cx="16" cy="16" r="6" fill="white"/>' : ''}
        </svg>
      `)}`,
      imageSize: new AMap.Size(32, 32)
    })
  }



  return (
    <div className="map-container">
      <div ref={mapRef} className="map"></div>
      {!mapLoaded && (
        <div className="map-loading">
          <div className="map-loading-spinner"></div>
          <p>正在加载地图...</p>
        </div>
      )}
    </div>
  )
}

export default Map
