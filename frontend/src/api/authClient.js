// src/api/authClient.js
import { grpc } from 'grpc-web'
import { 
  AuthServiceClient, 
  RegisterRequest, 
  LoginRequest, 
  RefreshTokenRequest, 
  LogoutRequest 
} from '../proto'

const GRPC_URL = import.meta.env.VITE_GRPC_URL || 'http://localhost:8080'

class AuthApiClient {
  constructor() {
    this.client = new AuthServiceClient(GRPC_URL)
    this.accessToken = localStorage.getItem('accessToken')
    this.refreshToken = localStorage.getItem('refreshToken')
    this.deviceId = this.getOrCreateDeviceId()
  }

  getOrCreateDeviceId() {
    let deviceId = localStorage.getItem('deviceId')
    if (!deviceId) {
      deviceId = 'web_' + Math.random().toString(36).substring(2, 15) + '_' + Date.now()
      localStorage.setItem('deviceId', deviceId)
    }
    return deviceId
  }

  setTokens(accessToken, refreshToken) {
    this.accessToken = accessToken
    this.refreshToken = refreshToken
    if (accessToken) localStorage.setItem('accessToken', accessToken)
    if (refreshToken) localStorage.setItem('refreshToken', refreshToken)
  }

  clearTokens() {
    this.accessToken = null
    this.refreshToken = null
    localStorage.removeItem('accessToken')
    localStorage.removeItem('refreshToken')
  }

  getMetadata() {
    const metadata = {}
    if (this.accessToken) {
      metadata['authorization'] = `Bearer ${this.accessToken}`
    }
    metadata['x-device-id'] = this.deviceId
    return metadata
  }

  register(data) {
    return new Promise((resolve, reject) => {
      const request = new RegisterRequest()
      request.setEmail(data.email)
      request.setName(data.name)
      request.setPassword(data.password)
      request.setInn(data.inn || '')
      request.setBusinesstype(data.businessType || '')
      request.setUserrole(data.userRole)

      this.client.register(request, this.getMetadata(), (err, response) => {
        if (err) {
          console.error('gRPC register error:', err)
          reject(new Error(this.parseGrpcError(err)))
        } else {
          const accessToken = response.getAccesstoken()
          const refreshToken = response.getRefreshtoken()
          this.setTokens(accessToken, refreshToken)
          resolve({ accessToken, refreshToken })
        }
      })
    })
  }

  login(credentials) {
    return new Promise((resolve, reject) => {
      const request = new LoginRequest()
      if (credentials.email) request.setEmail(credentials.email)
      if (credentials.inn) request.setInn(credentials.inn)
      request.setPassword(credentials.password)

      this.client.login(request, this.getMetadata(), (err, response) => {
        if (err) {
          console.error('gRPC login error:', err)
          reject(new Error(this.parseGrpcError(err)))
        } else {
          const accessToken = response.getAccesstoken()
          const refreshToken = response.getRefreshtoken()
          this.setTokens(accessToken, refreshToken)
          resolve({ accessToken, refreshToken })
        }
      })
    })
  }

  refreshToken() {
    return new Promise((resolve, reject) => {
      if (!this.refreshToken) {
        reject(new Error('No refresh token'))
        return
      }

      const request = new RefreshTokenRequest()
      request.setRefreshtoken(this.refreshToken)

      this.client.refreshToken(request, this.getMetadata(), (err, response) => {
        if (err) {
          console.error('gRPC refresh error:', err)
          this.clearTokens()
          reject(new Error(this.parseGrpcError(err)))
        } else {
          const accessToken = response.getAccesstoken()
          const refreshToken = response.getRefreshtoken()
          this.setTokens(accessToken, refreshToken)
          resolve({ accessToken, refreshToken })
        }
      })
    })
  }

  logout() {
    return new Promise((resolve, reject) => {
      const request = new LogoutRequest()
      
      this.client.logout(request, this.getMetadata(), (err, response) => {
        this.clearTokens()
        if (err) {
          console.error('gRPC logout error:', err)
          reject(new Error(this.parseGrpcError(err)))
        } else {
          resolve({ success: response.getSuccess() })
        }
      })
    })
  }

  parseGrpcError(err) {
    if (err.code === 16) return 'Неверные учетные данные'
    if (err.code === 6) return 'Пользователь уже существует'
    if (err.code === 5) return 'Не найдено'
    if (err.code === 14) return 'Не удалось подключиться к серверу'
    return err.message || 'Ошибка соединения с сервером'
  }

  isAuthenticated() {
    return !!this.accessToken
  }

  getUserFromToken() {
    if (!this.accessToken) return null
    try {
      const payload = JSON.parse(atob(this.accessToken.split('.')[1]))
      return {
        id: payload.user_id,
        email: payload.email,
        role: payload.role,
      }
    } catch {
      return null
    }
  }
}

export const authApi = new AuthApiClient()