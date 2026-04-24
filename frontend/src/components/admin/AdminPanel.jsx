import React, { useState, useEffect } from 'react';
import coreAPI from '../../api/core';
import './Admin.css';

const AdminPanel = () => {
  const [activeTab, setActiveTab] = useState('performers');
  const [performers, setPerformers] = useState([]);
  const [users, setUsers] = useState([]);
  const [bookings, setBookings] = useState([]);
  const [services, setServices] = useState([]);
  const [reviews, setReviews] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [editItem, setEditItem] = useState(null);
  const [editType, setEditType] = useState('');
  const [createType, setCreateType] = useState('user');
  const [createFormData, setCreateFormData] = useState({
    email: '',
    name: '',
    password: '',
    role: 'client',
    inn: '',
    business_type: '',
    title: '',
    description: '',
    price: '',
    duration_minutes: '',
    service_id: '',
    rating: 5,
    comment: '',
    booking_id: ''
  });
  
  const [pagination, setPagination] = useState({
    performers: { page: 1, pageSize: 20, total: 0 },
    users: { page: 1, pageSize: 20, total: 0 },
    bookings: { page: 1, pageSize: 20, total: 0 },
    services: { page: 1, pageSize: 20, total: 0 },
    reviews: { page: 1, pageSize: 20, total: 0 }
  });
  
  const [filters, setFilters] = useState({
    performers: { name: '', email: '', inn: '', business_type: '', verification_status: '' },
    users: { name: '', email: '', role: '', verification_status: '' },
    bookings: { status: '', client_name: '', service_name: '' },
    services: { title: '', performer_name: '', min_price: '', max_price: '' },
    reviews: { client_name: '', rating: '', comment: '' }
  });

  const [debouncedFilters, setDebouncedFilters] = useState(filters);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebouncedFilters(filters);
    }, 500);
    return () => clearTimeout(timer);
  }, [filters]);

  useEffect(() => {
    loadData();
  }, [activeTab, pagination[activeTab].page, debouncedFilters]);

  const loadData = async () => {
    setLoading(true);
    try {
      const currentPagination = pagination[activeTab];
      const currentFilters = debouncedFilters[activeTab];
      
      if (activeTab === 'performers') {
        const data = await coreAPI.getUnverifiedPerformers(currentPagination.page, currentPagination.pageSize);
        let filteredData = data.data || [];
        
        if (currentFilters.name) {
          filteredData = filteredData.filter(p => p.name?.toLowerCase().includes(currentFilters.name.toLowerCase()));
        }
        if (currentFilters.email) {
          filteredData = filteredData.filter(p => p.email?.toLowerCase().includes(currentFilters.email.toLowerCase()));
        }
        if (currentFilters.inn) {
          filteredData = filteredData.filter(p => p.inn?.includes(currentFilters.inn));
        }
        if (currentFilters.business_type) {
          filteredData = filteredData.filter(p => p.business_type === currentFilters.business_type);
        }
        if (currentFilters.verification_status) {
          filteredData = filteredData.filter(p => p.verification_status === currentFilters.verification_status);
        }
        
        setPerformers(filteredData);
        setPagination(prev => ({
          ...prev,
          performers: { ...prev.performers, total: filteredData.length }
        }));
      } 
      else if (activeTab === 'users') {
        const data = await coreAPI.getUsers();
        let filteredData = data.data || [];
        
        if (currentFilters.name) {
          filteredData = filteredData.filter(u => u.name?.toLowerCase().includes(currentFilters.name.toLowerCase()));
        }
        if (currentFilters.email) {
          filteredData = filteredData.filter(u => u.email?.toLowerCase().includes(currentFilters.email.toLowerCase()));
        }
        if (currentFilters.role) {
          filteredData = filteredData.filter(u => u.role === currentFilters.role);
        }
        if (currentFilters.verification_status) {
          filteredData = filteredData.filter(u => u.verification_status === currentFilters.verification_status);
        }
        
        setUsers(filteredData);
        setPagination(prev => ({
          ...prev,
          users: { ...prev.users, total: filteredData.length }
        }));
      } 
      else if (activeTab === 'bookings') {
        const data = await coreAPI.getAdminBookings();
        let filteredData = data.data || [];
        
        if (currentFilters.status) {
          filteredData = filteredData.filter(b => b.status === currentFilters.status);
        }
        if (currentFilters.client_name) {
          filteredData = filteredData.filter(b => b.client_name?.toLowerCase().includes(currentFilters.client_name.toLowerCase()));
        }
        if (currentFilters.service_name) {
          filteredData = filteredData.filter(b => b.service_name?.toLowerCase().includes(currentFilters.service_name.toLowerCase()));
        }
        
        setBookings(filteredData);
        setPagination(prev => ({
          ...prev,
          bookings: { ...prev.bookings, total: filteredData.length }
        }));
      } 
      else if (activeTab === 'services') {
        const data = await coreAPI.getAdminServices('', 1, 1000);
        let filteredData = data.data || [];
        
        if (currentFilters.title) {
          filteredData = filteredData.filter(s => s.title?.toLowerCase().includes(currentFilters.title.toLowerCase()));
        }
        if (currentFilters.performer_name) {
          filteredData = filteredData.filter(s => s.performer_name?.toLowerCase().includes(currentFilters.performer_name.toLowerCase()));
        }
        if (currentFilters.min_price) {
          filteredData = filteredData.filter(s => s.price >= parseFloat(currentFilters.min_price));
        }
        if (currentFilters.max_price) {
          filteredData = filteredData.filter(s => s.price <= parseFloat(currentFilters.max_price));
        }
        
        setServices(filteredData);
        setPagination(prev => ({
          ...prev,
          services: { ...prev.services, total: filteredData.length }
        }));
      } 
      else if (activeTab === 'reviews') {
        const data = await coreAPI.getAdminReviews();
        let filteredData = data.data || [];
        
        if (currentFilters.client_name) {
          filteredData = filteredData.filter(r => r.client_name?.toLowerCase().includes(currentFilters.client_name.toLowerCase()));
        }
        if (currentFilters.rating) {
          filteredData = filteredData.filter(r => r.rating === parseInt(currentFilters.rating));
        }
        if (currentFilters.comment) {
          filteredData = filteredData.filter(r => r.comment?.toLowerCase().includes(currentFilters.comment.toLowerCase()));
        }
        
        setReviews(filteredData);
        setPagination(prev => ({
          ...prev,
          reviews: { ...prev.reviews, total: filteredData.length }
        }));
      }
    } catch (err) {
      console.error('Ошибка загрузки данных:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleRefresh = () => {
    loadData();
  };

  const handlePageChange = (newPage) => {
    if (newPage < 1) return;
    setPagination(prev => ({
      ...prev,
      [activeTab]: { ...prev[activeTab], page: newPage }
    }));
  };

  const handlePageSizeChange = (newSize) => {
    setPagination(prev => ({
      ...prev,
      [activeTab]: { ...prev[activeTab], pageSize: newSize, page: 1 }
    }));
  };

  const handleFilterChange = (filterName, value) => {
    setFilters(prev => ({
      ...prev,
      [activeTab]: { ...prev[activeTab], [filterName]: value }
    }));
    setPagination(prev => ({
      ...prev,
      [activeTab]: { ...prev[activeTab], page: 1 }
    }));
  };

  const clearFilters = () => {
    const emptyFilters = {
      performers: { name: '', email: '', inn: '', business_type: '', verification_status: '' },
      users: { name: '', email: '', role: '', verification_status: '' },
      bookings: { status: '', client_name: '', service_name: '' },
      services: { title: '', performer_name: '', min_price: '', max_price: '' },
      reviews: { client_name: '', rating: '', comment: '' }
    };
    setFilters(emptyFilters);
    setDebouncedFilters(emptyFilters);
  };

  const handleUpdateVerificationStatus = async (userId, currentStatus) => {
    const newStatus = currentStatus === 'verified' ? 'rejected' : 'verified';
    if (window.confirm(`Изменить статус верификации для этого исполнителя с "${currentStatus === 'verified' ? 'Подтвержден' : currentStatus === 'pending' ? 'Ожидает' : 'Отклонен'}" на "${newStatus === 'verified' ? 'Подтвержден' : 'Отклонен'}"?`)) {
      try {
        await coreAPI.updateVerificationStatus(userId, newStatus);
        alert(`Статус верификации успешно изменен на "${newStatus === 'verified' ? 'Подтвержден' : 'Отклонен'}"!`);
        loadData();
      } catch (err) {
        alert('Не удалось обновить статус верификации');
      }
    }
  };

  const openEditModal = (type, item) => {
    setEditType(type);
    setEditItem(item);
    setShowEditModal(true);
  };

  const handleUpdateUser = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.updateProfile(editItem);
      alert('Пользователь успешно обновлен!');
      setShowEditModal(false);
      loadData();
    } catch (err) {
      alert('Не удалось обновить пользователя: ' + err.message);
    }
  };

  const handleUpdateService = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.updateService(editItem.id, {
        title: editItem.title,
        description: editItem.description,
        price: editItem.price,
        duration_minutes: editItem.duration_minutes
      });
      alert('Услуга успешно обновлена!');
      setShowEditModal(false);
      loadData();
    } catch (err) {
      alert('Не удалось обновить услугу: ' + err.message);
    }
  };

  const handleUpdateBooking = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.updateBooking(editItem.id, {
        status: editItem.status
      });
      alert('Бронирование успешно обновлено!');
      setShowEditModal(false);
      loadData();
    } catch (err) {
      alert('Не удалось обновить бронирование: ' + err.message);
    }
  };

  const handleUpdateReview = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.updateReview(editItem.id, {
        rating: editItem.rating,
        comment: editItem.comment
      });
      alert('Отзыв успешно обновлен!');
      setShowEditModal(false);
      loadData();
    } catch (err) {
      alert('Не удалось обновить отзыв: ' + err.message);
    }
  };

  const openCreateModal = (type) => {
    setCreateType(type);
    setCreateFormData({
      email: '',
      name: '',
      password: '',
      role: 'client',
      inn: '',
      business_type: '',
      title: '',
      description: '',
      price: '',
      duration_minutes: '',
      service_id: '',
      rating: 5,
      comment: '',
      booking_id: ''
    });
    setShowCreateModal(true);
  };

  const handleCreateUser = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.register(createFormData);
      alert('Пользователь успешно создан!');
      setShowCreateModal(false);
      loadData();
    } catch (err) {
      alert('Не удалось создать пользователя: ' + err.message);
    }
  };

  const handleCreateService = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.createService({
        title: createFormData.title,
        description: createFormData.description,
        price: parseFloat(createFormData.price),
        duration_minutes: parseInt(createFormData.duration_minutes)
      });
      alert('Услуга успешно создана!');
      setShowCreateModal(false);
      loadData();
    } catch (err) {
      alert('Не удалось создать услугу: ' + err.message);
    }
  };

  const handleCreateReview = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.createReview({
        booking_id: createFormData.booking_id,
        rating: parseInt(createFormData.rating),
        comment: createFormData.comment
      });
      alert('Отзыв успешно создан!');
      setShowCreateModal(false);
      loadData();
    } catch (err) {
      alert('Не удалось создать отзыв: ' + err.message);
    }
  };

  const handleDeleteUser = async (userId) => {
    if (window.confirm('Вы уверены, что хотите удалить этого пользователя?')) {
      try {
        await coreAPI.deleteUser(userId);
        alert('Пользователь успешно удален!');
        loadData();
      } catch (err) {
        alert('Не удалось удалить пользователя');
      }
    }
  };

  const handleDeleteService = async (serviceId, serviceTitle) => {
    if (window.confirm(`Вы уверены, что хотите удалить услугу "${serviceTitle}"?`)) {
      try {
        await coreAPI.deleteService(serviceId);
        alert('Услуга успешно удалена!');
        loadData();
      } catch (err) {
        alert('Не удалось удалить услугу');
      }
    }
  };

  const handleDeleteBooking = async (bookingId) => {
    if (window.confirm('Вы уверены, что хотите удалить это бронирование?')) {
      try {
        await coreAPI.deleteBooking(bookingId);
        alert('Бронирование успешно удалено!');
        loadData();
      } catch (err) {
        alert('Не удалось удалить бронирование');
      }
    }
  };

  const handleDeleteReview = async (reviewId) => {
    if (window.confirm('Вы уверены, что хотите удалить этот отзыв?')) {
      try {
        await coreAPI.deleteAdminReview(reviewId);
        alert('Отзыв успешно удален!');
        loadData();
      } catch (err) {
        alert('Не удалось удалить отзыв');
      }
    }
  };

  const handleVerifyPerformers = async () => {
    if (selectedUsers.length === 0) {
      alert('Пожалуйста, выберите исполнителей для подтверждения');
      return;
    }
    if (window.confirm(`Подтвердить ${selectedUsers.length} исполнителя(ей)?`)) {
      try {
        await coreAPI.batchVerifyPerformers(selectedUsers, 'verified');
        alert('Исполнители успешно подтверждены!');
        setSelectedUsers([]);
        loadData();
      } catch (err) {
        alert('Не удалось подтвердить исполнителей');
      }
    }
  };

  const handleRejectPerformers = async () => {
    if (selectedUsers.length === 0) {
      alert('Пожалуйста, выберите исполнителей для отклонения');
      return;
    }
    if (window.confirm(`Отклонить ${selectedUsers.length} исполнителя(ей)?`)) {
      try {
        await coreAPI.batchVerifyPerformers(selectedUsers, 'rejected');
        alert('Исполнители успешно отклонены!');
        setSelectedUsers([]);
        loadData();
      } catch (err) {
        alert('Не удалось отклонить исполнителей');
      }
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'Н/Д';
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU');
  };

  const getPaginatedData = (data) => {
    const start = (pagination[activeTab].page - 1) * pagination[activeTab].pageSize;
    const end = start + pagination[activeTab].pageSize;
    return data.slice(start, end);
  };

  const currentPagination = pagination[activeTab];
  const totalPages = Math.ceil(currentPagination.total / currentPagination.pageSize);

  const renderPagination = () => {
    if (currentPagination.total === 0) return null;
    
    return (
      <div className="pagination">
        <div className="pagination-info">
          Показано {((currentPagination.page - 1) * currentPagination.pageSize) + 1} -{' '}
          {Math.min(currentPagination.page * currentPagination.pageSize, currentPagination.total)} из{' '}
          {currentPagination.total} записей
        </div>
        <div className="pagination-controls">
          <button 
            onClick={() => handlePageChange(currentPagination.page - 1)}
            disabled={currentPagination.page === 1}
            className="btn-pagination"
          >
            Назад
          </button>
          <span className="page-number">Страница {currentPagination.page} из {totalPages || 1}</span>
          <button 
            onClick={() => handlePageChange(currentPagination.page + 1)}
            disabled={currentPagination.page >= totalPages}
            className="btn-pagination"
          >
            Вперед
          </button>
          <select 
            value={currentPagination.pageSize} 
            onChange={(e) => handlePageSizeChange(Number(e.target.value))}
            className="page-size-select"
          >
            <option value="10">10 на странице</option>
            <option value="20">20 на странице</option>
            <option value="50">50 на странице</option>
            <option value="100">100 на странице</option>
          </select>
        </div>
      </div>
    );
  };

  const renderEditModal = () => {
    if (!editItem) return null;
    
    if (editType === 'user') {
      return (
        <form onSubmit={handleUpdateUser}>
          <div className="form-group">
            <label>Имя</label>
            <input
              type="text"
              value={editItem.name || ''}
              onChange={(e) => setEditItem({ ...editItem, name: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>Email</label>
            <input
              type="email"
              value={editItem.email || ''}
              onChange={(e) => setEditItem({ ...editItem, email: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>Роль</label>
            <select
              value={editItem.role || 'client'}
              onChange={(e) => setEditItem({ ...editItem, role: e.target.value })}
            >
              <option value="client">Клиент</option>
              <option value="performer">Исполнитель</option>
              <option value="admin">Администратор</option>
            </select>
          </div>
          <div className="modal-actions">
            <button type="submit" className="btn-confirm">Обновить пользователя</button>
            <button type="button" onClick={() => setShowEditModal(false)} className="btn-cancel">Отмена</button>
          </div>
        </form>
      );
    }
    
    if (editType === 'service') {
      return (
        <form onSubmit={handleUpdateService}>
          <div className="form-group">
            <label>Название</label>
            <input
              type="text"
              value={editItem.title || ''}
              onChange={(e) => setEditItem({ ...editItem, title: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>Описание</label>
            <textarea
              value={editItem.description || ''}
              onChange={(e) => setEditItem({ ...editItem, description: e.target.value })}
              rows="3"
            />
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>Цена</label>
              <input
                type="number"
                value={editItem.price || 0}
                onChange={(e) => setEditItem({ ...editItem, price: parseFloat(e.target.value) })}
                required
                min="0"
                step="0.01"
              />
            </div>
            <div className="form-group">
              <label>Длительность (минуты)</label>
              <input
                type="number"
                value={editItem.duration_minutes || 0}
                onChange={(e) => setEditItem({ ...editItem, duration_minutes: parseInt(e.target.value) })}
                required
                min="15"
                step="15"
              />
            </div>
          </div>
          <div className="modal-actions">
            <button type="submit" className="btn-confirm">Обновить услугу</button>
            <button type="button" onClick={() => setShowEditModal(false)} className="btn-cancel">Отмена</button>
          </div>
        </form>
      );
    }
    
    if (editType === 'booking') {
      return (
        <form onSubmit={handleUpdateBooking}>
          <div className="form-group">
            <label>Статус</label>
            <select
              value={editItem.status || 'pending'}
              onChange={(e) => setEditItem({ ...editItem, status: e.target.value })}
            >
              <option value="pending">Ожидает</option>
              <option value="confirmed">Подтверждено</option>
              <option value="completed">Завершено</option>
              <option value="cancelled">Отменено</option>
            </select>
          </div>
          <div className="modal-actions">
            <button type="submit" className="btn-confirm">Обновить бронирование</button>
            <button type="button" onClick={() => setShowEditModal(false)} className="btn-cancel">Отмена</button>
          </div>
        </form>
      );
    }
    
    if (editType === 'review') {
      return (
        <form onSubmit={handleUpdateReview}>
          <div className="form-group">
            <label>Рейтинг</label>
            <select
              value={editItem.rating || 5}
              onChange={(e) => setEditItem({ ...editItem, rating: parseInt(e.target.value) })}
            >
              <option value="1">1 Звезда</option>
              <option value="2">2 Звезды</option>
              <option value="3">3 Звезды</option>
              <option value="4">4 Звезды</option>
              <option value="5">5 Звезд</option>
            </select>
          </div>
          <div className="form-group">
            <label>Комментарий</label>
            <textarea
              value={editItem.comment || ''}
              onChange={(e) => setEditItem({ ...editItem, comment: e.target.value })}
              rows="3"
            />
          </div>
          <div className="modal-actions">
            <button type="submit" className="btn-confirm">Обновить отзыв</button>
            <button type="button" onClick={() => setShowEditModal(false)} className="btn-cancel">Отмена</button>
          </div>
        </form>
      );
    }
    
    return null;
  };

  const renderCreateModal = () => {
    if (createType === 'user') {
      return (
        <form onSubmit={handleCreateUser}>
          <div className="form-group">
            <label>Имя *</label>
            <input
              type="text"
              value={createFormData.name}
              onChange={(e) => setCreateFormData({ ...createFormData, name: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>Email *</label>
            <input
              type="email"
              value={createFormData.email}
              onChange={(e) => setCreateFormData({ ...createFormData, email: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>Пароль *</label>
            <input
              type="password"
              value={createFormData.password}
              onChange={(e) => setCreateFormData({ ...createFormData, password: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>Роль *</label>
            <select
              value={createFormData.role}
              onChange={(e) => setCreateFormData({ ...createFormData, role: e.target.value })}
            >
              <option value="client">Клиент</option>
              <option value="performer">Исполнитель</option>
              <option value="admin">Администратор</option>
            </select>
          </div>
          {createFormData.role === 'performer' && (
            <>
              <div className="form-group">
                <label>ИНН</label>
                <input
                  type="text"
                  value={createFormData.inn}
                  onChange={(e) => setCreateFormData({ ...createFormData, inn: e.target.value })}
                  placeholder="12 цифр"
                  maxLength="12"
                />
              </div>
              <div className="form-group">
                <label>Тип бизнеса</label>
                <select
                  value={createFormData.business_type}
                  onChange={(e) => setCreateFormData({ ...createFormData, business_type: e.target.value })}
                >
                  <option value="">Выберите...</option>
                  <option value="IP">ИП</option>
                  <option value="self_employed">Самозанятый</option>
                </select>
              </div>
            </>
          )}
          <div className="modal-actions">
            <button type="submit" className="btn-confirm">Создать пользователя</button>
            <button type="button" onClick={() => setShowCreateModal(false)} className="btn-cancel">Отмена</button>
          </div>
        </form>
      );
    } 
    else if (createType === 'service') {
      return (
        <form onSubmit={handleCreateService}>
          <div className="form-group">
            <label>Название услуги *</label>
            <input
              type="text"
              value={createFormData.title}
              onChange={(e) => setCreateFormData({ ...createFormData, title: e.target.value })}
              required
            />
          </div>
          <div className="form-group">
            <label>Описание</label>
            <textarea
              value={createFormData.description}
              onChange={(e) => setCreateFormData({ ...createFormData, description: e.target.value })}
              rows="3"
            />
          </div>
          <div className="form-row">
            <div className="form-group">
              <label>Цена (₽) *</label>
              <input
                type="number"
                value={createFormData.price}
                onChange={(e) => setCreateFormData({ ...createFormData, price: e.target.value })}
                required
                min="0"
                step="0.01"
              />
            </div>
            <div className="form-group">
              <label>Длительность (минуты) *</label>
              <input
                type="number"
                value={createFormData.duration_minutes}
                onChange={(e) => setCreateFormData({ ...createFormData, duration_minutes: e.target.value })}
                required
                min="15"
                step="15"
              />
            </div>
          </div>
          <div className="modal-actions">
            <button type="submit" className="btn-confirm">Создать услугу</button>
            <button type="button" onClick={() => setShowCreateModal(false)} className="btn-cancel">Отмена</button>
          </div>
        </form>
      );
    }
    else if (createType === 'review') {
      return (
        <form onSubmit={handleCreateReview}>
          <div className="form-group">
            <label>ID бронирования *</label>
            <input
              type="text"
              value={createFormData.booking_id}
              onChange={(e) => setCreateFormData({ ...createFormData, booking_id: e.target.value })}
              required
              placeholder="UUID бронирования"
            />
          </div>
          <div className="form-group">
            <label>Рейтинг *</label>
            <select
              value={createFormData.rating}
              onChange={(e) => setCreateFormData({ ...createFormData, rating: e.target.value })}
            >
              <option value="1">1 Звезда</option>
              <option value="2">2 Звезды</option>
              <option value="3">3 Звезды</option>
              <option value="4">4 Звезды</option>
              <option value="5">5 Звезд</option>
            </select>
          </div>
          <div className="form-group">
            <label>Комментарий</label>
            <textarea
              value={createFormData.comment}
              onChange={(e) => setCreateFormData({ ...createFormData, comment: e.target.value })}
              rows="3"
              placeholder="Необязательный комментарий..."
            />
          </div>
          <div className="modal-actions">
            <button type="submit" className="btn-confirm">Создать отзыв</button>
            <button type="button" onClick={() => setShowCreateModal(false)} className="btn-cancel">Отмена</button>
          </div>
        </form>
      );
    }
    return null;
  };

  if (loading) return <div className="loading">Загрузка данных админки...</div>;

  return (
    <div className="admin-panel">
      <div className="admin-header">
        <h1>Панель администратора</h1>
        <button onClick={handleRefresh} className="btn-refresh">
          🔄 Обновить всё
        </button>
      </div>
      
      <div className="admin-tabs">
        <button className={activeTab === 'performers' ? 'active' : ''} onClick={() => setActiveTab('performers')}>
          Неподтвержденные исполнители ({pagination.performers.total})
        </button>
        <button className={activeTab === 'users' ? 'active' : ''} onClick={() => setActiveTab('users')}>
          Пользователи ({pagination.users.total})
        </button>
        <button className={activeTab === 'bookings' ? 'active' : ''} onClick={() => setActiveTab('bookings')}>
          Бронирования ({pagination.bookings.total})
        </button>
        <button className={activeTab === 'services' ? 'active' : ''} onClick={() => setActiveTab('services')}>
          Услуги ({pagination.services.total})
        </button>
        <button className={activeTab === 'reviews' ? 'active' : ''} onClick={() => setActiveTab('reviews')}>
          Отзывы ({pagination.reviews.total})
        </button>
      </div>

      {showEditModal && (
        <div className="modal-overlay" onClick={() => setShowEditModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Редактировать {editType === 'user' ? 'пользователя' : editType === 'service' ? 'услугу' : editType === 'booking' ? 'бронирование' : 'отзыв'}</h3>
            {renderEditModal()}
          </div>
        </div>
      )}

      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Создать {createType === 'user' ? 'пользователя' : createType === 'service' ? 'услугу' : 'отзыв'}</h3>
            {renderCreateModal()}
          </div>
        </div>
      )}

      {activeTab === 'performers' && (
        <div className="admin-section">
          <div className="section-header">
            <h2>Неподтвержденные исполнители</h2>
            <div className="header-buttons">
              <button onClick={handleVerifyPerformers} className="btn-verify">
                Подтвердить выбранных ({selectedUsers.length})
              </button>
              <button onClick={handleRejectPerformers} className="btn-reject">
                Отклонить выбранных ({selectedUsers.length})
              </button>
              <button onClick={clearFilters} className="btn-clear-filters">
                Очистить фильтры
              </button>
            </div>
          </div>
          
          <div className="filters-bar">
            <input type="text" placeholder="Фильтр по имени..." value={filters.performers.name} onChange={(e) => handleFilterChange('name', e.target.value)} className="filter-input" />
            <input type="text" placeholder="Фильтр по email..." value={filters.performers.email} onChange={(e) => handleFilterChange('email', e.target.value)} className="filter-input" />
            <input type="text" placeholder="Фильтр по ИНН..." value={filters.performers.inn} onChange={(e) => handleFilterChange('inn', e.target.value)} className="filter-input" />
            <select value={filters.performers.business_type} onChange={(e) => handleFilterChange('business_type', e.target.value)} className="filter-select">
              <option value="">Все типы бизнеса</option>
              <option value="IP">ИП</option>
              <option value="self_employed">Самозанятый</option>
            </select>
            <select value={filters.performers.verification_status} onChange={(e) => handleFilterChange('verification_status', e.target.value)} className="filter-select">
              <option value="">Все статусы</option>
              <option value="verified">Подтвержден</option>
              <option value="pending">Ожидает</option>
              <option value="rejected">Отклонен</option>
            </select>
          </div>
          
          
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th><input type="checkbox" onChange={(e) => { if (e.target.checked) { setSelectedUsers(performers.map(p => p.id)); } else { setSelectedUsers([]); } }} /></th>
                  <th>Имя</th>
                  <th>Email</th>
                  <th>ИНН</th>
                  <th>Тип бизнеса</th>
                  <th>Статус</th>
                  <th>Кол-во услуг</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {getPaginatedData(performers).map(performer => {
                  const status = performer.verification_status || 'pending';
                  return (
                    <tr key={performer.id}>
                      <td><input type="checkbox" checked={selectedUsers.includes(performer.id)} onChange={(e) => { if (e.target.checked) { setSelectedUsers([...selectedUsers, performer.id]); } else { setSelectedUsers(selectedUsers.filter(id => id !== performer.id)); } }} /></td>
                      <td>{performer.name}</td>                      <td>{performer.email}</td>
                      <td>{performer.inn || 'Н/Д'}</td>
                      <td>{performer.business_type || 'Н/Д'}</td>
                      <td><span className={`status-badge ${status}`}>{status === 'verified' ? '✅ Подтвержден' : status === 'pending' ? '⏳ Ожидает' : status === 'rejected' ? '❌ Отклонен' : status}</span></td>
                      <td>{performer.services_count || 0}</td>
                      <td><div className="action-buttons"><button onClick={() => handleUpdateVerificationStatus(performer.id, status)} className="btn-verify-sm">{status === 'verified' ? '↻ Отклонить' : '✓ Подтвердить'}</button><button onClick={() => handleDeleteUser(performer.id)} className="btn-delete-sm">Удалить</button></div></td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
          {renderPagination()}
        </div>
      )}

      {activeTab === 'users' && (
        <div className="admin-section">
          <div className="section-header">
            <h2>Все пользователи</h2>
            <div className="header-buttons">
              <button onClick={() => openCreateModal('user')} className="btn-create-sm">➕ Создать</button>
              <button onClick={clearFilters} className="btn-clear-filters">Очистить фильтры</button>
            </div>
          </div>
          
          <div className="filters-bar">
            <input type="text" placeholder="Фильтр по имени..." value={filters.users.name} onChange={(e) => handleFilterChange('name', e.target.value)} className="filter-input" />
            <input type="text" placeholder="Фильтр по email..." value={filters.users.email} onChange={(e) => handleFilterChange('email', e.target.value)} className="filter-input" />
            <select value={filters.users.role} onChange={(e) => handleFilterChange('role', e.target.value)} className="filter-select">
              <option value="">Все роли</option>
              <option value="admin">Админ</option>
              <option value="client">Клиент</option>
              <option value="performer">Исполнитель</option>
            </select>
            <select value={filters.users.verification_status} onChange={(e) => handleFilterChange('verification_status', e.target.value)} className="filter-select">
              <option value="">Все статусы</option>
              <option value="verified">Подтвержден</option>
              <option value="pending">Ожидает</option>
              <option value="rejected">Отклонен</option>
            </select>
          </div>
          
          
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Имя</th>
                  <th>Email</th>
                  <th>Роль</th>
                  <th>Статус</th>
                  <th>ИНН</th>
                  <th>Тип бизнеса</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {getPaginatedData(users).map(user => {
                  const isPerformer = user.role === 'performer';
                  return (
                    <tr key={user.id}>
                      <td>{user.name}</td>
                      <td>{user.email}</td>
                      <td><span className={`role-badge ${user.role}`}>{user.role === 'admin' ? 'Админ' : user.role === 'performer' ? 'Исполнитель' : 'Клиент'}</span></td>
                      <td><span className={`status-badge ${user.verification_status}`}>{user.verification_status === 'verified' ? 'Подтвержден' : user.verification_status === 'pending' ? 'Ожидает' : user.verification_status === 'rejected' ? 'Отклонен' : 'Н/Д'}</span></td>
                      <td>{user.inn || 'Н/Д'}</td>
                      <td>{user.business_type || 'Н/Д'}</td>
                      <td><div className="action-buttons"><button onClick={() => openEditModal('user', user)} className="btn-edit-sm">✏️ Ред.</button>{isPerformer && <button onClick={() => handleUpdateVerificationStatus(user.id, user.verification_status)} className="btn-verify-sm">{user.verification_status === 'verified' ? '↻ Отклонить' : '✓ Подтвердить'}</button>}<button onClick={() => handleDeleteUser(user.id)} className="btn-delete-sm">Удалить</button></div></td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
          {renderPagination()}
        </div>
      )}

      {activeTab === 'bookings' && (
        <div className="admin-section">
          <div className="section-header">
            <h2>Все бронирования</h2>
            <div className="header-buttons"><button onClick={clearFilters} className="btn-clear-filters">Очистить фильтры</button></div>
          </div>
          
          <div className="filters-bar">
            <select value={filters.bookings.status} onChange={(e) => handleFilterChange('status', e.target.value)} className="filter-select">
              <option value="">Все статусы</option>
              <option value="pending">Ожидает</option>
              <option value="confirmed">Подтверждено</option>
              <option value="completed">Завершено</option>
              <option value="cancelled">Отменено</option>
            </select>
            <input type="text" placeholder="Фильтр по имени клиента..." value={filters.bookings.client_name} onChange={(e) => handleFilterChange('client_name', e.target.value)} className="filter-input" />
            <input type="text" placeholder="Фильтр по названию услуги..." value={filters.bookings.service_name} onChange={(e) => handleFilterChange('service_name', e.target.value)} className="filter-input" />
          </div>
          
          
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Клиент</th>
                  <th>Услуга</th>
                  <th>Дата</th>
                  <th>Статус</th>
                  <th>Цена</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {getPaginatedData(bookings).map(booking => (
                  <tr key={booking.id}>
                    <td>{booking.id?.slice(0, 8)}...</td>
                    <td>{booking.client_name}</td>
                    <td>{booking.service_name}</td>
                    <td>{formatDate(booking.booking_time)}</td>
                    <td><span className={`status-badge ${booking.status}`}>{booking.status === 'pending' ? 'Ожидает' : booking.status === 'confirmed' ? 'Подтверждено' : booking.status === 'completed' ? 'Завершено' : 'Отменено'}</span></td>
                    <td>{booking.final_price}₽</td>
                    <td><div className="action-buttons"><button onClick={() => openEditModal('booking', booking)} className="btn-edit-sm">✏️ Ред.</button><button onClick={() => handleDeleteBooking(booking.id)} className="btn-delete-sm">Удалить</button></div></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {renderPagination()}
        </div>
      )}

      {activeTab === 'services' && (
        <div className="admin-section">
          <div className="section-header">
            <h2>Все услуги</h2>
            <div className="header-buttons">
              <button onClick={() => openCreateModal('service')} className="btn-create-sm">➕ Создать</button>
              <button onClick={clearFilters} className="btn-clear-filters">Очистить фильтры</button>
            </div>
          </div>
          
          <div className="filters-bar">
            <input type="text" placeholder="Фильтр по названию..." value={filters.services.title} onChange={(e) => handleFilterChange('title', e.target.value)} className="filter-input" />
            <input type="text" placeholder="Фильтр по исполнителю..." value={filters.services.performer_name} onChange={(e) => handleFilterChange('performer_name', e.target.value)} className="filter-input" />
            <input type="number" placeholder="Мин. цена" value={filters.services.min_price} onChange={(e) => handleFilterChange('min_price', e.target.value)} className="filter-input" style={{ width: '120px' }} />
            <input type="number" placeholder="Макс. цена" value={filters.services.max_price} onChange={(e) => handleFilterChange('max_price', e.target.value)} className="filter-input" style={{ width: '120px' }} />
          </div>
          
          
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Название</th>
                  <th>Исполнитель</th>
                  <th>Цена</th>
                  <th>Длительность</th>
                  <th>Создана</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {getPaginatedData(services).map(service => (
                  <tr key={service.id}>
                    <td>{service.title}</td>
                    <td>{service.performer_name || service.performer_id?.slice(0, 8)}</td>
                    <td>${service.price}</td>
                    <td>{service.duration_minutes} мин</td>
                    <td>{formatDate(service.created_at)}</td>
                    <td><div className="action-buttons"><button onClick={() => openEditModal('service', service)} className="btn-edit-sm">✏️ Ред.</button><button onClick={() => handleDeleteService(service.id, service.title)} className="btn-delete-sm">Удалить</button></div></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {renderPagination()}
        </div>
      )}

      {activeTab === 'reviews' && (
        <div className="admin-section">
          <div className="section-header">
            <h2>Все отзывы</h2>
            <div className="header-buttons">
              <button onClick={() => openCreateModal('review')} className="btn-create-sm">➕ Создать</button>
              <button onClick={clearFilters} className="btn-clear-filters">Очистить фильтры</button>
            </div>
          </div>
          
          <div className="filters-bar">
            <input type="text" placeholder="Фильтр по клиенту..." value={filters.reviews.client_name} onChange={(e) => handleFilterChange('client_name', e.target.value)} className="filter-input" />
            <select value={filters.reviews.rating} onChange={(e) => handleFilterChange('rating', e.target.value)} className="filter-select">
              <option value="">Все рейтинги</option>
              <option value="1">1 Звезда</option>
              <option value="2">2 Звезды</option>
              <option value="3">3 Звезды</option>
              <option value="4">4 Звезды</option>
              <option value="5">5 Звезд</option>
            </select>
            <input type="text" placeholder="Фильтр по комментарию..." value={filters.reviews.comment} onChange={(e) => handleFilterChange('comment', e.target.value)} className="filter-input" />
          </div>
          
          
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Клиент</th>
                  <th>Рейтинг</th>
                  <th>Комментарий</th>
                  <th>Дата</th>
                  <th>Действия</th>
                </tr>
              </thead>
              <tbody>
                {getPaginatedData(reviews).map(review => (
                  <tr key={review.id}>
                    <td>{review.client_name}</td>
                    <td>{'⭐'.repeat(review.rating)} ({review.rating}/5)</td>
                    <td>{review.comment || 'Нет комментария'}</td>
                    <td>{formatDate(review.created_at)}</td>
                    <td><div className="action-buttons"><button onClick={() => openEditModal('review', review)} className="btn-edit-sm">✏️ Ред.</button><button onClick={() => handleDeleteReview(review.id)} className="btn-delete-sm">Удалить</button></div></td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {renderPagination()}
        </div>
      )}
    </div>
  );
};

export default AdminPanel;