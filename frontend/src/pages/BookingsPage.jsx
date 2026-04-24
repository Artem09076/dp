import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import BookingList from '../components/bookings/BookingList';
import { useAuth } from '../contexts/AuthContext';
import './BookingsPage.css';

const BookingsPage = () => {
  const navigate = useNavigate();
  const { userRole } = useAuth();

  return (
    <div className="bookings-page">
      <div className="page-header">
        <button onClick={() => navigate('/')} className="btn-back-page">
          ← Назад
        </button>
        <h1>Мои бронирования</h1>
      </div>
      
      <BookingList />
    </div>
  );
};

export default BookingsPage;