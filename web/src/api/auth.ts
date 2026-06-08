import client from './client';

export interface RegisterParams {
  username: string;
  password: string;
  member_type: string;
  company_name?: string;
  contact_name?: string;
  contact_phone?: string;
  email?: string;
  business_license?: string;
}

export const authApi = {
  register: (params: RegisterParams) => client.post('/auth/register', params),
  login: (username: string, password: string) => client.post('/admin/login', { username, password }),
  refresh: (refreshToken: string) => client.post('/admin/refresh', { refresh_token: refreshToken }),
};
