import { useState } from 'react';
import { useNavigate, Link } from 'react-router-dom';
import { Form, Input, Button, Card, Typography, Select, message, Space, Steps } from 'antd';
import { UserOutlined, LockOutlined, PhoneOutlined, MailOutlined, ShopOutlined } from '@ant-design/icons';
import { authApi } from '../api/auth';
import { useAuthStore } from '../stores/authStore';

const { Title, Text } = Typography;

export default function Register() {
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<{ id: number; username: string; message: string } | null>(null);
  const { isLoggedIn } = useAuthStore();
  const navigate = useNavigate();

  if (isLoggedIn) {
    navigate('/admin', { replace: true });
    return null;
  }

  const onFinish = async (values: Record<string, string>) => {
    setLoading(true);
    try {
      const res = await authApi.register(values as any);
      setResult({ ...res.data.data, member_type: values.member_type });
      message.success('注册成功');
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  if (result) {
    return (
      <div
        style={{
          minHeight: '100vh',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        }}
      >
        <Card style={{ width: 480, borderRadius: 12, textAlign: 'center' }}>
          <Title level={3} style={{ color: '#52c41a' }}>注册成功</Title>
          <Text>您的账号 <strong>{result.username}</strong> 注册成功。</Text>
          <br />
          {result.member_type === 'consumer' ? (
            <Text type="success">核销方账号已自动激活，可直接登录使用。</Text>
          ) : (
            <Text type="secondary">请等待管理员审批后即可登录使用。</Text>
          )}
          <div style={{ marginTop: 24 }}>
            <Space>
              <Link to="/login">
                <Button type="primary">前往登录</Button>
              </Link>
            </Space>
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        justifyContent: 'center',
        alignItems: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '40px 0',
      }}
    >
      <Card style={{ width: 520, borderRadius: 12, boxShadow: '0 8px 40px rgba(0,0,0,0.12)' }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <Title level={3} style={{ marginBottom: 4 }}>商家入驻</Title>
          <Text type="secondary">注册成为优惠券平台商家</Text>
        </div>

        <Form
          name="register"
          onFinish={onFinish}
          size="large"
          layout="vertical"
          autoComplete="off"
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, max: 64, message: '用户名长度 3-64 个字符' },
            ]}
          >
            <Input prefix={<UserOutlined />} placeholder="3-64 个字符" />
          </Form.Item>

          <Form.Item
            name="password"
            label="密码"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, max: 128, message: '密码长度 6-128 个字符' },
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder="6-128 个字符" />
          </Form.Item>

          <Form.Item
            name="member_type"
            label="商家类型"
            rules={[{ required: true, message: '请选择商家类型' }]}
          >
            <Select
              placeholder="选择商家类型"
              options={[
                { value: 'issuer', label: '发券方 — 可创建模板并发券' },
                { value: 'consumer', label: '核销方 — 可核销优惠券' },
                { value: 'both', label: '综合商家 — 发券 + 核销' },
              ]}
            />
          </Form.Item>

          <Form.Item
            name="company_name"
            label="公司名称"
            rules={[{ max: 128 }]}
          >
            <Input prefix={<ShopOutlined />} placeholder="选填" />
          </Form.Item>

          <Form.Item
            name="contact_name"
            label="联系人"
            rules={[{ max: 64 }]}
          >
            <Input prefix={<UserOutlined />} placeholder="选填" />
          </Form.Item>

          <Form.Item
            name="contact_phone"
            label="联系电话"
            rules={[{ max: 20 }]}
          >
            <Input prefix={<PhoneOutlined />} placeholder="选填" />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱"
            rules={[{ max: 128 }]}
          >
            <Input prefix={<MailOutlined />} placeholder="选填" />
          </Form.Item>

          <Form.Item>
            <Button type="primary" htmlType="submit" loading={loading} block>
              提交注册
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <Space>
              <Text type="secondary">已有账号？</Text>
              <Link to="/login">立即登录</Link>
            </Space>
          </div>
        </Form>
      </Card>
    </div>
  );
}
