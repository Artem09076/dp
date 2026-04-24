import React from 'react';
import { useNavigate } from 'react-router-dom';
import './NotFoundPage.css';

const NotFoundPage = () => {
  const navigate = useNavigate();

  return (
    <div className="not-found-container">
      <div className="not-found-content">
        <div className="error-code">404</div>
        <h1>Страница не найдена</h1>
        <p>Извините, страница, которую вы ищете, не существует или была перемещена.</p>
        <div className="not-found-actions">
          <button onClick={() => navigate('/')} className="btn-home">
            🏠 На главную
          </button>
          <button onClick={() => navigate(-1)} className="btn-back">
            ← Вернуться назад
          </button>
        </div>
        <div className="not-found-suggestions">
          <h3>Возможно, вы искали:</h3>
          <ul>
            <li><button onClick={() => navigate('/services')}>Поиск услуг</button></li>
            <li><button onClick={() => navigate('/bookings')}>Мои бронирования</button></li>
            <li><button onClick={() => navigate('/profile')}>Мой профиль</button></li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default NotFoundPage;