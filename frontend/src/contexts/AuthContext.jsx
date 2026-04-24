import React, { createContext, useState, useContext, useEffect } from 'react';
import authAPI from '../api/auth';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

const isTokenValid = () => {
  const token = localStorage.getItem('accessToken');
  if (!token) return false;
  
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    const exp = payload.exp * 1000;
    return Date.now() < exp;
  } catch {
    return false;
  }
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [userRole, setUserRole] = useState(null);

  const isAuthenticated = () => {
    return !!localStorage.getItem('accessToken');
  };

  useEffect(() => {
    const token = localStorage.getItem('accessToken');
    if (token) {
      const savedRole = localStorage.getItem('userRole');
      if (savedRole) {
        setUserRole(savedRole);
        setUser({ authenticated: true, role: savedRole });
      } else {
        setUser({ authenticated: true });
      }
    }
    setLoading(false);

    const handleLogout = () => {
      setUser(null);
      setUserRole(null);
    };
    window.addEventListener('auth:logout', handleLogout);
    return () => window.removeEventListener('auth:logout', handleLogout);
  }, []);

  useEffect(() => {
    const checkToken = async () => {
      if (isAuthenticated() && !isTokenValid()) {
        const refreshed = await authAPI.refreshToken();
        if (!refreshed) {
          logout();
        }
      }
    };
    
    const interval = setInterval(checkToken, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []); 
  const login = async (credentials) => {
    try {
      const result = await authAPI.login(credentials);
      if (result.accessToken) {
        try {
          const payload = JSON.parse(atob(result.accessToken.split('.')[1]));
          const role = payload.role;
          if (role) {
            localStorage.setItem('userRole', role);
            setUserRole(role);
            setUser({ authenticated: true, role });
          } else {
            setUser({ authenticated: true });
          }
        } catch {
          setUser({ authenticated: true });
        }
        return { success: true };
      }
      return { success: false, error: result.error || 'Login failed' };
    } catch (error) {
      return { success: false, error: error.message };
    }
  };

  const register = async (userData) => {
    try {
      const result = await authAPI.register(userData);
      if (result.accessToken) {
        try {
          const payload = JSON.parse(atob(result.accessToken.split('.')[1]));
          const role = payload.role;
          if (role) {
            localStorage.setItem('userRole', role);
            setUserRole(role);
            setUser({ authenticated: true, role });
          } else {
            setUser({ authenticated: true });
          }
        } catch {
          setUser({ authenticated: true });
        }
        return { success: true };
      }
      return { success: false, error: result.error || 'Registration failed' };
    } catch (error) {
      return { success: false, error: error.message };
    }
  };

  const logout = async () => {
    await authAPI.logout();
    setUser(null);
    setUserRole(null);
  };

  const hasRole = (role) => {
    return userRole === role;
  };

  return (
    <AuthContext.Provider value={{ 
      user, 
      loading, 
      login, 
      register, 
      logout,
      userRole,
      hasRole,
      isAuthenticated: !!user
    }}>
      {children}
    </AuthContext.Provider>
  );
};