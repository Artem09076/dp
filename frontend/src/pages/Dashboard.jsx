import React from 'react';
import ServiceList from '../components/services/ServiceList';
import './Dashboard.css';

const Dashboard = () => {
  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>Welcome to Service Booking Platform</h1>
        <p>Find and book services from professional performers</p>
      </div>

      <div className="search-section">
        <h2>Search Services</h2>
        <ServiceList />
      </div>
    </div>
  );
};

export default Dashboard;