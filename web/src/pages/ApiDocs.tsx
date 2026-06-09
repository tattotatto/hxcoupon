import { useState, useMemo } from 'react';
import {
  Card,
  Collapse,
  Tag,
  Button,
  Input,
  Select,
  Space,
  Table,
  Typography,
  message,
  Spin,
  Row,
  Col,
  Empty,
  Tabs,
} from 'antd';
import {
  SendOutlined,
  ClearOutlined,
  CopyOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
} from '@ant-design/icons';
import client from '../api/client';
import axios from 'axios';

const { Text } = Typography;
const { TextArea } = Input;

// ─── Endpoint definitions ───────────────────────────────────────────────

interface ParamDef {
  name: string;
  type: string;
  required?: boolean;
  desc: string;
}

interface EndpointDef {
  method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  path: string;
  desc: string;
  auth: string;
  body?: Record<string, unknown>;
  queryParams?: ParamDef[];
  pathParams?: ParamDef[];
}

interface EndpointGroup {
  name: string;
  icon?: string;
  endpoints: EndpointDef[];
}

const METHOD_COLORS: Record<string, string> = {
  GET: '#52c41a',
  POST: '#1677ff',
  PUT: '#fa8c16',
  PATCH: '#722ed1',
  DELETE: '#ff4d4f',
};

