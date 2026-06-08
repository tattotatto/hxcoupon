import { useState } from 'react';
import { Card, Form, Input, Button, Descriptions, Tag, Divider, message } from 'antd';
import { userApi } from '../api/user';
import { useAuthStore } from '../stores/authStore';

const approvalStatusMap: Record<number, { color: string; text: string }> = {
  0: { color: 'processing', text: '待审批' },
  1: { color: 'success', text: '已通过' },
  2: { color: 'error', text: '已拒绝' },
  3: { color: 'warning', text: '已停用' },
};

const memberTypeMap: Record<string, string> = {
  issuer: '发券方',
  consumer: '核销方',
  both: '综合商家',
};

export default function Profile() {
  const { user, fetchProfile } = useAuthStore();
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [form] = Form.useForm();

  if (!user) return null;

  const approvalStatus = approvalStatusMap[user.approval_status] || { color: 'default', text: '未知' };

  const onFinish = async (values: Record<string, string>) => {
    const data: Record<string, string> = {};
    Object.entries(values).forEach(([k, v]) => {
      if (v !== undefined && v !== '') data[k] = v;
    });
    setLoading(true);
    try {
      await userApi.updateProfile(data);
      message.success('资料更新成功');
      await fetchProfile();
      setEditing(false);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: 720 }}>
      <Card title="个人资料" style={{ borderRadius: 12 }}>
        {!editing ? (
          <>
            <Descriptions column={2} bordered size="small">
              <Descriptions.Item label="用户名">{user.username}</Descriptions.Item>
              <Descriptions.Item label="角色">
                <Tag color={user.role === 'super_admin' ? 'red' : user.role === 'admin' ? 'blue' : 'default'}>
                  {user.role === 'super_admin' ? '超级管理员' : user.role === 'admin' ? '管理员' : '商家'}
                </Tag>
              </Descriptions.Item>
              {user.member_type && (
                <Descriptions.Item label="商家类型">
                  {memberTypeMap[user.member_type] || user.member_type}
                </Descriptions.Item>
              )}
              <Descriptions.Item label="审批状态">
                <Tag color={approvalStatus.color}>{approvalStatus.text}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="公司名称">{user.company_name || '-'}</Descriptions.Item>
              <Descriptions.Item label="联系人">{user.contact_name || '-'}</Descriptions.Item>
              <Descriptions.Item label="联系电话">{user.contact_phone || '-'}</Descriptions.Item>
              <Descriptions.Item label="邮箱">{user.email || '-'}</Descriptions.Item>
            </Descriptions>
            {user.reject_reason && (
              <Card style={{ marginTop: 16, background: '#fff2f0', border: '1px solid #ffccc7' }} size="small">
                <strong>拒绝原因：</strong>{user.reject_reason}
              </Card>
            )}
            <Button type="primary" style={{ marginTop: 24 }} onClick={() => setEditing(true)}>
              编辑资料
            </Button>
          </>
        ) : (
          <Form
            form={form}
            layout="vertical"
            initialValues={{
              contact_name: user.contact_name,
              contact_phone: user.contact_phone,
              email: user.email,
              company_name: user.company_name,
            }}
            onFinish={onFinish}
          >
            <Form.Item name="contact_name" label="联系人">
              <Input />
            </Form.Item>
            <Form.Item name="contact_phone" label="联系电话">
              <Input />
            </Form.Item>
            <Form.Item name="email" label="邮箱">
              <Input />
            </Form.Item>
            <Form.Item name="company_name" label="公司名称">
              <Input />
            </Form.Item>
            <Divider>修改密码</Divider>
            <Form.Item name="old_password" label="当前密码">
              <Input.Password placeholder="不修改密码可不填" />
            </Form.Item>
            <Form.Item name="new_password" label="新密码">
              <Input.Password placeholder="至少6个字符" />
            </Form.Item>
            <Form.Item>
              <Button type="primary" htmlType="submit" loading={loading}>
                保存
              </Button>
              <Button style={{ marginLeft: 12 }} onClick={() => setEditing(false)}>
                取消
              </Button>
            </Form.Item>
          </Form>
        )}
      </Card>
    </div>
  );
}
