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
  
  // Пагинация для каждой вкладки
  const [pagination, setPagination] = useState({
    performers: { page: 1, pageSize: 20, total: 0 },
    users: { page: 1, pageSize: 20, total: 0 },
    bookings: { page: 1, pageSize: 20, total: 0 },
    services: { page: 1, pageSize: 20, total: 0 },
    reviews: { page: 1, pageSize: 20, total: 0 }
  });
  
  // Фильтры
  const [filters, setFilters] = useState({
    bookings: { status: '', client_id: '', performer_id: '' },
    services: { performer_id: '', search: '' },  // Добавлен search
    reviews: { service_id: '', rating: '' }
  });

  useEffect(() => {
    loadData();
  }, [activeTab, pagination[activeTab].page, filters]);

  const loadData = async () => {
    setLoading(true);
    try {
      const currentPagination = pagination[activeTab];
      
      if (activeTab === 'performers') {
        const data = await coreAPI.getUnverifiedPerformers(currentPagination.page, currentPagination.pageSize);
        setPerformers(data.data || []);
        setPagination(prev => ({
          ...prev,
          performers: { ...prev.performers, total: data.total || 0 }
        }));
      } 
      else if (activeTab === 'users') {
        const data = await coreAPI.getUsers();
        setUsers(data.data || []);
        setPagination(prev => ({
          ...prev,
          users: { ...prev.users, total: data.total || 0 }
        }));
      } 
      else if (activeTab === 'bookings') {
        const params = {
          page: currentPagination.page,
          page_size: currentPagination.pageSize,
          ...filters.bookings
        };
        Object.keys(params).forEach(key => {
          if (!params[key]) delete params[key];
        });
        const data = await coreAPI.getAdminBookings(params);
        setBookings(data.data || []);
        setPagination(prev => ({
          ...prev,
          bookings: { ...prev.bookings, total: data.total || 0 }
        }));
      } 
      else if (activeTab === 'services') {
        // Передаем параметры поиска в API
        const params = {
          page: currentPagination.page,
          page_size: currentPagination.pageSize
        };
        if (filters.services.performer_id) params.performer_id = filters.services.performer_id;
        if (filters.services.search) params.search = filters.services.search;
        
        const data = await coreAPI.getAdminServices(
          filters.services.performer_id,
          currentPagination.page,
          currentPagination.pageSize,
          filters.services.search  // Добавляем поиск
        );
        console.log('Services with search:', data);
        setServices(data.data || []);
        setPagination(prev => ({
          ...prev,
          services: { ...prev.services, total: data.total || 0 }
        }));
      } 
      else if (activeTab === 'reviews') {
        const data = await coreAPI.getAdminReviews(
          filters.reviews.service_id,
          filters.reviews.rating,
          currentPagination.page,
          currentPagination.pageSize
        );
        setReviews(data.data || []);
        setPagination(prev => ({
          ...prev,
          reviews: { ...prev.reviews, total: data.total || 0 }
        }));
      }
    } catch (err) {
      console.error('Failed to load data:', err);
    } finally {
      setLoading(false);
    }
  };

  // Остальные функции остаются без изменений...
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

  const handleVerifyPerformers = async () => {
    if (selectedUsers.length === 0) {
      alert('Please select performers to verify');
      return;
    }
    
    if (window.confirm(`Verify ${selectedUsers.length} performer(s)?`)) {
      try {
        await coreAPI.batchVerifyPerformers(selectedUsers, 'verified');
        alert('Performers verified successfully!');
        setSelectedUsers([]);
        loadData();
      } catch (err) {
        alert('Failed to verify performers');
      }
    }
  };

  const handleRejectPerformers = async () => {
    if (selectedUsers.length === 0) {
      alert('Please select performers to reject');
      return;
    }
    
    if (window.confirm(`Reject ${selectedUsers.length} performer(s)?`)) {
      try {
        await coreAPI.batchVerifyPerformers(selectedUsers, 'rejected');
        alert('Performers rejected successfully!');
        setSelectedUsers([]);
        loadData();
      } catch (err) {
        alert('Failed to reject performers');
      }
    }
  };

  const handleUpdateVerificationStatus = async (userId, status) => {
    if (window.confirm(`Set verification status to ${status} for this performer?`)) {
      try {
        await coreAPI.updateVerificationStatus(userId, status);
        alert('Verification status updated successfully!');
        loadData();
      } catch (err) {
        alert('Failed to update status');
      }
    }
  };

  const handleDeleteUser = async (userId) => {
    if (window.confirm('Are you sure you want to delete this user?')) {
      try {
        await coreAPI.deleteUser(userId);
        loadData();
        alert('User deleted successfully');
      } catch (err) {
        alert('Failed to delete user');
      }
    }
  };

  const handleDeleteReview = async (reviewId) => {
    if (window.confirm('Are you sure you want to delete this review?')) {
      try {
        await coreAPI.deleteAdminReview(reviewId);
        loadData();
        alert('Review deleted successfully');
      } catch (err) {
        alert('Failed to delete review');
      }
    }
  };

  const formatDate = (dateString) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return date.toLocaleDateString('ru-RU');
  };

  const currentPagination = pagination[activeTab];
  const totalPages = Math.ceil(currentPagination.total / currentPagination.pageSize);

  if (loading) return <div className="loading">Loading admin data...</div>;

  return (
    <div className="admin-panel">
      <h1>Admin Dashboard</h1>
      
      <div className="admin-tabs">
        <button className={activeTab === 'performers' ? 'active' : ''} onClick={() => setActiveTab('performers')}>
          Unverified Performers ({pagination.performers.total})
        </button>
        <button className={activeTab === 'users' ? 'active' : ''} onClick={() => setActiveTab('users')}>
          Users ({pagination.users.total})
        </button>
        <button className={activeTab === 'bookings' ? 'active' : ''} onClick={() => setActiveTab('bookings')}>
          Bookings ({pagination.bookings.total})
        </button>
        <button className={activeTab === 'services' ? 'active' : ''} onClick={() => setActiveTab('services')}>
          Services ({pagination.services.total})
        </button>
        <button className={activeTab === 'reviews' ? 'active' : ''} onClick={() => setActiveTab('reviews')}>
          Reviews ({pagination.reviews.total})
        </button>
      </div>

      {/* Services Tab с поиском по названию */}
      {activeTab === 'services' && (
        <div className="admin-section">
          <h2>All Services</h2>
          <div className="filters-bar">
            <input
              type="text"
              placeholder="Search by service title..."
              value={filters.services.search}
              onChange={(e) => {
                setFilters({
                  ...filters,
                  services: { ...filters.services, search: e.target.value }
                });
                setPagination(prev => ({
                  ...prev,
                  services: { ...prev.services, page: 1 }
                }));
              }}
              className="search-input"
            />
            <input
              type="text"
              placeholder="Performer ID"
              value={filters.services.performer_id}
              onChange={(e) => {
                setFilters({
                  ...filters,
                  services: { ...filters.services, performer_id: e.target.value }
                });
                setPagination(prev => ({
                  ...prev,
                  services: { ...prev.services, page: 1 }
                }));
              }}
            />
            <button onClick={() => loadData()}>Apply Filters</button>
          </div>
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Title</th>
                  <th>Performer</th>
                  <th>Price</th>
                  <th>Duration</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {services.map(service => (
                  <tr key={service.id}>
                    <td>{service.title}</td>
                    <td>{service.performerName || service.performerID?.slice(0, 8)}</td>
                    <td>${service.price}</td>
                    <td>{service.durationMinutes} min</td>
                    <td>{formatDate(service.createdAt)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {renderPagination()}
        </div>
      )}

      {/* Остальные вкладки без изменений... */}
      {activeTab === 'performers' && (
        <div className="admin-section">
          <div className="section-header">
            <h2>Unverified Performers</h2>
            <div className="header-buttons">
              <button onClick={handleVerifyPerformers} className="btn-verify">
                Verify Selected ({selectedUsers.length})
              </button>
              <button onClick={handleRejectPerformers} className="btn-reject">
                Reject Selected ({selectedUsers.length})
              </button>
            </div>
          </div>
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th><input type="checkbox" onChange={(e) => {
                    if (e.target.checked) {
                      setSelectedUsers(performers.map(p => p.id));
                    } else {
                      setSelectedUsers([]);
                    }
                  }} /></th>
                  <th>Name</th>
                  <th>Email</th>
                  <th>INN</th>
                  <th>Business Type</th>
                  <th>Services Count</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {performers.map(performer => (
                  <tr key={performer.id}>
                    <td><input type="checkbox" checked={selectedUsers.includes(performer.id)} onChange={(e) => {
                      if (e.target.checked) {
                        setSelectedUsers([...selectedUsers, performer.id]);
                      } else {
                        setSelectedUsers(selectedUsers.filter(id => id !== performer.id));
                      }
                    }} /></td>
                    <td>{performer.name}</td>
                    <td>{performer.email}</td>
                    <td>{performer.inn || 'N/A'}</td>
                    <td>{performer.businessType || 'N/A'}</td>
                    <td>{performer.servicesCount || 0}</td>
                    <td>
                      <div className="action-buttons">
                        <button onClick={() => handleUpdateVerificationStatus(performer.id, 'verified')} className="btn-verify-sm">
                          Verify
                        </button>
                        <button onClick={() => handleUpdateVerificationStatus(performer.id, 'rejected')} className="btn-reject-sm">
                          Reject
                        </button>
                        <button onClick={() => handleDeleteUser(performer.id)} className="btn-delete-sm">
                          Delete
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {renderPagination()}
        </div>
      )}

      {activeTab === 'users' && (
        <div className="admin-section">
          <h2>All Users</h2>
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Role</th>
                  <th>Status</th>
                  <th>INN</th>
                  <th>Business Type</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map(user => (
                  <tr key={user.id}>
                    <td>{user.name}</td>
                    <td>{user.email}</td>
                    <td>{user.role}</td>
                    <td><span className={`status-badge ${user.verification_status}`}>{user.verification_status || 'N/A'}</span></td>
                    <td>{user.inn || 'N/A'}</td>
                    <td>{user.business_type || 'N/A'}</td>
                    <td>
                      <button onClick={() => handleDeleteUser(user.id)} className="btn-delete-sm">
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {renderPagination()}
        </div>
      )}

      {activeTab === 'bookings' && (
        <div className="admin-section">
          <h2>All Bookings</h2>
          <div className="filters-bar">
            <select 
              value={filters.bookings.status} 
              onChange={(e) => setFilters({
                ...filters,
                bookings: { ...filters.bookings, status: e.target.value }
              })}
            >
              <option value="">All Status</option>
              <option value="pending">Pending</option>
              <option value="confirmed">Confirmed</option>
              <option value="completed">Completed</option>
              <option value="cancelled">Cancelled</option>
            </select>
            <input
              type="text"
              placeholder="Client ID"
              value={filters.bookings.client_id}
              onChange={(e) => setFilters({
                ...filters,
                bookings: { ...filters.bookings, client_id: e.target.value }
              })}
            />
            <input
              type="text"
              placeholder="Performer ID"
              value={filters.bookings.performer_id}
              onChange={(e) => setFilters({
                ...filters,
                bookings: { ...filters.bookings, performer_id: e.target.value }
              })}
            />
            <button onClick={() => setPagination(prev => ({
              ...prev,
              bookings: { ...prev.bookings, page: 1 }
            }))}>Apply Filters</button>
          </div>
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>ID</th>
                  <th>Client Name</th>
                  <th>Service Name</th>
                  <th>Date</th>
                  <th>Status</th>
                  <th>Price</th>
                </tr>
              </thead>
              <tbody>
                {bookings.map(booking => (
                  <tr key={booking.id}>
                    <td>{booking.id?.slice(0, 8)}...</td>
                    <td>{booking.clientName}</td>
                    <td>{booking.serviceName}</td>
                    <td>{formatDate(booking.bookingTime)}</td>
                    <td><span className={`status-badge ${booking.status}`}>{booking.status}</span></td>
                    <td>${booking.finalPrice}</td>
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
          <h2>All Reviews</h2>
          <div className="filters-bar">
            <input
              type="text"
              placeholder="Service ID"
              value={filters.reviews.service_id}
              onChange={(e) => setFilters({
                ...filters,
                reviews: { ...filters.reviews, service_id: e.target.value }
              })}
            />
            <select 
              value={filters.reviews.rating} 
              onChange={(e) => setFilters({
                ...filters,
                reviews: { ...filters.reviews, rating: e.target.value }
              })}
            >
              <option value="">All Ratings</option>
              <option value="1">1 Star</option>
              <option value="2">2 Stars</option>
              <option value="3">3 Stars</option>
              <option value="4">4 Stars</option>
              <option value="5">5 Stars</option>
            </select>
            <button onClick={() => setPagination(prev => ({
              ...prev,
              reviews: { ...prev.reviews, page: 1 }
            }))}>Apply Filters</button>
          </div>
          <div className="table-responsive">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Client</th>
                  <th>Rating</th>
                  <th>Comment</th>
                  <th>Date</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {reviews.map(review => (
                  <tr key={review.id}>
                    <td>{review.clientName}</td>
                    <td>{'⭐'.repeat(review.rating)} ({review.rating}/5)</td>
                    <td>{review.comment || 'No comment'}</td>
                    <td>{formatDate(review.createdAt)}</td>
                    <td>
                      <button onClick={() => handleDeleteReview(review.id)} className="btn-delete-sm">
                        Delete
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {renderPagination()}
        </div>
      )}

      {/* Pagination */}
      {renderPagination()}
    </div>
  );

  function renderPagination() {
    if (currentPagination.total === 0) return null;
    
    return (
      <div className="pagination">
        <div className="pagination-info">
          Showing {((currentPagination.page - 1) * currentPagination.pageSize) + 1} to{' '}
          {Math.min(currentPagination.page * currentPagination.pageSize, currentPagination.total)} of{' '}
          {currentPagination.total} items
        </div>
        <div className="pagination-controls">
          <button 
            onClick={() => handlePageChange(currentPagination.page - 1)}
            disabled={currentPagination.page === 1}
            className="btn-pagination"
          >
            Previous
          </button>
          <span className="page-number">Page {currentPagination.page} of {totalPages || 1}</span>
          <button 
            onClick={() => handlePageChange(currentPagination.page + 1)}
            disabled={currentPagination.page >= totalPages}
            className="btn-pagination"
          >
            Next
          </button>
          <select 
            value={currentPagination.pageSize} 
            onChange={(e) => handlePageSizeChange(Number(e.target.value))}
            className="page-size-select"
          >
            <option value="10">10 per page</option>
            <option value="20">20 per page</option>
            <option value="50">50 per page</option>
            <option value="100">100 per page</option>
          </select>
        </div>
      </div>
    );
  }
};

export default AdminPanel;