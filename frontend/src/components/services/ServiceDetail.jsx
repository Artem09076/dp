import React, { useState, useEffect } from 'react';
import coreAPI from '../../api/core';
import bookingAPI from '../../api/booking';
import { useAuth } from '../../contexts/AuthContext';
import './Services.css';

const ServiceDetail = ({ serviceId, onClose }) => {
  const [service, setService] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [discounts, setDiscounts] = useState([]);
  const [showBookingForm, setShowBookingForm] = useState(false);
  const [bookingTime, setBookingTime] = useState('');
  const [selectedDiscount, setSelectedDiscount] = useState('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { userRole, isAuthenticated } = useAuth();

  useEffect(() => {
    loadServiceDetails();
  }, [serviceId]);

  const loadServiceDetails = async () => {
    try {
      const serviceData = await coreAPI.getService(serviceId);
      setService(serviceData);
      
      const reviewsData = await coreAPI.getServiceReviews(serviceId);
      setReviews(Array.isArray(reviewsData) ? reviewsData : []);
      
      setDiscounts([]);
    } catch (err) {
      setError('Failed to load service details');
    } finally {
      setLoading(false);
    }
  };

  const handleBookService = async (e) => {
    e.preventDefault();
    if (!bookingTime) {
      setError('Please select booking time');
      return;
    }

    try {
      await bookingAPI.createBooking({
        service_id: serviceId,
        booking_time: bookingTime,
        discount_id: selectedDiscount || undefined
      });
      alert('Booking created successfully!');
      setShowBookingForm(false);
      onClose();
    } catch (err) {
      setError(err.message || 'Failed to create booking');
    }
  };

  const calculateAverageRating = () => {
    if (reviews.length === 0) return 0;
    const sum = reviews.reduce((acc, review) => acc + review.rating, 0);
    return (sum / reviews.length).toFixed(1);
  };

  if (loading) return <div className="loading">Loading service details...</div>;
  if (error) return <div className="error-message">{error}</div>;
  if (!service) return <div className="error-message">Service not found</div>;

  return (
    <div className="service-detail-modal">
      <div className="service-detail-content">
        <button className="close-button" onClick={onClose}>×</button>
        
        <div className="service-header">
          <h2>{service.title}</h2>
          <div className="service-meta">
            <span className="price">₽{service.price}</span>
            <span className="duration">⏱️ {service.duration_minutes} minutes</span>
          </div>
        </div>
        
        {service.description && (
          <div className="service-description">
            <h3>Описание</h3>
            <p>{service.description}</p>
          </div>
        )}
        
        <div className="service-stats">
          <div className="stat">
            <span className="stat-label">Average Rating</span>
            <span className="stat-value">
              {calculateAverageRating()} ⭐ ({reviews.length} reviews)
            </span>
          </div>
          <div className="stat">
            <span className="stat-label">Performer</span>
            <span className="stat-value">{service.performerName || 'Professional'}</span>
          </div>
        </div>
        
        {discounts.length > 0 && (
          <div className="discounts-section">
            <h3>Available Discounts</h3>
            <div className="discounts-list">
              {discounts.map(discount => (
                <div key={discount.id} className="discount-card">
                  <span className="discount-value">
                    {discount.type === 'percentage' ? `${discount.value}% OFF` : `$${discount.value} OFF`}
                  </span>
                  <span className="discount-valid">
                    Valid until {new Date(discount.validTo).toLocaleDateString()}
                  </span>
                </div>
              ))}
            </div>
          </div>
        )}
        
        {reviews.length > 0 && (
          <div className="reviews-section">
            <h3>Customer Reviews</h3>
            <div className="reviews-list">
              {reviews.map(review => (
                <div key={review.id} className="review-card">
                  <div className="review-header">
                    <span className="review-rating">{'⭐'.repeat(review.rating)}</span>
                    <span className="review-date">{new Date(review.createdAt).toLocaleDateString()}</span>
                  </div>
                  {review.comment && <p className="review-comment">"{review.comment}"</p>}
                  <span className="review-author">- {review.clientName || 'Customer'}</span>
                </div>
              ))}
            </div>
          </div>
        )}
        
        {isAuthenticated && userRole === 'client' && !showBookingForm && (
          <div className="service-actions">
            <button onClick={() => setShowBookingForm(true)} className="btn-book-now">
              Book Now
            </button>
          </div>
        )}
        
        {showBookingForm && (
          <div className="booking-form">
            <h3>Book this Service</h3>
            <form onSubmit={handleBookService}>
              <div className="form-group">
                <label>Select Date and Time:</label>
                <input
                  type="datetime-local"
                  value={bookingTime}
                  onChange={(e) => setBookingTime(e.target.value)}
                  required
                  min={new Date().toISOString().slice(0, 16)}
                />
              </div>
              
              {discounts.length > 0 && (
                <div className="form-group">
                  <label>Apply Discount:</label>
                  <select value={selectedDiscount} onChange={(e) => setSelectedDiscount(e.target.value)}>
                    <option value="">No discount</option>
                    {discounts.map(discount => (
                      <option key={discount.id} value={discount.id}>
                        {discount.type === 'percentage' ? `${discount.value}% off` : `$${discount.value} off`}
                      </option>
                    ))}
                  </select>
                </div>
              )}
              
              {error && <div className="error-message">{error}</div>}
              
              <div className="form-actions">
                <button type="submit" className="btn-confirm">Confirm Booking</button>
                <button type="button" onClick={() => setShowBookingForm(false)} className="btn-cancel">
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}
      </div>
    </div>
  );
};

export default ServiceDetail;