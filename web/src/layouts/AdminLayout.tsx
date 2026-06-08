import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { ProLayout, PageContainer } from '@ant-design/pro-components';
import {
  DashboardOutlined,
  ShopOutlined,
  AppstoreOutlined,
  GiftOutlined,
  ProjectOutlined,
  BarChartOutlined,
  TeamOutlined,
  UserOutlined,
  LogoutOutlined,
  EyeOutlined,
  SendOutlined,
  FileTextOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons';
import { Dropdown, Avatar, Space } from 'antd';
import { useAuthStore } from '../stores/authStore';

export default function AdminLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const [collapsed, setCollapsed] = useState(false);

  const role = user?.role;
  const memberType = user?.member_type;
  const isSuperAdmin = role === 'super_admin';
  const isAdmin = role === 'admin';
  const isMember = role === 'member';
  const isIssuer = memberType === 'issuer' || memberType === 'both';
  const isConsumer = memberType === 'consumer' || memberType === 'both';
  const canManage = isSuperAdmin || isAdmin;

  const menuItems: Record<string, { path?: string; icon?: React.ReactNode; children?: Record<string, { path: string; icon: React.ReactNode }> }> = {
    dashboard: { path: '/', icon: <DashboardOutlined /> },
  };

  if (canManage || isIssuer) {
    menuItems['stores'] = { path: '/stores', icon: <ShopOutlined /> };
    menuItems['templates'] = { path: '/templates', icon: <ProjectOutlined /> };
    menuItems['browse'] = { path: '/browse', icon: <EyeOutlined /> };
  }

  menuItems['coupons'] = {
    icon: <GiftOutlined />,
    children: {},
  };
  if (isIssuer || canManage) {
    menuItems['coupons'].children!['/coupons/issue'] = { path: '/coupons/issue', icon: <SendOutlined /> };
    menuItems['coupons'].children!['/coupons/records'] = { path: '/coupons/records', icon: <FileTextOutlined /> };
  }
  if (isConsumer || canManage) {
    menuItems['coupons'].children!['/coupons/consume'] = { path: '/coupons/consume', icon: <CheckCircleOutlined /> };
  }

  if (isMember) {
    menuItems['apps'] = { path: '/apps', icon: <AppstoreOutlined /> };
  }

  menuItems['reports'] = { path: '/reports', icon: <BarChartOutlined /> };

  if (isSuperAdmin) {
    menuItems['users'] = { path: '/users', icon: <TeamOutlined /> };
  }

  const routeMap: Record<string, { name: string; icon?: React.ReactNode }> = {
    '/': { name: '仪表盘', icon: <DashboardOutlined /> },
    '/profile': { name: '个人设置' },
    '/stores': { name: '门店管理', icon: <ShopOutlined /> },
    '/apps': { name: '我的应用', icon: <AppstoreOutlined /> },
    '/templates': { name: '模板管理', icon: <ProjectOutlined /> },
    '/browse': { name: '模板浏览', icon: <EyeOutlined /> },
    '/coupons/issue': { name: '发券', icon: <SendOutlined /> },
    '/coupons/records': { name: '发券记录', icon: <FileTextOutlined /> },
    '/coupons/consume': { name: '核销', icon: <CheckCircleOutlined /> },
    '/reports': { name: '数据报表', icon: <BarChartOutlined /> },
    '/users': { name: '用户管理', icon: <TeamOutlined /> },
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人设置',
      onClick: () => navigate('/profile'),
    },
    { type: 'divider' as const },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => {
        logout();
        navigate('/login');
      },
    },
  ];

  const currentRoute = routeMap[location.pathname];

  return (
    <ProLayout
      title="优惠券管理"
      logo={null}
      collapsed={collapsed}
      onCollapse={setCollapsed}
      location={location}
      route={{ routes: Object.entries(routeMap).map(([path, { name, icon }]) => ({ path, name, icon })) }}
      menuItemRender={(item, dom) => (
        <div onClick={() => item.path && navigate(item.path)} style={{ cursor: 'pointer' }}>
          {dom}
        </div>
      )}
      headerContentRender={() => (
        <div style={{ fontSize: 14, color: '#666' }}>
          {currentRoute?.name || ''}
        </div>
      )}
      avatarProps={{
        icon: <UserOutlined />,
        render: (_: unknown, dom: React.ReactNode) => (
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <Space style={{ cursor: 'pointer' }}>
              <Avatar size="small" icon={<UserOutlined />} />
              <span>{user?.username || '用户'}</span>
              {dom}
            </Space>
          </Dropdown>
        ),
      }}
      menuDataRender={() => {
        const items: Array<{ path?: string; name: string; icon: React.ReactNode; children?: Array<{ path: string; name: string; icon: React.ReactNode }> }> = [
          { path: '/', name: '仪表盘', icon: <DashboardOutlined /> },
        ];

        if (canManage || isIssuer) {
          items.push({ path: '/stores', name: '门店管理', icon: <ShopOutlined /> });
          items.push({ path: '/templates', name: '模板管理', icon: <ProjectOutlined /> });
          items.push({ path: '/browse', name: '模板浏览', icon: <EyeOutlined /> });
        }

        const couponChildren: Array<{ path: string; name: string; icon: React.ReactNode }> = [];
        if (isIssuer || canManage) {
          couponChildren.push({ path: '/coupons/issue', name: '发券', icon: <SendOutlined /> });
          couponChildren.push({ path: '/coupons/records', name: '发券记录', icon: <FileTextOutlined /> });
        }
        if (isConsumer || canManage) {
          couponChildren.push({ path: '/coupons/consume', name: '核销', icon: <CheckCircleOutlined /> });
        }
        items.push({ name: '优惠券', icon: <GiftOutlined />, children: couponChildren });

        if (isMember) {
          items.push({ path: '/apps', name: '我的应用', icon: <AppstoreOutlined /> });
        }

        items.push({ path: '/reports', name: '数据报表', icon: <BarChartOutlined /> });

        if (isSuperAdmin) {
          items.push({ path: '/users', name: '用户管理', icon: <TeamOutlined /> });
        }

        return items;
      }}
      fixSiderbar
      layout="mix"
      splitMenus={false}
      contentStyle={{ margin: 0 }}
    >
      <PageContainer
        header={{ title: null, breadcrumb: {} }}
        style={{ minHeight: '100%' }}
      >
        <Outlet />
      </PageContainer>
    </ProLayout>
  );
}
