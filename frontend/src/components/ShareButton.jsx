import { useState } from 'react'
import { createSession } from '../services/api'
import './ShareButton.css'

function ShareButton({ userId, userLocation, sessionId, onShare }) {
  const [showShareModal, setShowShareModal] = useState(false)
  const [shareUrl, setShareUrl] = useState('')
  const [copied, setCopied] = useState(false)
  const [loading, setLoading] = useState(false)

  const handleShareClick = async () => {
    if (sessionId) {
      const url = `${window.location.origin}${window.location.pathname}?session=${sessionId}`
      setShareUrl(url)
      setShowShareModal(true)
    } else {
      setLoading(true)
      try {
        const response = await createSession({
          userId,
          latitude: userLocation.latitude,
          longitude: userLocation.longitude,
          heading: userLocation.heading || 0
        })

        if (response.sessionId) {
          const url = `${window.location.origin}${window.location.pathname}?session=${response.sessionId}`
          setShareUrl(url)
          setShowShareModal(true)
          onShare(response.sessionId)
        }
      } catch (error) {
        console.error('创建分享会话失败:', error)
        alert('创建分享会话失败，请重试')
      } finally {
        setLoading(false)
      }
    }
  }

  const handleCopyClick = async () => {
    try {
      // 尝试使用现代 Clipboard API
      await navigator.clipboard.writeText(shareUrl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (error) {
      console.error('Clipboard API 复制失败:', error)
      // 降级方案：使用传统的 document.execCommand('copy')
      try {
        // 创建一个临时的 textarea 元素
        const textarea = document.createElement('textarea')
        textarea.value = shareUrl
        textarea.style.position = 'fixed'
        textarea.style.opacity = '0'
        document.body.appendChild(textarea)
        
        // 选择并复制文本
        textarea.select()
        textarea.setSelectionRange(0, 99999) // 兼容移动设备
        
        // 执行复制命令
        const successful = document.execCommand('copy')
        if (successful) {
          setCopied(true)
          setTimeout(() => setCopied(false), 2000)
        } else {
          throw new Error('execCommand 复制失败')
        }
      } catch (execError) {
        console.error('execCommand 复制失败:', execError)
        // 最终方案：选中输入框文本，提示用户手动复制
        const inputElement = document.querySelector('.share-url-input')
        if (inputElement) {
          inputElement.select()
          inputElement.setSelectionRange(0, 99999) // 兼容移动设备
        }
        alert('复制失败，请长按输入框手动复制链接')
      } finally {
        // 清理临时元素
        const textarea = document.querySelector('textarea[style*="position: fixed"]')
        if (textarea) {
          document.body.removeChild(textarea)
        }
      }
    }
  }

  const handleCloseModal = () => {
    setShowShareModal(false)
    setCopied(false)
  }

  return (
    <>
      <button
        className="share-button"
        onClick={handleShareClick}
        disabled={loading}
      >
        {loading ? (
          <>
            <span className="share-button-spinner"></span>
            创建中...
          </>
        ) : (
          <>
            <svg
              className="share-icon"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M8.684 13.342C8.886 12.938 9 12.482 9 12c0-.482-.114-.938-.316-1.342m0 2.684a3 3 0 110-2.684m0 2.684l6.632 3.316m-6.632-6l6.632-3.316m0 0a3 3 0 105.367-2.684 3 3 0 00-5.367 2.684zm0 9.316a3 3 0 105.368 2.684 3 3 0 00-5.368-2.684z"
              />
            </svg>
            分享位置
          </>
        )}
      </button>

      {showShareModal && (
        <div className="share-modal-overlay" onClick={handleCloseModal}>
          <div className="share-modal" onClick={(e) => e.stopPropagation()}>
            <div className="share-modal-header">
              <h2>分享位置</h2>
              <button className="close-button" onClick={handleCloseModal}>
                <svg
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M6 18L18 6M6 6l12 12"
                  />
                </svg>
              </button>
            </div>
            <div className="share-modal-body">
              <p className="share-description">
                将此链接分享给好友，他们就能看到您的位置
              </p>
              <div className="share-url-container">
                <input
                  type="text"
                  value={shareUrl}
                  readOnly
                  className="share-url-input"
                />
                <button
                  className={`copy-button ${copied ? 'copied' : ''}`}
                  onClick={handleCopyClick}
                >
                  {copied ? (
                    <>
                      <svg
                        className="check-icon"
                        fill="none"
                        stroke="currentColor"
                        viewBox="0 0 24 24"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M5 13l4 4L19 7"
                        />
                      </svg>
                      已复制
                    </>
                  ) : (
                    '复制链接'
                  )}
                </button>
              </div>
              <div className="share-tips">
                <h3>使用提示：</h3>
                <ul>
                  <li>链接有效期：24小时</li>
                  <li>可以分享给多个好友</li>
                  <li>好友点击链接后即可看到您的实时位置</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

export default ShareButton
