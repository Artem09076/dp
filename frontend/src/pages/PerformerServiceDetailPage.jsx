import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import coreAPI from '../api/core';
import { useAuth } from '../contexts/AuthContext';
import './ServiceDetailPage.css';

const PerformerServiceDetailPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  
  const [service, setService] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [discounts, setDiscounts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [averageRating, setAverageRating] = useState(0);
  const [showEditForm, setShowEditForm] = useState(false);
  const [showDiscountForm, setShowDiscountForm] = useState(false);
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    price: '',
    duration_minutes: ''
  });
  const [discountForm, setDiscountForm] = useState({
    type: 'percentage',
    value: '',
    valid_from: '',
    valid_to: '',
    max_uses: ''
  });

  useEffect(() => {
    if (id) {
      loadServiceDetails();
    }
  }, [id]);

  const loadServiceDetails = async () => {
    try {
      setLoading(true);
      setError('');
      
      console.log('Fetching service with ID:', id);
      const serviceData = await coreAPI.getService(id);
      console.log('Service data received:', serviceData);
      
      if (!serviceData || !serviceData.id) {
        setError('Service not found');
        setLoading(false);
        return;
      }
      
      setService(serviceData);
      setFormData({
        title: serviceData.title || '',
        description: serviceData.description || '',
        price: serviceData.price?.toString() || '0',
        duration_minutes: serviceData.duration_minutes?.toString() || serviceData.durationMinutes?.toString() || '0'
      });
      
      try {
        const reviewsData = await coreAPI.getServiceReviews(id);
        const reviewsArray = Array.isArray(reviewsData) ? reviewsData : [];
        setReviews(reviewsArray);
        
        if (reviewsArray.length > 0) {
          const sum = reviewsArray.reduce((acc, review) => acc + (review.rating || 0), 0);
          setAverageRating(sum / reviewsArray.length);
        }
      } catch (err) {
        console.error('Failed to load reviews:', err);
      }
      
      try {
        const discountsData = await coreAPI.getServiceDiscounts(id);
        setDiscounts(Array.isArray(discountsData) ? discountsData : []);
      } catch (err) {
        console.error('Failed to load discounts:', err);
      }
      
    } catch (err) {
      console.error('Failed to load service:', err);
      setError(err.message || 'Service not found');
    } finally {
      setLoading(false);
    }
  };

  const handleUpdateService = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.updateService(id, {
        title: formData.title,
        description: formData.description || undefined,
        price: parseFloat(formData.price),
        duration_minutes: parseInt(formData.duration_minutes)
      });
      alert('Service updated successfully!');
      setShowEditForm(false);
      loadServiceDetails();
    } catch (err) {
      console.error('Update error:', err);
      alert('Failed to update service');
    }
  };

  const handleDeleteService = async () => {
    if (window.confirm('Are you sure you want to delete this service? This action cannot be undone.')) {
      try {
        await coreAPI.deleteService(id);
        alert('Service deleted successfully!');
        navigate('/my-services');
      } catch (err) {
        console.error('Delete error:', err);
        alert('Failed to delete service');
      }
    }
  };

const formatDateForBackend = (dateTimeLocal) => {
  if (!dateTimeLocal) return null;
  
  const date = new Date(dateTimeLocal);
  if (isNaN(date.getTime())) {
    throw new Error('Invalid date');
  }
  
  const offset = -date.getTimezoneOffset();
  const offsetHours = String(Math.floor(Math.abs(offset) / 60)).padStart(2, '0');
  const offsetMinutes = String(Math.abs(offset) % 60).padStart(2, '0');
  const offsetSign = offset >= 0 ? '+' : '-';
  
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  const hours = String(date.getHours()).padStart(2, '0');
  const minutes = String(date.getMinutes()).padStart(2, '0');
  const seconds = '00';
  
  return `${year}-${month}-${day}T${hours}:${minutes}:${seconds}${offsetSign}${offsetHours}:${offsetMinutes}`;
};

