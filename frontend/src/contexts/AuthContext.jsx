import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { authApi } from '../api/authClient';

const AuthContext = createContext(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('accessToken');
      if (token) {
        const userData = authApi.getUserFromToken();
        if (userData) {
          setUser(userData);
          setIsAuthenticated(true);
        } else {
          // Токен невалидный
          authApi.clearTokens();
        }
      }
      setLoading(false);
    };

    initAuth();
  }, []);

  const register = useCallback(async (userData) => {
    try {
      const response = await authApi.register(userData);
      const userFromToken = authApi.getUserFromToken();
      setUser(userFromToken);
      setIsAuthenticated(true);
      return { success: true, data: response };
    } catch (error) {
      console.error('Registration failed:', error);
      return { 
        success: false, 
        error: error.response?.data?.message || error.message || 'Ошибка регистрации' 
      };
    }
  }, []);

  const login = useCallback(async (credentials) => {
    try {
      const response = await authApi.login(credentials);
      const userFromToken = authApi.getUserFromToken();
      setUser(userFromToken);
      setIsAuthenticated(true);
      return { success: true, data: response };
    } catch (error) {
      console.error('Login failed:', error);
      return { 
        success: false, 
        error: error.response?.data?.message || error.message || 'Ошибка входа' 
      };
    }
  }, []);

  const logout = useCallback(async () => {
    try {
      await authApi.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      setUser(null);
      setIsAuthenticated(false);
    }
  }, []);

  const refreshToken = useCallback(async () => {
    try {
      const success = await authApi.refreshToken();
      if (success) {
        const userFromToken = authApi.getUserFromToken();
        setUser(userFromToken);
        setIsAuthenticated(true);
      } else {
        logout();
      }
      return success;
    } catch (error) {
      logout();
      return false;
    }
  }, [logout]);

  const value = {
    user,
    loading,
    isAuthenticated,
    register,
    login,
    logout,
    refreshToken,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};