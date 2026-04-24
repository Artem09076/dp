import React, { useState, useEffect } from 'react';
import bookingAPI from '../../api/booking';
import coreAPI from '../../api/core';
import './Booking.css';

const BookingCreate = ({ service, onSuccess, onCancel }) => {
  const [bookingTime, setBookingTime] = useState('');
  const [discounts, setDiscounts] =  useState([]);
  const [selectedDiscount, setSelectedDiscount] = useState('');
  const [loading, setLoading] = useState(false);
  const [loadingDiscounts, setLoadingDiscounts] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (service) {
      loadDiscounts();
    }
  }, [service]);

  const loadDiscounts = async () => {
    if (!service) return;
    
    setLoadingDiscounts(true);
    try {
      const discountsData = await coreAPI.getServiceDiscounts(service.id);
      setDiscounts(Array.isArray(discountsData) ? discountsData : []);
    } catch (err) {
      console.error('Failed to load discounts:', err);
      setDiscounts([]);
    } finally {
      setLoadingDiscounts(false);
    }
  };

  const calculateFinalPrice = () => {
    if (!service) return 0;
    
    let price = service.price;
    
    if (selectedDiscount) {
      const discount = discounts.find(d => d.id === selectedDiscount);
      if (discount) {
        if (discount.type === 'percentage') {
          price = price * (1 - discount.value / 100);
        } else if (discount.type === 'fixed') {
          price = price - discount.value;
        }
      }
    }
    
    return Math.max(0, price);
  };

  const formatDateForBackend = (dateTimeLocal) => {
    if (!dateTimeLocal) return null;
    
    return dateTimeLocal + ':00Z';
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    
    if (!bookingTime) {
      setError('Please select booking time');
      return;
    }

    setLoading(true);
    setError('');

    try {
      const formattedTime = formatDateForBackend(bookingTime);
      console.log('Selected time:', bookingTime);
      console.log('Formatted time:', formattedTime);
      
      const bookingData = {
        service_id: service.id,
        booking_time: formattedTime,
      };
      
      if (selectedDiscount) {
        bookingData.discount_id = selectedDiscount;
      }
      
      console.log('Sending booking data:', JSON.stringify(bookingData, null, 2));
      
      const result = await bookingAPI.createBooking(bookingData);
      console.log('Booking result:', result);
      
      alert('Booking created successfully!');
      if (onSuccess) onSuccess();
    } catch (err) {
      console.error('Booking creation error:', err);
      
      let errorMessage = '';
      
      if (typeof err === 'string') {
        errorMessage = err;
      } else if (err.message) {
        errorMessage = err.message;
      } else if (err.error) {
        if (typeof err.error === 'string') {
          errorMessage = err.error;
        } else if (err.error.message) {
          errorMessage = err.error.message;
        } else {
          errorMessage = JSON.stringify(err.error);
        }
      } else {
        errorMessage = 'Failed to create booking';
      }
      
      if (errorMessage.toLowerCase().includes('time') || errorMessage.toLowerCase().includes('busy')) {
        errorMessage = 'This time slot is already booked. Please select another time.';
      } else if (errorMessage.toLowerCase().includes('invalid')) {
        errorMessage = 'Invalid booking time. Please select a future time.';
      }
      
      setError(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const finalPrice = calculateFinalPrice();
  const originalPrice = service?.price || 0;
  const hasDiscount = selectedDiscount && finalPrice < originalPrice;

  const getMinDateTime = () => {
    const now = new Date();
    now.setHours(now.getHours() + 1);
    return now.toISOString().slice(0, 16);
  };

  return (
    <div className="modal" onClick={onCancel}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        <button className="modal-close" onClick={onCancel}>×</button>
        
        <h3>Забронировать: {service?.title}</h3>
        
        <div className="service-info">
          <p className="service-duration">⏱️ Продолжительность: {service?.durationMinutes} минут</p>
          <p className="service-price">Начальная цена: ${originalPrice}</p>
          {hasDiscount && (
            <p className="service-final-price">Финальная цена: <strong>${finalPrice.toFixed(2)}</strong></p>
          )}
        </div>
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Выберите дату и время бронирования:</label>
            <input
              type="datetime-local"
              value={bookingTime}
              onChange={(e) => setBookingTime(e.target.value)}
              required
              min={getMinDateTime()}
            />
            <small className="form-hint">Пожалуйста, выберите время минимум через 1 час.</small>
          </div>

          {loadingDiscounts ? (
            <div className="form-group">
              <label>Загрузка скидок...</label>
            </div>
          ) : discounts.length > 0 ? (
            <div className="form-group">
              <label>Применить скидку:</label>
              <select 
                value={selectedDiscount} 
                onChange={(e) => setSelectedDiscount(e.target.value)}
                className="discount-select"
              >
                <option value="">Нет скидок</option>
                {discounts.map(discount => {
                  let discountText = '';
                  let savings = 0;
                  
                  if (discount.type === 'percentage') {
                    discountText = `${discount.value}% off`;
                    savings = originalPrice * discount.value / 100;
                  } else if (discount.type === 'fixed') {
                    discountText = `$${discount.value} off`;
                    savings = discount.value;
                  }
                  
                  return (
                    <option key={discount.id} value={discount.id}>
                      {discountText} - Сохранить ${savings.toFixed(2)} (Скидка деуствительна до {new Date(discount.validTo).toLocaleDateString()})
                    </option>
                  );
                })}
              </select>
              {selectedDiscount && (
                <p className="savings-info">
                  Разница со скидкой ${(originalPrice - finalPrice).toFixed(2)}!
                </p>
              )}
            </div>
          ) : (
            <div className="form-group">
              <p className="no-discounts">Нет доступных скидок</p>
            </div>
          )}

          {error && (
            <div className="error-message">
              ❌ {error}
            </div>
          )}

          <button type="submit" disabled={loading} className="btn-confirm-booking">
            {loading ? 'Создание бронирования...' : `Подтвердить бронирования - ${finalPrice.toFixed(2)}`}
          </button>
        </form>
      </div>
    </div>
  );
};

export default BookingCreate;