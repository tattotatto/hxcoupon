import { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, DatePicker, Button, Space, Spin, Empty } from 'antd';
import { DownloadOutlined, ReloadOutlined, RiseOutlined, SendOutlined, CheckCircleOutlined } from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import dayjs, { Dayjs } from 'dayjs';
import { reportApi } from '../../api/report';

const { RangePicker } = DatePicker;

export default function ReportDashboard() {
  const [overview, setOverview] = useState<any>(null);
  const [trend, setTrend] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [dates, setDates] = useState<[Dayjs, Dayjs]>([dayjs().subtract(30, 'day'), dayjs()]);

  const fetchData = async () => {
    setLoading(true);
    try {
      const [overviewRes, trendRes] = await Promise.all([
        reportApi.overview(),
        reportApi.trend(dates[0].format('YYYY-MM-DD'), dates[1].format('YYYY-MM-DD')),
      ]);
      setOverview(overviewRes.data.data);
      setTrend(Array.isArray(trendRes.data.data) ? trendRes.data.data : []);
    } catch { /* handled */ }
    finally { setLoading(false); }
  };

  useEffect(() => { fetchData(); }, [dates]);

  const handleExport = async (type: 'coupons' | 'usage') => {
    try {
      const res = type === 'coupons' ? await reportApi.exportCoupons() : await reportApi.exportUsage();
      const blob = new Blob([res.data], { type: 'text/csv;charset=utf-8' });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${type}_${dayjs().format('YYYYMMDD')}.csv`;
      a.click();
      URL.revokeObjectURL(url);
    } catch { /* handled */ }
  };

  const chartData = (Array.isArray(trend) ? trend : []).map((item) => ({
    date: item.date,
    发券: item.issued,
    核销: item.used,
  }));

  return (
    <Spin spinning={loading}>
      <div>
        <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: 8 }}>
          <h2 style={{ margin: 0, fontSize: 20 }}>数据报表</h2>
          <Space wrap>
            <RangePicker value={dates} onChange={(vals) => vals && setDates([vals[0]!, vals[1]!])} allowClear={false} />
            <Button icon={<ReloadOutlined />} onClick={fetchData}>刷新</Button>
            <Button icon={<DownloadOutlined />} onClick={() => handleExport('coupons')}>导出券数据</Button>
            <Button icon={<DownloadOutlined />} onClick={() => handleExport('usage')}>导出核销数据</Button>
          </Space>
        </div>

        <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
          <Col xs={24} sm={12} md={6}>
            <Card hoverable style={{ borderRadius: 12, borderTop: '4px solid #1677ff' }}>
              <Statistic title="累计发券" value={overview?.total_issued ?? '-'} prefix={<SendOutlined />} />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card hoverable style={{ borderRadius: 12, borderTop: '4px solid #52c41a' }}>
              <Statistic title="累计核销" value={overview?.total_used ?? '-'} prefix={<CheckCircleOutlined />} />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card hoverable style={{ borderRadius: 12, borderTop: '4px solid #722ed1' }}>
              <Statistic
                title="核销率"
                value={overview?.usage_rate != null ? `${(overview.usage_rate * 100).toFixed(1)}` : '-'}
                suffix="%"
                prefix={<RiseOutlined />}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card hoverable style={{ borderRadius: 12, borderTop: '4px solid #fa8c16' }}>
              <Statistic title="今日核销" value={overview?.today_used ?? '-'} prefix={<CheckCircleOutlined />} />
            </Card>
          </Col>
        </Row>

        <Card title="趋势图" style={{ borderRadius: 12 }}>
          {chartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={350}>
              <LineChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="date" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Line type="monotone" dataKey="发券" stroke="#1677ff" strokeWidth={2} dot={{ r: 3 }} />
                <Line type="monotone" dataKey="核销" stroke="#52c41a" strokeWidth={2} dot={{ r: 3 }} />
              </LineChart>
            </ResponsiveContainer>
          ) : (
            <Empty description="暂无趋势数据" />
          )}
        </Card>
      </div>
    </Spin>
  );
}
