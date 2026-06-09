import { useState, useEffect } from 'react';
import { Row, Col, Card, Statistic, DatePicker, Spin, Empty } from 'antd';
import {
  ShopOutlined,
  ProjectOutlined,
  SendOutlined,
  CheckCircleOutlined,
  RiseOutlined,
} from '@ant-design/icons';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import dayjs, { Dayjs } from 'dayjs';
import { DndContext, closestCenter, PointerSensor, useSensor, useSensors } from '@dnd-kit/core';
import { SortableContext, rectSortingStrategy, useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { reportApi } from '../api/report';
import { useAuthStore } from '../stores/authStore';

const { RangePicker } = DatePicker;

interface StatData {
  total_stores: number;
  total_templates: number;
  total_issued: number;
  total_used: number;
  usage_rate: number;
  today_issued: number;
  today_used: number;
}

interface TrendItem {
  date: string;
  issued: number;
  used: number;
}

interface CardDef {
  key: string;
  title: string;
  icon: React.ReactNode;
  color: string;
  bgColor: string;
  getValue: (data: StatData | null) => number | string;
  suffix?: string;
}

function DraggableCard({ card, data }: { card: CardDef; data: StatData | null }) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({ id: card.key });
  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

  return (
    <div ref={setNodeRef} style={style} {...attributes} {...listeners}>
      <Card
        hoverable
        style={{
          borderRadius: 12,
          borderTop: `4px solid ${card.color}`,
          cursor: 'grab',
        }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <div style={{ color: '#8c8c8c', fontSize: 14, marginBottom: 8 }}>{card.title}</div>
            <Statistic
              value={data ? card.getValue(data) : '-'}
              suffix={card.suffix}
              valueStyle={{ fontSize: 28, fontWeight: 600 }}
            />
          </div>
          <div
            style={{
              width: 48,
              height: 48,
              borderRadius: 12,
              background: card.bgColor,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontSize: 24,
              color: card.color,
            }}
          >
            {card.icon}
          </div>
        </div>
      </Card>
    </div>
  );
}

export default function Dashboard() {
  const user = useAuthStore((s) => s.user);
  const [stats, setStats] = useState<StatData | null>(null);
  const [trend, setTrend] = useState<TrendItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [dates, setDates] = useState<[Dayjs, Dayjs]>([dayjs().subtract(7, 'day'), dayjs()]);

  const cards: CardDef[] = [
    { key: 'stores', title: '门店总数', icon: <ShopOutlined />, color: '#1677ff', bgColor: '#e6f4ff', getValue: (d) => d?.total_stores ?? '-' },
    { key: 'templates', title: '模板总数', icon: <ProjectOutlined />, color: '#722ed1', bgColor: '#f9f0ff', getValue: (d) => d?.total_templates ?? '-' },
    { key: 'issued', title: '累计发券', icon: <SendOutlined />, color: '#52c41a', bgColor: '#f6ffed', getValue: (d) => d?.total_issued ?? '-' },
    { key: 'used', title: '累计核销', icon: <CheckCircleOutlined />, color: '#fa8c16', bgColor: '#fff7e6', getValue: (d) => d?.total_used ?? '-' },
    { key: 'rate', title: '核销率', icon: <RiseOutlined />, color: '#eb2f96', bgColor: '#fff0f6', getValue: (d) => d?.usage_rate ? `${(d.usage_rate * 100).toFixed(1)}` : '-', suffix: '%' },
    { key: 'today_issued', title: '今日发券', icon: <SendOutlined />, color: '#13c2c2', bgColor: '#e6fffb', getValue: (d) => d?.today_issued ?? '-' },
    { key: 'today_used', title: '今日核销', icon: <CheckCircleOutlined />, color: '#f5222d', bgColor: '#fff1f0', getValue: (d) => d?.today_used ?? '-' },
  ];

  const [cardOrder, setCardOrder] = useState<string[]>(cards.map((c) => c.key));

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 5 } }),
  );

  const fetchData = async () => {
    setLoading(true);
    try {
      const [overviewRes, trendRes] = await Promise.all([
        reportApi.overview(),
        reportApi.trend(dates[0].format('YYYY-MM-DD'), dates[1].format('YYYY-MM-DD')),
      ]);
      setStats(overviewRes.data.data);
      setTrend(Array.isArray(trendRes.data.data) ? trendRes.data.data : []);
    } catch {
      // handled by interceptor
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [dates]);

  const transformedTrend = (Array.isArray(trend) ? trend : []).flatMap((item) => [
    { date: item.date, value: item.issued, type: '发券' },
    { date: item.date, value: item.used, type: '核销' },
  ]);

  // Pivot: group by date for recharts
  const chartData = (Array.isArray(trend) ? trend : []).map((item) => ({
    date: item.date,
    发券: item.issued,
    核销: item.used,
  }));

  return (
    <Spin spinning={loading}>
      <div style={{ padding: '0 0 24px' }}>
        <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h2 style={{ margin: 0, fontSize: 20 }}>欢迎回来，{user?.username}</h2>
          <RangePicker
            value={dates}
            onChange={(vals) => vals && setDates([vals[0]!, vals[1]!])}
            allowClear={false}
          />
        </div>

        <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={({ active, over }) => {
          if (over && active.id !== over.id) {
            setCardOrder((prev) => {
              const oldIdx = prev.indexOf(active.id as string);
              const newIdx = prev.indexOf(over.id as string);
              const newOrder = [...prev];
              newOrder.splice(oldIdx, 1);
              newOrder.splice(newIdx, 0, active.id as string);
              return newOrder;
            });
          }
        }}>
          <SortableContext items={cardOrder} strategy={rectSortingStrategy}>
            <Row gutter={[16, 16]}>
              {cardOrder.map((key) => {
                const card = cards.find((c) => c.key === key)!;
                return (
                  <Col xs={24} sm={12} lg={6} key={key}>
                    <DraggableCard card={card} data={stats} />
                  </Col>
                );
              })}
            </Row>
          </SortableContext>
        </DndContext>

        <Card title="趋势统计" style={{ marginTop: 24, borderRadius: 12 }}>
          {chartData.length > 0 ? (
            <ResponsiveContainer width="100%" height={300}>
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
