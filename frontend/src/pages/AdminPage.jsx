import React from 'react';
import AdminPanel from '../components/admin/AdminPanel';
import PrivateRoute from '../components/common/PrivateRoute';
import './AdminPage.css';

const AdminPage = () => {
  return (
    <PrivateRoute requiredRole="admin">
      <div className="admin-page">
        <AdminPanel />
      </div>
    </PrivateRoute>
  );
};

export default AdminPage;