import { useEffect, useState } from 'react';
import { RouterProvider } from 'react-router-dom';
import { Spin } from 'antd';
import router from './router';
import { useAuthStore } from './stores/authStore';

export default function App() {
  const { hydrate, isLoggedIn, accessToken } = useAuthStore();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    hydrate();
    setReady(true);
  }, []);

  // Re-hydrate when localStorage changes from other tabs (refresh interceptor)
  useEffect(() => {
    const handler = () => hydrate();
    window.addEventListener('auth-storage', handler);
    return () => window.removeEventListener('auth-storage', handler);
  }, [hydrate]);

  if (!ready) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    );
  }

  return <RouterProvider router={router} />;
}
