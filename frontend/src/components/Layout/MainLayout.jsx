import React from 'react';
import { Outlet, useNavigate } from 'react-router-dom';
import { Layout, Menu, Avatar, Dropdown, Button, Space, Typography } from 'antd';
import {
  HomeOutlined,
  UserOutlined,
  LogoutOutlined,
  DashboardOutlined,
  MenuOutlined,
} from '@ant-design/icons';
import { useAuth } from '../../contexts/AuthContext';

const { Header, Content, Footer } = Layout;
const { Text } = Typography;

const MainLayout = () => {
  const { user, isAuthenticated, logout } = useAuth();
  const navigate = useNavigate();

  const menuItems = [
    {
      key: '/',
      icon: <HomeOutlined />,
      label: 'Главная',
      onClick: () => navigate('/'),
    },
  ];

  if (isAuthenticated) {
    menuItems.push({
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: 'Личный кабинет',
      onClick: () => navigate('/dashboard'),
    });
  }

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: 'Профиль',
      onClick: () => navigate('/profile'),
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Выйти',
      onClick: logout,
      danger: true,
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Header
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          position: 'sticky',
          top: 0,
          zIndex: 1,
          padding: '0 24px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
          <div
            style={{
              fontSize: 20,
              fontWeight: 'bold',
              color: 'white',
              cursor: 'pointer',
            }}
            onClick={() => navigate('/')}
          >
            Ростелеком Услуги
          </div>
          <Menu
            theme="dark"
            mode="horizontal"
            selectedKeys={[location.pathname]}
            items={menuItems}
            style={{ flex: 1, minWidth: 0, background: 'transparent', borderBottom: 'none' }}
          />
        </div>

        {isAuthenticated ? (
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Space style={{ cursor: 'pointer' }}>
              <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#FF6B00' }} />
              <Text style={{ color: 'white' }}>{user?.email}</Text>
            </Space>
          </Dropdown>
        ) : (
          <Space>
            <Button type="link" onClick={() => navigate('/login')} style={{ color: 'white' }}>
              Войти
            </Button>
            <Button type="primary" onClick={() => navigate('/register')} style={{ backgroundColor: '#FF6B00' }}>
              Регистрация
            </Button>
          </Space>
        )}
      </Header>

      <Content style={{ padding: '24px 48px', minHeight: 'calc(100vh - 134px)' }}>
        <Outlet />
      </Content>

      <Footer style={{ textAlign: 'center', background: '#f5f5f5' }}>
        Ростелеком ©{new Date().getFullYear()} - Все права защищены
      </Footer>
    </Layout>
  );
};

export default MainLayout;