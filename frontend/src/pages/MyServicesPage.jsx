import React, { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import coreAPI from '../api/core';
import { useAuth } from '../contexts/AuthContext';
import './MyServicesPage.css';

const MyServicesPage = () => {
  const navigate = useNavigate();
  const { userRole } = useAuth();
  const [services, setServices] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreateForm, setShowCreateForm] = useState(false);
  const [editingService, setEditingService] = useState(null);
  
  const [formData, setFormData] = useState({
    title: '',
    description: '',
    price: '',
    duration_minutes: ''
  });

  useEffect(() => {
    if (userRole === 'performer' || userRole === 'admin') {
      loadMyServices();
    }
  }, [userRole]);

  const loadMyServices = async () => {
    try {
      setLoading(true);
      const data = await coreAPI.getMyServices();
      setServices(Array.isArray(data) ? data : []);
    } catch (err) {
      setError('Failed to load your services');
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateService = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.createService({
        title: formData.title,
        description: formData.description || undefined,
        price: parseFloat(formData.price),
        duration_minutes: parseInt(formData.duration_minutes)
      });
      alert('Service created successfully!');
      setShowCreateForm(false);
      setFormData({ title: '', description: '', price: '', duration_minutes: '' });
      loadMyServices();
    } catch (err) {
      alert('Failed to create service');
    }
  };

  const handleUpdateService = async (e) => {
    e.preventDefault();
    try {
      await coreAPI.updateService(editingService.id, {
        title: formData.title,
        description: formData.description || undefined,
        price: parseFloat(formData.price),
        duration_minutes: parseInt(formData.duration_minutes)
      });
      alert('Service updated successfully!');
      setEditingService(null);
      setFormData({ title: '', description: '', price: '', duration_minutes: '' });
      loadMyServices();
    } catch (err) {
      alert('Failed to update service');
    }
  };

  const handleDeleteService = async (serviceId) => {
    if (window.confirm('Are you sure you want to delete this service? This action cannot be undone.')) {
      try {
        await coreAPI.deleteService(serviceId);
        alert('Service deleted successfully!');
        loadMyServices();
      } catch (err) {
        alert('Failed to delete service');
      }
    }
  };

  const startEdit = (service) => {
    setEditingService(service);
    setFormData({
      title: service.title,
      description: service.description || '',
      price: service.price.toString(),
      duration_minutes: service.durationMinutes.toString()
    });
  };

  const cancelForm = () => {
    setShowCreateForm(false);
    setEditingService(null);
    setFormData({ title: '', description: '', price: '', duration_minutes: '' });
  };

  if (userRole !== 'performer' && userRole !== 'admin') {
    return (
      <div className="my-services-page">
        <div className="access-denied">
          <h2>Access Denied</h2>
          <p>Only performers can manage services.</p>
          <button onClick={() => navigate('/')} className="btn-back">Back to Home</button>
        </div>
      </div>
    );
  }

  if (loading) {
    return <div className="loading">Loading your services...</div>;
  }

  return (
    <div className="my-services-page">
      <div className="page-header">
        <button onClick={() => navigate('/')} className="btn-back-page">
          ← Back to Home
        </button>
        <h1>My Services</h1>
        <button onClick={() => setShowCreateForm(true)} className="btn-create-service">
          + Create New Service
        </button>
      </div>

      {error && <div className="error-message">{error}</div>}

      {showCreateForm && (
        <div className="modal" onClick={cancelForm}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Create New Service</h3>
            <form onSubmit={handleCreateService}>
              <div className="form-group">
                <label>Service Title *</label>
                <input
                  type="text"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  required
                  placeholder="e.g., Web Development, Plumbing..."
                />
              </div>
              <div className="form-group">
                <label>Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows="3"
                  placeholder="Describe your service..."
                />
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>Price ($) *</label>
                  <input
                    type="number"
                    value={formData.price}
                    onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                    required
                    min="0"
                    step="0.01"
                  />
                </div>
                <div className="form-group">
                  <label>Duration (minutes) *</label>
                  <input
                    type="number"
                    value={formData.duration_minutes}
                    onChange={(e) => setFormData({ ...formData, duration_minutes: e.target.value })}
                    required
                    min="15"
                    step="15"
                  />
                </div>
              </div>
              <div className="modal-actions">
                <button type="submit" className="btn-confirm">Create Service</button>
                <button type="button" onClick={cancelForm} className="btn-cancel">Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {editingService && (
        <div className="modal" onClick={cancelForm}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Edit Service: {editingService.title}</h3>
            <form onSubmit={handleUpdateService}>
              <div className="form-group">
                <label>Service Title *</label>
                <input
                  type="text"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  required
                />
              </div>
              <div className="form-group">
                <label>Description</label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  rows="3"
                />
              </div>
              <div className="form-row">
                <div className="form-group">
                  <label>Price ($) *</label>
                  <input
                    type="number"
                    value={formData.price}
                    onChange={(e) => setFormData({ ...formData, price: e.target.value })}
                    required
                    min="0"
                    step="0.01"
                  />
                </div>
                <div className="form-group">
                  <label>Duration (minutes) *</label>
                  <input
                    type="number"
                    value={formData.duration_minutes}
                    onChange={(e) => setFormData({ ...formData, duration_minutes: e.target.value })}
                    required
                    min="15"
                    step="15"
                  />
                </div>
              </div>
              <div className="modal-actions">
                <button type="submit" className="btn-confirm">Update Service</button>
                <button type="button" onClick={cancelForm} className="btn-cancel">Cancel</button>
              </div>
            </form>
          </div>
        </div>
      )}

      {services.length === 0 ? (
        <div className="no-services">
          <p>You haven't created any services yet.</p>
          <button onClick={() => setShowCreateForm(true)} className="btn-create-first">
            Create Your First Service
          </button>
        </div>
      ) : (
        <div className="services-management-list">
          {services.map(service => (
            <div key={service.id} className="service-management-card">
              <div className="service-info">
                <h3>{service.title}</h3>
                {service.description && <p className="service-description">{service.description}</p>}
                <div className="service-stats">
                  <span className="price">💰 ${service.price}</span>
                  <span className="duration">⏱️ {service.duration_minutes} min</span>
                </div>
              </div>
              <div className="service-actions">
                <button onClick={() => navigate(`/services/performer/${service.id}`)} className="btn-view">
                    View
                </button>
                <button onClick={() => startEdit(service)} className="btn-edit">
                  Edit
                </button>
                <button onClick={() => handleDeleteService(service.id)} className="btn-delete">
                  Delete
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default MyServicesPage;