import { useState, useEffect } from 'react'
import Map from './components/Map'
import ShareButton from './components/ShareButton'
import './App.css'

function App() {
  const [userLocation, setUserLocation] = useState(null)
  const [otherLocations, setOtherLocations] = useState([])
  const [sessionId, setSessionId] = useState(null)
  const [userId, setUserId] = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState(null)

  useEffect(() => {
    const urlParams = new URLSearchParams(window.location.search)
    const sharedSessionId = urlParams.get('session')

    // 从本地存储获取用户ID，如果不存在则生成新的
    let storedUserId = localStorage.getItem('map_location_user_id')
    if (!storedUserId) {
      storedUserId = 'user_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
      localStorage.setItem('map_location_user_id', storedUserId)
    }
    setUserId(storedUserId)

    if (sharedSessionId) {
      setSessionId(sharedSessionId)
    }

    if (navigator.geolocation) {
      console.log('开始获取位置...')
      let lastAccurateLocation = null
      let locationAttempts = 0
      const maxAttempts = 5
      const accuracyThreshold = 200 // 只接受精度优于200米的位置

      const updateLocation = (position) => {
        const { latitude, longitude, accuracy, heading } = position.coords
        console.log('位置更新:', { latitude, longitude, accuracy, heading })
        
        // 位置过滤：只接受精度优于阈值的位置
        if (accuracy > accuracyThreshold && locationAttempts < maxAttempts) {
          console.log(`位置精度不足 (${accuracy}米)，继续尝试...`)
          locationAttempts++
          return
        }

        // 如果之前有准确位置，可以考虑平滑处理
        let finalLatitude = latitude
        let finalLongitude = longitude
        
        if (lastAccurateLocation) {
          // 简单的加权平均平滑，新位置权重0.7，旧位置权重0.3
          const weight = 0.7
          finalLatitude = latitude * weight + lastAccurateLocation.latitude * (1 - weight)
          finalLongitude = longitude * weight + lastAccurateLocation.longitude * (1 - weight)
        }

        const location = {
          userId: storedUserId,
          latitude: finalLatitude,
          longitude: finalLongitude,
          heading: heading || 0,
          timestamp: Date.now(),
          accuracy: accuracy
        }

        setUserLocation(location)
        lastAccurateLocation = location
        setLoading(false)
      }

      navigator.geolocation.getCurrentPosition(
        updateLocation,
        (err) => {
          console.error('位置获取失败:', err)
          const errorMessage = `无法获取位置信息: ${err.message || err.code || '未知错误'}`
          setError(errorMessage)
          setLoading(false)
        },
        {
          enableHighAccuracy: true,
          timeout: 15000,
          maximumAge: 0
        }
      )

      const watchId = navigator.geolocation.watchPosition(
        updateLocation,
        (err) => {
          console.error('位置监听失败:', err)
        },
        {
          enableHighAccuracy: true,
          timeout: 15000,
          maximumAge: 0,
          // 增加距离过滤，只有当位置变化超过5米时才更新
          distanceFilter: 5
        }
      )

      return () => {
        if (watchId) {
          navigator.geolocation.clearWatch(watchId)
        }
      }
    } else {
      const errorMessage = '您的浏览器不支持地理定位'
      console.error(errorMessage)
      setError(errorMessage)
      setLoading(false)
    }
  }, [])

  const handleShare = (newSessionId) => {
    console.log('分享会话ID:', newSessionId)
    setSessionId(newSessionId)
  }

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loading-spinner"></div>
        <p>正在获取您的位置...</p>
        <p className="loading-hint">请允许浏览器访问您的位置</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="error-container">
        <p className="error-message">{error}</p>
        <button onClick={() => window.location.reload()} className="retry-button">
          重试
        </button>
        <p className="error-hint">
          提示：请确保浏览器允许位置访问权限
        </p>
      </div>
    )
  }

  const allLocations = [...otherLocations]
  if (userLocation) {
    allLocations.push(userLocation)
  }

  return (
    <div className="app">
      <header className="app-header">
        <h1>地图位置共享</h1>
        {userLocation && (
          <ShareButton
            userId={userId}
            userLocation={userLocation}
            sessionId={sessionId}
            onShare={handleShare}
          />
        )}
      </header>
      <main className="app-main">
        <Map
          userLocation={userLocation}
          otherLocations={otherLocations}
          userId={userId}
          sessionId={sessionId}
          onLocationUpdate={setOtherLocations}
        />
      </main>
    </div>
  )
}

export default App
