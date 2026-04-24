import { API_CONFIG } from '../utils/config';

class AuthAPI {
  constructor() {
    this.baseURL = API_CONFIG.AUTH;
    this.isRefreshing = false;
    this.refreshSubscribers = [];
  }

  getHeaders() {
    const headers = {
      'Content-Type': 'application/json',
    };
    
    const token = localStorage.getItem('accessToken');
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    let deviceId = localStorage.getItem('deviceId');
    if (!deviceId) {
      deviceId = this.generateDeviceId();
      localStorage.setItem('deviceId', deviceId);
    }
    headers['X-Device-ID'] = deviceId;
    
    return headers;
  }

  generateDeviceId() {
    return 'web_' + Math.random().toString(36).substring(2, 15);
  }

  subscribeToRefresh(callback) {
    this.refreshSubscribers.push(callback);
  }

  onRefreshSuccess(token) {
    this.refreshSubscribers.forEach(callback => callback(token));
    this.refreshSubscribers = [];
  }

  async refreshToken() {
    const refreshToken = localStorage.getItem('refreshToken');
    if (!refreshToken) {
      return false;
    }

    try {
      const response = await fetch(`${this.baseURL}/api/v1/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Device-ID': localStorage.getItem('deviceId') || 'web_default',
        },
        body: JSON.stringify({ refreshToken }),
      });

      const data = await response.json();
      
      if (data.accessToken && data.refreshToken) {
        localStorage.setItem('accessToken', data.accessToken);
        localStorage.setItem('refreshToken', data.refreshToken);
        return true;
      }
      return false;
    } catch (error) {
      console.error('Refresh token failed:', error);
      return false;
    }
  }

  async request(endpoint, options = {}, retryCount = 0) {
    const makeRequest = async () => {
      const response = await fetch(`${this.baseURL}${endpoint}`, {
        ...options,
        headers: {
          ...this.getHeaders(),
          ...options.headers,
        },
      });

      if (response.status === 401 && retryCount === 0) {
        const errorText = await response.text();
        let errorData;
        try {
          errorData = JSON.parse(errorText);
        } catch {
          errorData = { error: errorText };
        }
        
        if (errorData.error && (
          errorData.error.includes('expired') || 
          errorData.error.includes('invalid token') ||
          errorData.error.includes('token has invalid claims')
        )) {
          const refreshed = await this.refreshToken();
          
          if (refreshed) {
            return this.request(endpoint, options, retryCount + 1);
          } else {
            localStorage.removeItem('accessToken');
            localStorage.removeItem('refreshToken');
            window.dispatchEvent(new Event('auth:logout'));
            throw new Error('Session expired. Please login again.');
          }
        }
      }

      const responseText = await response.text();
      let result;
      try {
        result = responseText ? JSON.parse(responseText) : {};
      } catch {
        result = {};
      }

      return result;
    };

    return makeRequest();
  }

  async register(data) {
    return this.request('/api/v1/register', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async login(data) {
    const result = await this.request('/api/v1/login', {
      method: 'POST',
      body: JSON.stringify(data),
    });
    
    if (result.accessToken && result.refreshToken) {
      localStorage.setItem('accessToken', result.accessToken);
      localStorage.setItem('refreshToken', result.refreshToken);
    }
    
    return result;
  }

  async logout() {
    try {
      await this.request('/api/v1/logout', { method: 'POST' });
    } finally {
      localStorage.removeItem('accessToken');
      localStorage.removeItem('refreshToken');
      localStorage.removeItem('userRole');
      window.dispatchEvent(new Event('auth:logout'));
    }
  }

  isAuthenticated() {
    const token = localStorage.getItem('accessToken');
    if (!token) return false;
    
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const exp = payload.exp * 1000;
      return Date.now() < exp;
    } catch {
      return true;
    }
  }

  getUserRole() {
    const token = localStorage.getItem('accessToken');
    if (!token) return null;
    
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      return payload.role || localStorage.getItem('userRole');
    } catch {
      return localStorage.getItem('userRole');
    }
  }
}

export default new AuthAPI();