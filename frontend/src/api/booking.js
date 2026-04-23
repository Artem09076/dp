import { API_CONFIG } from '../utils/config';
import authAPI from './auth';

class BookingAPI {
  constructor() {
    this.baseURL = API_CONFIG.BOOKING;
  }

  getHeaders() {
    const headers = {
      'Content-Type': 'application/json',
    };
    
    const token = localStorage.getItem('accessToken');
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    
    return headers;
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

      console.log(`Request to ${endpoint}:`, {
        status: response.status,
        statusText: response.statusText,
        ok: response.ok
      });

      if (response.status === 401 && retryCount === 0) {
        const refreshed = await authAPI.refreshToken();
        
        if (refreshed) {
          return this.request(endpoint, options, retryCount + 1);
        } else {
          localStorage.removeItem('accessToken');
          localStorage.removeItem('refreshToken');
          window.dispatchEvent(new Event('auth:logout'));
          throw new Error('Session expired. Please login again.');
        }
      }

      // Получаем текст ответа
      const responseText = await response.text();
      console.log('Response text:', responseText);
      
      // Парсим JSON если возможно
      let errorData;
      try {
        errorData = responseText ? JSON.parse(responseText) : {};
      } catch {
        errorData = { message: responseText };
      }
      
      // Если ответ не успешный (status не 2xx)
      if (!response.ok) {
        // Извлекаем сообщение об ошибке из разных возможных форматов
        let errorMessage = '';
        
        if (errorData.error) {
          if (typeof errorData.error === 'string') {
            errorMessage = errorData.error;
          } else if (errorData.error.message) {
            errorMessage = errorData.error.message;
          } else if (errorData.error.details) {
            errorMessage = errorData.error.details;
          } else {
            errorMessage = JSON.stringify(errorData.error);
          }
        } else if (errorData.message) {
          errorMessage = errorData.message;
        } else if (errorData.details) {
          errorMessage = errorData.details;
        } else {
          errorMessage = responseText || `HTTP ${response.status}: ${response.statusText}`;
        }
        
        throw new Error(errorMessage);
      }

      if (response.status === 204) {
        return null;
      }

      return errorData;
    };

    return makeRequest();
  }

  async createBooking(data) {
    console.log('Creating booking with data:', data);
    return this.request('/api/v1/bookings', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getBookings() {
    return this.request('/api/v1/bookings');
  }

  async getBooking(id) {
    return this.request(`/api/v1/bookings/${id}`);
  }

  async cancelBooking(id) {
    return this.request(`/api/v1/bookings/cancel/${id}`, {
      method: 'PATCH',
    });
  }

  async submitBooking(id) {
    return this.request(`/api/v1/bookings/submit/${id}`, {
      method: 'PATCH',
    });
  }

  async updateBooking(id, data) {
    return this.request(`/api/v1/bookings/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async deleteBooking(id) {
    return this.request(`/api/v1/bookings/${id}`, {
      method: 'DELETE',
    });
  }
}

export default new BookingAPI();