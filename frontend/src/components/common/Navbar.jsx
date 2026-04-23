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
        🏠 ServiceBooking
      </div>
      
      <div className="navbar-menu">
        <button 
          className={isActive('/bookings') ? 'active' : ''}
          onClick={() => navigate('/bookings')}
        >
          My Bookings
        </button>
        
        <button 
          className={isActive('/profile') ? 'active' : ''}
          onClick={() => navigate('/profile')}
        >
          Profile
        </button>
        
        {userRole === 'admin' && (
          <button 
            className={isActive('/admin') ? 'active' : ''}
            onClick={() => navigate('/admin')}
          >
            Admin Panel
          </button>
        )}
        
        <button onClick={logout} className="logout-btn">
          Logout
        </button>
      </div>
    </nav>
  );
};

export default Navbar;