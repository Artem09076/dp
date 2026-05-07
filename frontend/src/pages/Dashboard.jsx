import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';
import ServiceList from '../components/services/ServiceList';
import bookingAPI from '../api/booking';
import coreAPI from '../api/core';
import './Dashboard.css';

const Dashboard = () => {
  const navigate = useNavigate();
  const { userRole } = useAuth();
  const [todayBookings, setTodayBookings] = useState([]);
  const [upcomingBookings, setUpcomingBookings] = useState([]);
  const [stats, setStats] = useState({
    totalBookings: 0,
    completedToday: 0,
    pendingBookings: 0,
    totalServices: 0
  });
  const [loading, setLoading] = useState(true);
  const [serviceCache, setServiceCache] = useState({});
  const [userCache, setUserCache] = useState({});

  useEffect(() => {
    if (userRole === 'performer') {
      loadPerformerDashboard();
    } else {
      setLoading(false);
    }
  }, [userRole]);

  // Загрузка названия услуги
  const loadServiceTitle = async (serviceId) => {
    if (!serviceId) return 'Услуга не указана';
    
    if (serviceCache[serviceId]) {
      return serviceCache[serviceId];
    }
    
    try {
      const service = await coreAPI.getService(serviceId);
      const title = service?.title || `Услуга #${serviceId.slice(0, 8)}`;
      setServiceCache(prev => ({ ...prev, [serviceId]: title }));
      return title;
    } catch (err) {
      console.error('Failed to load service:', serviceId, err);
      return `Услуга #${serviceId.slice(0, 8)}`;
    }
  };

  // Загрузка имени клиента
  const loadClientName = async (clientId) => {
    if (!clientId) return 'Клиент не указан';
    
    if (userCache[clientId]) {
      return userCache[clientId];
    }
    
    try {
      const user = await coreAPI.getUserById(clientId);
      const name = user?.name || `Клиент #${clientId.slice(0, 8)}`;
      setUserCache(prev => ({ ...prev, [clientId]: name }));
      return name;
    } catch (err) {
      console.error('Failed to load client:', clientId, err);
      return `Клиент #${clientId.slice(0, 8)}`;
    }
  };

  const loadPerformerDashboard = async () => {
    try {
      setLoading(true);
      
      // Получаем все бронирования исполнителя
      const bookings = await bookingAPI.getBookings();
      const bookingsArray = Array.isArray(bookings) ? bookings : [];
      
      // Получаем услуги исполнителя
      const services = await coreAPI.getMyServices();
      const servicesArray = Array.isArray(services) ? services : [];
      
      const today = new Date();
      today.setHours(0, 0, 0, 0);
      const tomorrow = new Date(today);
      tomorrow.setDate(tomorrow.getDate() + 1);
      const endOfWeek = new Date(today);
      endOfWeek.setDate(endOfWeek.getDate() + 7);
      
      // Бронирования на сегодня
      const todayList = bookingsArray.filter(booking => {
        const bookingDate = new Date(booking.booking_time);
        return bookingDate >= today && bookingDate < tomorrow && booking.status !== 'cancelled';
      });
      
      // Предстоящие бронирования
      const upcomingList = bookingsArray.filter(booking => {
        const bookingDate = new Date(booking.booking_time);
        return bookingDate >= tomorrow && bookingDate <= endOfWeek && booking.status !== 'cancelled';
      });
      
      // Загружаем названия услуг и имена клиентов для бронирований
      const todayWithDetails = await Promise.all(todayList.map(async (booking) => ({
        ...booking,
        service_title: await loadServiceTitle(booking.service_id),
        client_name: await loadClientName(booking.client_id)
      })));
      
      const upcomingWithDetails = await Promise.all(upcomingList.map(async (booking) => ({
        ...booking,
        service_title: await loadServiceTitle(booking.service_id),
        client_name: await loadClientName(booking.client_id)
      })));
      
      // Статистика
      const completedTodayCount = bookingsArray.filter(booking => {
        const bookingDate = new Date(booking.booking_time);
        return bookingDate >= today && bookingDate < tomorrow && booking.status === 'completed';
      }).length;
      
      const pendingCount = bookingsArray.filter(booking => booking.status === 'pending').length;
      
      setTodayBookings(todayWithDetails);
      setUpcomingBookings(upcomingWithDetails);
      setStats({
        totalBookings: bookingsArray.length,
        completedToday: completedTodayCount,
        pendingBookings: pendingCount,
        totalServices: servicesArray.length
      });
      
    } catch (err) {
      console.error('Failed to load performer dashboard:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatTime = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });
  };

  const getStatusText = (status) => {
    const statuses = {
      pending: '⏳ Ожидает',
      confirmed: '✅ Подтверждено',
      completed: '🎉 Завершено',
      cancelled: '❌ Отменено'
    };
    return statuses[status] || status;
  };

  if (userRole === 'performer') {
    if (loading) {
      return <div className="loading">Загрузка дашборда...</div>;
    }

    return (
      <div className="dashboard performer-dashboard">
        <div className="dashboard-header">
          <h1>С возвращением</h1>
          <p>Управляйте своими услугами и бронированиями</p>
        </div>

        <div className="stats-grid">
          <div className="stat-card">
            <div className="stat-icon">📊</div>
            <div className="stat-info">
              <h3>Всего бронирований</h3>
              <p className="stat-number">{stats.totalBookings}</p>
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-icon">✅</div>
            <div className="stat-info">
              <h3>Выполнено сегодня</h3>
              <p className="stat-number">{stats.completedToday}</p>
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-icon">⏳</div>
            <div className="stat-info">
              <h3>Ожидают подтверждения</h3>
              <p className="stat-number">{stats.pendingBookings}</p>
            </div>
          </div>
          <div className="stat-card">
            <div className="stat-icon">🛠️</div>
            <div className="stat-info">
              <h3>Мои услуги</h3>
              <p className="stat-number">{stats.totalServices}</p>
            </div>
          </div>
        </div>

        <div className="quick-actions-section">
          <h3>Быстрые действия</h3>
          <div className="quick-actions">
            <button onClick={() => navigate('/my-services')} className="action-btn">
              <span className="action-icon">➕</span>
              Управление услугами
            </button>
            <button onClick={() => navigate('/bookings')} className="action-btn">
              <span className="action-icon">📋</span>
              Все бронирования
            </button>
          </div>
        </div>

        <div className="today-bookings-section">
          <h3>📅 Бронирования на сегодня</h3>
          {todayBookings.length === 0 ? (
            <div className="empty-state">
              <p>На сегодня нет бронирований</p>
            </div>
          ) : (
            <div className="bookings-list">
              {todayBookings.map(booking => (
                <div key={booking.id} className="booking-item">
                  <div className="booking-time">{formatTime(booking.booking_time)}</div>
                  <div className="booking-info">
                    <div className="booking-service">{booking.service_title}</div>
                  </div>
                  <div className={`booking-status ${booking.status}`}>
                    {getStatusText(booking.status)}
                  </div>
                  <button 
                    onClick={() => navigate(`/bookings/performer/${booking.id}`)} 
                    className="btn-view"
                  >
                    Подробнее
                  </button>
                </div>
              ))}
            </div>
          )}
        </div>

        {upcomingBookings.length > 0 && (
          <div className="upcoming-bookings-section">
            <h3>🗓️ Предстоящие бронирования на неделю</h3>
            <div className="bookings-list">
              {upcomingBookings.map(booking => (
                <div key={booking.id} className="booking-item">
                  <div className="booking-date">{formatDate(booking.booking_time)}</div>
                  <div className="booking-time">{formatTime(booking.booking_time)}</div>
                  <div className="booking-info">
                    <div className="booking-service">{booking.service_title}</div>
                  </div>
                  <div className={`booking-status ${booking.status}`}>
                    {getStatusText(booking.status)}
                  </div>
                  <button 
                    onClick={() => navigate(`/bookings/performer/${booking.id}`)} 
                    className="btn-view"
                  >
                    Подробнее
                  </button>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    );
  }

  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>Добро пожаловать на платформу по бронированию услуг</h1>
        <p>Найдите и забронируйте услугу от проверенных поставщиков</p>
      </div>

      <div className="search-section">
        <h2>Поиск сервисов</h2>
        <ServiceList />
      </div>
    </div>
  );
};

export default Dashboard;