import { create } from 'zustand';
import client from '../api/client';

export interface UserInfo {
  id: number;
  username: string;
  role: string;
  member_type: string | null;
  approval_status: number;
  company_name: string | null;
  contact_name: string | null;
  contact_phone: string | null;
  email: string | null;
  reject_reason?: string | null;
}

interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  expiresIn: number | null;
  user: UserInfo | null;
  isLoggedIn: boolean;

  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  fetchProfile: () => Promise<void>;
  setAccessToken: (token: string) => void;
  hydrate: () => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  accessToken: null,
  refreshToken: null,
  expiresIn: null,
  user: null,
  isLoggedIn: false,

  setAccessToken: (token: string) => {
    set({ accessToken: token });
  },

  login: async (username: string, password: string) => {
    const res = await client.post('/admin/login', { username, password });
    const { access_token, refresh_token, expires_in } = res.data.data;
    set({
      accessToken: access_token,
      refreshToken: refresh_token,
      expiresIn: expires_in,
      isLoggedIn: true,
    });
    try {
      const profileRes = await client.get('/admin/profile');
      set({ user: profileRes.data.data });
    } catch {
      // profile fetch can fail, user still logged in
    }
  },

  logout: () => {
    set({
      accessToken: null,
      refreshToken: null,
      expiresIn: null,
      user: null,
      isLoggedIn: false,
    });
  },

  fetchProfile: async () => {
    const res = await client.get('/admin/profile');
    set({ user: res.data.data, isLoggedIn: true });
  },

  hydrate: () => {
    try {
      const raw = localStorage.getItem('auth-storage');
      if (!raw) return;
      const parsed = JSON.parse(raw);
      const state = parsed?.state;
      if (state?.accessToken && state?.refreshToken) {
        set({
          accessToken: state.accessToken,
          refreshToken: state.refreshToken,
          expiresIn: state.expiresIn,
          user: state.user,
          isLoggedIn: true,
        });
      }
    } catch {
      // ignore
    }
  },
}));

// Persist auth state to localStorage
useAuthStore.subscribe((state) => {
  localStorage.setItem(
    'auth-storage',
    JSON.stringify({
      state: {
        accessToken: state.accessToken,
        refreshToken: state.refreshToken,
        expiresIn: state.expiresIn,
        user: state.user,
        isLoggedIn: state.isLoggedIn,
      },
    }),
  );
});