const handleCreateDiscount = async (e) => {
  e.preventDefault();
  
  if (!discountForm.valid_from || !discountForm.valid_to) {
    alert('Please select valid from and valid to dates');
    return;
  }
  
  let formattedValidFrom, formattedValidTo;
  try {
    formattedValidFrom = formatDateForBackend(discountForm.valid_from);
    formattedValidTo = formatDateForBackend(discountForm.valid_to);
    console.log('Formatted dates:', { formattedValidFrom, formattedValidTo });
  } catch (err) {
    alert('Invalid date format');
    return;
  }
  
  try {
    const discountData = {
      type: discountForm.type,
      value: parseInt(discountForm.value),
      valid_from: formattedValidFrom,
      valid_to: formattedValidTo,
      max_uses: parseInt(discountForm.max_uses)
    };
    console.log('Sending discount data:', discountData);
    
    await coreAPI.createDiscount(id, discountData);
    alert('Discount created successfully!');
    setShowDiscountForm(false);
    setDiscountForm({ type: 'percentage', value: '', valid_from: '', valid_to: '', max_uses: '' });
    loadServiceDetails();
  } catch (err) {
    console.error('Create discount error:', err);
    alert('Failed to create discount: ' + (err.message || 'Unknown error'));
  }
};

  const handleDeleteDiscount = async (discountId) => {
    if (window.confirm('Are you sure you want to delete this discount?')) {
      try {
        await coreAPI.deleteDiscount(id, discountId);
        alert('Discount deleted successfully!');
        loadServiceDetails();
      } catch (err) {
        console.error('Delete discount error:', err);
        alert('Failed to delete discount');
      }
    }
  };

  const parseDate = (dateString) => {
    if (!dateString) return null;
    try {
      const date = new Date(dateString);
      if (!isNaN(date.getTime())) return date;
      
      const parts = dateString.split(' ');
      if (parts.length >= 3) {
        const datePart = parts[0];
        const timePart = parts[1].split('.')[0];
        const isoString = `${datePart}T${timePart}+00:00`;
        const date2 = new Date(isoString);
        if (!isNaN(date2.getTime())) return date2;
      }
      return null;
    } catch (e) {
      return null;
    }
  };

  const formatDate = (dateString) => {
    const date = parseDate(dateString);
    if (!date) return 'Date not available';
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  if (loading) {
    return (
      <div className="service-detail-page">
        <div className="loading-container">
          <div className="loader"></div>
          <p>Loading service details...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="service-detail-page">
        <div className="error-container">
          <h2>Error</h2>
          <p>{error}</p>
          <p>Service ID: {id}</p>
          <button onClick={() => navigate('/my-services')} className="btn-back">
            ← Back to My Services
          </button>
        </div>
      </div>
    );
  }

  if (!service) {
    return (
      <div className="service-detail-page">
        <div className="error-container">
          <h2>Service Not Found</h2>
          <p>Service ID: {id}</p>
          <button onClick={() => navigate('/my-services')} className="btn-back">
            ← Back to My Services
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="service-detail-page performer-page">
      <div className="service-detail-container">
        <button onClick={() => navigate('/my-services')} className="btn-back">
          ← Back to My Services
        </button>
        
        <div className="service-detail-header">
          <h1>{service.title}</h1>
          <div className="service-actions-buttons">
            <button onClick={() => setShowEditForm(true)} className="btn-edit-service">
              ✏️ Edit Service
            </button>
            <button onClick={handleDeleteService} className="btn-delete-service">
              🗑️ Delete Service
            </button>
          </div>
          <div className="service-meta">
            <div className="rating-large">
              {'⭐'.repeat(Math.round(averageRating))}
              <span className="rating-value">
                {averageRating > 0 ? ` ${averageRating.toFixed(1)}` : ' No ratings yet'}
              </span>
            </div>
            <div className="price-large">${service.price}</div>
            <div className="duration-large">⏱️ {service.duration_minutes || service.durationMinutes} minutes</div>
          </div>
        </div>

        {service.description && (
          <div className="service-description">
            <h2>Description</h2>
            <p>{service.description}</p>
          </div>
        )}

        {/* Discounts Management */}
        <div className="service-discounts">
          <div className="discounts-header">
            <h2>Discounts</h2>
            <button onClick={() => setShowDiscountForm(true)} className="btn-add-discount">
              + Add Discount
            </button>
          </div>
          {discounts.length === 0 ? (
            <p className="no-discounts">No discounts available for this service.</p>
          ) : (
            <div className="discounts-list">
              {discounts.map(discount => (
                <div key={discount.id} className="discount-card">
                  <div className="discount-badge">
                    {discount.type === 'percentage' ? `${discount.value}% OFF` : `$${discount.value} OFF`}
                  </div>
                  <div className="discount-info">
                    <p>Valid from {formatDate(discount.valid_from || discount.validFrom)} to {formatDate(discount.valid_to || discount.validTo)}</p>
                    <p>Used {discount.used_count || discount.usedCount} of {discount.max_uses || discount.maxUses} times</p>
                  </div>
                  <button onClick={() => handleDeleteDiscount(discount.id)} className="btn-delete-discount">
                    Delete
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Reviews Section */}
        {reviews.length > 0 && (
          <div className="service-reviews">
            <h2>Customer Reviews ({reviews.length})</h2>
            <div className="reviews-list">
              {reviews.map(review => (
                <div key={review.id} className="review-card">
                  <div className="review-header">
                    <div className="review-rating">
                      {'⭐'.repeat(review.rating)} ({review.rating}/5)
                    </div>
                    <div className="review-date">{formatDate(review.created_at || review.createdAt)}</div>
                  </div>
                  {review.comment && (
                    <p className="review-comment">"{review.comment}"</p>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {/* Edit Service Modal */}
      {showEditForm && (
        <div className="modal" onClick={() => setShowEditForm(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Edit Service</h3>
            <form onSubmit={handleUpdateService}>
              <div className="form-group">
                <label>Service Title *</label>
                <input
                  type="text"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows="3"
                />
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>Price ($) *</label>
                  <input
                    type="number"
                    value={formData.price}
                    onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                    required
                    min="0"
                    step="0.01"
                  />
                </div>
                <div className="form-group">
                  <label>Duration (minutes) *</label>
                  <input
                    type="number"
                    value={formData.duration_minutes}
                    onChange={(e) => setFormData({ ...formData, duration_minutes: e.target.value })}
                    required
                    min="15"
                    step="15"
                  />
                </div>
              </div>
              <div className="modal-actions">
                <button type="submit" className="btn-confirm">Update Service</button>
                <button type="button" onClick={() => setShowEditForm(false)} className="btn-cancel">Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Add Discount Modal */}
      {showDiscountForm && (
        <div className="modal" onClick={() => setShowDiscountForm(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Add Discount</h3>
            <form onSubmit={handleCreateDiscount}>
              <div className="form-group">
                <label>Discount Type</label>
                <select
                  value={discountForm.type}
                  onChange={(e) => setDiscountForm({ ...discountForm, type: e.target.value })}
                >
                  <option value="percentage">Percentage (%)</option>
                  <option value="fixed_amount">Fixed Amount ($)</option>
                </select>
              </div>
              <div className="form-group">
                <label>Value</label>
                <input
                  type="number"
                  value={discountForm.value}
                  onChange={(e) => setDiscountForm({ ...discountForm, value: e.target.value })}
                  required
                  min="1"
                  placeholder={discountForm.type === 'percentage' ? 'e.g., 10' : 'e.g., 50'}
                />
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>Valid From</label>
                  <input
                    type="datetime-local"
                    value={discountForm.valid_from}
                    onChange={(e) => setDiscountForm({ ...discountForm, valid_from: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label>Valid To</label>
                  <input
                    type="datetime-local"
                    value={discountForm.valid_to}
                    onChange={(e) => setDiscountForm({ ...discountForm, valid_to: e.target.value })}
                    required
                  />
                </div>
              </div>
              <div className="form-group">
                <label>Max Uses</label>
                <input
                  type="number"
                  value={discountForm.max_uses}
                  onChange={(e) => setDiscountForm({ ...discountForm, max_uses: e.target.value })}
                  required
                  min="1"
                  placeholder="e.g., 10"
                />
              </div>
              <div className="modal-actions">
                <button type="submit" className="btn-confirm">Create Discount</button>
                <button type="button" onClick={() => setShowDiscountForm(false)} className="btn-cancel">Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default PerformerServiceDetailPage;