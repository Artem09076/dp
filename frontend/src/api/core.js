import { API_CONFIG } from '../utils/config';
import authAPI from './auth';

class CoreAPI {
  constructor() {
    this.baseURL = API_CONFIG.CORE;
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

      if (response.status === 204) {
        return null;
      }

      const responseText = await response.text();
      try {
        return responseText ? JSON.parse(responseText) : {};
      } catch {
        return {};
      }
    };

    return makeRequest();
  }

  async getProfile() {
    return this.request('/api/v1/profile');
  }

  async updateProfile(data) {
    return this.request('/api/v1/profile', {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async deleteProfile() {
    return this.request('/api/v1/profile', {
      method: 'DELETE',
    });
  }

  async searchServices(query, page = 1, limit = 10) {
    const params = new URLSearchParams({ query, page, limit });
    return this.request(`/api/v1/services/search?${params}`);
  }

  async getService(id) {
    return this.request(`/api/v1/services/${id}`);
  }

  async getMyServices() {
    return this.request('/api/v1/services');
  }

  async createService(data) {
    return this.request('/api/v1/services', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateService(id, data) {
    return this.request(`/api/v1/services/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async deleteService(id) {
    return this.request(`/api/v1/services/${id}`, {
      method: 'DELETE',
    });
  }

  async getDiscount(id) {
    return this.request(`/api/v1/discounts/${id}`);
  }

  async getServiceDiscounts(serviceId) {
    return this.request(`/api/v1/services/${serviceId}/discounts`);
  }

  async createDiscount(serviceId, data) {
    return this.request(`/api/v1/services/${serviceId}/discounts`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateDiscount(serviceId, discountId, data) {
    return this.request(`/api/v1/services/${serviceId}/discounts/${discountId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async deleteDiscount(serviceId, discountId) {
    return this.request(`/api/v1/services/${serviceId}/discounts/${discountId}`, {
      method: 'DELETE',
    });
  }
  async deleteService(serviceId) {
  return this.request(`/api/v1/services/${serviceId}`, {
    method: 'DELETE',
  });
}

  async deleteBooking(bookingId) {
    return this.request(`/api/v1/bookings/${bookingId}`, {
      method: 'DELETE',
    });
  }


  async createReview(data) {
    return this.request('/api/v1/reviews', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

 async getReviewByBooking(bookingId) {
  try {
    const result = await this.request(`/api/v1/booking/${bookingId}/reviews`);
    return result;
  } catch (error) {
    if (error.message === 'Not Found' || error.message?.includes('404') || error.status === 404) {
      console.log('No review found for booking:', bookingId);
      return null;
    }
    throw error;
  }
}

  async getServiceReviews(serviceId, page = 1, limit = 10) {
    const params = new URLSearchParams({ page, limit });
    return this.request(`/api/v1/service/${serviceId}/reviews?${params}`);
  }

  async updateReview(reviewId, data) {
    return this.request(`/api/v1/reviews/${reviewId}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
    });
  }

  async deleteReview(reviewId) {
    return this.request(`/api/v1/reviews/${reviewId}`, {
      method: 'DELETE',
    });
  }

  async getUnverifiedPerformers(page = 1, pageSize = 20) {
  const params = new URLSearchParams({ 
    page: page.toString(), 
    page_size: pageSize.toString() 
  });
  const result = await this.request(`/api/v1/admin/performers/unverified?${params}`);
  console.log('Unverified performers response:', result);
  return result;
}

  async getUsers(filters = {}) {
  const params = new URLSearchParams();
  if (filters.role) params.append('role', filters.role);
  if (filters.verification_status) params.append('verification_status', filters.verification_status);
  if (filters.search) params.append('search', filters.search);
  if (filters.page) params.append('page', filters.page);
  if (filters.page_size) params.append('page_size', filters.page_size);
  
  const result = await this.request(`/api/v1/admin/users?${params}`);
  console.log('Users response:', result);
  return result;
}

  async getUserById(userId) {
    return this.request(`/api/v1/admin/users/${userId}`);
  }

  async deleteUser(userId) {
    return this.request(`/api/v1/admin/users/${userId}`, {
      method: 'DELETE',
    });
  }

  async batchVerifyPerformers(userIds, status) {
    return this.request('/api/v1/admin/users/verify/batch', {
      method: 'POST',
      body: JSON.stringify({ user_ids: userIds, status }),
    });
  }

  async getAdminServices(performerId = '', page = 1, pageSize = 20) {
  const params = new URLSearchParams({ 
    page: page.toString(), 
    page_size: pageSize.toString() 
  });
  if (performerId) params.append('performer_id', performerId);
  const result = await this.request(`/api/v1/admin/services?${params}`);
  console.log('Admin services response:', result);
  return result;
}
  async getAdminBookings(filters = {}) {
  const params = new URLSearchParams();
  if (filters.status) params.append('status', filters.status);
  if (filters.client_id) params.append('client_id', filters.client_id);
  if (filters.performer_id) params.append('performer_id', filters.performer_id);
  if (filters.page) params.append('page', filters.page);
  if (filters.page_size) params.append('page_size', filters.page_size);
  
  const result = await this.request(`/api/v1/admin/bookings?${params}`);
  console.log('Admin bookings response:', result);
  return result;
}

  async getAdminReviews(serviceId = '', rating = '', page = 1, pageSize = 20) {
  const params = new URLSearchParams({ 
    page: page.toString(), 
    page_size: pageSize.toString() 
  });
  if (serviceId) params.append('service_id', serviceId);
  if (rating) params.append('rating', rating);
  const result = await this.request(`/api/v1/admin/reviews?${params}`);
  console.log('Admin reviews response:', result);
  return result;
}

  async deleteAdminReview(reviewId) {
    return this.request(`/api/v1/admin/reviews/${reviewId}`, {
      method: 'DELETE',
    });
  }

  async updateVerificationStatus(userId, status) {
    return this.request('/api/v1/admin/performers/verification_status', {
      method: 'PATCH',
      body: JSON.stringify({ user_id: userId, verification_status: status }),
    });
  }
}

export default new CoreAPI();