import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ServiceList from '../components/services/ServiceList';
import ServiceCreate from '../components/services/ServiceCreate';
import { useAuth } from '../contexts/AuthContext';
import './ServicesPage.css';

const ServicesPage = () => {
  const navigate = useNavigate();
  const [activeTab, setActiveTab] = useState('list');
  const { userRole } = useAuth();
  const canCreateServices = userRole === 'performer' || userRole === 'admin';

  return (
    <div className="services-page">
      <div className="page-header">
        <button onClick={() => navigate('/')} className="btn-back-page">
          ← Back to Home
        </button>
        <h1>Services</h1>
        {canCreateServices && (
          <div className="tab-buttons">
            <button 
              className={activeTab === 'list' ? 'active' : ''}
              onClick={() => setActiveTab('list')}
            >
              Browse Services
            </button>
            <button 
              className={activeTab === 'create' ? 'active' : ''}
              onClick={() => setActiveTab('create')}
            >
              Create Service
            </button>
          </div>
        )}
      </div>
      
      {activeTab === 'list' && <ServiceList />}
      {activeTab === 'create' && canCreateServices && <ServiceCreate onSuccess={() => setActiveTab('list')} />}
    </div>
  );
};

export default ServicesPage;