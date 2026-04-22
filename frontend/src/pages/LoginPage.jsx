import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, Segmented, Alert, message } from 'antd';
import { MailOutlined, IdcardOutlined, LockOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const { Title, Text, Link } = Typography;

const LoginPage = () => {
  const { login } = useAuth();
  const navigate = useNavigate();
  const [loginType, setLoginType] = useState('email');
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  const onFinish = async (values) => {
    setLoading(true);
    try {
      const credentials = {
        password: values.password,
        ...(loginType === 'email' ? { email: values.email } : { inn: values.inn }),
      };
      
      const result = await login(credentials);
      
      if (result.success) {
        message.success('Успешный вход!');
        navigate('/');
      } else {
        message.error(result.error || 'Неверный email/ИНН или пароль');
      }
    } catch (error) {
      message.error('Ошибка при входе');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 'calc(100vh - 200px)' }}>
      <Card style={{ width: 450, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}>
        <Title level={2} style={{ textAlign: 'center', marginBottom: 8 }}>
          Добро пожаловать
        </Title>
        <Text type="secondary" style={{ display: 'block', textAlign: 'center', marginBottom: 32 }}>
          Войдите в свой аккаунт
        </Text>

        <Segmented
          options={[
            { label: 'По Email', value: 'email', icon: <MailOutlined /> },
            { label: 'По ИНН', value: 'inn', icon: <IdcardOutlined /> },
          ]}
          value={loginType}
          onChange={setLoginType}
          block
          style={{ marginBottom: 24 }}
        />

        <Form form={form} layout="vertical" onFinish={onFinish}>
          {loginType === 'email' ? (
            <Form.Item
              name="email"
              label="Email"
              rules={[
                { required: true, message: 'Введите email' },
                { type: 'email', message: 'Введите корректный email' },
              ]}
            >
              <Input prefix={<MailOutlined />} placeholder="example@mail.com" size="large" />
            </Form.Item>
          ) : (
            <Form.Item
              name="inn"
              label="ИНН"
              rules={[
                { required: true, message: 'Введите ИНН' },
                { len: 12, message: 'ИНН должен содержать 12 цифр' },
              ]}
            >
              <Input prefix={<IdcardOutlined />} placeholder="123456789012" size="large" maxLength={12} />
            </Form.Item>
          )}

          <Form.Item
            name="password"
            label="Пароль"
            rules={[{ required: true, message: 'Введите пароль' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="Введите пароль" size="large" />
          </Form.Item>

          <Form.Item>
            <Button
              type="primary"
              htmlType="submit"
              size="large"
              block
              loading={loading}
              style={{ backgroundColor: '#FF6B00' }}
            >
              Войти
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <Text type="secondary">
              Нет аккаунта? <Link href="/register">Зарегистрироваться</Link>
            </Text>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default LoginPage;