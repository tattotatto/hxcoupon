import { useState, useEffect, useCallback } from 'react';
import { Card, Form, Input, Button, Select, message, Result, Descriptions, Tag } from 'antd';
import { couponApi } from '../../api/coupon';
import { storeApi } from '../../api/store';
import { templateApi } from '../../api/template';

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

export default function IssueForm() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);
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
      setResult(res.data.data);
      message.success('发券成功');
    } catch { /* handled */ }
    finally { setLoading(false); }
  };

  if (result) {
    return (
      <Card style={{ borderRadius: 12, maxWidth: 600 }}>
        <Result
          status="success"
          title="发券成功"
          subTitle={`优惠券已成功发放`}
          extra={[
            <Button key="again" type="primary" onClick={() => { setResult(null); setSelectedStoreId(undefined); setTemplates([]); form.resetFields(); }}>继续发券</Button>,
          ]}
        />
        <Descriptions bordered size="small" style={{ marginTop: 16 }}>
          <Descriptions.Item label="券码">{result.coupon_code}</Descriptions.Item>
          <Descriptions.Item label="模板">{result.template_name}</Descriptions.Item>
          <Descriptions.Item label="状态">
            <Tag color="success">已发放</Tag>
          </Descriptions.Item>
          <Descriptions.Item label="类型">{result.type}</Descriptions.Item>
          <Descriptions.Item label="优惠值">{result.type === 'discount' ? `${(result.discount_value / 10).toFixed(1)}折` : `¥${result.discount_value}`}</Descriptions.Item>
          <Descriptions.Item label="有效期">{result.valid_start} ~ {result.valid_end}</Descriptions.Item>
        </Descriptions>
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
        <Form.Item name="user_phone" label="用户手机号" rules={[{ required: true, message: '请输入用户手机号' }, { max: 20 }]}>
          <Input placeholder="用户手机号" />
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
