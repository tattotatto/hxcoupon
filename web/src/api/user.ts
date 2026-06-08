import client from './client';

export const userApi = {
  list: (params: Record<string, unknown>) => client.get('/admin/users', { params }),
  get: (id: number) => client.get(`/admin/users/${id}`),
  approve: (id: number, reason?: string) => client.post(`/admin/users/${id}/approve`, { reason }),
  reject: (id: number, reason: string) => client.post(`/admin/users/${id}/reject`, { reason }),
  suspend: (id: number, reason: string) => client.post(`/admin/users/${id}/suspend`, { reason }),
  unsuspend: (id: number) => client.post(`/admin/users/${id}/unsuspend`),
  getProfile: () => client.get('/admin/profile'),
  updateProfile: (data: Record<string, unknown>) => client.put('/admin/profile', data),
};
