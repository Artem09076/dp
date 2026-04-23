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
    
    // Добавляем секунды и UTC
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
      
      // Обрабатываем разные типы ошибок
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
      
      // Русские сообщения для понятных ошибок
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
        
        <h3>Book Service: {service?.title}</h3>
        
        <div className="service-info">
          <p className="service-duration">⏱️ Duration: {service?.durationMinutes} minutes</p>
          <p className="service-price">Original Price: ${originalPrice}</p>
          {hasDiscount && (
            <p className="service-final-price">Final Price: <strong>${finalPrice.toFixed(2)}</strong></p>
          )}
        </div>
        
        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label>Select Date and Time:</label>
            <input
              type="datetime-local"
              value={bookingTime}
              onChange={(e) => setBookingTime(e.target.value)}
              required
              min={getMinDateTime()}
            />
            <small className="form-hint">Please select a time at least 1 hour from now</small>
          </div>

          {loadingDiscounts ? (
            <div className="form-group">
              <label>Loading discounts...</label>
            </div>
          ) : discounts.length > 0 ? (
            <div className="form-group">
              <label>Apply Discount (Optional):</label>
              <select 
                value={selectedDiscount} 
                onChange={(e) => setSelectedDiscount(e.target.value)}
                className="discount-select"
              >
                <option value="">No discount</option>
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
                      {discountText} - Save ${savings.toFixed(2)} (Valid until {new Date(discount.validTo).toLocaleDateString()})
                    </option>
                  );
                })}
              </select>
              {selectedDiscount && (
                <p className="savings-info">
                  🎉 You save ${(originalPrice - finalPrice).toFixed(2)}!
                </p>
              )}
            </div>
          ) : (
            <div className="form-group">
              <p className="no-discounts">No active discounts available</p>
            </div>
          )}

          {error && (
            <div className="error-message">
              ❌ {error}
            </div>
          )}

          <button type="submit" disabled={loading} className="btn-confirm-booking">
            {loading ? 'Creating Booking...' : `Confirm Booking - $${finalPrice.toFixed(2)}`}
          </button>
        </form>
      </div>
    </div>
  );
};

export default BookingCreate;