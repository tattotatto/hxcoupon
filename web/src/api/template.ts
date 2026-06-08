import client from './client';

export const templateApi = {
  list: (params: Record<string, unknown>) => client.get('/admin/templates', { params }),
  get: (id: number) => client.get(`/admin/templates/${id}`),
  create: (data: Record<string, unknown>) => client.post('/admin/templates', data),
  update: (id: number, data: Record<string, unknown>) => client.put(`/admin/templates/${id}`, data),
  updateStatus: (id: number, status: number) => client.patch(`/admin/templates/${id}/status`, { status }),
  delete: (id: number) => client.delete(`/admin/templates/${id}`),
  browse: (params: Record<string, unknown>) => client.get('/admin/browse/templates', { params }),
  browseDetail: (id: number) => client.get(`/admin/browse/templates/${id}`),
};
