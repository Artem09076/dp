import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import coreAPI from '../../api/core';
import { useAuth } from '../../contexts/AuthContext';
import './Profile.css';

const Profile = () => {
  const navigate = useNavigate();
  const [profile, setProfile] = useState(null);
  const [loading, setLoading] = useState(true);
  const [editing, setEditing] = useState(false);
  const [formData, setFormData] = useState({ name: '', email: '' });
  const { logout, userRole } = useAuth();

  useEffect(() => {
    loadProfile();
  }, []);

  const loadProfile = async () => {
    try {
      const data = await coreAPI.getProfile();
      setProfile(data);
      setFormData({ name: data.name, email: data.email });
    } catch (err) {
      console.error('Failed to load profile:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleUpdate = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.updateProfile(formData);
      await loadProfile();
      setEditing(false);
      alert('Profile updated successfully!');
    } catch (err) {
      alert('Failed to update profile');
    }
  };

  const handleDelete = async () => {
    if (window.confirm('Are you sure you want to delete your account? This action cannot be undone.')) {
      try {
        await coreAPI.deleteProfile();
        logout();
      } catch (err) {
        alert('Failed to delete account');
      }
    }
  };

  if (loading) return <div className="loading">Loading profile...</div>;
  if (!profile) return <div className="error">Failed to load profile</div>;

  return (
    <div className="profile-container">
      <div className="profile-card">
        <h2>Мой профиль</h2>
        
        <div className="quick-actions-section">
          <h3>Действия</h3>
          <div className="quick-actions">
            {userRole === 'performer' && profile.verification_status === 'verified' && (
              <>
                <button onClick={() => navigate('/my-services')} className="action-btn">
                  <span className="action-icon">🛠️</span>
                  Manage Services
                </button>
                <button onClick={() => navigate('/bookings')} className="action-btn">
                  <span className="action-icon">📋</span>
                  Manage Bookings
                </button>
              </>
            )}
            {userRole === 'performer' && profile.verification_status != 'verified' && (
              <>
              <span>Вы не можете пока создавать или смотреть ваши бронирования, подождите пока ваш аккаунт подтвердят</span>
              </>
            )}
            {userRole === 'client' && (
              <button onClick={() => navigate('/bookings')} className="action-btn">
                <span className="action-icon">📅</span>
                Мои бронирования
              </button>
            )}
            {userRole === 'admin' && (
              <button onClick={() => navigate('/admin')} className="action-btn">
                <span className="action-icon">👥</span>
                Admin Panel
              </button>
            )}

          </div>
        </div>

        <div className="profile-divider"></div>
        
        {editing ? (
          <form onSubmit={handleUpdate}>
            <div className="form-group">
              <label>Имя</label>
              <input
                type="text"
                value={formData.name}
                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                required
              />
            </div>
            <div className="form-group">
              <label>Почта</label>
              <input
                type="email"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                required
              />
            </div>
            <div className="profile-actions">
              <button type="submit" className="btn-save">Save</button>
              <button type="button" onClick={() => setEditing(false)} className="btn-cancel">
                Cancel
              </button>
            </div>
          </form>
        ) : (
          <>
            <div className="profile-info">
              <p><strong>Имя:</strong> {profile.name}</p>
              <p><strong>Почта:</strong> {profile.email}</p>
              

              
              {userRole === 'performer' && (
                <p>
                  <strong>Статус:</strong> 
                  <span className={`status-badge ${profile.verification_status}`}>
                    {profile.verificationStatus === 'verified' ? '✅ Verified' : 
                     profile.verificationStatus === 'pending' ? '⏳ Pending' : 
                     profile.verificationStatus === 'rejected' ? '❌ Rejected' : profile.verification_status}
                  </span>
                </p>
              )}
              
              {profile.inn && userRole === 'performer' && (
                <p><strong>Инн:</strong> {profile.inn}</p>
              )}
              {profile.businessType && userRole === 'performer' && (
                <p><strong>Тип бизнесса:</strong> {profile.businessType}</p>
              )}
            </div>
            <div className="profile-actions">
              <button onClick={() => setEditing(true)} className="btn-edit">Изменить профиль</button>
              <button onClick={handleDelete} className="btn-delete">Удалить аккаунт</button>
            </div>
          </>
        )}
      </div>
    </div>
  );
};

export default Profile;