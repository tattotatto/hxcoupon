import { useState } from 'react';
import { Card, Form, Input, InputNumber, Button, message, Result, Descriptions, Tag } from 'antd';
import { couponApi } from '../../api/coupon';

export default function ConsumeForm() {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<any>(null);

  const onFinish = async (values: Record<string, any>) => {
    setLoading(true);
    try {
      const res = await couponApi.consume(values);
      setResult(res.data.data);
      message.success('核销成功');
    } catch { /* handled */ }
    finally { setLoading(false); }
  };

  if (result) {
    return (
      <Card style={{ borderRadius: 12, maxWidth: 600 }}>
        <Result
          status="success"
          title="核销成功"
          extra={[
            <Button key="again" type="primary" onClick={() => { setResult(null); form.resetFields(); }}>继续核销</Button>,
          ]}
        />
        <Descriptions bordered size="small" style={{ marginTop: 16 }}>
          <Descriptions.Item label="券码">{result.coupon_code}</Descriptions.Item>
          <Descriptions.Item label="优惠金额">¥{result.discount_value}</Descriptions.Item>
          <Descriptions.Item label="实付金额">¥{result.actual_amount}</Descriptions.Item>
          <Descriptions.Item label="核销时间">{result.used_at}</Descriptions.Item>
        </Descriptions>
      </Card>
    );
  }

  return (
    <Card title="管理端核销" style={{ borderRadius: 12, maxWidth: 600 }}>
      <Form form={form} layout="vertical" onFinish={onFinish}>
        <Form.Item name="coupon_code" label="券码" rules={[{ required: true, message: '请输入券码' }]}>
          <Input placeholder="输入优惠券码" />
        </Form.Item>
        <Form.Item name="user_phone" label="用户手机号" rules={[{ required: true, max: 20 }]}>
          <Input placeholder="用户手机号" />
        </Form.Item>
        <Form.Item name="store_id" label="门店ID" rules={[{ required: true }]}>
          <Input placeholder="核销门店ID" />
        </Form.Item>
        <Form.Item name="order_id" label="订单号" rules={[{ required: true, max: 64 }]}>
          <Input placeholder="关联的订单号" />
        </Form.Item>
        <Form.Item name="order_amount" label="订单金额" rules={[{ required: true }]}>
          <InputNumber min={0.01} precision={2} style={{ width: '100%' }} placeholder="订单金额" />
        </Form.Item>
        <Form.Item>
          <Button type="primary" htmlType="submit" loading={loading} size="large">
            确认核销
          </Button>
        </Form.Item>
      </Form>
    </Card>
  );
}
