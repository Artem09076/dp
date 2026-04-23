import React, { useState, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import bookingAPI from '../api/booking';
import coreAPI from '../api/core';
import './BookingDetailPage.css';

const PerformerBookingDetailPage = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  
  const [booking, setBooking] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [serviceTitle, setServiceTitle] = useState('');
  const [clientInfo, setClientInfo] = useState(null);

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

  const loadClientInfo = async (clientId) => {
    try {
      const user = await coreAPI.getUserById(clientId);
      return user;
    } catch (err) {
      console.error('Failed to load client info:', err);
      return null;
    }
  };

  const loadBookingDetail = async () => {
    try {
      setLoading(true);
      const data = await bookingAPI.getBooking(id);
      console.log('Booking details:', data);
      
      const title = await loadServiceTitle(data.service_id);
      setServiceTitle(title);
      
      const client = await loadClientInfo(data.client_id);
      setClientInfo(client);
      
      setBooking(data);
    } catch (err) {
      setError('Не удалось загрузить информацию о бронировании');
      console.error(err);
    } finally {
      setLoading(false);
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

  const handleConfirmBooking = async () => {
    try {
      await bookingAPI.submitBooking(id);
      await loadBookingDetail();
      alert('Бронирование подтверждено');
    } catch (err) {
      setError('Не удалось подтвердить бронирование');
    }
  };

  const handleCompleteBooking = async () => {
    if (window.confirm('Отметить это бронирование как выполненное?')) {
      try {
        await bookingAPI.submitBooking(id);
        await loadBookingDetail();
        alert('Бронирование отмечено как выполненное');
      } catch (err) {
        setError('Не удалось отметить бронирование');
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
  const canConfirm = booking.status === 'pending' && !isExpired;
  const canComplete = booking.status === 'confirmed';

  return (
    <div className="booking-detail-page performer-page">
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

        {clientInfo && (
          <div className="client-info-section">
            <h2>Информация о клиенте</h2>
            <div className="info-row">
              <span className="label">Имя:</span>
              <span className="value">{clientInfo.name}</span>
            </div>
            <div className="info-row">
              <span className="label">Email:</span>
              <span className="value">{clientInfo.email}</span>
            </div>
          </div>
        )}

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
          {canConfirm && (
            <button onClick={handleConfirmBooking} className="btn-confirm-booking">
              Подтвердить бронирование
            </button>
          )}
          {canComplete && (
            <button onClick={handleCompleteBooking} className="btn-complete-booking">
              Отметить как выполненное
            </button>
          )}
          {canConfirm && (
            <button onClick={handleCancelBooking} className="btn-cancel-booking">
              Отменить бронирование
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default PerformerBookingDetailPage;