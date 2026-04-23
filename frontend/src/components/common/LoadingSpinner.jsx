import React from 'react';
import './LoadingSpinner.css';

const LoadingSpinner = ({ size = 'medium', message = 'Loading...' }) => {
  const sizes = {
    small: '20px',
    medium: '40px',
    large: '60px'
  };

  return (
    <div className="spinner-container">
      <div 
        className="spinner" 
        style={{ 
          width: sizes[size], 
          height: sizes[size],
          borderWidth: size === 'small' ? '2px' : size === 'large' ? '6px' : '4px'
        }}
      />
      {message && <p className="spinner-message">{message}</p>}
    </div>
  );
};

export default LoadingSpinner;