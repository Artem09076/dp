import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import bookingAPI from '../../api/booking';
import coreAPI from '../../api/core';
import './Booking.css';
import { useAuth } from '../../contexts/AuthContext';

const BookingList = () => {
  const navigate = useNavigate();
  const [bookings, setBookings] = useState([]);
  const [filteredBookings, setFilteredBookings] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [filter, setFilter] = useState('all');
  const [sortBy, setSortBy] = useState('status');
  const [serviceCache, setServiceCache] = useState({});
  const { userRole } = useAuth();


  useEffect(() => {
    loadBookings();
  }, []);

  useEffect(() => {
    filterAndSortBookings();
  }, [bookings, filter, sortBy]);

  const loadServiceTitle = async (serviceId) => {
    if (!serviceId) return 'Unknown Service';
    
    if (serviceCache[serviceId]) {
      return serviceCache[serviceId];
    }
    
    try {
      const service = await coreAPI.getService(serviceId);
      const title = service?.title || `Service #${serviceId.slice(0, 8)}`;
      setServiceCache(prev => ({ ...prev, [serviceId]: title }));
      return title;
    } catch (err) {
      console.error('Failed to load service:', serviceId, err);
      return `Service #${serviceId.slice(0, 8)}`;
    }
  };

  const loadBookings = async () => {
    try {
      setLoading(true);
      const data = await bookingAPI.getBookings();
      console.log('Raw bookings data:', data);
      
      const bookingsArray = Array.isArray(data) ? data : [];
      
      const formattedBookings = await Promise.all(bookingsArray.map(async (booking) => {
        const serviceTitle = await loadServiceTitle(booking.service_id);
        return {
          id: booking.id,
          serviceID: booking.service_id,
          serviceTitle: serviceTitle,
          clientID: booking.client_id,
          performerID: booking.performer_id,
          basePrice: booking.base_price || 0,
          finalPrice: booking.final_price || 0,
          bookingTime: booking.booking_time,
          status: booking.status,
          discountID: booking.discount_id,
          createdAt: booking.created_at,
          updatedAt: booking.updated_at
        };
      }));
      
      setBookings(formattedBookings);
      
    } catch (err) {
      setError('Failed to load bookings');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const filterAndSortBookings = () => {
    let filtered = [...bookings];
    
    if (filter !== 'all') {
      filtered = filtered.filter(booking => booking.status === filter);
    }
    
    const now = new Date();
    filtered = filtered.filter(booking => {
      const bookingDate = new Date(booking.bookingTime);
      return bookingDate > now || booking.status === 'completed' || booking.status === 'cancelled';
    });
    
    filtered.sort((a, b) => {
      if (sortBy === 'status') {
        const statusOrder = { pending: 1, confirmed: 2, completed: 3, cancelled: 4 };
        const orderA = statusOrder[a.status] || 5;
        const orderB = statusOrder[b.status] || 5;
        
        if (orderA !== orderB) {
          return orderA - orderB;
        }
        return new Date(a.bookingTime) - new Date(b.bookingTime);
      } else {
        return new Date(a.bookingTime) - new Date(b.bookingTime);
      }
    });
    
    setFilteredBookings(filtered);
  };

  const handleCancel = async (id) => {
    if (window.confirm('Are you sure you want to cancel this booking?')) {
      try {
        await bookingAPI.cancelBooking(id);
        await loadBookings();
        alert('Booking cancelled successfully');
      } catch (err) {
        setError('Failed to cancel booking');
      }
    }
  };

  const handleBookingClick = (bookingId) => {
    if (userRole === 'performer') {
      navigate(`/bookings/performer/${bookingId}`);
    } else {
      navigate(`/bookings/client/${bookingId}`);
    }
  };

  const handleSubmit = async (id) => {
    try {
      await bookingAPI.submitBooking(id);
      await loadBookings();
      alert('Booking marked as completed');
    } catch (err) {
      setError('Failed to submit booking');
    }
  };

  const handleCompleted = async (id) => {
    try {
      await bookingAPI.completedBooking(id);
      await loadBookings();
      alert('Booking marked as completed');
    } catch (err) {
      setError('Failed to submit booking');
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

  const formatDateOnly = (dateString) => {
    if (!dateString) return 'Дата не указана';
    const date = new Date(dateString);
    if (isNaN(date.getTime())) return 'Неверная дата';
    
    return date.toLocaleDateString('ru-RU', {
      year: 'numeric',
      month: 'long',
      day: 'numeric'
    });
  };

  const formatPrice = (price) => {
    if (!price && price !== 0) return '0';
    if (price > 10000) {
      return (price / 100).toFixed(2);
    }
    return price.toString();
  };

  const getStatusBadge = (status) => {
    const statuses = {
      pending: { text: '⏳ Ожидает', class: 'pending' },
      confirmed: { text: '✅ Подтверждено', class: 'confirmed' },
      completed: { text: '🎉 Завершено', class: 'completed' },
      cancelled: { text: '❌ Отменено', class: 'cancelled' },
    };
    const s = statuses[status?.toLowerCase()] || { text: status || 'Неизвестно', class: 'pending' };
    return s;
  };

  const isBookingExpired = (bookingTime, status) => {
    const now = new Date();
    const bookingDate = new Date(bookingTime);
    return bookingDate < now && status !== 'completed' && status !== 'cancelled';
  };

  if (loading) return <div className="loading">Загрузка бронирований...</div>;
  if (error) return <div className="error-message">{error}</div>;

  return (
    <div className="booking-list">
      <div className="booking-filters">
        <div className="filter-group">
          <label>Фильтр по статусу:</label>
          <select value={filter} onChange={(e) => setFilter(e.target.value)}>
            <option value="all">Все бронирования</option>
            <option value="pending">Ожидающие</option>
            <option value="confirmed">Подтвержденные</option>
            <option value="completed">Завершенные</option>
            <option value="cancelled">Отмененные</option>
          </select>
        </div>
        
        <div className="filter-group">
          <label>Сортировка:</label>
          <select value={sortBy} onChange={(e) => setSortBy(e.target.value)}>
            <option value="status">По статусу</option>
            <option value="date">По дате (сначала ближайшие)</option>
          </select>
        </div>
      </div>

      <div className="bookings-stats">
        <span>Всего: {filteredBookings.length}</span>
        <span>Ожидают: {filteredBookings.filter(b => b.status === 'pending').length}</span>
        <span>Подтверждено: {filteredBookings.filter(b => b.status === 'confirmed').length}</span>
        <span>Завершено: {filteredBookings.filter(b => b.status === 'completed').length}</span>
      </div>

      {filteredBookings.length === 0 ? (
        <p className="no-bookings">Бронирования не найдены</p>
      ) : (
        <div className="bookings-grid">
          {filteredBookings.map((booking) => {
            const status = getStatusBadge(booking.status);
            const isExpired = isBookingExpired(booking.bookingTime, booking.status);
            
            return (
              <div key={booking.id} className={`booking-card ${isExpired ? 'expired' : ''}`} onClick={() => handleBookingClick(booking.id)} style={{cursor: "pointer"}}>
                <div className="booking-header">
                  <h3>{booking.serviceTitle}</h3>
                  <span className={`status ${status.class}`}>
                    {status.text}
                    {isExpired && <span className="expired-badge"> (Просрочено)</span>}
                  </span>
                </div>
                <div className="booking-details">
                  <p><strong>Дата и время:</strong> {formatDate(booking.bookingTime)}</p>
                  <p><strong>Базовая цена:</strong> {formatPrice(booking.basePrice)} ₽</p>
                  <p><strong>Итоговая цена:</strong> {formatPrice(booking.finalPrice)} ₽</p>
                  {booking.discountID && (
                    <p className="discount">🎉 Применена скидка!</p>
                  )}
                </div>
                <div className="booking-actions">
                  {booking.status === 'pending' && !isExpired && (
                    <button onClick={() => handleCancel(booking.id)} className="btn-cancel">
                      Отменить бронирование
                    </button>
                  )}
                  {booking.status === 'confirmed' && (
                    <button onClick={() => handleCompleted(booking.id)} className="btn-submit">
                      Отметить как выполненное
                    </button>
                  )}
                  {isExpired && booking.status === 'pending' && (
                    <div className="expired-message">
                      ⚠️ Время бронирования прошло
                    </div>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default BookingList;