import { useState, useEffect, useCallback } from 'react';
import { Card, Form, Input, Button, Select, message, Result, Table, Tag, Typography, Space } from 'antd';
import { couponApi } from '../../api/coupon';
import { storeApi } from '../../api/store';
import { templateApi } from '../../api/template';

const { Text } = Typography;

interface StoreOption {
  id: number;
  name: string;
  code: string;
}

interface TemplateOption {
  id: number;
  name: string;
  type: string;
  applicable_scope: string;
}

interface BatchItem {
  user_phone: string;
  success: boolean;
  error_code?: number;
  error_message?: string;
  coupon?: {
    coupon_id: number;
    coupon_code: string;
    template_name: string;
    type: string;
    discount_value: number;
    threshold_amount: number;
    valid_start: string;
    valid_end: string;
    status: string;
    qr_code_url: string;
  };
}

interface BatchResult {
  total_count: number;
  success_count: number;
  failed_count: number;
  items: BatchItem[];
}

export default function IssueForm() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<BatchResult | null>(null);
  const [stores, setStores] = useState<StoreOption[]>([]);
  const [templates, setTemplates] = useState<TemplateOption[]>([]);
  const [templatesLoading, setTemplatesLoading] = useState(false);
  const [selectedStoreId, setSelectedStoreId] = useState<number | undefined>();

  // Load store options on mount
  useEffect(() => {
    storeApi.options().then(res => {
      setStores(res.data.data || []);
    }).catch(() => {});
  }, []);

  // Load templates when store changes
  const onStoreChange = useCallback(async (storeId: number) => {
    setSelectedStoreId(storeId);
    form.setFieldValue('template_id', undefined);
    setTemplates([]);
    if (!storeId) return;
    setTemplatesLoading(true);
    try {
      const res = await templateApi.list({ store_id: storeId });
      setTemplates((res.data.data?.items as TemplateOption[]) || []);
    } catch { /* handled */ }
    finally { setTemplatesLoading(false); }
  }, [form]);

  const onFinish = async (values: Record<string, string | number>) => {
    setLoading(true);
    try {
      const res = await couponApi.issue(values);
      const data: BatchResult = res.data.data;
      setResult(data);
      if (data.failed_count === 0) {
        message.success(`发券成功，共 ${data.success_count} 张`);
      } else if (data.success_count === 0) {
        message.warning(`发券全部失败，共 ${data.failed_count} 条`);
      } else {
        message.warning(`部分成功：成功 ${data.success_count}，失败 ${data.failed_count}`);
      }
    } catch { /* handled */ }
    finally { setLoading(false); }
  };

  if (result) {
    const allOk = result.failed_count === 0;
    const allFail = result.success_count === 0;
    const status: 'success' | 'warning' | 'error' = allOk ? 'success' : allFail ? 'error' : 'warning';
    const title = allOk
      ? `发券成功（${result.success_count} 张）`
      : allFail
      ? `发券全部失败（${result.failed_count} 条）`
      : `部分成功（成功 ${result.success_count} / 失败 ${result.failed_count}）`;

    return (
      <Card style={{ borderRadius: 12, maxWidth: 720 }}>
        <Result
          status={status}
          title={title}
          extra={[
            <Button key="again" type="primary" onClick={() => { setResult(null); setSelectedStoreId(undefined); setTemplates([]); form.resetFields(); }}>继续发券</Button>,
          ]}
        />
        <Table<BatchItem>
          size="small"
          style={{ marginTop: 16 }}
          rowKey={(r) => `${r.user_phone}-${r.coupon?.coupon_code ?? 'fail'}`}
          dataSource={result.items}
          pagination={false}
          columns={[
            { title: '手机号', dataIndex: 'user_phone', width: 140 },
            {
              title: '结果',
              dataIndex: 'success',
              width: 100,
              render: (ok: boolean, row) =>
                ok ? <Tag color="success">成功</Tag> : <Tag color="error">失败</Tag>,
            },
            {
              title: '详情',
              render: (_: unknown, row) =>
                row.success && row.coupon ? (
                  <Space direction="vertical" size={0}>
                    <Text>券码：{row.coupon.coupon_code}</Text>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {row.coupon.template_name} · {row.coupon.valid_start} ~ {row.coupon.valid_end}
                    </Text>
                  </Space>
                ) : (
                  <Text type="danger">{row.error_message || '未知错误'}</Text>
                ),
            },
          ]}
        />
      </Card>
    );
  }

  return (
    <Card title="管理端发券" style={{ borderRadius: 12, maxWidth: 600 }}>
      <Form form={form} layout="vertical" onFinish={onFinish}>
        <Form.Item name="store_id" label="发券门店" rules={[{ required: true, message: '请选择发券门店' }]}>
          <Select
            placeholder="选择门店"
            showSearch
            optionFilterProp="label"
            onChange={onStoreChange}
            options={stores.map(s => ({
              value: s.id,
              label: `${s.name} (${s.code})`,
            }))}
          />
        </Form.Item>
        <Form.Item name="template_id" label="优惠券模板" rules={[{ required: true, message: '请选择模板' }]}>
          <Select
            placeholder={selectedStoreId ? '选择模板' : '请先选择门店'}
            loading={templatesLoading}
            showSearch
            optionFilterProp="label"
            disabled={!selectedStoreId}
            options={templates.map(t => ({
              value: t.id,
              label: `[${t.id}] ${t.name}`,
            }))}
          />
        </Form.Item>
        <Form.Item
          name="user_phone"
          label="用户手机号"
          extra="多个手机号请用英文逗号 , 分隔，单个失败不影响其他"
          rules={[{ required: true, message: '请输入用户手机号' }]}
        >
          <Input placeholder="例如：13800001111,13800002222" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} size="large">
            确认发券
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
}
