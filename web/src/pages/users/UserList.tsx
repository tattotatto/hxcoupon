import { useState, useEffect, useCallback } from 'react';
import { Table, Card, Tag, Space, Button, Input, Select, Modal, Form, message, Popconfirm, Tooltip } from 'antd';
import { CheckOutlined, CloseOutlined, PauseCircleOutlined, PlayCircleOutlined, SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { userApi } from '../../api/user';

const approvalMap: Record<number, { color: string; text: string }> = {
  0: { color: 'processing', text: '待审批' },
  1: { color: 'success', text: '已通过' },
  2: { color: 'error', text: '已拒绝' },
  3: { color: 'warning', text: '已停用' },
};

const memberTypeMap: Record<string, string> = {
  issuer: '发券方',
  consumer: '用券方',
  both: '综合',
};

export default function UserList() {
  const [data, setData] = useState<{ items: any[]; total: number }>({ items: [], total: 0 });
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState<Record<string, any>>({ page: 1, page_size: 20 });
  const [reasonModal, setReasonModal] = useState<{ open: boolean; id: number; action: string; title: string } | null>(null);
  const [reason, setReason] = useState('');
  const [actionLoading, setActionLoading] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await userApi.list(filters);
      setData(res.data.data);
    } catch {
      // handled
    } finally {
      setLoading(false);
    }
  }, [filters]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const handleAction = async (id: number, action: string) => {
    if (action === 'unsuspend') {
      setActionLoading(true);
      try {
        await userApi.unsuspend(id);
        message.success('操作成功');
        fetchData();
      } catch { /* handled */ }
      finally { setActionLoading(false); }
      return;
    }
    const titles: Record<string, string> = {
      approve: '审批通过',
      reject: '拒绝',
      suspend: '停用',
    };
    setReasonModal({ open: true, id, action, title: titles[action] || action });
  };

  const handleReasonSubmit = async () => {
    if (!reasonModal) return;
    setActionLoading(true);
    try {
      const { id, action } = reasonModal;
      if (action === 'approve') await userApi.approve(id, reason || undefined);
      else if (action === 'reject') await userApi.reject(id, reason);
      else if (action === 'suspend') await userApi.suspend(id, reason);
      message.success('操作成功');
      setReasonModal(null);
      setReason('');
      fetchData();
    } catch { /* handled */ }
    finally { setActionLoading(false); }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '用户名', dataIndex: 'username', width: 120 },
    {
      title: '角色', dataIndex: 'role', width: 80,
      render: (r: string) => (
        <Tag color={r === 'super_admin' ? 'red' : r === 'admin' ? 'blue' : 'default'}>
          {r === 'super_admin' ? '超管' : r === 'admin' ? '管理员' : '商家'}
        </Tag>
      ),
    },
    {
      title: '商家类型', dataIndex: 'member_type', width: 80,
      render: (t: string | null) => t ? memberTypeMap[t] || t : '-',
    },
    {
      title: '审批状态', dataIndex: 'approval_status', width: 90,
      render: (s: number) => {
        const item = approvalMap[s];
        return item ? <Tag color={item.color}>{item.text}</Tag> : '-';
      },
    },
    { title: '公司', dataIndex: 'company_name', width: 140, ellipsis: true },
    { title: '联系人', dataIndex: 'contact_name', width: 100, ellipsis: true },
    { title: '电话', dataIndex: 'contact_phone', width: 120 },
    {
      title: '注册时间', dataIndex: 'created_at', width: 160,
      render: (t: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '最后登录', dataIndex: 'last_login_at', width: 160,
      render: (t: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '操作', key: 'actions', fixed: 'right' as const, width: 200,
      render: (_: any, record: any) => (
        <Space size="small">
          {record.approval_status === 0 && (
            <>
              <Tooltip title="通过"><Button type="link" size="small" icon={<CheckOutlined style={{ color: '#52c41a' }} />} onClick={() => handleAction(record.id, 'approve')} /></Tooltip>
              <Tooltip title="拒绝"><Button type="link" size="small" icon={<CloseOutlined style={{ color: '#ff4d4f' }} />} onClick={() => handleAction(record.id, 'reject')} /></Tooltip>
            </>
          )}
          {record.approval_status === 1 && (
            <Tooltip title="停用"><Button type="link" size="small" icon={<PauseCircleOutlined style={{ color: '#faad14' }} />} onClick={() => handleAction(record.id, 'suspend')} /></Tooltip>
          )}
          {record.approval_status === 3 && (
            <Tooltip title="恢复"><Button type="link" size="small" icon={<PlayCircleOutlined style={{ color: '#52c41a' }} />} onClick={() => handleAction(record.id, 'unsuspend')} /></Tooltip>
          )}
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card style={{ borderRadius: 12 }}>
        <Space style={{ marginBottom: 16 }} wrap>
          <Input
            placeholder="搜索用户名/公司"
            prefix={<SearchOutlined />}
            allowClear
            style={{ width: 200 }}
            onChange={(e) => setFilters((f: any) => ({ ...f, keyword: e.target.value || undefined, page: 1 }))}
          />
          <Select
            placeholder="角色"
            allowClear
            style={{ width: 120 }}
            options={[
              { value: 'super_admin', label: '超管' },
              { value: 'admin', label: '管理员' },
              { value: 'member', label: '商家' },
            ]}
            onChange={(v) => setFilters((f: any) => ({ ...f, role: v, page: 1 }))}
          />
          <Select
            placeholder="审批状态"
            allowClear
            style={{ width: 120 }}
            options={[
              { value: 0, label: '待审批' },
              { value: 1, label: '已通过' },
              { value: 2, label: '已拒绝' },
              { value: 3, label: '已停用' },
            ]}
            onChange={(v) => setFilters((f: any) => ({ ...f, approval_status: v, page: 1 }))}
          />
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>

        <Table
          rowKey="id"
          columns={columns}
          dataSource={data.items}
          loading={loading}
          scroll={{ x: 1200 }}
          pagination={{
            current: filters.page,
            pageSize: filters.page_size,
            total: data.total,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => setFilters((f: any) => ({ ...f, page: p, page_size: ps })),
          }}
        />
      </Card>

      <Modal
        open={reasonModal?.open}
        title={reasonModal?.title}
        onOk={handleReasonSubmit}
        onCancel={() => { setReasonModal(null); setReason(''); }}
        confirmLoading={actionLoading}
        okText="确认"
        cancelText="取消"
      >
        <Form.Item label="原因" style={{ marginTop: 16 }}>
          <Input.TextArea
            rows={3}
            value={reason}
            onChange={(e) => setReason(e.target.value)}
            placeholder={reasonModal?.action === 'approve' ? '可选填' : '必填'}
            maxLength={512}
          />
        </Form.Item>
      </Modal>
    </>
  );
}
