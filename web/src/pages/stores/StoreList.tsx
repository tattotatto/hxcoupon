import { useState, useEffect, useCallback } from 'react';
import { Table, Card, Tag, Space, Button, Modal, Form, Input, Select, message, Popconfirm } from 'antd';
import { PlusOutlined, EditOutlined, KeyOutlined, ReloadOutlined } from '@ant-design/icons';
import { storeApi } from '../../api/store';
import CredentialsModal from '../../components/CredentialsModal';

export default function StoreList() {
  const [data, setData] = useState<{ items: any[]; total: number }>({ items: [], total: 0 });
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState<Record<string, any>>({ page: 1, page_size: 20 });
  const [modal, setModal] = useState<{ open: boolean; editing?: any }>({ open: false });
  const [credentials, setCredentials] = useState<{ app_key: string; app_secret: string } | null>(null);
  const [saveLoading, setSaveLoading] = useState(false);
  const [form] = Form.useForm();

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await storeApi.list(filters);
      setData(res.data.data);
    } catch { /* handled */ }
    finally { setLoading(false); }
  }, [filters]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const openCreate = () => {
    form.resetFields();
    setModal({ open: true });
  };

  const openEdit = (record: any) => {
    form.setFieldsValue(record);
    setModal({ open: true, editing: record });
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    setSaveLoading(true);
    try {
      if (modal.editing) {
        await storeApi.update(modal.editing.id, values);
        message.success('更新成功');
      } else {
        await storeApi.create(values);
        message.success('创建成功');
      }
      setModal({ open: false });
      fetchData();
    } catch { /* validation or API error */ }
    finally { setSaveLoading(false); }
  };

  const handleStatus = async (id: number, status: number) => {
    try {
      await storeApi.updateStatus(id, status);
      message.success('状态更新成功');
      fetchData();
    } catch { /* handled */ }
  };

  const handleCredentials = async (id: number) => {
    try {
      const res = await storeApi.generateCredentials(id);
      setCredentials(res.data.data);
    } catch { /* handled */ }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name', width: 150 },
    { title: '编码', dataIndex: 'code', width: 80 },
    { title: 'App ID', dataIndex: 'app_id', width: 120, ellipsis: true },
    {
      title: '类型', dataIndex: 'type', width: 80,
      render: (t: string) => <Tag>{t === 'miniprogram' ? '小程序' : 'H5'}</Tag>,
    },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (s: number, record: any) => (
        <Popconfirm
          title={`确认${s === 1 ? '禁用' : '启用'}？`}
          onConfirm={() => handleStatus(record.id, s === 1 ? 0 : 1)}
        >
          <Tag color={s === 1 ? 'success' : 'error'} style={{ cursor: 'pointer' }}>
            {s === 1 ? '启用' : '禁用'}
          </Tag>
        </Popconfirm>
      ),
    },
    { title: '联系人', dataIndex: 'contact_name', width: 100 },
    { title: '电话', dataIndex: 'contact_phone', width: 120 },
    {
      title: '操作', key: 'actions', fixed: 'right' as const, width: 160,
      render: (_: any, record: any) => (
        <Space>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>编辑</Button>
          <Button type="link" size="small" icon={<KeyOutlined />} onClick={() => handleCredentials(record.id)}>密钥</Button>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card style={{ borderRadius: 12 }}>
        <Space style={{ marginBottom: 16 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增门店</Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>

        <Table
          rowKey="id"
          columns={columns}
          dataSource={data.items}
          loading={loading}
          scroll={{ x: 900 }}
          pagination={{
            current: filters.page,
            pageSize: filters.page_size,
            total: data.total,
            showTotal: (t) => `共 ${t} 条`,
            onChange: (p, ps) => setFilters((f) => ({ ...f, page: p, page_size: ps })),
          }}
        />
      </Card>

      <Modal
        open={modal.open}
        title={modal.editing ? '编辑门店' : '新增门店'}
        onOk={handleSave}
        onCancel={() => setModal({ open: false })}
        confirmLoading={saveLoading}
        width={520}
        okText="保存"
        cancelText="取消"
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true, max: 128 }]}>
            <Input />
          </Form.Item>
          {!modal.editing && (
            <>
              <Form.Item name="app_id" label="App ID" rules={[{ max: 64 }]} tooltip="可选，留空则自动生成">
                <Input placeholder="留空自动生成" />
              </Form.Item>
              <Form.Item name="type" label="类型" rules={[{ required: true }]}>
                <Select options={[{ value: 'miniprogram', label: '小程序' }, { value: 'h5', label: 'H5' }]} />
              </Form.Item>
            </>
          )}
          <Form.Item name="contact_name" label="联系人" rules={[{ max: 64 }]}>
            <Input />
          </Form.Item>
          <Form.Item name="contact_phone" label="电话" rules={[{ max: 20 }]}>
            <Input />
          </Form.Item>
          <Form.Item name="remark" label="备注" rules={[{ max: 512 }]}>
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="mp_appid" label="小程序 AppID" rules={[{ max: 64 }]} tooltip="用于优惠券'去用券'跳转目标小程序">
            <Input placeholder="wxXXXXXXXX" />
          </Form.Item>
          <Form.Item name="mp_page_path" label="小程序页面路径" rules={[{ max: 256 }]} tooltip="用户点击'去用券'时跳转的页面路径">
            <Input placeholder="pages/coupon/use" />
          </Form.Item>
        </Form>
      </Modal>

      <CredentialsModal
        open={!!credentials}
        credentials={credentials}
        onClose={() => setCredentials(null)}
      />
    </>
  );
}
