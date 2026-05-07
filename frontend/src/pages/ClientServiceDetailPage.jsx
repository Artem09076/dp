import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import coreAPI from '../api/core';
import bookingAPI from '../api/booking';
import { useAuth } from '../contexts/AuthContext';
import BookingCreate from '../components/bookings/BookingCreate';
import './ServiceDetailPage.css';

const ClientServiceDetailPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user, userRole } = useAuth();
  
  const [service, setService] = useState(null);
  const [reviews, setReviews] = useState([]);
  const [discounts, setDiscounts] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showBookingModal, setShowBookingModal] = useState(false);
  const [averageRating, setAverageRating] = useState(0);

  useEffect(() => {
    if (userRole === 'performer') {
      navigate(`/services/performer/${id}`, { replace: true });
    }
  }, [userRole, id, navigate]);

  useEffect(() => {
    if (userRole !== 'performer') {
      loadServiceDetails();
    }
  }, [id, userRole]);

  const loadServiceDetails = async () => {
    try {
      setLoading(true);
      setError('');
      
      const serviceData = await coreAPI.getService(id);
      setService(serviceData);
      
      const reviewsData = await coreAPI.getServiceReviews(id);
      const reviewsArray = Array.isArray(reviewsData) ? reviewsData : [];
      setReviews(reviewsArray);
      
      if (reviewsArray.length > 0) {
        const sum = reviewsArray.reduce((acc, review) => acc + review.rating, 0);
        setAverageRating(sum / reviewsArray.length);
      }
      
      try {
        const discountsData = await coreAPI.getServiceDiscounts(id);
        setDiscounts(Array.isArray(discountsData) ? discountsData : []);
      } catch (err) {
        console.error('Failed to load discounts:', err);
        setDiscounts([]);
      }
      
    } catch (err) {
      console.error('Failed to load service:', err);
      setError('Service not found');
    } finally {
      setLoading(false);
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

  const parseDate = (dateString) => {
    if (!dateString) return null;
    try {
      const parts = dateString.split(' ');
      if (parts.length >= 3) {
        const datePart = parts[0];
        const timePart = parts[1].split('.')[0];
        const offsetPart = parts[2];
        const isoString = `${datePart}T${timePart}${offsetPart}`;
        const date = new Date(isoString);
        if (!isNaN(date.getTime())) return date;
      }
      const simpleDate = new Date(dateString);
      if (!isNaN(simpleDate.getTime())) return simpleDate;
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

  if (userRole === 'performer') {
    return (
      <div className="loading-container">
        <div className="loader"></div>
        <p>Перенаправление...</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="service-detail-page">
        <div className="loading-container">
          <div className="loader"></div>
          <p>Загрузка информации об услуге...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="service-detail-page">
        <div className="error-container">
          <h2>Ошибка</h2>
          <p>{error}</p>
          <button onClick={() => navigate('/')} className="btn-back">
            ← На главную
          </button>
        </div>
      </div>
    );
  }

  if (!service) {
    return (
      <div className="service-detail-page">
        <div className="error-container">
          <h2>Услуга не найдена</h2>
          <button onClick={() => navigate('/')} className="btn-back">
            ← На главную
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="service-detail-page">
      <div className="service-detail-container">
        <button onClick={() => navigate('/')} className="btn-back">
          ← Назад к поиску
        </button>
        
        <div className="service-detail-header">
          <h1>{service.title}</h1>
          <div className="service-meta">
            <div className="rating-large">
              {'⭐'.repeat(Math.round(averageRating))}
              <span className="rating-value">
                {averageRating > 0 ? ` ${averageRating.toFixed(1)}` : ' Нет отзывов'}
              </span>
            </div>
            <div className="price-large">{service.price}₽</div>
            <div className="duration-large">⏱️ {service.duration_minutes} минут</div>
          </div>
        </div>

        {service.description && (
          <div className="service-description">
            <h2>Описание</h2>
            <p>{service.description}</p>
          </div>
        )}

        {discounts.length > 0 && (
          <div className="service-discounts">
            <h2>Доступные скидки</h2>
            <div className="discounts-list">
              {discounts.map(discount => (
                <div key={discount.id} className="discount-card">
                  <div className="discount-badge">
                    {discount.type === 'percentage' ? `${discount.value}% OFF` : `${discount.value}₽ OFF`}
                  </div>
                  <div className="discount-info">
                    <p>Действительна до {formatDate(discount.valid_to)}</p>
                    <p>Использована {discount.used_count} из {discount.max_uses} раз</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="booking-action">
          <button onClick={() => setShowBookingModal(true)} className="btn-book-now">
            Забронировать
          </button>
        </div>

        {reviews.length > 0 && (
          <div className="service-reviews">
            <h2>Отзывы ({reviews.length})</h2>
            <div className="reviews-list">
              {reviews.map(review => (
                <div key={review.id} className="review-card">
                  <div className="review-header">
                    <div className="review-rating">
                      {'⭐'.repeat(review.rating)} ({review.rating}/5)
                    </div>
                    <div className="review-date">{formatDate(review.created_at)}</div>
                  </div>
                  {review.comment && (
                    <p className="review-comment">"{review.comment}"</p>
                  )}
                  <div className="review-author">
                    — Пользователь
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>

      {showBookingModal && (
        <BookingCreate 
          service={service}
          onSuccess={() => {
            setShowBookingModal(false);
            alert('Бронирование успешно создано!');
          }}
          onCancel={() => setShowBookingModal(false)}
        />
      )}
    </div>
  );
};

export default ClientServiceDetailPage;