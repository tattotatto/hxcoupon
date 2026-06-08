import { useState } from 'react';
import { Card, Form, Input, Button, Select, message, Result, Descriptions, Tag } from 'antd';
import { couponApi } from '../../api/coupon';

export default function IssueForm() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);

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
            <Button key="again" type="primary" onClick={() => { setResult(null); form.resetFields(); }}>继续发券</Button>,
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
        <Form.Item name="store_id" label="门店ID" rules={[{ required: true, message: '请选择发券门店' }]}>
          <Input placeholder="输入门店ID" />
        </Form.Item>
        <Form.Item name="template_id" label="模板ID" rules={[{ required: true, message: '请选择模板' }]}>
          <Input placeholder="输入模板ID" />
        </Form.Item>
        <Form.Item name="user_phone" label="用户手机号" rules={[{ required: true, message: '请输入用户手机号' }, { max: 20 }]}>
          <Input placeholder="用户手机号" />
        </Form.Item>
        <Form.Item name="idempotency_key" label="幂等键" rules={[{ required: true, message: '请输入幂等键' }, { max: 128 }]} tooltip="用于防止重复发券，每次请求需唯一">
          <Input placeholder="唯一的幂等键，如 UUID" />
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
