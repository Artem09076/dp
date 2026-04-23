import React, { useState, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import coreAPI from '../../api/core';
import { useAuth } from '../../contexts/AuthContext';
import './Services.css';

const ServiceList = () => {
  const navigate = useNavigate();
  const [services, setServices] = useState([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [debouncedQuery, setDebouncedQuery] = useState('');
  const [error, setError] = useState('');
  const [initialLoading, setInitialLoading] = useState(true);
  const { user, userRole } = useAuth();
  
  const searchInputRef = useRef(null);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (searchQuery !== debouncedQuery) {
        setDebouncedQuery(searchQuery);
      }
    }, 500);

    return () => clearTimeout(timer);
  }, [searchQuery]);

  useEffect(() => {
    if (debouncedQuery !== undefined) {
      loadServices();
    }
  }, [debouncedQuery]);

  useEffect(() => {
    loadInitialServices();
  }, []);

  const loadInitialServices = async () => {
    try {
      setInitialLoading(true);
      const data = await coreAPI.searchServices('');
      setServices(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Failed to load services:', err);
      setError('Failed to load services. Please try again.');
    } finally {
      setInitialLoading(false);
    }
  };

  const loadServices = async () => {
    try {
      setLoading(true);
      setError('');
      const data = await coreAPI.searchServices(debouncedQuery);
      setServices(Array.isArray(data) ? data : []);
    } catch (err) {
      console.error('Failed to load services:', err);
      setError('Failed to load services. Please try again.');
      setServices([]);
    } finally {
      setLoading(false);
    }
  };

  const handleSearchChange = (e) => {
    const value = e.target.value;
    setSearchQuery(value);
  };

  const handleClearSearch = () => {
    setSearchQuery('');
    setDebouncedQuery('');
    searchInputRef.current?.focus();
  };

  // Обработчик клика по карточке услуги
  const handleServiceClick = (serviceId) => {
    if (userRole === 'performer') {
      navigate(`/services/performer/${serviceId}`);
    } else {
      navigate(`/services/client/${serviceId}`);
    }
  };

  if (initialLoading) {
    return <div className="loading">Loading services...</div>;
  }

  return (
    <div className="service-list">
      <div className="search-bar">
        <div className="search-wrapper">
          <input
            ref={searchInputRef}
            type="text"
            placeholder="Search services by title, description..."
            value={searchQuery}
            onChange={handleSearchChange}
            className="search-input"
            autoFocus
          />
          {loading && (
            <div className="search-spinner">
              <div className="spinner-small"></div>
            </div>
          )}
          {searchQuery && !loading && (
            <button 
              className="search-clear" 
              onClick={handleClearSearch}
              type="button"
            >
              ✕
            </button>
          )}
        </div>
      </div>

      {error && <div className="error-message">{error}</div>}

      {!loading && services.length === 0 && (
        <div className="no-services">
          {searchQuery ? (
            <>No services found for "<strong>{searchQuery}</strong>"</>
          ) : (
            <>No services available. Check back later!</>
          )}
        </div>
      )}

      <div className="services-grid">
        {services.map((service) => (
          <div 
            key={service.id} 
            className="service-card"
            onClick={() => handleServiceClick(service.id)}
            style={{ cursor: 'pointer' }}
          >
            <h3>{service.title}</h3>
            <p className="price">${service.price}</p>
            <p className="duration">⏱️ {service.durationMinutes} minutes</p>
            {service.description && <p className="description">{service.description}</p>}
            <div className="rating">
              {'⭐'.repeat(Math.round(service.averageRating || 0))}
              {service.averageRating ? ` (${service.averageRating})` : ' No reviews yet'}
            </div>
            {!user && (
              <div className="login-hint">Login to book</div>
            )}
          </div>
        ))}
      </div>
    </div>
  );
};

export default ServiceList;