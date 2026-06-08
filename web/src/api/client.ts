import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { message } from 'antd';

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
});

let isRefreshing = false;
let pendingQueue: Array<{
  resolve: (token: string) => void;
  reject: (err: unknown) => void;
}> = [];

function getAccessToken(): string | null {
  try {
    const raw = localStorage.getItem('auth-storage');
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    return parsed?.state?.accessToken ?? null;
  } catch {
    return null;
  }
}

function getRefreshToken(): string | null {
  try {
    const raw = localStorage.getItem('auth-storage');
    if (!raw) return null;
    const parsed = JSON.parse(raw);
    return parsed?.state?.refreshToken ?? null;
  } catch {
    return null;
  }
}

function setAccessToken(token: string) {
  try {
    const raw = localStorage.getItem('auth-storage');
    if (!raw) return;
    const parsed = JSON.parse(raw);
    parsed.state.accessToken = token;
    localStorage.setItem('auth-storage', JSON.stringify(parsed));
    window.dispatchEvent(new Event('auth-storage'));
  } catch {
    // ignore
  }
}

function clearAuth() {
  localStorage.removeItem('auth-storage');
  window.dispatchEvent(new Event('auth-storage'));
}

async function doRefresh(): Promise<string> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) throw new Error('no refresh token');
  const res = await axios.post('/api/v1/admin/refresh', { refresh_token: refreshToken });
  const newToken = res.data.data.access_token;
  setAccessToken(newToken);
  return newToken;
}

client.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const token = getAccessToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

const SKIP_REFRESH_PATHS = ['/admin/login', '/admin/refresh'];

client.interceptors.response.use(
  (res) => res,
  async (error: AxiosError<{ code?: number; message?: string }>) => {
    const status = error.response?.status;
    const msg = error.response?.data?.message || error.message || '网络请求失败';
    const isAuthPath = error.config?.url ? SKIP_REFRESH_PATHS.some((p) => error.config!.url!.includes(p)) : false;

    if (status === 401 && error.config && !error.config.headers['x-retry'] && !isAuthPath) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          pendingQueue.push({
            resolve: (token: string) => {
              error.config!.headers.Authorization = `Bearer ${token}`;
              resolve(client(error.config!));
            },
            reject,
          });
        });
      }

      isRefreshing = true;
      try {
        const newToken = await doRefresh();
        pendingQueue.forEach(({ resolve }) => resolve(newToken));
        pendingQueue = [];
        error.config.headers.Authorization = `Bearer ${newToken}`;
        error.config.headers['x-retry'] = '1';
        return client(error.config);
      } catch (refreshErr) {
        pendingQueue.forEach(({ reject }) => reject(refreshErr));
        pendingQueue = [];
        clearAuth();
        window.location.href = '/login';
        return Promise.reject(refreshErr);
      } finally {
        isRefreshing = false;
      }
    }

    // Always show error for non-401, and for 401 on auth paths
    if (status !== 401 || isAuthPath) {
      message.error(msg);
    }

    return Promise.reject(error);
  },
);

export default client;
