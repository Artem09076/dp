import React, { useState } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import './Auth.css';
import { useNavigate } from 'react-router-dom';
const Login = ({ onSuccess, onSwitchToRegister }) => {
  const [isEmailLogin, setIsEmailLogin] = useState(true);
  const [formData, setFormData] = useState({
    email: '',
    inn: '',
    password: '',
  });
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);
  const [showPassword, setShowPassword] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleChange = (e) => {
    const { name, value } = e.target;
    
    if (name === 'inn') {
      const onlyNumbers = value.replace(/[^\d]/g, '').slice(0, 12);
      setFormData(prev => ({ ...prev, [name]: onlyNumbers }));
    } else {
      setFormData(prev => ({ ...prev, [name]: value }));
    }
    setError('');
  };

  const handleSubmit = async (e) => {
  e.preventDefault();
  setLoading(true);
  setError('');

  if (!isEmailLogin && formData.inn.length !== 12) {
    setError('INN must be exactly 12 digits');
    setLoading(false);
    return;
  }

  const credentials = isEmailLogin
    ? { email: formData.email, password: formData.password }
    : { inn: formData.inn, password: formData.password };

  const result = await login(credentials);

  if (result.success) {
    navigate('/');  // Редирект на главную
  } else {
    setError(result.error || 'Login failed. Please check your credentials.');
  }
  setLoading(false);
};

  return (
    <div className="auth-container">
      <div className="auth-card">
        <div className="auth-header">
          <h2>С возвращением</h2>
          <p>Войдите в ваш аккаунт</p>
        </div>
        
        <div className="auth-tabs">
          <button
            type="button"
            className={isEmailLogin ? 'active' : ''}
            onClick={() => {
              setIsEmailLogin(true);
              setError('');
              setFormData(prev => ({ ...prev, email: '', inn: '' }));
            }}
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
              <polyline points="22,6 12,13 2,6"/>
            </svg>
            Почта
          </button>
          <button
            type="button"
            className={!isEmailLogin ? 'active' : ''}
            onClick={() => {
              setIsEmailLogin(false);
              setError('');
              setFormData(prev => ({ ...prev, email: '', inn: '' }));
            }}
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
              <line x1="16" y1="2" x2="16" y2="6"/>
              <line x1="8" y1="2" x2="8" y2="6"/>
              <line x1="3" y1="10" x2="21" y2="10"/>
            </svg>
            Инн
          </button>
        </div>

        <form onSubmit={handleSubmit}>
          {isEmailLogin ? (
            <div className="form-group">
              <label htmlFor="email">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/>
                  <polyline points="22,6 12,13 2,6"/>
                </svg>
                Email
              </label>
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={handleChange}
                required
                placeholder="your@email.com"
                autoComplete="email"
              />
            </div>
          ) : (
            <div className="form-group">
              <label htmlFor="inn">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <rect x="3" y="4" width="18" height="18" rx="2" ry="2"/>
                  <line x1="16" y1="2" x2="16" y2="6"/>
                  <line x1="8" y1="2" x2="8" y2="6"/>
                  <line x1="3" y1="10" x2="21" y2="10"/>
                </svg>
                Инн (12 цифр)
              </label>
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
                autoComplete="off"
              />
              {formData.inn && formData.inn.length === 12 && (
                <span className="input-valid">✓ Допустимый формат ИНН</span>
              )}
            </div>
          )}

          <div className="form-group">
            <label htmlFor="password">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
                <path d="M7 11V7a5 5 0 0 1 10 0v4"/>
              </svg>
              Пароль
            </label>
            <div className="password-input-wrapper">
              <input
                type={showPassword ? "text" : "password"}
                id="password"
                name="password"
                value={formData.password}
                onChange={handleChange}
                required
                placeholder="••••••••"
                autoComplete="current-password"
              />
              <button
                type="button"
                className="toggle-password"
                onClick={() => setShowPassword(!showPassword)}
              >
                {showPassword ? (
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                    <circle cx="12" cy="12" r="3"/>
                    <line x1="3" y1="3" x2="21" y2="21"/>
                  </svg>
                ) : (
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/>
                    <circle cx="12" cy="12" r="3"/>
                  </svg>
                )}
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
                Входим в аккаунт
              </>
            ) : (
              'Войти'
            )}
          </button>
        </form>

        <div className="auth-footer">
          <p>
            Нет аккаунта?{' '}
            <button onClick={() => navigate('/register')} className="link-button">
              Создать аккаунт
            </button>
          </p>
        </div>
      </div>
    </div>
  );
};

export default Login;