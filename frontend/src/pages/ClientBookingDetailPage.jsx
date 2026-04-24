import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import bookingAPI from '../api/booking';
import coreAPI from '../api/core';
import './BookingDetailPage.css';

const ClientBookingDetailPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  
  const [booking, setBooking] = useState(null);
  const [review, setReview] = useState(null);
  const [showReviewForm, setShowReviewForm] = useState(false);
  const [showEditForm, setShowEditForm] = useState(false);
  const [reviewData, setReviewData] = useState({ rating: 5, comment: '' });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [serviceTitle, setServiceTitle] = useState('');
  const [loadingReview, setLoadingReview] = useState(false);

  useEffect(() => {
    loadBookingDetail();
  }, [id]);

  const loadServiceTitle = async (serviceId) => {
    try {
      const service = await coreAPI.getService(serviceId);
      return service?.title || `Сервис #${serviceId.slice(0, 8)}`;
    } catch (err) {
      console.error('Failed to load service:', err);
      return `Сервис #${serviceId.slice(0, 8)}`;
    }
  };

  const loadBookingDetail = async () => {
    try {
      setLoading(true);
      const data = await bookingAPI.getBooking(id);
      console.log('Booking details:', data);
      
      const title = await loadServiceTitle(data.service_id);
      setServiceTitle(title);
      setBooking(data);
      
      await loadReview();
      
    } catch (err) {
      setError('Не удалось загрузить информацию о бронировании');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const loadReview = async () => {
    try {
      setLoadingReview(true);
      const reviewData = await coreAPI.getReviewByBooking(id);
      console.log('Review data:', reviewData);
      
      if (reviewData && reviewData.id && !reviewData.error) {
        setReview(reviewData);
        setReviewData({
          rating: reviewData.rating,
          comment: reviewData.comment || ''
        });
      } else {
        setReview(null);
      }
    } catch (err) {
      console.log('No review yet for this booking');
      setReview(null);
    } finally {
      setLoadingReview(false);
    }
  };

  const handleCancelBooking = async () => {
    if (window.confirm('Вы уверены, что хотите отменить это бронирование?')) {
      try {
        await bookingAPI.cancelBooking(id);
        await loadBookingDetail();
        alert('Бронирование успешно отменено');
      } catch (err) {
        setError('Не удалось отменить бронирование');
      }
    }
  };

  const handleSubmitReview = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.createReview({
        booking_id: id,
        rating: reviewData.rating,
        comment: reviewData.comment
      });
      await loadReview();
      setShowReviewForm(false);
      setShowEditForm(false);
      setReviewData({ rating: 5, comment: '' });
      alert('Отзыв успешно отправлен');
    } catch (err) {
      console.error('Review error:', err);
      setError('Не удалось отправить отзыв. Попробуйте позже.');
    }
  };

  const handleUpdateReview = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.updateReview(review.id, {
        rating: reviewData.rating,
        comment: reviewData.comment
      });
      await loadReview();
      setShowEditForm(false);
      alert('Отзыв успешно обновлен');
    } catch (err) {
      console.error('Update review error:', err);
      setError('Не удалось обновить отзыв');
    }
  };

  const handleDeleteReview = async () => {
    if (window.confirm('Вы уверены, что хотите удалить этот отзыв?')) {
      try {
        await coreAPI.deleteReview(review.id);
        setReview(null);
        alert('Отзыв успешно удален');
      } catch (err) {
        console.error('Delete review error:', err);
        setError('Не удалось удалить отзыв');
      }
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'Дата не указана';
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return 'Неверная дата';
    
    return date.toLocaleString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  };

  const formatPrice = (price) => {
    if (!price && price !== 0) return '0';
    if (price > 10000) {
      return (price / 100).toFixed(2);
    }
    return price.toString();
  };

  const getStatusText = (status) => {
    const statuses = {
      pending: 'Ожидает подтверждения',
      confirmed: 'Подтверждено',
      completed: 'Завершено',
      cancelled: 'Отменено',
    };
    return statuses[status?.toLowerCase()] || status || 'Неизвестно';
  };

  const getStatusClass = (status) => {
    const statuses = {
      pending: 'pending',
      confirmed: 'confirmed',
      completed: 'completed',
      cancelled: 'cancelled',
    };
    return statuses[status?.toLowerCase()] || 'pending';
  };

  const canReview = (booking?.status?.toLowerCase() === 'completed' || booking?.status?.toLowerCase() === 'cancelled') && !review;
  const canCancel = booking?.status?.toLowerCase() === 'pending';

  if (loading) {
    return (
      <div className="booking-detail-page">
        <div className="loading-container">
          <div className="loader"></div>
          <p>Загрузка информации о бронировании...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="booking-detail-page">
        <div className="error-container">
          <h2>Ошибка</h2>
          <p>{error}</p>
          <button onClick={() => navigate('/bookings')} className="btn-back">
            ← Вернуться к списку
          </button>
        </div>
      </div>
    );
  }

  if (!booking) {
    return (
      <div className="booking-detail-page">
        <div className="error-container">
          <h2>Бронирование не найдено</h2>
          <button onClick={() => navigate('/bookings')} className="btn-back">
            ← Вернуться к списку
          </button>
        </div>
      </div>
    );
  }

  const statusClass = getStatusClass(booking.status);
  const statusText = getStatusText(booking.status);
  const isExpired = new Date(booking.booking_time) < new Date() && booking.status !== 'completed' && booking.status !== 'cancelled';

  return (
    <div className="booking-detail-page">
      <div className="booking-detail-container">
        <button onClick={() => navigate('/bookings')} className="btn-back">
          ← Вернуться к списку
        </button>
        
        <div className="booking-detail-header">
          <h1>Детали бронирования</h1>
          <div className={`status-badge ${statusClass}`}>
            {statusText}
            {isExpired && <span className="expired-badge"> (Просрочено)</span>}
          </div>
        </div>

        <div className="booking-info-section">
          <h2>Информация о бронировании</h2>
          <div className="info-row">
            <span className="label">Услуга:</span>
            <span className="value">{serviceTitle}</span>
          </div>
          <div className="info-row">
            <span className="label">Дата и время:</span>
            <span className="value">{formatDate(booking.booking_time)}</span>
          </div>
          {isExpired && booking.status === 'pending' && (
            <div className="warning-message">
              ⚠️ Время этого бронирования уже прошло
            </div>
          )}
        </div>

        <div className="price-section">
          <h2>Стоимость</h2>
          <div className="info-row">
            <span className="label">Базовая цена:</span>
            <span className="value">{formatPrice(booking.base_price)} ₽</span>
          </div>
          {booking.discount_id && (
            <div className="info-row discount">
              <span className="label">Скидка:</span>
              <span className="value">
                {booking.discount_type === 'percentage' 
                  ? `${booking.discount_value}%` 
                  : `${formatPrice(booking.discount_value)} ₽`}
              </span>
            </div>
          )}
          <div className="info-row total">
            <span className="label">Итоговая цена:</span>
            <span className="value">{formatPrice(booking.final_price)} ₽</span>
          </div>
        </div>

        <div className="actions-section">
          {canCancel && (
            <button onClick={handleCancelBooking} className="btn-cancel-booking">
              Отменить бронирование
            </button>
          )}
        </div>

        {review && !showEditForm && (
          <div className="review-section">
            <div className="review-header-with-buttons">
              <h2>Ваш отзыв</h2>
              <div className="review-actions-buttons">
                <button onClick={() => {
                  setShowEditForm(true);
                  setShowReviewForm(false);
                }} className="btn-edit-review">
                  ✏️ Редактировать
                </button>
                <button onClick={handleDeleteReview} className="btn-delete-review">
                  🗑️ Удалить
                </button>
              </div>
            </div>
            <div className="review-card">
              <div className="review-rating">
                {'⭐'.repeat(review.rating)} ({review.rating}/5)
              </div>
              {review.comment && <p className="review-comment">"{review.comment}"</p>}
              <div className="review-date">
                Оставлен {formatDate(review.created_at)}
              </div>
            </div>
          </div>
        )}

        {showReviewForm && (
          <div className="review-form-section">
            <h2>Оставить отзыв</h2>
            <form onSubmit={handleSubmitReview}>
              <div className="form-group">
                <label>Оценка:</label>
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
                <label>Комментарий (необязательно):</label>
                <textarea
                  value={reviewData.comment}
                  onChange={(e) => setReviewData({ ...reviewData, comment: e.target.value })}
                  rows="4"
                  placeholder="Поделитесь впечатлением о услуге..."
                />
              </div>
              
              <div className="form-actions">
                <button type="submit" className="btn-submit-review">Отправить отзыв</button>
                <button type="button" onClick={() => setShowReviewForm(false)} className="btn-cancel">
                  Отмена
                </button>
              </div>
            </form>
          </div>
        )}

        {showEditForm && (
          <div className="review-form-section">
            <h2>Редактировать отзыв</h2>
            <form onSubmit={handleUpdateReview}>
              <div className="form-group">
                <label>Оценка:</label>
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
                <label>Комментарий (необязательно):</label>
                <textarea
                  value={reviewData.comment}
                  onChange={(e) => setReviewData({ ...reviewData, comment: e.target.value })}
                  rows="4"
                  placeholder="Поделитесь впечатлением о услуге..."
                />
              </div>
              
              <div className="form-actions">
                <button type="submit" className="btn-submit-review">Сохранить изменения</button>
                <button type="button" onClick={() => setShowEditForm(false)} className="btn-cancel">
                  Отмена
                </button>
              </div>
            </form>
          </div>
        )}

        {canReview && !showReviewForm && !showEditForm && (
          <div className="review-prompt">
            <button onClick={() => setShowReviewForm(true)} className="btn-write-review">
              ✍️ Написать отзыв
            </button>
          </div>
        )}

        {loadingReview && (
          <div className="loading-review">
            <div className="spinner-small"></div>
            <p>Загрузка отзыва...</p>
          </div>
        )}
      </div>
    </div>
  );
};

export default ClientBookingDetailPage;