const endpointGroups: EndpointGroup[] = [
  {
    name: '认证 Auth',
    icon: '🔐',
    endpoints: [
      { method: 'POST', path: '/auth/register', desc: '用户注册', auth: '无', body: { username: '', password: '', member_type: 'issuer', company_name: '', contact_name: '', contact_phone: '', email: '' } },
      { method: 'POST', path: '/admin/login', desc: '用户登录', auth: '无', body: { username: '', password: '' } },
      { method: 'POST', path: '/admin/refresh', desc: '刷新令牌', auth: '无', body: { refresh_token: '' } },
    ],
  },
  {
    name: '个人 Profile',
    icon: '👤',
    endpoints: [
      { method: 'GET', path: '/admin/profile', desc: '获取当前用户信息', auth: 'JWT' },
      { method: 'PUT', path: '/admin/profile', desc: '更新当前用户信息', auth: 'JWT', body: { company_name: '', contact_name: '', contact_phone: '', email: '' } },
    ],
  },
  {
    name: '用户管理 Users',
    icon: '👥',
    endpoints: [
      { method: 'GET', path: '/admin/users', desc: '用户列表', auth: 'JWT + manage_users', queryParams: [{ name: 'page', type: 'number', desc: '页码' }, { name: 'page_size', type: 'number', desc: '每页条数' }] },
      { method: 'GET', path: '/admin/users/:id', desc: '用户详情', auth: 'JWT + manage_users', pathParams: [{ name: 'id', type: 'number', required: true, desc: '用户ID' }] },
      { method: 'POST', path: '/admin/users/:id/approve', desc: '审批通过用户', auth: 'JWT + manage_users', pathParams: [{ name: 'id', type: 'number', required: true, desc: '用户ID' }], body: { reason: '' } },
      { method: 'POST', path: '/admin/users/:id/reject', desc: '驳回用户', auth: 'JWT + manage_users', pathParams: [{ name: 'id', type: 'number', required: true, desc: '用户ID' }], body: { reason: '' } },
      { method: 'POST', path: '/admin/users/:id/suspend', desc: '停用用户', auth: 'JWT + manage_users', pathParams: [{ name: 'id', type: 'number', required: true, desc: '用户ID' }], body: { reason: '' } },
      { method: 'POST', path: '/admin/users/:id/unsuspend', desc: '启用用户', auth: 'JWT + manage_users', pathParams: [{ name: 'id', type: 'number', required: true, desc: '用户ID' }] },
    ],
  },
  {
    name: '门店管理 Stores',
    icon: '🏪',
    endpoints: [
      { method: 'GET', path: '/admin/stores', desc: '门店列表', auth: 'JWT', queryParams: [{ name: 'page', type: 'number', desc: '页码' }, { name: 'page_size', type: 'number', desc: '每页条数' }] },
      { method: 'GET', path: '/admin/stores/:id', desc: '门店详情', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }] },
      { method: 'POST', path: '/admin/stores', desc: '创建门店', auth: 'JWT', body: { name: '', description: '' } },
      { method: 'PUT', path: '/admin/stores/:id', desc: '更新门店', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }], body: { name: '', description: '' } },
      { method: 'PATCH', path: '/admin/stores/:id/status', desc: '更新门店状态', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }], body: { status: 1 } },
      { method: 'DELETE', path: '/admin/stores/:id', desc: '删除门店', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }] },
      { method: 'POST', path: '/admin/stores/:id/credentials', desc: '生成门店凭证', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }] },
    ],
  },
  {
    name: '模板管理 Templates',
    icon: '📋',
    endpoints: [
      { method: 'GET', path: '/admin/templates', desc: '模板列表', auth: 'JWT', queryParams: [{ name: 'page', type: 'number', desc: '页码' }, { name: 'page_size', type: 'number', desc: '每页条数' }] },
      { method: 'GET', path: '/admin/templates/:id', desc: '模板详情', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }] },
      { method: 'POST', path: '/admin/templates', desc: '创建模板', auth: 'JWT', body: { name: '', description: '', coupon_type: '', amount: 0 } },
      { method: 'PUT', path: '/admin/templates/:id', desc: '更新模板', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }], body: { name: '', description: '' } },
      { method: 'PATCH', path: '/admin/templates/:id/status', desc: '更新模板状态', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }], body: { status: 1 } },
      { method: 'DELETE', path: '/admin/templates/:id', desc: '删除模板', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }] },
      { method: 'GET', path: '/admin/browse/templates', desc: '浏览模板列表', auth: 'JWT', queryParams: [{ name: 'page', type: 'number', desc: '页码' }, { name: 'page_size', type: 'number', desc: '每页条数' }] },
      { method: 'GET', path: '/admin/browse/templates/:id', desc: '浏览模板详情', auth: 'JWT', pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }] },
    ],
  },
  {
    name: '优惠券 Coupons',
    icon: '🎫',
    endpoints: [
      { method: 'POST', path: '/admin/coupons/issue', desc: '发放优惠券', auth: 'JWT + 已审批', body: { template_id: 0, store_id: 0, user_id: '', quantity: 1 } },
      { method: 'GET', path: '/admin/coupons/records', desc: '发券记录列表', auth: 'JWT + 已审批', queryParams: [{ name: 'page', type: 'number', desc: '页码' }, { name: 'page_size', type: 'number', desc: '每页条数' }] },
      { method: 'GET', path: '/admin/coupons/records/:id', desc: '发券记录详情', auth: 'JWT + 已审批', pathParams: [{ name: 'id', type: 'number', required: true, desc: '记录ID' }] },
      { method: 'POST', path: '/admin/coupons/consume', desc: '核销优惠券', auth: 'JWT + 已审批', body: { coupon_code: '', store_id: 0 } },
      { method: 'GET', path: '/admin/coupons/consume-records', desc: '核销记录列表', auth: 'JWT + 已审批', queryParams: [{ name: 'page', type: 'number', desc: '页码' }, { name: 'page_size', type: 'number', desc: '每页条数' }] },
    ],
  },
  {
    name: '报表 Reports',
    icon: '📊',
    endpoints: [
      { method: 'GET', path: '/admin/reports/overview', desc: '报表概览', auth: 'JWT + 已审批' },
      { method: 'GET', path: '/admin/reports/trend', desc: '趋势数据', auth: 'JWT + 已审批', queryParams: [{ name: 'start_date', type: 'string', desc: '开始日期' }, { name: 'end_date', type: 'string', desc: '结束日期' }] },
      { method: 'GET', path: '/admin/reports/export/coupons', desc: '导出券数据CSV', auth: 'JWT + 已审批' },
      { method: 'GET', path: '/admin/reports/export/usage', desc: '导出核销数据CSV', auth: 'JWT + 已审批' },
    ],
  },
  {
    name: '统计 Statistics',
    icon: '📈',
    endpoints: [
      { method: 'GET', path: '/admin/statistics/overview', desc: '统计概览', auth: 'JWT' },
      { method: 'GET', path: '/admin/statistics/trend', desc: '统计趋势', auth: 'JWT', queryParams: [{ name: 'start_date', type: 'string', desc: '开始日期' }, { name: 'end_date', type: 'string', desc: '结束日期' }] },
    ],
  },
  {
    name: '应用管理 Apps',
    icon: '📱',
    endpoints: [
      { method: 'GET', path: '/admin/apps', desc: '应用列表', auth: 'JWT + 已审批', queryParams: [{ name: 'page', type: 'number', desc: '页码' }, { name: 'page_size', type: 'number', desc: '每页条数' }] },
      { method: 'GET', path: '/admin/apps/:id', desc: '应用详情', auth: 'JWT + 已审批', pathParams: [{ name: 'id', type: 'number', required: true, desc: '应用ID' }] },
      { method: 'POST', path: '/admin/apps', desc: '创建应用', auth: 'JWT + 已审批', body: { name: '', description: '' } },
      { method: 'PUT', path: '/admin/apps/:id', desc: '更新应用', auth: 'JWT + 已审批', pathParams: [{ name: 'id', type: 'number', required: true, desc: '应用ID' }], body: { name: '', description: '' } },
      { method: 'DELETE', path: '/admin/apps/:id', desc: '删除应用', auth: 'JWT + 已审批', pathParams: [{ name: 'id', type: 'number', required: true, desc: '应用ID' }] },
      { method: 'POST', path: '/admin/apps/:id/credentials', desc: '生成应用凭证', auth: 'JWT + 已审批', pathParams: [{ name: 'id', type: 'number', required: true, desc: '应用ID' }] },
    ],
  },
  {
    name: 'Open API (HMAC)',
    icon: '🔗',
    endpoints: [
      { method: 'POST', path: '/coupons/issue', desc: '发放优惠券 (HMAC签名)', auth: 'HMAC + 频率限制', body: { template_id: 0, user_id: '', quantity: 1 } },
      { method: 'GET', path: '/coupons/available', desc: '查询可用优惠券 (HMAC签名)', auth: 'HMAC + 频率限制', queryParams: [{ name: 'user_id', type: 'string', desc: '用户ID' }] },
      { method: 'GET', path: '/coupons/user', desc: '用户优惠券列表 (HMAC签名)', auth: 'HMAC + 频率限制', queryParams: [{ name: 'user_id', type: 'string', desc: '用户ID' }, { name: 'status', type: 'string', desc: '状态' }] },
      { method: 'GET', path: '/coupons/:coupon_code', desc: '优惠券详情 (HMAC签名)', auth: 'HMAC + 频率限制', pathParams: [{ name: 'coupon_code', type: 'string', required: true, desc: '券码' }] },
      { method: 'POST', path: '/coupons/consume', desc: '核销优惠券 (HMAC签名)', auth: 'HMAC + 频率限制', body: { coupon_code: '', store_id: 0 } },
      { method: 'POST', path: '/coupons/refund', desc: '退券 (HMAC签名)', auth: 'HMAC + 频率限制', body: { coupon_code: '' } },
    ],
  },
];

