import { useState, useEffect, useCallback } from 'react';
import { Row, Col, Card, Tag, Typography, Spin, Empty, Pagination, Space } from 'antd';
import { templateApi } from '../../api/template';
import dayjs from 'dayjs';

const { Title, Text } = Typography;

const typeMap: Record<string, { label: string; color: string }> = {
  full_reduction: { label: '满减', color: 'blue' },
  discount: { label: '折扣', color: 'purple' },
  fixed_amount: { label: '固定金额', color: 'green' },
};

export default function TemplateBrowse() {
  const [items, setItems] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [loading, setLoading] = useState(false);
  const [filters, setFilters] = useState({ page: 1, page_size: 12 });

  const fetchData = useCallback(async () => {
    setLoading(true);
    try {
      const res = await templateApi.browse(filters);
      setItems(res.data.data.items || []);
      setTotal(res.data.data.total || 0);
    } catch { /* handled */ }
    finally { setLoading(false); }
  }, [filters]);

  useEffect(() => { fetchData(); }, [fetchData]);

  return (
    <Spin spinning={loading}>
      <div>
        <Title level={4} style={{ marginBottom: 24 }}>模板浏览</Title>
        {items.length === 0 && !loading ? (
          <Empty description="暂无可用模板" />
        ) : (
          <>
            <Row gutter={[16, 16]}>
              {items.map((item) => {
                const typeInfo = typeMap[item.type] || { label: item.type, color: 'default' };
                return (
                  <Col xs={24} sm={12} lg={8} key={item.id}>
                    <Card
                      hoverable
                      style={{ borderRadius: 12 }}
                      title={
                        <Space>
                          <Tag color={typeInfo.color}>{typeInfo.label}</Tag>
                          <Text strong>{item.name}</Text>
                        </Space>
                      }
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 12 }}>
                        <div>
                          <Text type="secondary" style={{ fontSize: 12 }}>优惠</Text>
                          <div style={{ fontSize: 24, fontWeight: 600, color: '#f5222d' }}>
                            {item.type === 'discount' ? `${(item.discount_value / 10).toFixed(1)}折` : `¥${item.discount_value}`}
                          </div>
                        </div>
                        <div style={{ textAlign: 'right' }}>
                          <Text type="secondary" style={{ fontSize: 12 }}>门槛</Text>
                          <div style={{ fontSize: 24, fontWeight: 600 }}>¥{item.threshold_amount}</div>
                        </div>
                      </div>
                      <div style={{ color: '#8c8c8c', fontSize: 13 }}>
                        <div>库存: {item.total_quantity}</div>
                        <div>有效期: {item.validity_type === 'days_after_receive'
                          ? `领取后${item.validity_days}天`
                          : `${item.valid_start ? dayjs(item.valid_start).format('MM/DD') : ''} - ${item.valid_end ? dayjs(item.valid_end).format('MM/DD') : ''}`}</div>
                      </div>
                    </Card>
                  </Col>
                );
              })}
            </Row>
            {total > filters.page_size && (
              <div style={{ textAlign: 'center', marginTop: 24 }}>
                <Pagination
                  current={filters.page}
                  pageSize={filters.page_size}
                  total={total}
                  onChange={(p) => setFilters((f) => ({ ...f, page: p }))}
                />
              </div>
            )}
          </>
        )}
      </div>
    </Spin>
  );
}
