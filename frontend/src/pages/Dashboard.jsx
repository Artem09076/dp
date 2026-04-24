import React from 'react';
import ServiceList from '../components/services/ServiceList';
import './Dashboard.css';

const Dashboard = () => {
  return (
    <div className="dashboard">
      <div className="dashboard-header">
        <h1>Добро пожаловать на платформу по бронированию услуг</h1>
        <p>Найдите и забранируйте услугу от провериных поставщиков</p>
      </div>

      <div className="search-section">
        <h2>Поиск сервисов</h2>
        <ServiceList />
      </div>
    </div>
  );
};

export default Dashboard;