import { useState, useMemo, useCallback } from 'react';
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
  Alert,
} from 'antd';
import {
  SendOutlined,
  ClearOutlined,
  CopyOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  CodeOutlined,
} from '@ant-design/icons';
import client from '../api/client';
import axios from 'axios';

const { Text, Paragraph } = Typography;
const { TextArea } = Input;

// ─── Types ──────────────────────────────────────────────────────────────

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
  responseExample?: Record<string, unknown>;
}

interface EndpointGroup {
  name: string;
  icon?: string;
  endpoints: EndpointDef[];
}

// ─── Constants ──────────────────────────────────────────────────────────

const METHOD_COLORS: Record<string, string> = {
  GET: '#52c41a',
  POST: '#1677ff',
  PUT: '#fa8c16',
  PATCH: '#722ed1',
  DELETE: '#ff4d4f',
};

const BASE_URL = '/api/v1';
const BASE_DOMAIN = 'https://coupon.mx.yn.cn';

// ─── Endpoint data ──────────────────────────────────────────────────────

const endpointGroups: EndpointGroup[] = [
  {
    name: '门店管理 Stores',
    icon: '🏪',
    endpoints: [
      {
        method: 'GET', path: '/admin/stores', desc: '门店列表', auth: 'JWT',
        queryParams: [
          { name: 'page', type: 'number', desc: '页码，默认1' },
          { name: 'page_size', type: 'number', desc: '每页条数，默认20' },
        ],
        responseExample: { code: 0, message: 'success', data: { total: 1, page: 1, page_size: 20, items: [{ id: 1, name: '旗舰店', code: 'ABC12', app_id: 'wxabc123', type: 'miniprogram', status: 1, contact_name: '张三', contact_phone: '13800138000', remark: '' }] } },
      },
      {
        method: 'GET', path: '/admin/stores/:id', desc: '门店详情', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }],
        responseExample: { code: 0, message: 'success', data: { id: 1, name: '旗舰店', code: 'ABC12', app_id: 'wxabc123', type: 'miniprogram', status: 1, contact_name: '张三', contact_phone: '13800138000', remark: '' } },
      },
      {
        method: 'GET', path: '/admin/stores/options', desc: '门店下拉选项', auth: 'JWT',
        responseExample: { code: 0, message: 'success', data: [{ id: 1, name: '旗舰店', code: 'ABC12' }] },
      },
      {
        method: 'POST', path: '/admin/stores', desc: '创建门店（编码自动生成）', auth: 'JWT',
        body: { name: '新门店', app_id: 'wxabc123', type: 'miniprogram', contact_name: '张三', contact_phone: '13800138000', remark: '' },
        responseExample: { code: 0, message: 'success', data: { id: 2, name: '新门店', code: 'XK9M3', app_id: 'wxabc123', type: 'miniprogram', status: 1, contact_name: '', contact_phone: '', remark: '', credentials: { app_key: 'ak_xxx', app_secret: 'sk_xxx' } } },
      },
      {
        method: 'PUT', path: '/admin/stores/:id', desc: '更新门店', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }],
        body: { name: '更新名称', contact_name: '张三', contact_phone: '13800138000', remark: '' },
        responseExample: { code: 0, message: 'success', data: { id: 1, name: '更新名称', code: 'ABC12', app_id: 'wxabc123', type: 'miniprogram', status: 1, contact_name: '张三', contact_phone: '13800138000', remark: '' } },
      },
      {
        method: 'PATCH', path: '/admin/stores/:id/status', desc: '更新门店状态', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }],
        body: { status: 1 },
        responseExample: { code: 0, message: 'success' },
      },
      {
        method: 'DELETE', path: '/admin/stores/:id', desc: '删除门店', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }],
        responseExample: { code: 0, message: 'success' },
      },
      {
        method: 'POST', path: '/admin/stores/:id/credentials', desc: '生成门店凭证', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '门店ID' }],
        responseExample: { code: 0, message: 'success', data: { app_key: 'ak_xxxxxxxxxxxx', app_secret: 'sk_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx' } },
      },
    ],
  },
  {
    name: '模板管理 Templates',
    icon: '📋',
    endpoints: [
      {
        method: 'GET', path: '/admin/templates', desc: '模板列表', auth: 'JWT',
        queryParams: [
          { name: 'page', type: 'number', desc: '页码' },
          { name: 'page_size', type: 'number', desc: '每页条数' },
          { name: 'store_id', type: 'number', desc: '按门店ID筛选模板' },
        ],
        responseExample: { code: 0, message: 'success', data: { total: 1, page: 1, page_size: 20, items: [{ id: 1, name: '满50减20优惠券', type: 'full_reduction', discount_value: 20, threshold_amount: 50, applicable_scope: 'all', stackable: false, max_stack_count: 1, validity_type: 'days_after_receive', validity_days: 30, total_quantity: 100, issued_count: 0, per_user_limit: 1, status: 1, created_at: '2025-01-01T00:00:00Z' }] } },
      },
      {
        method: 'GET', path: '/admin/templates/:id', desc: '模板详情', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }],
        responseExample: { code: 0, message: 'success', data: { id: 1, name: '满50减20优惠券', type: 'full_reduction', discount_value: 20, threshold_amount: 50, applicable_scope: 'all', store_ids: [], stackable: false, max_stack_count: 1, validity_type: 'days_after_receive', validity_days: 30, valid_start: null, valid_end: null, total_quantity: 100, issued_count: 0, per_user_limit: 1, product_restriction: null, status: 1, created_at: '2025-01-01T00:00:00Z' } },
      },
      {
        method: 'POST', path: '/admin/templates', desc: '创建模板', auth: 'JWT',
        body: { name: '满50减20', type: 'full_reduction', discount_value: 20, threshold_amount: 50, applicable_scope: 'all', stackable: false, max_stack_count: 1, validity_type: 'days_after_receive', validity_days: 30, total_quantity: 100, per_user_limit: 1 },
        responseExample: { code: 0, message: 'success', data: { id: 2, name: '满50减20', type: 'full_reduction', discount_value: 20, threshold_amount: 50, applicable_scope: 'all', store_ids: [], stackable: false, max_stack_count: 1, validity_type: 'days_after_receive', validity_days: 30, total_quantity: 100, issued_count: 0, per_user_limit: 1, status: 0, created_at: '2025-01-01T00:00:00Z' } },
      },
      {
        method: 'PUT', path: '/admin/templates/:id', desc: '更新模板（仅草稿状态可编辑）', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }],
        body: { name: '满50减30', discount_value: 30, threshold_amount: 50, stackable: false, max_stack_count: 1, validity_days: 30, valid_start: '', valid_end: '', total_quantity: 200, per_user_limit: 2 },
        responseExample: { code: 0, message: 'success', data: { id: 1, name: '满50减30', type: 'full_reduction', discount_value: 30, threshold_amount: 50, applicable_scope: 'all', stackable: false, max_stack_count: 1, validity_type: 'days_after_receive', validity_days: 30, total_quantity: 200, issued_count: 0, per_user_limit: 2, status: 0, created_at: '2025-01-01T00:00:00Z' } },
      },
      {
        method: 'PATCH', path: '/admin/templates/:id/status', desc: '更新模板状态', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }],
        body: { status: 1 },
        responseExample: { code: 0, message: 'success' },
      },
      {
        method: 'DELETE', path: '/admin/templates/:id', desc: '删除模板', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }],
        responseExample: { code: 0, message: 'success' },
      },
      {
        method: 'GET', path: '/admin/browse/templates', desc: '浏览模板列表', auth: 'JWT',
        queryParams: [
          { name: 'page', type: 'number', desc: '页码' },
          { name: 'page_size', type: 'number', desc: '每页条数' },
        ],
        responseExample: { code: 0, message: 'success', data: { total: 1, page: 1, page_size: 20, items: [{ id: 1, name: '满50减20优惠券', type: 'full_reduction', discount_value: 20, threshold_amount: 50, applicable_scope: 'all', stackable: false, validity_type: 'days_after_receive', validity_days: 30, total_quantity: 100, issued_count: 0, per_user_limit: 1, status: 1, mp_appid: 'wxabc123', mp_page_path: 'pages/coupon/use' }] } },
      },
      {
        method: 'GET', path: '/admin/browse/templates/:id', desc: '浏览模板详情', auth: 'JWT',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '模板ID' }],
        responseExample: { code: 0, message: 'success', data: { id: 1, name: '满50减20优惠券', type: 'full_reduction', discount_value: 20, threshold_amount: 50, applicable_scope: 'all', store_ids: [], stackable: false, max_stack_count: 1, validity_type: 'days_after_receive', validity_days: 30, valid_start: null, valid_end: null, total_quantity: 100, issued_count: 0, per_user_limit: 1, product_restriction: null, status: 1, created_at: '2025-01-01T00:00:00Z' } },
      },
    ],
  },
  {
    name: '优惠券 Coupons',
    icon: '🎫',
    endpoints: [
      {
        method: 'POST', path: '/admin/coupons/issue', desc: '发放优惠券（幂等键自动生成）', auth: 'JWT + 已审批',
        body: { store_id: 1, template_id: 1, user_phone: '13800138000' },
        responseExample: { code: 0, message: 'success', data: { coupon_id: 1, coupon_code: 'ABC12DEF34GH', template_name: '满减券', type: 'full_reduction', discount_value: 10, threshold_amount: 100, valid_start: '2025-01-01T00:00:00Z', valid_end: '2025-02-01T00:00:00Z', status: 'unused' } },
      },
      {
        method: 'GET', path: '/admin/coupons/records', desc: '发券记录列表', auth: 'JWT + 已审批',
        queryParams: [
          { name: 'page', type: 'number', desc: '页码' },
          { name: 'page_size', type: 'number', desc: '每页条数' },
          { name: 'store_id', type: 'number', desc: '按门店筛选' },
          { name: 'status', type: 'string', desc: '按状态筛选' },
        ],
        responseExample: { code: 0, message: 'success', data: { total: 1, page: 1, page_size: 20, items: [{ id: 1, coupon_code: 'ABC12DEF34GH', template_name: '满50减20优惠券', type: 'full_reduction', discount_value: 20, user_phone: '13800138000', status: 'unused', source_store_name: '旗舰店', receive_time: '2025-01-01T12:00:00Z', use_time: null }] } },
      },
      {
        method: 'GET', path: '/admin/coupons/records/:id', desc: '发券记录详情', auth: 'JWT + 已审批',
        pathParams: [{ name: 'id', type: 'number', required: true, desc: '记录ID' }],
        responseExample: { code: 0, message: 'success', data: { coupon_id: 1, coupon_code: 'ABC12DEF34GH', template_name: '满50减20优惠券', type: 'full_reduction', discount_value: 20, threshold_amount: 50, status: 'unused', user_phone: '13800138000', source_store_name: '旗舰店', valid_start: '2025-01-01T00:00:00Z', valid_end: '2025-02-01T00:00:00Z', receive_time: '2025-01-01T12:00:00Z', use_time: null, used_at_store_name: '', use_order_id: '', records: [] } },
      },
      {
        method: 'POST', path: '/admin/coupons/consume', desc: '核销优惠券', auth: 'JWT + 已审批',
        body: { coupon_code: 'ABC12DEF34GH', user_phone: '13800138000', store_id: 1, order_id: 'ORD_001', order_amount: 199.00 },
        responseExample: { code: 0, message: 'success', data: { coupon_id: 1, coupon_code: 'ABC12DEF34GH', discount_value: 10, actual_amount: 189.00, used_at: '2025-01-01T14:00:00Z' } },
      },
      {
        method: 'GET', path: '/admin/coupons/consume-records', desc: '核销记录列表', auth: 'JWT + 已审批',
        queryParams: [
          { name: 'page', type: 'number', desc: '页码' },
          { name: 'page_size', type: 'number', desc: '每页条数' },
        ],
        responseExample: { code: 0, message: 'success', data: { total: 1, page: 1, page_size: 20, items: [{ id: 1, coupon_id: 1, user_phone: '13800138000', store_name: '旗舰店', action: 'consume', order_info: { order_id: 'ORD_001', order_amount: 199 }, created_at: '2025-01-01T14:00:00Z' }] } },
      },
    ],
  },
  {
    name: '报表 Reports',
    icon: '📊',
    endpoints: [
      {
        method: 'GET', path: '/admin/reports/overview', desc: '报表概览', auth: 'JWT + 已审批',
        responseExample: { code: 0, message: 'success', data: { total_stores: 5, total_templates: 10, total_issued: 1000, total_used: 650, usage_rate: 0.65, today_issued: 25, today_used: 12 } },
      },
      {
        method: 'GET', path: '/admin/reports/trend', desc: '趋势数据', auth: 'JWT + 已审批',
        queryParams: [
          { name: 'start_date', type: 'string', desc: '开始日期 (YYYY-MM-DD)', required: true },
          { name: 'end_date', type: 'string', desc: '结束日期 (YYYY-MM-DD)', required: true },
        ],
        responseExample: { code: 0, message: 'success', data: { items: [{ date: '2025-01-01', issued: 30, used: 20 }, { date: '2025-01-02', issued: 45, used: 32 }] } },
      },
      {
        method: 'GET', path: '/admin/reports/export/coupons', desc: '导出券数据CSV文件', auth: 'JWT + 已审批',
        responseExample: { code: 0, message: 'success', data: { note: '返回 CSV 文件下载流' } },
      },
      {
        method: 'GET', path: '/admin/reports/export/usage', desc: '导出核销数据CSV文件', auth: 'JWT + 已审批',
        responseExample: { code: 0, message: 'success', data: { note: '返回 CSV 文件下载流' } },
      },
    ],
  },
  {
    name: '统计 Statistics',
    icon: '📈',
    endpoints: [
      {
        method: 'GET', path: '/admin/statistics/overview', desc: '统计概览', auth: 'JWT',
        responseExample: { code: 0, message: 'success', data: { total_stores: 5, total_templates: 10, total_issued: 1000, total_used: 650, usage_rate: 0.65, today_issued: 25, today_used: 12 } },
      },
      {
        method: 'GET', path: '/admin/statistics/trend', desc: '统计趋势', auth: 'JWT',
        queryParams: [
          { name: 'start_date', type: 'string', desc: '开始日期' },
          { name: 'end_date', type: 'string', desc: '结束日期' },
        ],
        responseExample: { code: 0, message: 'success', data: [{ date: '2025-01-01', issued: 30, used: 20 }] },
      },
    ],
  },
  {
    name: 'Open API (HMAC)',
    icon: '🔗',
    endpoints: [
      {
        method: 'GET', path: '/coupons/templates', desc: '浏览可发券模板（含小程序跳转信息）', auth: 'HMAC + 频率限制',
        queryParams: [
          { name: 'page', type: 'number', desc: '页码，默认1' },
          { name: 'page_size', type: 'number', desc: '每页条数，默认20' },
        ],
        responseExample: { code: 0, message: 'success', data: { total: 1, page: 1, page_size: 20, items: [{ id: 1, name: '满50减20优惠券', type: 'full_reduction', discount_value: 20, threshold_amount: 50, applicable_scope: 'all', stackable: false, validity_type: 'days_after_receive', validity_days: 30, total_quantity: 100, per_user_limit: 1, status: 1, mp_appid: 'wxabc123', mp_page_path: 'pages/coupon/use' }] } },
      },
      {
        method: 'POST', path: '/coupons/issue', desc: '发放优惠券（幂等键自动生成）', auth: 'HMAC + 频率限制',
        body: { template_id: 1, user_phone: '13800138000' },
        responseExample: { code: 0, message: 'success', data: { coupon_id: 1, coupon_code: 'XYZ12ABC34DE', template_name: '满减券', type: 'full_reduction', discount_value: 10, threshold_amount: 100, valid_start: '2025-01-01T00:00:00Z', valid_end: '2025-02-01T00:00:00Z', status: 'unused' } },
      },
      {
        method: 'GET', path: '/coupons/available', desc: '查询可用优惠券 (HMAC)', auth: 'HMAC + 频率限制',
        queryParams: [
          { name: 'user_phone', type: 'string', desc: '用户手机号', required: true },
          { name: 'store_id', type: 'number', desc: '门店ID', required: true },
          { name: 'order_amount', type: 'number', desc: '订单金额（用于门槛校验）' },
          { name: 'page', type: 'number', desc: '页码，默认1' },
          { name: 'page_size', type: 'number', desc: '每页条数，默认20' },
        ],
        responseExample: { code: 0, message: 'success', data: { total: 1, items: [{ coupon_id: 1, coupon_code: 'XYZ12ABC34DE', template_name: '满减券', type: 'full_reduction', discount_value: 10, threshold_amount: 100, valid_end: '2025-12-31T23:59:59Z', stackable: false }] } },
      },
      {
        method: 'GET', path: '/coupons/user', desc: '用户优惠券列表 (HMAC)', auth: 'HMAC + 频率限制',
        queryParams: [
          { name: 'user_phone', type: 'string', desc: '用户手机号', required: true },
          { name: 'status', type: 'string', desc: '状态: unused/used/expired/revoke' },
          { name: 'page', type: 'number', desc: '页码，默认1' },
          { name: 'page_size', type: 'number', desc: '每页条数，默认20' },
        ],
        responseExample: { code: 0, message: 'success', data: { total: 1, page: 1, page_size: 20, items: [{ coupon_id: 1, coupon_code: 'TY000260613LZSPH0', template_name: '满50减20优惠券', type: 'full_reduction', discount_value: 20, threshold_amount: 50, status: 'unused', valid_start: '2026-06-13T01:10:20.543+08:00', valid_end: '2027-06-13T01:10:20.543+08:00', mp_appid: 'wxabc123', mp_page_path: 'pages/coupon/use' }] } },
      },
      {
        method: 'GET', path: '/coupons/:coupon_code', desc: '优惠券详情 (HMAC)', auth: 'HMAC + 频率限制',
        pathParams: [{ name: 'coupon_code', type: 'string', required: true, desc: '券码' }],
        responseExample: { code: 0, message: 'success', data: { coupon_id: 1, coupon_code: 'XYZ12ABC34DE', template_name: '满减券', type: 'full_reduction', discount_value: 10, threshold_amount: 100, status: 'unused', user_phone: '13800138000', source_store_name: '旗舰店', valid_start: '2025-01-01T00:00:00Z', valid_end: '2025-02-01T00:00:00Z', receive_time: '2025-01-01T12:00:00Z', mp_appid: 'wxabc123', mp_page_path: 'pages/coupon/use', records: [] } },
      },
      {
        method: 'POST', path: '/coupons/consume', desc: '核销优惠券 (HMAC)', auth: 'HMAC + 频率限制',
        body: { coupon_code: 'XYZ12ABC34DE', user_phone: '13800138000', store_id: 1, order_id: 'ORD_001', order_amount: 199.00 },
        responseExample: { code: 0, message: 'success', data: { coupon_id: 1, coupon_code: 'XYZ12ABC34DE', discount_value: 10, actual_amount: 189.00, used_at: '2025-01-01T14:00:00Z' } },
      },
      {
        method: 'POST', path: '/coupons/refund', desc: '退券 (HMAC)', auth: 'HMAC + 频率限制',
        body: { coupon_code: 'XYZ12ABC34DE', user_phone: '13800138000', store_id: 1, order_id: 'ORD_001' },
        responseExample: { code: 0, message: 'success', data: { coupon_id: 1, coupon_code: 'XYZ12ABC34DE', new_status: 'unused', restored: true } },
      },
    ],
  },
];

