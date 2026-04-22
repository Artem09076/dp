import React, { useState } from 'react';
import { Form, Input, Button, Card, Typography, Steps, Radio, message, Alert } from 'antd';
import { UserOutlined, MailOutlined, LockOutlined, IdcardOutlined, BuildOutlined } from '@ant-design/icons';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

const { Title, Text, Link } = Typography;
const { Step } = Steps;

const RegisterPage = () => {
  const { register } = useAuth();
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(0);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();
  const [userRole, setUserRole] = useState('customer');

  const steps = [
    { title: 'Выберите роль', content: 'RoleSelection' },
    { title: 'Личные данные', content: 'PersonalInfo' },
    { title: 'Данные компании', content: 'CompanyInfo' },
  ];

  const onRoleChange = (e) => {
    setUserRole(e.target.value);
    setCurrentStep(1);
  };

  const onNext = async () => {
    try {
      if (currentStep === 0) {
        setCurrentStep(1);
      } else if (currentStep === 1) {
        await form.validateFields(['email', 'name', 'password', 'confirmPassword']);
        setCurrentStep(2);
      } else if (currentStep === 2) {
        if (userRole === 'performer') {
          await form.validateFields(['inn', 'businessType']);
        }
        
        setLoading(true);
        const values = form.getFieldsValue();
        
        const result = await register({
          email: values.email,
          name: values.name,
          password: values.password,
          inn: userRole === 'performer' ? values.inn : '',
          businessType: userRole === 'performer' ? values.businessType : '',
          userRole: userRole,
        });
        
        if (result.success) {
          message.success('Регистрация успешна!');
          navigate('/');
        } else {
          message.error(result.error || 'Ошибка регистрации');
        }
        setLoading(false);
      }
    } catch (error) {
      console.error('Validation failed:', error);
    }
  };

  const onPrev = () => {
    setCurrentStep(currentStep - 1);
  };

  const renderStepContent = () => {
    switch (currentStep) {
      case 0:
        return (
          <div style={{ textAlign: 'center', padding: '24px 0' }}>
            <Radio.Group onChange={onRoleChange} value={userRole} size="large">
              <Space direction="vertical" size="large">
                <Radio value="customer" style={{ fontSize: 16 }}>
                  <div>
                    <strong>Клиент</strong>
                    <div style={{ fontSize: 14, color: '#666' }}>Физическое лицо</div>
                  </div>
                </Radio>
                <Radio value="performer" style={{ fontSize: 16 }}>
                  <div>
                    <strong>Исполнитель</strong>
                    <div style={{ fontSize: 14, color: '#666' }}>Юридическое лицо / ИП</div>
                  </div>
                </Radio>
              </Space>
            </Radio.Group>
          </div>
        );
      
      case 1:
        return (
          <>
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
            <Form.Item
              name="name"
              label="Имя"
              rules={[{ required: true, message: 'Введите имя' }]}
            >
              <Input prefix={<UserOutlined />} placeholder="Иван Иванов" size="large" />
            </Form.Item>
            <Form.Item
              name="password"
              label="Пароль"
              rules={[
                { required: true, message: 'Введите пароль' },
                { min: 6, message: 'Пароль должен содержать минимум 6 символов' },
              ]}
              hasFeedback
            >
              <Input.Password prefix={<LockOutlined />} placeholder="Введите пароль" size="large" />
            </Form.Item>
            <Form.Item
              name="confirmPassword"
              label="Подтвердите пароль"
              dependencies={['password']}
              rules={[
                { required: true, message: 'Подтвердите пароль' },
                ({ getFieldValue }) => ({
                  validator(_, value) {
                    if (!value || getFieldValue('password') === value) {
                      return Promise.resolve();
                    }
                    return Promise.reject(new Error('Пароли не совпадают'));
                  },
                }),
              ]}
              hasFeedback
            >
              <Input.Password prefix={<LockOutlined />} placeholder="Подтвердите пароль" size="large" />
            </Form.Item>
          </>
        );
      
      case 2:
        if (userRole === 'performer') {
          return (
            <>
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
              <Form.Item
                name="businessType"
                label="Тип бизнеса"
                rules={[{ required: true, message: 'Введите тип бизнеса' }]}
              >
                <Input prefix={<BuildOutlined />} placeholder="ООО, ИП, ЗАО и т.д." size="large" />
              </Form.Item>
            </>
          );
        }
        return (
          <Alert
            message="Проверьте данные"
            description="Пожалуйста, проверьте введенные данные перед завершением регистрации"
            type="info"
            showIcon
          />
        );
      
      default:
        return null;
    }
  };

  return (
    <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', minHeight: 'calc(100vh - 200px)' }}>
      <Card style={{ width: 500, boxShadow: '0 4px 12px rgba(0,0,0,0.1)' }}>
        <Title level={2} style={{ textAlign: 'center', marginBottom: 8 }}>
          Регистрация
        </Title>
        <Text type="secondary" style={{ display: 'block', textAlign: 'center', marginBottom: 32 }}>
          Присоединяйтесь к платформе услуг Ростелеком
        </Text>

        <Steps current={currentStep} style={{ marginBottom: 32 }}>
          {steps.map((step) => (
            <Step key={step.title} title={step.title} />
          ))}
        </Steps>

        <Form form={form} layout="vertical">
          {renderStepContent()}
        </Form>

        <div style={{ display: 'flex', justifyContent: 'space-between', marginTop: 24 }}>
          {currentStep > 0 && (
            <Button onClick={onPrev} size="large">
              Назад
            </Button>
          )}
          <Button
            type="primary"
            onClick={onNext}
            loading={loading}
            size="large"
            style={{ marginLeft: currentStep === 0 ? 'auto' : 0, backgroundColor: currentStep === steps.length - 1 ? '#FF6B00' : undefined }}
          >
            {currentStep === steps.length - 1 ? 'Зарегистрироваться' : 'Далее'}
          </Button>
        </div>

        {currentStep === 0 && (
          <div style={{ marginTop: 24, textAlign: 'center' }}>
            <Text type="secondary">
              Уже есть аккаунт? <Link href="/login">Войти</Link>
            </Text>
          </div>
        )}
      </Card>
    </div>
  );
};

export default RegisterPage;