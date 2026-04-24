import React, { useState } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import './Auth.css';
import { useNavigate } from 'react-router-dom';

const Register = ({ onSuccess, onSwitchToLogin }) => {
  const [formData, setFormData] = useState({
    email: '',
    name: '',
    password: '',
    confirmPassword: '',
    userRole: 'client',
    inn: '',
    businessType: '',
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [passwordStrength, setPasswordStrength] = useState(0);
  const { register } = useAuth();
  const navigate = useNavigate();


  const checkPasswordStrength = (password) => {
    let strength = 0;
    if (password.length >= 6) strength++;
    if (password.match(/[a-z]/) && password.match(/[A-Z]/)) strength++;
    if (password.match(/\d/)) strength++;
    if (password.match(/[^a-zA-Z\d]/)) strength++;
    setPasswordStrength(strength);
  };

  const handleChange = (e) => {
    const { name, value } = e.target;
    
    if (name === 'inn') {
      const onlyNumbers = value.replace(/[^\d]/g, '').slice(0, 12);
      setFormData(prev => ({ ...prev, [name]: onlyNumbers }));
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
    
    if (name === 'password') {
      checkPasswordStrength(value);
    }
    
    setError('');
  };

  const handleRoleChange = (role) => {
    setFormData({
      ...formData,
      userRole: role,
      inn: '',
      businessType: '',
    });
  };

  const validateForm = () => {
    if (formData.password !== formData.confirmPassword) {
      setError('Пароли не совпадают');
      return false;
    }
    
    if (formData.password.length < 6) {
      setError('В пароле должно быть как минимум 6 символов');
      return false;
    }

    if (passwordStrength < 2) {
      setError('Пожалуйста придумайте более сильный пароль');
      return false;
    }

    if (formData.userRole === 'performer') {
      if (!formData.inn || formData.inn.length !== 12) {
        setError('ИНН должен содержать ровно 12 цифр');
        return false;
      }
      if (!formData.businessType) {
        setError('Пожалуйста укажите тип вашего бизнесса');
        return false;
      }
    }

    return true;
  };


const handleSubmit = async (e) => {
  e.preventDefault();
  
  if (!validateForm()) return;
  
  setLoading(true);
  setError('');

  const registerData = {
    email: formData.email,
    name: formData.name,
    password: formData.password,
    userRole: formData.userRole,
    inn: formData.userRole === 'performer' ? formData.inn : '',
    businessType: formData.userRole === 'performer' ? formData.businessType : '',
  };

  const result = await register(registerData);

  if (result.success) {
    navigate('/'); 
    if (onSuccess) onSuccess();
  } else {
    setError(result.error || 'Регистрация провалена. Попробуйте снова.');
  }
  setLoading(false);
};

  const getPasswordStrengthText = () => {
    if (passwordStrength === 0) return 'Очень слабый';
    if (passwordStrength === 1) return 'Слабый';
    if (passwordStrength === 2) return 'Средний';
    if (passwordStrength === 3) return 'Сильный';
    return 'Очень сильный';
  };

  const getPasswordStrengthColor = () => {
    if (passwordStrength === 0) return '#ff4444';
    if (passwordStrength === 1) return '#ff8844';
    if (passwordStrength === 2) return '#ffcc00';
    if (passwordStrength === 3) return '#88cc44';
    return '#44cc88';
  };

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h2>Создать аккаунт</h2>
        </div>

        <form onSubmit={handleSubmit}>
          <div className="form-group">
            <label htmlFor="name">ФИО</label>
            <input
              type="text"
              id="name"
              name="name"
              value={formData.name}
              onChange={handleChange}
              required
              placeholder="Иван Иванович Иванов"
            />
          </div>

          <div className="form-group">
            <label htmlFor="email">Почта</label>
            <input
              type="email"
              id="email"
              name="email"
              value={formData.email}
              onChange={handleChange}
              required
              placeholder="me@email.com"
            />
          </div>

          <div className="form-group">
            <label>Тип аккаунта</label>
            <div className="role-buttons">
              <button
                type="button"
                className={`role-button ${formData.userRole === 'client' ? 'active' : ''}`}
                onClick={() => handleRoleChange('client')}
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
                  <circle cx="12" cy="7" r="4"/>
                </svg>
                Пользователь
              </button>
              <button
                type="button"
                className={`role-button ${formData.userRole === 'performer' ? 'active' : ''}`}
                onClick={() => handleRoleChange('performer')}
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
                  <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
                </svg>
                Поставщик
              </button>
            </div>
          </div>

          {formData.userRole === 'performer' && (
            <div className="form-group slide-down">
              <label htmlFor="inn">ИНН (12 цифр)</label>
              <input
                type="text"
                id="inn"
                name="inn"
                value={formData.inn}
                onChange={handleChange}
                required
                placeholder="123456789012"
                maxLength="12"
                pattern="\d{12}"
              />
              {formData.inn && formData.inn.length === 12 && (
                <span className="input-valid">✓ Инн валидный</span>
              )}

              <label htmlFor="businessType" style={{ marginTop: '1rem' }}>Тип вашего бизнесса</label>
              <select
                id="businessType"
                name="businessType"
                value={formData.businessType}
                onChange={handleChange}
                required
              >
                <option value="">Выберите тип вашего бизнеса</option>
                <option value="self_employed">Самозанятый</option>
                <option value="ip">ИП (Индивидуальный предприниматель)</option>
              </select>
            </div>
          )}

          <div className="form-group">
            <label htmlFor="password">Пароль</label>
            <div className="password-input-wrapper">
              <input
                type={showPassword ? "text" : "password"}
                id="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                required
                placeholder="Придумайте пароль"
              />
              <button
                type="button"
                className="toggle-password"
                onClick={() => setShowPassword(!showPassword)}
              >
                {showPassword ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
            {formData.password && (
              <div className="password-strength">
                <div className="strength-bar">
                  <div 
                    className="strength-fill" 
                    style={{ 
                      width: `${(passwordStrength / 4) * 100}%`,
                      backgroundColor: getPasswordStrengthColor()
                    }}
                  />
                </div>
                <span style={{ color: getPasswordStrengthColor() }}>
                  Пароль: {getPasswordStrengthText()}
                </span>
              </div>
            )}
          </div>

          <div className="form-group">
            <label htmlFor="confirmPassword">Подтвердите пароль</label>
            <div className="password-input-wrapper">
              <input
                type={showConfirmPassword ? "text" : "password"}
                id="confirmPassword"
                name="confirmPassword"
                value={formData.confirmPassword}
                onChange={handleChange}
                required
                placeholder="Подтвердите ваш пароль"
              />
              <button
                type="button"
                className="toggle-password"
                onClick={() => setShowConfirmPassword(!showConfirmPassword)}
              >
                {showConfirmPassword ? '👁️' : '👁️‍🗨️'}
              </button>
            </div>
          </div>

          {error && (
            <div className="error-message">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              {error}
            </div>
          )}

          <button type="submit" disabled={loading} className="auth-button">
            {loading ? (
              <>
                <span className="spinner"></span>
                Создание аккаунта...
              </>
            ) : (
              'Создать аккаунт'
            )}
          </button>
        </form>

        <div className="auth-footer">
          <p>
            Уже есть аккаунт?{' '}
            <button onClick={() => navigate('/login')} className="link-button">
              Войти
            </button>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Register;