// ─── Helpers ────────────────────────────────────────────────────────────

function extractPathParams(path: string): string[] {
  const re = /:(\w+)/g;
  const params: string[] = [];
  let m: RegExpExecArray | null;
  while ((m = re.exec(path)) !== null) {
    params.push(m[1]);
  }
  return params;
}

function getAccessToken(): string | null {
  try {
    const raw = localStorage.getItem('auth-storage');
    if (!raw) return null;
    return JSON.parse(raw)?.state?.accessToken ?? null;
  } catch { return null; }
}

// ─── Code generators ────────────────────────────────────────────────────

function buildFinalUrl(path: string, pathParams: Record<string, string>, queryParams: { key: string; value: string }[]): string {
  let p = path;
  Object.entries(pathParams).forEach(([k, v]) => { if (v) p = p.replace(`:${k}`, v); });
  const qps = queryParams.filter((q) => q.key.trim());
  const qs = qps.length ? '?' + new URLSearchParams(qps.map((q) => [q.key, q.value])).toString() : '';
  return p + qs;
}

function generateCode(method: string, urlPath: string, pathParams: Record<string, string>, queryParams: { key: string; value: string }[], bodyText: string, isHMAC: boolean, token: string | null): Record<string, string> {
  const fullPath = buildFinalUrl(urlPath, pathParams, queryParams);
  const hasBody = ['POST', 'PUT', 'PATCH'].includes(method);
  const bodyObj = hasBody ? (() => { try { return JSON.parse(bodyText); } catch { return {}; } })() : null;

  const cUrlParts: string[] = ['curl'];
  if (method !== 'GET') cUrlParts.push(`-X ${method}`);

  if (isHMAC) {
    cUrlParts.push(`-H "X-App-Key: YOUR_APP_KEY"`);
    cUrlParts.push(`-H "X-Timestamp: $(date +%s)"`);
    cUrlParts.push(`-H "X-Nonce: $(uuidgen)"`);
    cUrlParts.push(`-H "X-Signature: <GENERATED_SIGNATURE>"`);
  } else if (token) {
    cUrlParts.push(`-H "Authorization: Bearer ${token}"`);
  }
  cUrlParts.push(`-H "Content-Type: application/json"`);
  if (hasBody && bodyObj) {
    cUrlParts.push(`-d '${JSON.stringify(bodyObj)}'`);
  }
  cUrlParts.push(`"${BASE_DOMAIN}${BASE_URL}${fullPath}"`);
  const curl = cUrlParts.join(' \\\n  ');

  // JavaScript (fetch)
  const jsLines: string[] = [];
  if (isHMAC) {
    jsLines.push(`// 需要先调用 signRequest() 生成签名头（见下方 HMAC 签名示例）`, '');
  }
  jsLines.push(`fetch('${BASE_DOMAIN}${BASE_URL}${fullPath}', {`);
  jsLines.push(`  method: '${method}',`);
  if (!isHMAC && token) {
    jsLines.push(`  headers: { 'Authorization': 'Bearer ${token}', 'Content-Type': 'application/json' },`);
  } else if (!isHMAC) {
    jsLines.push(`  headers: { 'Content-Type': 'application/json' },`);
  } else {
    jsLines.push(`  headers, // signRequest() 返回的 headers`);
  }
  if (hasBody && bodyObj) {
    jsLines.push(`  body: JSON.stringify(${JSON.stringify(bodyObj, null, 2).split('\n').map((l, i) => i === 0 ? l : '  ' + l).join('\n')}),`);
  }
  jsLines.push('})');
  jsLines.push(`  .then(r => r.json())`);
  jsLines.push(`  .then(data => console.log(data));`);
  const js = jsLines.join('\n');

  // Python
  const pyLines: string[] = [];
  pyLines.push('import requests');
  if (isHMAC) {
    pyLines.push('# 需要先调用 sign_request() 生成签名头（见下方 HMAC 签名示例）');
  }
  pyLines.push('');
  if (!isHMAC && token) {
    pyLines.push(`headers = {"Authorization": "Bearer ${token}", "Content-Type": "application/json"}`);
  } else if (isHMAC) {
    pyLines.push(`headers = sign_request('${method}', '${fullPath}', ${hasBody && bodyObj ? JSON.stringify(bodyObj) : 'None'}, 'YOUR_APP_KEY', 'YOUR_APP_SECRET')`);
  } else {
    pyLines.push(`headers = {"Content-Type": "application/json"}`);
  }
  pyLines.push('');
  if (hasBody && bodyObj) {
    pyLines.push(`data = ${JSON.stringify(bodyObj, null, 2).split('\n').map(l => l).join('\n')}`);
    pyLines.push('');
  }
  pyLines.push(`resp = requests.${method.toLowerCase()}('${BASE_DOMAIN}${BASE_URL}${fullPath}'${hasBody && bodyObj ? ', json=data' : ''}, headers=headers)`);
  pyLines.push('print(resp.json())');
  const py = pyLines.join('\n');

  // Go
  const goLines: string[] = [];
  goLines.push('import (');
  goLines.push('    "bytes"');
  goLines.push('    "encoding/json"');
  goLines.push('    "fmt"');
  goLines.push('    "io"');
  goLines.push('    "net/http"');
  goLines.push(')');
  goLines.push('');
  if (hasBody && bodyObj) {
    goLines.push(`body := map[string]interface{}{`);
    Object.entries(bodyObj).forEach(([k]) => { goLines.push(`    "${k}": "",`); });
    goLines.push('}');
    goLines.push('bodyJSON, _ := json.Marshal(body)');
    goLines.push('');
  }
  goLines.push(`url := "${BASE_DOMAIN}${BASE_URL}${fullPath}"`);
  if (hasBody && bodyObj) {
    goLines.push('req, _ := http.NewRequest("' + method + '", url, bytes.NewBuffer(bodyJSON))');
  } else {
    goLines.push('req, _ := http.NewRequest("' + method + '", url, nil)');
  }
  if (!isHMAC && token) {
    goLines.push(`req.Header.Set("Authorization", "Bearer ${token}")`);
  }
  if (isHMAC) {
    goLines.push('// 需要先调用 SignRequest() 设置 HMAC 签名头（见下方 HMAC 签名示例）');
    goLines.push('// headers := SignRequest("' + method + '", "' + fullPath + '", bodyJSONString, appKey, appSecret)');
    goLines.push('// for k, v := range headers { req.Header.Set(k, v) }');
  }
  goLines.push('req.Header.Set("Content-Type", "application/json")');
  goLines.push('');
  goLines.push('client := &http.Client{}');
  goLines.push('resp, _ := client.Do(req)');
  goLines.push('defer resp.Body.Close()');
  goLines.push('respBody, _ := io.ReadAll(resp.Body)');
  goLines.push('fmt.Println(string(respBody))');
  const go = goLines.join('\n');

  // PHP
  const phpLines: string[] = [];
  phpLines.push('<?php');
  phpLines.push('');
  if (isHMAC) {
    phpLines.push('// 需要先调用 sign_request() 生成签名头（见下方 HMAC 签名示例）');
    phpLines.push('');
  }
  if (hasBody && bodyObj) {
    phpLines.push(`$body = json_encode(${JSON.stringify(bodyObj)});`);
    phpLines.push('');
  }
  phpLines.push('$ch = curl_init();');
  phpLines.push(`curl_setopt($ch, CURLOPT_URL, '${BASE_DOMAIN}${BASE_URL}${fullPath}');`);
  phpLines.push(`curl_setopt($ch, CURLOPT_CUSTOMREQUEST, '${method}');`);
  phpLines.push('curl_setopt($ch, CURLOPT_HTTPHEADER, [');
  if (!isHMAC && token) {
    phpLines.push(`    'Authorization: Bearer ${token}',`);
  }
  if (isHMAC) {
    phpLines.push('    // 使用 sign_request() 返回的 headers');
    phpLines.push('    // "X-App-Key: YOUR_APP_KEY",');
    phpLines.push('    // "X-Timestamp: " . $timestamp,');
    phpLines.push('    // "X-Nonce: " . $nonce,');
    phpLines.push('    // "X-Signature: " . $signature,');
  }
  phpLines.push(`    'Content-Type: application/json',`);
  phpLines.push(']);');
  if (hasBody && bodyObj) {
    phpLines.push('curl_setopt($ch, CURLOPT_POSTFIELDS, $body);');
  }
  phpLines.push('curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);');
  phpLines.push('$response = curl_exec($ch);');
  phpLines.push('curl_close($ch);');
  phpLines.push('');
  phpLines.push('echo $response;');
  const php = phpLines.join('\n');

  // C#
  const csLines: string[] = [];
  csLines.push('using System.Net.Http;');
  csLines.push('using System.Text;');
  csLines.push('using Newtonsoft.Json;');
  csLines.push('');
  if (isHMAC) {
    csLines.push('// 需要先调用 SignRequest() 生成签名头（见下方 HMAC 签名示例）');
    csLines.push('');
  }
  csLines.push('var client = new HttpClient();');
  if (!isHMAC && token) {
    csLines.push(`client.DefaultRequestHeaders.Authorization = new AuthenticationHeaderValue("Bearer", "${token}");`);
  }
  if (isHMAC) {
    csLines.push('// client.DefaultRequestHeaders.Add("X-App-Key", appKey);');
    csLines.push('// client.DefaultRequestHeaders.Add("X-Timestamp", timestamp);');
    csLines.push('// client.DefaultRequestHeaders.Add("X-Nonce", nonce);');
    csLines.push('// client.DefaultRequestHeaders.Add("X-Signature", signature);');
  }
  csLines.push('');
  if (hasBody && bodyObj) {
    csLines.push(`var body = new ${JSON.stringify(bodyObj).replace(/"/g, '').replace(/:([^,}]+)/g, ': "$1"')};`);
    // Use simple anonymous object representation
    const csBodyLines: string[] = [];
    csBodyLines.push('var body = new');
    csBodyLines.push('{');
    const entries = Object.entries(bodyObj);
    entries.forEach(([k, v], i) => {
      const comma = i < entries.length - 1 ? ',' : '';
      if (typeof v === 'number') {
        csBodyLines.push(`    ${k} = ${v}${comma}`);
      } else {
        csBodyLines.push(`    ${k} = "${v}"${comma}`);
      }
    });
    csBodyLines.push('};');
    csLines[csLines.length - 1] = csBodyLines.join('\n');
    csLines.push('var content = new StringContent(JsonConvert.SerializeObject(body), Encoding.UTF8, "application/json");');
  }
  csLines.push('');
  const csMethodLower = method.toLowerCase();
  if (hasBody && bodyObj) {
    if (csMethodLower === 'post') {
      csLines.push(`var response = await client.PostAsync("${BASE_DOMAIN}${BASE_URL}${fullPath}", content);`);
    } else if (csMethodLower === 'put') {
      csLines.push(`var response = await client.PutAsync("${BASE_DOMAIN}${BASE_URL}${fullPath}", content);`);
    } else {
      csLines.push(`var request = new HttpRequestMessage(HttpMethod.${method}, "${BASE_DOMAIN}${BASE_URL}${fullPath}") { Content = content };`);
      csLines.push('var response = await client.SendAsync(request);');
    }
  } else {
    csLines.push(`var response = await client.GetAsync("${BASE_DOMAIN}${BASE_URL}${fullPath}");`);
  }
  csLines.push('var result = await response.Content.ReadAsStringAsync();');
  csLines.push('Console.WriteLine(result);');
  const cs = csLines.join('\n');

  return { curl, js, py, go, php, cs };
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
  const [rightTab, setRightTab] = useState<string>('tester');

  const activePathParams = useMemo(() => extractPathParams(urlPath), [urlPath]);

  const codeSamples = useMemo(() => {
    if (!selectedEndpoint) return null;
    const isHMAC = selectedEndpoint.auth.includes('HMAC');
    const token = getAccessToken();
    return generateCode(method, urlPath, pathParamValues, queryParams, bodyText, isHMAC, token);
  }, [selectedEndpoint, method, urlPath, pathParamValues, queryParams, bodyText]);

  const handleSelectEndpoint = useCallback((ep: EndpointDef) => {
    setSelectedEndpoint(ep);
    setMethod(ep.method);
    setUrlPath(ep.path);
    setResponse(null);
    setRightTab('tester');

    const pp: Record<string, string> = {};
    (ep.pathParams || []).forEach((p) => { pp[p.name] = ''; });
    setPathParamValues(pp);

    const qps = (ep.queryParams || []).map((p) => ({ key: p.name, value: '' }));
    setQueryParams(qps.length > 0 ? qps : [{ key: '', value: '' }]);

    setBodyText(ep.body ? JSON.stringify(ep.body, null, 2) : '');
    setUseRawClient(ep.auth.includes('HMAC'));
  }, []);

  const handleSend = async () => {
    if (!urlPath) return;
    setLoading(true);
    const startTime = performance.now();

    let finalPath = urlPath;
    Object.entries(pathParamValues).forEach(([k, v]) => {
      if (v) finalPath = finalPath.replace(`:${k}`, v);
    });

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
        try { data = bodyText ? JSON.parse(bodyText) : {}; } catch {
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
      setResponse({ status: res.status, statusText: res.statusText, headers: res.headers as Record<string, string>, body: res.data, time: elapsed });
      setRightTab('tester');
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
      setRightTab('tester');
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

  const handleCopyCode = (code: string) => {
    navigator.clipboard.writeText(code);
    message.success('已复制到剪贴板');
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

      {/* ── Right: Tester + Examples ── */}
      <Card
        title={
          <Space>
            <span>{selectedEndpoint ? selectedEndpoint.desc : 'API 测试 & 示例'}</span>
            {selectedEndpoint && <Tag>{selectedEndpoint.method} {selectedEndpoint.path}</Tag>}
          </Space>
        }
        size="small"
        style={{ flex: 1, overflow: 'auto', maxHeight: 'calc(100vh - 160px)' }}
        extra={
          selectedEndpoint && (
            <Space>
              {selectedEndpoint.auth.includes('HMAC') && (
                <Tag color="warning">HMAC — 需生成签名</Tag>
              )}
              {selectedEndpoint.auth.includes('JWT') && (
                <Tag color="green">JWT — 自动携带 Token</Tag>
              )}
            </Space>
          )
        }
      >
        {selectedEndpoint ? (
          <Tabs
            activeKey={rightTab}
            onChange={setRightTab}
            size="small"
            items={[
              {
                key: 'tester',
                label: 'API 测试工具',
                children: (
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
                          placeholder="/admin/..."
                          addonBefore={`${BASE_DOMAIN}/api/v1`}
                          style={{ fontFamily: 'monospace' }}
                        />
                        <Button type="primary" icon={<SendOutlined />} loading={loading} onClick={handleSend}>
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
                                  addonBefore={<span style={{ color: '#ff4d4f', fontSize: 12 }}>*{p}</span>}
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
                          <Button type="dashed" size="small" onClick={() => setQueryParams((prev) => [...prev, { key: '', value: '' }])}>
                            + 添加
                          </Button>
                        </Space>
                        {queryParams.map((qp, i) => (
                          <Row gutter={8} key={i} style={{ marginBottom: 4 }}>
                            <Col span={8}>
                              <Input size="small" placeholder="参数名" value={qp.key}
                                onChange={(e) => { const next = [...queryParams]; next[i].key = e.target.value; setQueryParams(next); }} />
                            </Col>
                            <Col span={14}>
                              <Input size="small" placeholder="参数值" value={qp.value}
                                onChange={(e) => { const next = [...queryParams]; next[i].value = e.target.value; setQueryParams(next); }} />
                            </Col>
                            <Col span={2}>
                              <Button size="small" danger type="text" onClick={() => setQueryParams((prev) => prev.filter((_, j) => j !== i))}>✕</Button>
                            </Col>
                          </Row>
                        ))}
                      </div>

                      {/* Request Body */}
                      {['POST', 'PUT', 'PATCH'].includes(method) && (
                        <div style={{ marginBottom: 12 }}>
                          <Text type="secondary" style={{ fontSize: 12, marginBottom: 4, display: 'block' }}>请求体 (JSON):</Text>
                          <TextArea rows={8} value={bodyText} onChange={(e) => setBodyText(e.target.value)}
                            style={{ fontFamily: 'monospace', fontSize: 13 }} />
                        </div>
                      )}

                      {/* Endpoint param info */}
                      <div style={{ marginBottom: 12 }}>
                        <Space size={12} wrap>
                          {selectedEndpoint.pathParams && selectedEndpoint.pathParams.length > 0 && (
                            <div>
                              <Text type="secondary" style={{ fontSize: 11 }}>路径参数: </Text>
                              {selectedEndpoint.pathParams.map((p) => (
                                <Tag key={p.name} style={{ fontSize: 11 }}>
                                  {p.required ? <span style={{ color: '#ff4d4f' }}>*</span> : null}
                                  {p.name} ({p.type}) — {p.desc}
                                </Tag>
                              ))}
                            </div>
                          )}
                          {selectedEndpoint.queryParams && selectedEndpoint.queryParams.length > 0 && (
                            <div>
                              <Text type="secondary" style={{ fontSize: 11 }}>Query: </Text>
                              {selectedEndpoint.queryParams.map((p) => (
                                <Tag key={p.name} style={{ fontSize: 11 }}>
                                  {p.required ? <span style={{ color: '#ff4d4f' }}>*</span> : null}
                                  {p.name} ({p.type}) — {p.desc}
                                </Tag>
                              ))}
                            </div>
                          )}
                        </Space>
                      </div>

                      {/* Expected Response */}
                      {selectedEndpoint.responseExample && (
                        <div style={{ marginBottom: 12, padding: 12, background: '#f6ffed', borderRadius: 8, border: '1px solid #b7eb8f' }}>
                          <Text type="secondary" style={{ fontSize: 12, marginBottom: 4, display: 'block' }}>预期返回:</Text>
                          <pre style={{ margin: 0, fontSize: 12, lineHeight: 1.5, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                            {JSON.stringify(selectedEndpoint.responseExample, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>

                    {/* Response */}
                    <div style={{ borderTop: '1px solid #f0f0f0', paddingTop: 12 }}>
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                        <Space>
                          <Text strong>响应结果</Text>
                          {response && (
                            <>
                              <Tag color={response.status >= 200 && response.status < 300 ? 'green' : 'red'}
                                icon={response.status >= 200 && response.status < 300 ? <CheckCircleOutlined /> : <CloseCircleOutlined />}>
                                {response.status} {response.statusText}
                              </Tag>
                              <Text type="secondary" style={{ fontSize: 12 }}>{response.time}ms</Text>
                            </>
                          )}
                        </Space>
                        {response && (
                          <Button size="small" icon={<CopyOutlined />} onClick={handleCopyResponse}>复制</Button>
                        )}
                      </div>
                      <Spin spinning={loading}>
                        {response ? (
                          <Tabs size="small" items={[
                            {
                              key: 'body',
                              label: 'Body',
                              children: (
                                <pre style={{ background: '#1e1e1e', color: '#d4d4d4', padding: 16, borderRadius: 8, maxHeight: 300, overflow: 'auto', fontSize: 13, lineHeight: 1.6, margin: 0 }}>
                                  {JSON.stringify(response.body, null, 2)}
                                </pre>
                              ),
                            },
                            {
                              key: 'headers',
                              label: 'Headers',
                              children: (
                                <pre style={{ background: '#f5f5f5', padding: 16, borderRadius: 8, maxHeight: 300, overflow: 'auto', fontSize: 13, lineHeight: 1.6, margin: 0 }}>
                                  {JSON.stringify(response.headers, null, 2)}
                                </pre>
                              ),
                            },
                          ]} />
                        ) : (
                          <div style={{ background: '#fafafa', borderRadius: 8, height: 200, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                            <Text type="secondary">点击左侧接口，填写参数后点击「发送」查看响应</Text>
                          </div>
                        )}
                      </Spin>
                    </div>
                  </div>
                ),
              },
              {
                key: 'examples',
                label: (
                  <Space>
                    <CodeOutlined />
                    代码示例
                  </Space>
                ),
                children: (
                  <div>
                    {/* Endpoint doc */}
                    <div style={{ marginBottom: 16, padding: 12, background: '#f0f5ff', borderRadius: 8, border: '1px solid #adc6ff' }}>
                      <Text strong style={{ fontSize: 13 }}>
                        {selectedEndpoint.method} {selectedEndpoint.path}
                      </Text>
                      <br />
                      <Text type="secondary" style={{ fontSize: 12 }}>{selectedEndpoint.desc} | 认证: {selectedEndpoint.auth}</Text>
                    </div>

                    {/* Request body display */}
                    {['POST', 'PUT', 'PATCH'].includes(method) && (
                      <div style={{ marginBottom: 16 }}>
                        <Text strong style={{ fontSize: 13 }}>请求参数:</Text>
                        <pre style={{ background: '#fafafa', padding: 12, borderRadius: 8, fontSize: 12, lineHeight: 1.5, margin: '8px 0' }}>
                          {JSON.stringify(
                            (() => { try { return JSON.parse(bodyText); } catch { return {}; } })(),
                            null,
                            2
                          )}
                        </pre>
                      </div>
                    )}

                    {/* Expected response */}
                    {selectedEndpoint.responseExample && (
                      <div style={{ marginBottom: 16 }}>
                        <Text strong style={{ fontSize: 13 }}>预期返回:</Text>
                        <pre style={{ background: '#f6ffed', padding: 12, borderRadius: 8, border: '1px solid #b7eb8f', fontSize: 12, lineHeight: 1.5, margin: '8px 0', whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
                          {JSON.stringify(selectedEndpoint.responseExample, null, 2)}
                        </pre>
                      </div>
                    )}

                    {/* HMAC signature doc for Open API */}
                    {selectedEndpoint.auth.includes('HMAC') && (
                      <div style={{ marginBottom: 16 }}>
                        <Alert
                          type="warning"
                          message="HMAC 签名算法"
                          description={
                            <div>
                              <Paragraph style={{ marginBottom: 4, fontSize: 12 }}>
                                签名串 = HTTP方法 + "\n" + URL路径 + "\n" + Timestamp + "\n" + Nonce + "\n" + 请求体
                              </Paragraph>
                              <Paragraph style={{ marginBottom: 0, fontSize: 12 }}>
                                Signature = Base64( HMAC-SHA256( AppSecret, 签名串 ) )
                              </Paragraph>
                            </div>
                          }
                          style={{ marginBottom: 8 }}
                        />
                        <Tabs
                          size="small"
                          items={[
                            {
                              key: 'hmac-js',
                              label: 'JavaScript',
                              children: (
                                <pre style={codeBlockStyle}>
{`const crypto = require('crypto');

function signRequest(method, path, body, appKey, appSecret) {
  const timestamp = Math.floor(Date.now() / 1000).toString();
  const nonce = crypto.randomUUID();
  const bodyStr = body ? JSON.stringify(body) : '';

  const signingStr = method + '\\n' + path +
    '\\n' + timestamp + '\\n' + nonce + '\\n' + bodyStr;
  const signature = crypto.createHmac('sha256', appSecret)
    .update(signingStr).digest('base64');

  return {
    'X-App-Key': appKey,
    'X-Timestamp': timestamp,
    'X-Nonce': nonce,
    'X-Signature': signature,
    'Content-Type': 'application/json',
  };
}`}
                                </pre>
                              ),
                            },
                            {
                              key: 'hmac-py',
                              label: 'Python',
                              children: (
                                <pre style={codeBlockStyle}>
{`import hmac, hashlib, base64, time, uuid, json

def sign_request(method, path, body, app_key, app_secret):
    timestamp = str(int(time.time()))
    nonce = str(uuid.uuid4())
    body_str = json.dumps(body) if body else ''

    signing_str = (method + '\\n' + path + '\\n' +
                   timestamp + '\\n' + nonce + '\\n' + body_str)
    sig = hmac.new(app_secret.encode(),
                   signing_str.encode(),
                   hashlib.sha256).digest()
    signature = base64.b64encode(sig).decode()

    return {
        'X-App-Key': app_key,
        'X-Timestamp': timestamp,
        'X-Nonce': nonce,
        'X-Signature': signature,
        'Content-Type': 'application/json',
    }`}
                                </pre>
                              ),
                            },
                            {
                              key: 'hmac-go',
                              label: 'Go',
                              children: (
                                <pre style={codeBlockStyle}>
{`import (
    "crypto/hmac"
    "crypto/sha256"
    "encoding/base64"
    "fmt"
    "time"
    "github.com/google/uuid"
)

func SignRequest(method, path, body, appKey, appSecret string) map[string]string {
    timestamp := fmt.Sprintf("%d", time.Now().Unix())
    nonce := uuid.New().String()
    signingStr := method + "\\n" + path + "\\n" +
        timestamp + "\\n" + nonce + "\\n" + body
    mac := hmac.New(sha256.New, []byte(appSecret))
    mac.Write([]byte(signingStr))
    signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
    return map[string]string{
        "X-App-Key":    appKey,
        "X-Timestamp":  timestamp,
        "X-Nonce":      nonce,
        "X-Signature":  signature,
        "Content-Type": "application/json",
    }
}`}
                                </pre>
                              ),
                            },
                            {
                              key: 'hmac-php',
                              label: 'PHP',
                              children: (
                                <pre style={codeBlockStyle}>
{`<?php

function sign_request($method, $path, $body, $appKey, $appSecret) {
    $timestamp = (string)time();
    $nonce = bin2hex(random_bytes(16));
    $bodyStr = $body ? json_encode($body) : '';

    $signingStr = $method . "\\n" . $path . "\\n" .
        $timestamp . "\\n" . $nonce . "\\n" . $bodyStr;
    $signature = base64_encode(
        hash_hmac('sha256', $signingStr, $appSecret, true)
    );

    return [
        'X-App-Key'    => $appKey,
        'X-Timestamp'  => $timestamp,
        'X-Nonce'      => $nonce,
        'X-Signature'  => $signature,
        'Content-Type' => 'application/json',
    ];
}`}
                                </pre>
                              ),
                            },
                            {
                              key: 'hmac-cs',
                              label: 'C#',
                              children: (
                                <pre style={codeBlockStyle}>
{`using System.Security.Cryptography;
using System.Text;

public static Dictionary<string, string> SignRequest(
    string method, string path, string body,
    string appKey, string appSecret)
{
    var timestamp = DateTimeOffset.UtcNow.ToUnixTimeSeconds().ToString();
    var nonce = Guid.NewGuid().ToString();

    var signingStr = method + "\\n" + path + "\\n" +
        timestamp + "\\n" + nonce + "\\n" + body;

    using var hmac = new HMACSHA256(Encoding.UTF8.GetBytes(appSecret));
    var hash = hmac.ComputeHash(Encoding.UTF8.GetBytes(signingStr));
    var signature = Convert.ToBase64String(hash);

    return new Dictionary<string, string>
    {
        ["X-App-Key"]    = appKey,
        ["X-Timestamp"]  = timestamp,
        ["X-Nonce"]      = nonce,
        ["X-Signature"]  = signature,
        ["Content-Type"] = "application/json",
    };
}`}
                                </pre>
                              ),
                            },
                          ]}
                        />
                      </div>
                    )}

                    {/* Code examples for this endpoint */}
                    <Text strong style={{ fontSize: 13 }}>调用示例:</Text>
                    {codeSamples && (
                      <Tabs
                        size="small"
                        style={{ marginTop: 8 }}
                        items={[
                          {
                            key: 'curl',
                            label: 'cURL',
                            children: (
                              <div>
                                <div style={{ textAlign: 'right', marginBottom: 4 }}>
                                  <Button size="small" icon={<CopyOutlined />} onClick={() => handleCopyCode(codeSamples.curl)}>复制</Button>
                                </div>
                                <pre style={codeBlockStyle}>{codeSamples.curl}</pre>
                              </div>
                            ),
                          },
                          {
                            key: 'js',
                            label: 'JavaScript',
                            children: (
                              <div>
                                <div style={{ textAlign: 'right', marginBottom: 4 }}>
                                  <Button size="small" icon={<CopyOutlined />} onClick={() => handleCopyCode(codeSamples.js)}>复制</Button>
                                </div>
                                <pre style={codeBlockStyle}>{codeSamples.js}</pre>
                              </div>
                            ),
                          },
                          {
                            key: 'py',
                            label: 'Python',
                            children: (
                              <div>
                                <div style={{ textAlign: 'right', marginBottom: 4 }}>
                                  <Button size="small" icon={<CopyOutlined />} onClick={() => handleCopyCode(codeSamples.py)}>复制</Button>
                                </div>
                                <pre style={codeBlockStyle}>{codeSamples.py}</pre>
                              </div>
                            ),
                          },
                          {
                            key: 'go',
                            label: 'Go',
                            children: (
                              <div>
                                <div style={{ textAlign: 'right', marginBottom: 4 }}>
                                  <Button size="small" icon={<CopyOutlined />} onClick={() => handleCopyCode(codeSamples.go)}>复制</Button>
                                </div>
                                <pre style={codeBlockStyle}>{codeSamples.go}</pre>
                              </div>
                            ),
                          },
                          {
                            key: 'php',
                            label: 'PHP',
                            children: (
                              <div>
                                <div style={{ textAlign: 'right', marginBottom: 4 }}>
                                  <Button size="small" icon={<CopyOutlined />} onClick={() => handleCopyCode(codeSamples.php)}>复制</Button>
                                </div>
                                <pre style={codeBlockStyle}>{codeSamples.php}</pre>
                              </div>
                            ),
                          },
                          {
                            key: 'cs',
                            label: 'C#',
                            children: (
                              <div>
                                <div style={{ textAlign: 'right', marginBottom: 4 }}>
                                  <Button size="small" icon={<CopyOutlined />} onClick={() => handleCopyCode(codeSamples.cs)}>复制</Button>
                                </div>
                                <pre style={codeBlockStyle}>{codeSamples.cs}</pre>
                              </div>
                            ),
                          },
                        ]}
                      />
                    )}
                  </div>
                ),
              },
            ]}
          />
        ) : (
          <div style={{ height: 300, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
            <Empty description="请从左侧选择一个接口开始测试" />
          </div>
        )}
      </Card>
    </div>
  );
}

const codeBlockStyle: React.CSSProperties = {
  background: '#1e1e1e',
  color: '#d4d4d4',
  padding: 16,
  borderRadius: 8,
  maxHeight: 350,
  overflow: 'auto',
  fontSize: 12,
  lineHeight: 1.7,
  margin: 0,
  whiteSpace: 'pre-wrap',
  wordBreak: 'break-all',
};
