import { useState, useEffect, useCallback } from 'react';
import { Table, Card, Tag, Space, Button, Input } from 'antd';
import { SearchOutlined, ReloadOutlined } from '@ant-design/icons';
import dayjs from 'dayjs';
import { couponApi } from '../../api/coupon';

const statusMap: Record<string, { color: string; text: string }> = {
  unused: { color: 'default', text: '未使用' },
  used: { color: 'success', text: '已使用' },
  expired: { color: 'error', text: '已过期' },
  revoke: { color: 'warning', text: '已作废' },
};

export default function CouponRecords() {
  const [data, setData] = useState<{ items: any[]; total: number }>({ items: [], total: 0 });
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState<Record<string, any>>({ page: 1, page_size: 20 });

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await couponApi.listRecords(filters);
      setData(res.data.data);
    } catch { /* handled */ }
    finally { setLoading(false); }
  }, [filters]);

  useEffect(() => { fetchData(); }, [fetchData]);

  const columns = [
    { title: 'ID', dataIndex: 'id', width: 60 },
    { title: '券码', dataIndex: 'coupon_code', width: 160 },
    { title: '模板名称', dataIndex: 'template_name', width: 130, ellipsis: true },
    {
      title: '类型', dataIndex: 'type', width: 70,
      render: (t: string) => {
        const m: Record<string, string> = { full_reduction: '满减', discount: '折扣', fixed_amount: '固定金额' };
        return <Tag>{m[t] || t}</Tag>;
      },
    },
    {
      title: '优惠值', dataIndex: 'discount_value', width: 70,
      render: (v: number, r: any) => r.type === 'discount' ? `${(v / 10).toFixed(1)}折` : `¥${v}`,
    },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (s: string) => {
        const item = statusMap[s];
        return item ? <Tag color={item.color}>{item.text}</Tag> : '-';
      },
    },
    { title: '用户手机', dataIndex: 'user_phone', width: 120 },
    { title: '发券门店', dataIndex: 'source_store_name', width: 120, ellipsis: true },
    {
      title: '领取时间', dataIndex: 'receive_time', width: 160,
      render: (t: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-',
    },
    {
      title: '使用时间', dataIndex: 'use_time', width: 160,
      render: (t: string) => t ? dayjs(t).format('YYYY-MM-DD HH:mm') : '-',
    },
  ];

  return (
    <Card style={{ borderRadius: 12 }}>
      <Space style={{ marginBottom: 16 }}>
        <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
      </Space>
      <Table
        rowKey="id"
        columns={columns}
        dataSource={data.items}
        loading={loading}
        scroll={{ x: 1100 }}
        pagination={{
          current: filters.page,
          pageSize: filters.page_size,
          total: data.total,
          showTotal: (t) => `共 ${t} 条`,
          onChange: (p, ps) => setFilters((f) => ({ ...f, page: p, page_size: ps })),
        }}
      />
    </Card>
  );
}
