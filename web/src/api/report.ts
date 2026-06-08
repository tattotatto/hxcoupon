import client from './client';

export const reportApi = {
  overview: () => client.get('/admin/reports/overview'),
  trend: (startDate: string, endDate: string) => client.get('/admin/reports/trend', { params: { start_date: startDate, end_date: endDate } }),
  exportCoupons: () => client.get('/admin/reports/export/coupons', { responseType: 'blob' }),
  exportUsage: () => client.get('/admin/reports/export/usage', { responseType: 'blob' }),
  statsOverview: () => client.get('/admin/statistics/overview'),
  statsTrend: (startDate: string, endDate: string) => client.get('/admin/statistics/trend', { params: { start_date: startDate, end_date: endDate } }),
};