// ─── Helper ──────────────────────────────────────────────────────────────

function extractPathParams(path: string): string[] {
  const re = /:(\w+)/g;
  const params: string[] = [];
  let m: RegExpExecArray | null;
  while ((m = re.exec(path)) !== null) {
    params.push(m[1]);
  }
  return params;
}

// ─── Component ───────────────────────────────────────────────────────────

export default function ApiDocs() {
  const [selectedEndpoint, setSelectedEndpoint] = useState<EndpointDef | null>(null);
  const [method, setMethod] = useState<string>('GET');
  const [urlPath, setUrlPath] = useState('');
  const [pathParamValues, setPathParamValues] = useState<Record<string, string>>({});
  const [queryParams, setQueryParams] = useState<{ key: string; value: string }[]>([{ key: '', value: '' }]);
  const [bodyText, setBodyText] = useState('');
  const [useRawClient, setUseRawClient] = useState(false);
  const [loading, setLoading] = useState(false);
  const [response, setResponse] = useState<{
    status: number;
    statusText: string;
    headers: Record<string, string>;
    body: unknown;
    time: number;
  } | null>(null);

  const activePathParams = useMemo(() => extractPathParams(urlPath), [urlPath]);

  const handleSelectEndpoint = (ep: EndpointDef) => {
    setSelectedEndpoint(ep);
    setMethod(ep.method);
    setUrlPath(ep.path);
    setResponse(null);

    // Reset path param values
    const pp: Record<string, string> = {};
    (ep.pathParams || []).forEach((p) => { pp[p.name] = ''; });
    setPathParamValues(pp);

    // Reset query params
    const qps = (ep.queryParams || []).map((p) => ({ key: p.name, value: '' }));
    setQueryParams(qps.length > 0 ? qps : [{ key: '', value: '' }]);

    // Reset body
    setBodyText(ep.body ? JSON.stringify(ep.body, null, 2) : '');
    setUseRawClient(ep.auth.includes('HMAC'));
  };

  const handleSend = async () => {
    if (!urlPath) return;

    setLoading(true);
    const startTime = performance.now();

    // Build final URL with path params replaced
    let finalPath = urlPath;
    Object.entries(pathParamValues).forEach(([k, v]) => {
      if (v) finalPath = finalPath.replace(`:${k}`, v);
    });

    // Build query string
    const qps = queryParams.filter((q) => q.key.trim());
    const qs = qps.length > 0
      ? '?' + new URLSearchParams(qps.map((q) => [q.key, q.value])).toString()
      : '';

    const fullUrl = finalPath + qs;

    try {
      let res;
      const axiosInstance = useRawClient ? axios.create({ baseURL: '/api/v1' }) : client;

      if (['POST', 'PUT', 'PATCH'].includes(method)) {
        let data: unknown;
        try {
          data = bodyText ? JSON.parse(bodyText) : {};
        } catch {
          message.error('请求体 JSON 格式错误');
          setLoading(false);
          return;
        }
        res = await (axiosInstance as ReturnType<typeof axios.create>).request({
          method: method.toLowerCase() as 'post' | 'put' | 'patch',
          url: fullUrl,
          data,
          params: undefined,
        });
      } else {
        res = await (axiosInstance as ReturnType<typeof axios.create>).request({
          method: method.toLowerCase() as 'get' | 'delete',
          url: fullUrl,
        });
      }

      const elapsed = Math.round(performance.now() - startTime);
      setResponse({
        status: res.status,
        statusText: res.statusText,
        headers: res.headers as Record<string, string>,
        body: res.data,
        time: elapsed,
      });
    } catch (err: unknown) {
      const elapsed = Math.round(performance.now() - startTime);
      const e = err as { response?: { status: number; statusText: string; headers: Record<string, string>; data: unknown }; message?: string };
      setResponse({
        status: e.response?.status || 0,
        statusText: e.response?.statusText || 'Network Error',
        headers: e.response?.headers || {},
        body: e.response?.data || { error: e.message || '请求失败' },
        time: elapsed,
      });
    } finally {
      setLoading(false);
    }
  };

  const handleCopyResponse = () => {
    if (response) {
      navigator.clipboard.writeText(JSON.stringify(response.body, null, 2));
      message.success('已复制到剪贴板');
    }
  };

  const columns = [
    {
      title: '',
      dataIndex: 'method',
      key: 'method',
      width: 70,
      render: (m: string) => (
        <Tag color={METHOD_COLORS[m] || '#999'} style={{ fontWeight: 700, minWidth: 56, textAlign: 'center' }}>
          {m}
        </Tag>
      ),
    },
    {
      title: '路径',
      dataIndex: 'path',
      key: 'path',
      render: (p: string) => <Text code style={{ fontSize: 13 }}>{p}</Text>,
    },
    {
      title: '',
      dataIndex: 'auth',
      key: 'auth',
      width: 170,
      render: (a: string) => <Text type="secondary" style={{ fontSize: 12 }}>{a}</Text>,
    },
  ];

  return (
    <div style={{ display: 'flex', gap: 16, minHeight: 'calc(100vh - 200px)' }}>
      {/* ── Left: API Catalog ── */}
      <Card
        title="API 接口列表"
        size="small"
        style={{ width: 500, flexShrink: 0, overflow: 'auto', maxHeight: 'calc(100vh - 160px)' }}
        bodyStyle={{ padding: 0 }}
      >
        <Collapse
          ghost
          defaultActiveKey={endpointGroups.map((_, i) => String(i))}
          expandIconPosition="end"
        >
          {endpointGroups.map((group, gi) => (
            <Collapse.Panel
              key={String(gi)}
              header={
                <Space>
                  <span>{group.icon}</span>
                  <Text strong>{group.name}</Text>
                  <Tag>{group.endpoints.length} 个接口</Tag>
                </Space>
              }
            >
              <Table
                dataSource={group.endpoints.map((ep, ei) => ({ ...ep, key: `${gi}-${ei}` }))}
                columns={columns}
                size="small"
                showHeader={false}
                pagination={false}
                onRow={(record) => ({
                  onClick: () => handleSelectEndpoint(record),
                  style: {
                    cursor: 'pointer',
                    background: selectedEndpoint === record ? '#e6f4ff' : undefined,
                  },
                })}
                locale={{ emptyText: <Empty description="暂无接口" image={Empty.PRESENTED_IMAGE_SIMPLE} /> }}
              />
            </Collapse.Panel>
          ))}
        </Collapse>
      </Card>

      {/* ── Right: API Tester ── */}
      <Card
        title={
          <Space>
            <span>API 测试工具</span>
            {selectedEndpoint && <Tag>{selectedEndpoint.desc}</Tag>}
          </Space>
        }
        size="small"
        style={{ flex: 1, overflow: 'auto', maxHeight: 'calc(100vh - 160px)' }}
        extra={
          selectedEndpoint && (
            <Space>
              {selectedEndpoint.auth.includes('HMAC') && (
                <Tag color="warning">HMAC 认证 — 使用原始 axios 请求</Tag>
              )}
              {selectedEndpoint.auth.includes('JWT') && (
                <Tag color="green">JWT 认证 — 自动携带 Token</Tag>
              )}
            </Space>
          )
        }
      >
        {selectedEndpoint ? (
          <div>
            {/* Request Builder */}
            <div style={{ marginBottom: 16 }}>
              <Space.Compact style={{ width: '100%', marginBottom: 12 }}>
                <Select
                  value={method}
                  onChange={(v) => setMethod(v)}
                  style={{ width: 100 }}
                  options={Object.keys(METHOD_COLORS).map((m) => ({ value: m, label: m }))}
                />
                <Input
                  value={urlPath}
                  onChange={(e) => setUrlPath(e.target.value)}
                  placeholder="/api/v1/admin/..."
                  addonBefore="/api/v1"
                  style={{ fontFamily: 'monospace' }}
                />
                <Button
                  type="primary"
                  icon={<SendOutlined />}
                  loading={loading}
                  onClick={handleSend}
                >
                  发送
                </Button>
                <Button icon={<ClearOutlined />} onClick={() => { setResponse(null); setPathParamValues({}); setQueryParams([{ key: '', value: '' }]); setBodyText(selectedEndpoint.body ? JSON.stringify(selectedEndpoint.body, null, 2) : ''); }}>
                  重置
                </Button>
              </Space.Compact>

              {/* Path Parameters */}
              {activePathParams.length > 0 && (
                <div style={{ marginBottom: 12 }}>
                  <Text type="secondary" style={{ fontSize: 12, marginBottom: 4, display: 'block' }}>路径参数:</Text>
                  <Row gutter={[8, 8]}>
                    {activePathParams.map((p) => (
                      <Col span={12} key={p}>
                        <Input
                          size="small"
                          placeholder={`:${p}`}
                          addonBefore={
                            <span style={{ color: '#ff4d4f', fontSize: 12 }}>*{p}</span>
                          }
                          value={pathParamValues[p] || ''}
                          onChange={(e) => setPathParamValues((prev) => ({ ...prev, [p]: e.target.value }))}
                        />
                      </Col>
                    ))}
                  </Row>
                </div>
              )}

              {/* Query Parameters */}
              <div style={{ marginBottom: 12 }}>
                <Space style={{ marginBottom: 4 }}>
                  <Text type="secondary" style={{ fontSize: 12 }}>Query 参数:</Text>
                  <Button
                    type="dashed"
                    size="small"
                    onClick={() => setQueryParams((prev) => [...prev, { key: '', value: '' }])}
                  >
                    + 添加
                  </Button>
                </Space>
                {queryParams.map((qp, i) => (
                  <Row gutter={8} key={i} style={{ marginBottom: 4 }}>
                    <Col span={8}>
                      <Input
                        size="small"
                        placeholder="参数名"
                        value={qp.key}
                        onChange={(e) => {
                          const next = [...queryParams];
                          next[i].key = e.target.value;
                          setQueryParams(next);
                        }}
                      />
                    </Col>
                    <Col span={14}>
                      <Input
                        size="small"
                        placeholder="参数值"
                        value={qp.value}
                        onChange={(e) => {
                          const next = [...queryParams];
                          next[i].value = e.target.value;
                          setQueryParams(next);
                        }}
                      />
                    </Col>
                    <Col span={2}>
                      <Button
                        size="small"
                        danger
                        type="text"
                        onClick={() => setQueryParams((prev) => prev.filter((_, j) => j !== i))}
                      >
                        ✕
                      </Button>
                    </Col>
                  </Row>
                ))}
              </div>

              {/* Request Body */}
              {['POST', 'PUT', 'PATCH'].includes(method) && (
                <div style={{ marginBottom: 12 }}>
                  <Text type="secondary" style={{ fontSize: 12, marginBottom: 4, display: 'block' }}>请求体 (JSON):</Text>
                  <TextArea
                    rows={8}
                    value={bodyText}
                    onChange={(e) => setBodyText(e.target.value)}
                    placeholder='{"key": "value"}'
                    style={{ fontFamily: 'monospace', fontSize: 13 }}
                  />
                </div>
              )}

              {/* Endpoint Info */}
              <div style={{ marginBottom: 12 }}>
                <Space size={12}>
                  {selectedEndpoint.pathParams && selectedEndpoint.pathParams.length > 0 && (
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>路径参数: </Text>
                      {selectedEndpoint.pathParams.map((p) => (
                        <Tag key={p.name} style={{ fontSize: 11 }}>
                          {p.required ? <span style={{ color: '#ff4d4f' }}>*</span> : null}
                          {p.name} <Text type="secondary">({p.type})</Text> — {p.desc}
                        </Tag>
                      ))}
                    </div>
                  )}
                  {selectedEndpoint.queryParams && selectedEndpoint.queryParams.length > 0 && (
                    <div>
                      <Text type="secondary" style={{ fontSize: 11 }}>Query参数: </Text>
                      {selectedEndpoint.queryParams.map((p) => (
                        <Tag key={p.name} style={{ fontSize: 11 }}>
                          {p.name} <Text type="secondary">({p.type})</Text> — {p.desc}
                        </Tag>
                      ))}
                    </div>
                  )}
                </Space>
              </div>
            </div>

            {/* Response */}
            <div style={{ borderTop: '1px solid #f0f0f0', paddingTop: 12 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                <Space>
                  <Text strong>响应结果</Text>
                  {response && (
                    <>
                      <Tag color={response.status >= 200 && response.status < 300 ? 'green' : 'red'} icon={response.status >= 200 && response.status < 300 ? <CheckCircleOutlined /> : <CloseCircleOutlined />}>
                        {response.status} {response.statusText}
                      </Tag>
                      <Text type="secondary" style={{ fontSize: 12 }}>{response.time}ms</Text>
                    </>
                  )}
                </Space>
                {response && (
                  <Button size="small" icon={<CopyOutlined />} onClick={handleCopyResponse}>
                    复制
                  </Button>
                )}
              </div>
              <Spin spinning={loading}>
                {response ? (
                  <Tabs
                    size="small"
                    items={[
                      {
                        key: 'body',
                        label: 'Body',
                        children: (
                          <pre
                            style={{
                              background: '#1e1e1e',
                              color: '#d4d4d4',
                              padding: 16,
                              borderRadius: 8,
                              maxHeight: 400,
                              overflow: 'auto',
                              fontSize: 13,
                              lineHeight: 1.6,
                              margin: 0,
                            }}
                          >
                            {JSON.stringify(response.body, null, 2)}
                          </pre>
                        ),
                      },
                      {
                        key: 'headers',
                        label: 'Headers',
                        children: (
                          <pre
                            style={{
                              background: '#f5f5f5',
                              padding: 16,
                              borderRadius: 8,
                              maxHeight: 400,
                              overflow: 'auto',
                              fontSize: 13,
                              lineHeight: 1.6,
                              margin: 0,
                            }}
                          >
                            {JSON.stringify(response.headers, null, 2)}
                          </pre>
                        ),
                      },
                    ]}
                  />
                ) : (
                  <div
                    style={{
                      background: '#fafafa',
                      borderRadius: 8,
                      height: 200,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                  >
                    <Text type="secondary">
                      点击左侧接口，填写参数后点击「发送」查看响应
                    </Text>
                  </div>
                )}
              </Spin>
            </div>
          </div>
        ) : (
          <div
            style={{
              height: 300,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Empty description="请从左侧选择一个接口开始测试" />
          </div>
        )}
      </Card>
    </div>
  );
}
