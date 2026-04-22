import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ConfigProvider, App as AntApp, Spin } from 'antd';
import ruRU from 'antd/locale/ru_RU';
import { antdTheme } from './theme/antdTheme';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import MainLayout from './components/Layout/MainLayout';
import HomePage from './pages/HomePage';
import LoginPage from './pages/LoginPage';
import RegisterPage from './pages/RegisterPage';
import './App.css';

const PrivateRoute = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();
  
  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }
  
  return isAuthenticated ? children : <Navigate to="/login" />;
};

function AppContent() {
  return (
    <Routes>
      <Route path="/" element={<MainLayout />}>
        <Route index element={<HomePage />} />
        <Route path="login" element={<LoginPage />} />
        <Route path="register" element={<RegisterPage />} />
        <Route
          path="dashboard"
          element={
            <PrivateRoute>
              <div style={{ padding: 24 }}>Dashboard - в разработке</div>
            </PrivateRoute>
          }
        />
      </Route>
    </Routes>
  );
}

function App() {
  return (
    <ConfigProvider theme={antdTheme} locale={ruRU}>
      <AntApp>
        <Router>
          <AuthProvider>
            <AppContent />
          </AuthProvider>
        </Router>
      </AntApp>
    </ConfigProvider>
  );
}

export default App;