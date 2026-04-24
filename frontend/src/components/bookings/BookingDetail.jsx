import React, { useState, useEffect } from 'react';
import bookingAPI from '../../api/booking';
import coreAPI from '../../api/core';
import { useAuth } from '../../contexts/AuthContext';
import './Booking.css';

const BookingDetail = ({ bookingId, onClose }) => {
  const [booking, setBooking] = useState(null);
  const [review, setReview] = useState(null);
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [reviewData, setReviewData] = useState({ rating: 5, comment: '' });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const { userRole } = useAuth();

  useEffect(() => {
    loadBookingDetail();
  }, [bookingId]);

  const loadBookingDetail = async () => {
    try {
      const data = await bookingAPI.getBooking(bookingId);
      setBooking(data);
      
      try {
        const reviewData = await coreAPI.getReviewByBooking(bookingId);
        if (reviewData && reviewData.id) {
          setReview(reviewData);
        }
      } catch (err) {
      }
    } catch (err) {
      setError('Failed to load booking details');
    } finally {
      setLoading(false);
    }
  };

  const handleCancelBooking = async () => {
    if (window.confirm('Are you sure you want to cancel this booking?')) {
      try {
        await bookingAPI.cancelBooking(bookingId);
        await loadBookingDetail();
        alert('Booking cancelled successfully');
      } catch (err) {
        setError('Failed to cancel booking');
      }
    }
  };

  const handleSubmitBooking = async () => {
    if (window.confirm('Mark this booking as completed?')) {
      try {
        await bookingAPI.submitBooking(bookingId);
        await loadBookingDetail();
        alert('Booking marked as completed');
      } catch (err) {
        setError('Failed to submit booking');
      }
    }
  };

  const handleUpdateTime = async (newTime) => {
    try {
      await bookingAPI.updateBooking(bookingId, { booking_time: newTime });
      await loadBookingDetail();
      alert('Booking time updated successfully');
    } catch (err) {
      setError('Failed to update booking time');
    }
  };

  const handleSubmitReview = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.createReview({
        booking_id: bookingId,
        rating: reviewData.rating,
        comment: reviewData.comment
      });
      await loadBookingDetail();
      setShowReviewForm(false);
      alert('Review submitted successfully');
    } catch (err) {
      setError('Failed to submit review');
    }
  };

  const getStatusActions = () => {
    if (!booking) return null;
    
    switch(booking.status) {
      case 'pending':
        return (
          <div className="status-actions">
            <button onClick={handleCancelBooking} className="btn-cancel-booking">
              Cancel Booking
            </button>
            {userRole === 'performer' && (
              <button onClick={() => {
                const newTime = prompt('Enter new date and time (YYYY-MM-DD HH:MM):');
                if (newTime) handleUpdateTime(newTime);
              }} className="btn-update-time">
                Suggest New Time
              </button>
            )}
          </div>
        );
      case 'confirmed':
        return (
          <div className="status-actions">
            {userRole === 'performer' && (
              <button onClick={handleSubmitBooking} className="btn-complete">
                Mark as Completed
              </button>
            )}
            <button onClick={handleCancelBooking} className="btn-cancel-booking">
              Cancel Booking
            </button>
          </div>
        );
      case 'completed':
        if (!review && userRole === 'client') {
          return (
            <div className="status-actions">
              <button onClick={() => setShowReviewForm(true)} className="btn-review">
                Write a Review
              </button>
            </div>
          );
        }
        return null;
      default:
        return null;
    }
  };

  if (loading) return <div className="loading">Loading booking details...</div>;
  if (error) return <div className="error-message">{error}</div>;
  if (!booking) return <div className="error-message">Booking not found</div>;

  return (
    <div className="booking-detail-modal">
      <div className="booking-detail-content">
        <button className="close-button" onClick={onClose}>×</button>
        
        <h2>Booking Details</h2>
        
        <div className="detail-section">
          <div className="detail-row">
            <span className="label">Booking ID:</span>
            <span className="value">{booking.id}</span>
          </div>
          
          <div className="detail-row">
            <span className="label">Service:</span>
            <span className="value">{booking.serviceTitle || `Service #${booking.serviceID}`}</span>
          </div>
          
          <div className="detail-row">
            <span className="label">Date & Time:</span>
            <span className="value">{new Date(booking.bookingTime).toLocaleString()}</span>
          </div>
          
          <div className="detail-row">
            <span className="label">Base Price:</span>
            <span className="value">${booking.basePrice}</span>
          </div>
          
          {booking.discountID && (
            <div className="detail-row discount">
              <span className="label">Discount Applied:</span>
              <span className="value">
                {booking.discountType === 'percentage' 
                  ? `${booking.discountValue}% off` 
                  : `$${booking.discountValue} off`}
              </span>
            </div>
          )}
          
          <div className="detail-row total">
            <span className="label">Final Price:</span>
            <span className="value">${booking.finalPrice}</span>
          </div>
          
          <div className="detail-row">
            <span className="label">Status:</span>
            <span className={`status-badge ${booking.status}`}>
              {booking.status.toUpperCase()}
            </span>
          </div>
          
          <div className="detail-row">
            <span className="label">Created:</span>
            <span className="value">{new Date(booking.createdAt).toLocaleString()}</span>
          </div>
          
          <div className="detail-row">
            <span className="label">Last Updated:</span>
            <span className="value">{new Date(booking.updatedAt).toLocaleString()}</span>
          </div>
        </div>

        {getStatusActions()}

        {review && (
          <div className="review-section">
            <h3>Your Review</h3>
            <div className="review-display">
              <div className="review-rating">
                {'⭐'.repeat(review.rating)} ({review.rating}/5)
              </div>
              {review.comment && <p className="review-comment">"{review.comment}"</p>}
              <div className="review-date">
                Posted on {new Date(review.createdAt).toLocaleDateString()}
              </div>
            </div>
          </div>
        )}

        {showReviewForm && (
          <div className="review-form">
            <h3>Write a Review</h3>
            <form onSubmit={handleSubmitReview}>
              <div className="form-group">
                <label>Rating:</label>
                <div className="rating-input">
                  {[1, 2, 3, 4, 5].map(star => (
                    <button
                      key={star}
                      type="button"
                      onClick={() => setReviewData({ ...reviewData, rating: star })}
                      className={`star ${reviewData.rating >= star ? 'active' : ''}`}
                    >
                      ★
                    </button>
                  ))}
                </div>
              </div>
              
              <div className="form-group">
                <label>Comment (Optional):</label>
                <textarea
                  value={reviewData.comment}
                  onChange={(e) => setReviewData({ ...reviewData, comment: e.target.value })}
                  rows="4"
                  placeholder="Share your experience..."
                />
              </div>
              
              <div className="form-actions">
                <button type="submit" className="btn-submit-review">Submit Review</button>
                <button type="button" onClick={() => setShowReviewForm(false)} className="btn-cancel">
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

export default BookingDetail;