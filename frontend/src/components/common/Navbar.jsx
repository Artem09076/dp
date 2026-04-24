import React from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { useAuth } from '../../contexts/AuthContext';
import './Navbar.css';

const Navbar = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { logout, userRole, isAuthenticated } = useAuth();

  if (!isAuthenticated) return null;

  const isActive = (path) => location.pathname === path;

  return (
    <nav className="navbar">
      <div className="navbar-brand" onClick={() => navigate('/')}>
        ServiceBooking
      </div>
      
      <div className="navbar-menu">
        {userRole != 'admin' &&
        (<button 
          className={isActive('/bookings') ? 'active' : ''}
          onClick={() => navigate('/bookings')}
        >
          Мои бронирования
        </button>)}
        <button 
          className={isActive('/profile') ? 'active' : ''}
          onClick={() => navigate('/profile')}
        >
          Профиль
        </button>
        
        {userRole === 'admin' && (
          <button 
            className={isActive('/admin') ? 'active' : ''}
            onClick={() => navigate('/admin')}
          >
            Панель админа
          </button>
        )}
        
        <button onClick={logout} className="logout-btn">
          Выйти
        </button>
      </div>
    </nav>
  );
};

export default Navbar;