import React, { useState, useEffect } from 'react';
import coreAPI from '../../api/core';
import './Admin.css';

const UserManagement = () => {
  const [users, setUsers] = useState([]);
  const [performers, setPerformers] = useState([]);
  const [loading, setLoading] = useState(true);
  const [selectedTab, setSelectedTab] = useState('all');
  const [filters, setFilters] = useState({ role: '', verification_status: '', search: '' });
  const [selectedUsers, setSelectedUsers] = useState([]);
  const [showVerifyModal, setShowVerifyModal] = useState(false);

  useEffect(() => {
    loadUsers();
    loadUnverifiedPerformers();
  }, [filters, selectedTab]);

  const loadUsers = async () => {
    try {
      setLoading(true);
      const params = { ...filters };
      if (selectedTab === 'performers') params.role = 'performer';
      if (selectedTab === 'clients') params.role = 'client';
      
      const data = await coreAPI.getUsers(params);
      setUsers(data.data || []);
    } catch (err) {
      console.error('Failed to load users:', err);
    } finally {
      setLoading(false);
    }
  };

  const loadUnverifiedPerformers = async () => {
    try {
      const data = await coreAPI.getUnverifiedPerformers(1, 100);
      setPerformers(data.data || []);
    } catch (err) {
      console.error('Failed to load unverified performers:', err);
    }
  };

  const handleDeleteUser = async (userId) => {
    if (window.confirm('Are you sure you want to delete this user? This action cannot be undone.')) {
      try {
        await coreAPI.deleteUser(userId);
        await loadUsers();
        alert('User deleted successfully');
      } catch (err) {
        alert('Failed to delete user');
      }
    }
  };

  const handleVerifyPerformers = async () => {
    if (selectedUsers.length === 0) {
      alert('Please select performers to verify');
      return;
    }
    
    try {
      await coreAPI.batchVerifyPerformers(selectedUsers, 'verified');
      alert(`${selectedUsers.length} performer(s) verified successfully`);
      setSelectedUsers([]);
      setShowVerifyModal(false);
      await loadUsers();
      await loadUnverifiedPerformers();
    } catch (err) {
      alert('Failed to verify performers');
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
        alert(`${selectedUsers.length} performer(s) rejected`);
        setSelectedUsers([]);
        setShowVerifyModal(false);
        await loadUsers();
        await loadUnverifiedPerformers();
      } catch (err) {
        alert('Failed to reject performers');
      }
    }
  };

  const handleUpdateRole = async (userId, newRole) => {
    try {
      await coreAPI.updateUserRole(userId, newRole);
      await loadUsers();
      alert('User role updated successfully');
    } catch (err) {
      alert('Failed to update user role');
    }
  };

  const getStatusBadge = (status) => {
    const statusColors = {
      verified: '✅ Verified',
      pending: '⏳ Pending',
      rejected: '❌ Rejected'
    };
    return statusColors[status] || status;
  };

  if (loading) return <div className="loading">Loading users...</div>;

  return (
    <div className="user-management">
      <div className="management-header">
        <h2>User Management</h2>
        {performers.length > 0 && (
          <button onClick={() => setShowVerifyModal(true)} className="btn-verify-batch">
            Verify Performers ({performers.length} pending)
          </button>
        )}
      </div>

      <div className="management-tabs">
        <button className={selectedTab === 'all' ? 'active' : ''} onClick={() => setSelectedTab('all')}>
          All Users
        </button>
        <button className={selectedTab === 'performers' ? 'active' : ''} onClick={() => setSelectedTab('performers')}>
          Performers
        </button>
        <button className={selectedTab === 'clients' ? 'active' : ''} onClick={() => setSelectedTab('clients')}>
          Clients
        </button>
        <button className={selectedTab === 'unverified' ? 'active' : ''} onClick={() => setSelectedTab('unverified')}>
          Unverified
        </button>
      </div>

      <div className="filters-bar">
        <input
          type="text"
          placeholder="Search by name or email..."
          value={filters.search}
          onChange={(e) => setFilters({ ...filters, search: e.target.value })}
          className="search-input"
        />
        <select 
          value={filters.verification_status} 
          onChange={(e) => setFilters({ ...filters, verification_status: e.target.value })}
        >
          <option value="">All Status</option>
          <option value="verified">Verified</option>
          <option value="pending">Pending</option>
          <option value="rejected">Rejected</option>
        </select>
      </div>

      <div className="users-table-container">
        <table className="users-table">
          <thead>
            <tr>
              <th>User</th>
              <th>Email</th>
              <th>Role</th>
              <th>Status</th>
              <th>Details</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {users.map(user => (
              <tr key={user.id}>
                <td>
                  <div className="user-info">
                    <div className="user-avatar">
                      {user.name.charAt(0).toUpperCase()}
                    </div>
                    <div>
                      <div className="user-name">{user.name}</div>
                      <div className="user-id">ID: {user.id.slice(0, 8)}...</div>
                    </div>
                  </div>
                </td>
                <td>{user.email}</td>
                <td>
                  <select 
                    value={user.role} 
                    onChange={(e) => handleUpdateRole(user.id, e.target.value)}
                    className="role-select"
                  >
                    <option value="client">Client</option>
                    <option value="performer">Performer</option>
                    <option value="admin">Admin</option>
                  </select>
                </td>
                <td>
                  <span className={`status-badge ${user.verificationStatus}`}>
                    {getStatusBadge(user.verificationStatus)}
                  </span>
                </td>
                <td>
                  {user.inn && <div className="detail-item">INN: {user.inn}</div>}
                  {user.businessType && <div className="detail-item">Business: {user.businessType}</div>}
                  {user.servicesCount !== undefined && (
                    <div className="detail-item">Services: {user.servicesCount}</div>
                  )}
                </td>
                <td>
                  <button onClick={() => handleDeleteUser(user.id)} className="btn-delete-user">
                    Delete
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showVerifyModal && (
        <div className="modal-overlay" onClick={() => setShowVerifyModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Verify Performers</h3>
            <p>Select performers to verify or reject:</p>
            
            <div className="performers-list">
              {performers.map(performer => (
                <label key={performer.id} className="performer-checkbox">
                  <input
                    type="checkbox"
                    checked={selectedUsers.includes(performer.id)}
                    onChange={(e) => {
                      if (e.target.checked) {
                        setSelectedUsers([...selectedUsers, performer.id]);
                      } else {
                        setSelectedUsers(selectedUsers.filter(id => id !== performer.id));
                      }
                    }}
                  />
                  <div className="performer-info">
                    <strong>{performer.name}</strong>
                    <span>{performer.email}</span>
                    {performer.inn && <span>INN: {performer.inn}</span>}
                    {performer.businessType && <span>Business: {performer.businessType}</span>}
                  </div>
                </label>
              ))}
            </div>
            
            <div className="modal-actions">
              <button onClick={handleVerifyPerformers} className="btn-verify">
                Verify Selected ({selectedUsers.length})
              </button>
              <button onClick={handleRejectPerformers} className="btn-reject">
                Reject Selected
              </button>
              <button onClick={() => setShowVerifyModal(false)} className="btn-cancel">
                Cancel
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default UserManagement;