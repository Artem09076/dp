import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider, useAuth } from './contexts/AuthContext';
import Login from './components/auth/Login';
import Register from './components/auth/Register';
import Dashboard from './pages/Dashboard';
import BookingsPage from './pages/BookingsPage';
import MyServicesPage from './pages/MyServicesPage';
import ClientServiceDetailPage from './pages/ClientServiceDetailPage';
import PerformerServiceDetailPage from './pages/PerformerServiceDetailPage';
import Profile from './components/profile/Profile';
import AdminPanel from './components/admin/AdminPanel';
import Navbar from './components/common/Navbar';
import PrivateRoute from './components/common/PrivateRoute';
import ClientBookingDetailPage from './pages/ClientBookingDetailPage';
import PerformerBookingDetailPage from './pages/PerformerBookingDetailPage';
import './App.css';

const AppContent = () => {
  const { user, loading, userRole } = useAuth();

  if (loading) {
    return (
      <div className="loading-container">
        <div className="loader"></div>
      </div>
    );
  }

  return (
    <>
      {user && <Navbar />}
      <Routes>
        <Route path="/login" element={!user ? <Login /> : <Navigate to="/" />} />
        <Route path="/register" element={!user ? <Register /> : <Navigate to="/" />} />
        <Route path="/" element={user ? <Dashboard /> : <Navigate to="/login" />} />
        <Route path="/bookings" element={<PrivateRoute><BookingsPage /></PrivateRoute>} />
        <Route path="/my-services" element={<PrivateRoute><MyServicesPage /></PrivateRoute>} />
        <Route path="/services/client/:id" element={<PrivateRoute><ClientServiceDetailPage /></PrivateRoute>} />
        <Route path="/services/performer/:id" element={<PrivateRoute requiredRole="performer"><PerformerServiceDetailPage /></PrivateRoute>} />
        <Route path="/bookings/client/:id" element={<PrivateRoute><ClientBookingDetailPage /></PrivateRoute>} />
        <Route path="/bookings/performer/:id" element={<PrivateRoute requiredRole="performer"><PerformerBookingDetailPage /></PrivateRoute>} />
        <Route path="/profile" element={<PrivateRoute><Profile /></PrivateRoute>} />
        <Route path="/admin" element={<PrivateRoute requiredRole="admin"><AdminPanel /></PrivateRoute>} />
      </Routes>
    </>
  );
};

function App() {
  return (
    <Router>
      <AuthProvider>
        <AppContent/>
      </AuthProvider>
    </Router>
  );
}

export default App;