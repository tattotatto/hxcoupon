import { useState, useEffect, useCallback } from 'react';
import { Table, Card, Tag, Space, Button, Modal, Form, Input, InputNumber, Select, Switch, DatePicker, message, Popconfirm, Row, Col } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { templateApi } from '../../api/template';
import { storeApi } from '../../api/store';

const typeMap: Record<string, string> = { full_reduction: '满减', discount: '折扣', fixed_amount: '固定金额' };
const statusMap: Record<number, { color: string; text: string }> = {
  0: { color: 'default', text: '草稿' },
  1: { color: 'success', text: '已发布' },
  2: { color: 'error', text: '已停用' },
};

export default function TemplateList() {
  const [data, setData] = useState<{ items: any[]; total: number }>({ items: [], total: 0 });
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState<Record<string, any>>({ page: 1, page_size: 20 });
  const [modal, setModal] = useState<{ open: boolean; editing?: any }>({ open: false });
  const [saveLoading, setSaveLoading] = useState(false);
  const [stores, setStores] = useState<{ id: number; name: string; code: string }[]>([]);
  const [form] = Form.useForm();
  const validityType = Form.useWatch('validity_type', form);
  const applicableScope = Form.useWatch('applicable_scope', form);

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await templateApi.list(filters);
      setData(res.data.data);
    } catch { /* handled */ }
    finally { setLoading(false); }
  }, [filters]);

  useEffect(() => { fetchData(); }, [fetchData]);

  useEffect(() => {
    storeApi.options().then(res => setStores(res.data.data || [])).catch(() => {});
  }, []);

  const openCreate = () => {
    form.resetFields();
    form.setFieldsValue({ stackable: false, max_stack_count: 1, threshold_amount: 0, validity_type: 'fixed_date', applicable_scope: 'all' });
    setModal({ open: true });
  };

  const openEdit = (record: any) => {
    const vals = { ...record };
    if (vals.valid_start) vals.valid_start = dayjs(vals.valid_start);
    if (vals.valid_end) vals.valid_end = dayjs(vals.valid_end);
    form.setFieldsValue(vals);
    setModal({ open: true, editing: record });
  };

  const handleSave = async () => {
    const values = await form.validateFields();
    const data: Record<string, any> = { ...values };
    if (data.valid_start) data.valid_start = data.valid_start.format('YYYY-MM-DD');
    if (data.valid_end) data.valid_end = data.valid_end.format('YYYY-MM-DD');
    setSaveLoading(true);
    try {
      if (modal.editing) {
        await templateApi.update(modal.editing.id, data);
        message.success('更新成功');
      } else {
        await templateApi.create(data);
        message.success('创建成功');
      }
      setModal({ open: false });
      fetchData();
    } catch { /* validation or API error */ }
    finally { setSaveLoading(false); }
  };

  const handleStatus = async (id: number, status: number) => {
    try {
      await templateApi.updateStatus(id, status);
      message.success('状态更新成功');
      fetchData();
    } catch { /* handled */ }
  };

  const handleDelete = async (id: number) => {
    try {
      await templateApi.delete(id);
      message.success('删除成功');
      fetchData();
    } catch { /* handled */ }
  };

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '名称', dataIndex: 'name', width: 150 },
    {
      title: '类型', dataIndex: 'type', width: 80,
      render: (t: string) => <Tag>{typeMap[t] || t}</Tag>,
    },
    {
      title: '优惠值', dataIndex: 'discount_value', width: 80,
      render: (v: number, r: any) => r.type === 'discount' ? `${(v / 10).toFixed(1)}折` : `¥${v}`,
    },
    { title: '门槛', dataIndex: 'threshold_amount', width: 80, render: (v: number) => `¥${v}` },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (s: number) => {
        const item = statusMap[s];
        return item ? <Tag color={item.color}>{item.text}</Tag> : '-';
      },
    },
    {
      title: '库存', dataIndex: 'total_quantity', width: 80,
      render: (v: number, r: any) => `${r.issued_count || 0} / ${v}`,
    },
    {
      title: '操作', key: 'actions', fixed: 'right' as const, width: 240,
      render: (_: any, record: any) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => openEdit(record)}>编辑</Button>
          {record.status === 0 && (
            <Popconfirm title="确认发布？" onConfirm={() => handleStatus(record.id, 1)}>
              <Button type="link" size="small" style={{ color: '#52c41a' }}>发布</Button>
            </Popconfirm>
          )}
          {record.status === 1 && (
            <Popconfirm title="确认停用？" onConfirm={() => handleStatus(record.id, 2)}>
              <Button type="link" size="small" style={{ color: '#faad14' }}>停用</Button>
            </Popconfirm>
          )}
          {record.status === 2 && (
            <Popconfirm title="确认发布？" onConfirm={() => handleStatus(record.id, 1)}>
              <Button type="link" size="small" style={{ color: '#52c41a' }}>重新发布</Button>
            </Popconfirm>
          )}
          <Popconfirm title="确定删除？" onConfirm={() => handleDelete(record.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Card style={{ borderRadius: 12 }}>
        <Space style={{ marginBottom: 16 }}>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>新增模板</Button>
          <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
        </Space>

        <Table
          rowKey="id"
          columns={columns}
          dataSource={data.items}
          loading={loading}
          scroll={{ x: 1000 }}
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
        title={modal.editing ? '编辑模板' : '新增模板'}
        onOk={handleSave}
        onCancel={() => setModal({ open: false })}
        confirmLoading={saveLoading}
        width={640}
        okText="保存"
        cancelText="取消"
      >
        <Form form={form} layout="vertical" style={{ marginTop: 16 }}>
          <Form.Item name="name" label="模板名称" rules={[{ required: true }]}>
            <Input />
          </Form.Item>
          <Form.Item name="applicable_scope" label="适用范围" rules={[{ required: true }]}>
            <Select
              disabled={!!modal.editing}
              options={[
                { value: 'all', label: '全部门店' },
                { value: 'specific', label: '指定门店' },
              ]}
            />
          </Form.Item>
          {applicableScope === 'specific' && (
            <Form.Item name="store_ids" label="选择门店" rules={[{ required: true, type: 'array', min: 1, message: '请至少选择一个门店' }]}>
              <Select
                mode="multiple"
                placeholder="选择适用门店"
                showSearch
                optionFilterProp="label"
                disabled={!!modal.editing}
                options={stores.map(s => ({ value: s.id, label: `${s.name} (${s.code})` }))}
              />
            </Form.Item>
          )}
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="type" label="优惠类型" rules={[{ required: true }]}>
                <Select options={[
                  { value: 'full_reduction', label: '满减' },
                  { value: 'discount', label: '折扣' },
                  { value: 'fixed_amount', label: '固定金额' },
                ]} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="discount_value" label="优惠值" rules={[{ required: true }]} tooltip="满减/固定金额填元，折扣填分（如85折填850）">
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="threshold_amount" label="使用门槛（元）">
            <InputNumber min={0} precision={2} style={{ width: '100%' }} />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="validity_type" label="有效期类型" rules={[{ required: true }]}>
                <Select options={[
                  { value: 'fixed_date', label: '固定日期' },
                  { value: 'days_after_receive', label: '领取后N天' },
                ]} />
              </Form.Item>
            </Col>
            <Col span={12}>
              {validityType === 'days_after_receive' ? (
                <Form.Item name="validity_days" label="有效天数" rules={[{ required: true }]}>
                  <InputNumber min={1} style={{ width: '100%' }} />
                </Form.Item>
              ) : (
                <Row gutter={8}>
                  <Col span={12}>
                    <Form.Item name="valid_start" label="开始日期" rules={[{ required: true }]}>
                      <DatePicker style={{ width: '100%' }} />
                    </Form.Item>
                  </Col>
                  <Col span={12}>
                    <Form.Item name="valid_end" label="结束日期" rules={[{ required: true }]}>
                      <DatePicker style={{ width: '100%' }} />
                    </Form.Item>
                  </Col>
                </Row>
              )}
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="total_quantity" label="发行总量" rules={[{ required: true }]}>
                <InputNumber min={0} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="per_user_limit" label="每人限领" rules={[{ required: true }]}>
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="stackable" label="可叠加" valuePropName="checked">
                <Switch />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="max_stack_count" label="最大叠加数">
                <InputNumber min={1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Modal>
    </>
  );
}
