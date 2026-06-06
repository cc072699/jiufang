import { useState } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, message } from 'antd';
import { UserOutlined, LockOutlined } from '@ant-design/icons';
import { login } from '../shared/api';
import { useAuthStore } from '../shared/auth/store';
import { encryptPassword } from '../shared/utils/encrypt';
import { ApiError } from '../shared/api/client';
import type { LoginRequest } from '../shared/api/types';

const { Title, Text } = Typography;

export default function LoginPage() {
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const setAuth = useAuthStore((s) => s.setAuth);

  const from = (location.state as { from?: Location })?.from?.pathname || '/workbench';

  const onFinish = async (values: LoginRequest) => {
    setLoading(true);
    try {
      const encrypted = await encryptPassword(values.password);
      const data = await login({ username: values.username, password: encrypted });
      setAuth(data.token, data.user, data.expires_at);
      message.success('登录成功');
      navigate(from, { replace: true });
    } catch (err: unknown) {
      // detailed-design.md API-002: 40101=用户名或密码错误, 40301=用户已停用
      if (err instanceof ApiError) {
        if (err.code === 40301 || err.code === 40104) {
          message.error('账号已停用，请联系管理员');
        } else if (err.code === 40101 || err.code === 40103 || err.code === 40105 || err.code === 401) {
          message.error('用户名或密码错误');
        } else {
          message.error(err.message || '登录失败');
        }
      } else {
        message.error('登录失败');
      }
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: '#f0f2f5',
      }}
    >
      <Card style={{ width: 400, boxShadow: '0 2px 8px rgba(0,0,0,0.09)' }}>
        <div style={{ textAlign: 'center', marginBottom: 32 }}>
          <Title level={3} style={{ marginBottom: 4 }}>久方查询助手</Title>
          <Text type="secondary">ERP 智能数据查询平台</Text>
        </div>
        <Form name="login" onFinish={onFinish} size="large" autoComplete="off">
          <Form.Item
            name="username"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, max: 50, message: '用户名长度 3~50 个字符' },
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder="用户名" />
          </Form.Item>
          <Form.Item
            name="password"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 8, max: 100, message: '密码长度 8~100 个字符' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="密码" />
          </Form.Item>
          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              登录
            </Button>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
