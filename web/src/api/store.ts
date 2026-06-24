import client from './client';

export const storeApi = {
  list: (params: Record<string, unknown>) => client.get('/admin/stores', { params }),
  options: () => client.get('/admin/stores/options'),
  get: (id: number) => client.get(`/admin/stores/${id}`),
  create: (data: Record<string, unknown>) => client.post('/admin/stores', data),
  update: (id: number, data: Record<string, unknown>) => client.put(`/admin/stores/${id}`, data),
  updateStatus: (id: number, status: number) => client.patch(`/admin/stores/${id}/status`, { status }),
  delete: (id: number) => client.delete(`/admin/stores/${id}`),
  generateCredentials: (id: number) => client.post(`/admin/stores/${id}/credentials`),
  uploadQrCode: (id: number, file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return client.post(`/admin/stores/${id}/qrcode`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },

  // Apps (member's own stores)
  listApps: (params: Record<string, unknown>) => client.get('/admin/apps', { params }),
  getApp: (id: number) => client.get(`/admin/apps/${id}`),
  createApp: (data: Record<string, unknown>) => client.post('/admin/apps', data),
  updateApp: (id: number, data: Record<string, unknown>) => client.put(`/admin/apps/${id}`, data),
  deleteApp: (id: number) => client.delete(`/admin/apps/${id}`),
  generateAppCredentials: (id: number) => client.post(`/admin/apps/${id}/credentials`),
  uploadAppQrCode: (id: number, file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return client.post(`/admin/apps/${id}/qrcode`, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    });
  },
};
