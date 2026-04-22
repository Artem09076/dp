import React from 'react';
import { Typography, Button, Row, Col, Card, Space } from 'antd';
import { RocketOutlined, SafetyOutlined, CustomerServiceOutlined, ArrowRightOutlined } from '@ant-design/icons';
import { useAuth } from '../contexts/AuthContext';

const { Title, Paragraph, Text } = Typography;

const features = [
  {
    icon: <SafetyOutlined style={{ fontSize: 48, color: '#6B2EE8' }} />,
    title: 'Безопасность',
    description: 'Ваши данные защищены современными технологиями шифрования',
  },
  {
    icon: <RocketOutlined style={{ fontSize: 48, color: '#6B2EE8' }} />,
    title: 'Быстрота',
    description: 'Мгновенное бронирование и подтверждение услуг',
  },
  {
    icon: <CustomerServiceOutlined style={{ fontSize: 48, color: '#6B2EE8' }} />,
    title: 'Поддержка 24/7',
    description: 'Круглосуточная поддержка клиентов',
  },
];

const HomePage = () => {
  const { isAuthenticated, user } = useAuth();

  return (
    <div>
      {/* Hero Section */}
      <div
        style={{
          background: 'linear-gradient(135deg, #6B2EE8 0%, #4C1D95 100%)',
          borderRadius: 24,
          padding: '64px 48px',
          marginBottom: 48,
          textAlign: 'center',
          color: 'white',
        }}
      >
        <Title level={1} style={{ color: 'white', marginBottom: 16 }}>
          Бронирование услуг
        </Title>
        <Paragraph style={{ fontSize: 18, color: 'rgba(255,255,255,0.9)', marginBottom: 32 }}>
          Просто, быстро и надежно
        </Paragraph>
        {!isAuthenticated && (
          <Space size="middle">
            <Button
              type="primary"
              size="large"
              href="/register"
              style={{ backgroundColor: '#FF6B00', borderColor: '#FF6B00' }}
            >
              Начать <ArrowRightOutlined />
            </Button>
            <Button
              size="large"
              href="/login"
              style={{ color: 'white', borderColor: 'white' }}
              ghost
            >
              Войти
            </Button>
          </Space>
        )}
      </div>

      {/* Features Section */}
      <Title level={2} style={{ textAlign: 'center', marginBottom: 48 }}>
        Почему выбирают нас
      </Title>
      <Row gutter={[24, 24]}>
        {features.map((feature, index) => (
          <Col xs={24} md={8} key={index}>
            <Card
              hoverable
              style={{ textAlign: 'center', height: '100%' }}
              bodyStyle={{ padding: 32 }}
            >
              <div style={{ marginBottom: 16 }}>{feature.icon}</div>
              <Title level={4}>{feature.title}</Title>
              <Paragraph type="secondary">{feature.description}</Paragraph>
            </Card>
          </Col>
        ))}
      </Row>

      {/* Welcome Section for Authenticated Users */}
      {isAuthenticated && (
        <Card
          style={{
            marginTop: 48,
            textAlign: 'center',
            background: 'linear-gradient(135deg, #6B2EE8 0%, #4C1D95 100%)',
          }}
          bodyStyle={{ padding: 48 }}
        >
          <Title level={3} style={{ color: 'white' }}>
            Добро пожаловать, {user?.email}!
          </Title>
          <Paragraph style={{ color: 'rgba(255,255,255,0.9)', fontSize: 16 }}>
            Ваша роль: {user?.role === 'performer' ? 'Исполнитель' : 'Клиент'}
          </Paragraph>
          <Button
            type="primary"
            size="large"
            href="/dashboard"
            style={{ backgroundColor: '#FF6B00', borderColor: '#FF6B00' }}
          >
            Перейти в личный кабинет
          </Button>
        </Card>
      )}
    </div>
  );
};

export default HomePage;