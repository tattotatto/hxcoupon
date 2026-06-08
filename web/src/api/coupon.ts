import client from './client';

export const couponApi = {
  issue: (data: Record<string, unknown>) => client.post('/admin/coupons/issue', data),
  listRecords: (params: Record<string, unknown>) => client.get('/admin/coupons/records', { params }),
  getRecord: (id: number) => client.get(`/admin/coupons/records/${id}`),
  consume: (data: Record<string, unknown>) => client.post('/admin/coupons/consume', data),
  listConsumeRecords: (params: Record<string, unknown>) => client.get('/admin/coupons/consume-records', { params }),
};
