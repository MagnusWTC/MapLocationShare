import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

export const createSession = async (data) => {
  try {
    const response = await api.post('/session', data)
    return response.data
  } catch (error) {
    console.error('创建会话失败:', error)
    throw error
  }
}

export const getSession = async (sessionId) => {
  try {
    const response = await api.get(`/session/${sessionId}`)
    return response.data
  } catch (error) {
    console.error('获取会话失败:', error)
    throw error
  }
}

export const updateLocation = async (data) => {
  try {
    const response = await api.post('/location', data)
    return response.data
  } catch (error) {
    console.error('更新位置失败:', error)
    throw error
  }
}

export default api
