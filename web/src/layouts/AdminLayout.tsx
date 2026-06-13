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
import { Dropdown, Avatar, Space, Result, Button, Typography } from 'antd';
import { useAuthStore } from '../stores/authStore';

const { Text } = Typography;

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
    dashboard: { path: '/admin', icon: <DashboardOutlined /> },
  };

  if (canManage || isIssuer) {
    menuItems['stores'] = { path: '/admin/stores', icon: <ShopOutlined /> };
    menuItems['templates'] = { path: '/admin/templates', icon: <ProjectOutlined /> };
    menuItems['browse'] = { path: '/admin/browse', icon: <EyeOutlined /> };
  }

  menuItems['coupons'] = {
    icon: <GiftOutlined />,
    children: {},
  };
  if (isIssuer || canManage) {
    menuItems['coupons'].children!['/admin/coupons/issue'] = { path: '/admin/coupons/issue', icon: <SendOutlined /> };
    menuItems['coupons'].children!['/admin/coupons/records'] = { path: '/admin/coupons/records', icon: <FileTextOutlined /> };
  }
  if (isConsumer || canManage) {
    menuItems['coupons'].children!['/admin/coupons/consume'] = { path: '/admin/coupons/consume', icon: <CheckCircleOutlined /> };
  }

  if (isMember) {
    menuItems['apps'] = { path: '/admin/apps', icon: <AppstoreOutlined /> };
  }

  menuItems['reports'] = { path: '/admin/reports', icon: <BarChartOutlined /> };

  if (isSuperAdmin) {
    menuItems['users'] = { path: '/admin/users', icon: <TeamOutlined /> };
  }

  const routeMap: Record<string, { name: string; icon?: React.ReactNode }> = {
    '/admin': { name: '仪表盘', icon: <DashboardOutlined /> },
    '/admin/profile': { name: '个人设置' },
    '/admin/stores': { name: '门店管理', icon: <ShopOutlined /> },
    '/admin/apps': { name: '我的应用', icon: <AppstoreOutlined /> },
    '/admin/templates': { name: '模板管理', icon: <ProjectOutlined /> },
    '/admin/browse': { name: '模板浏览', icon: <EyeOutlined /> },
    '/admin/coupons/issue': { name: '发券', icon: <SendOutlined /> },
    '/admin/coupons/records': { name: '发券记录', icon: <FileTextOutlined /> },
    '/admin/coupons/consume': { name: '核销', icon: <CheckCircleOutlined /> },
    '/admin/reports': { name: '数据报表', icon: <BarChartOutlined /> },
    '/admin/users': { name: '用户管理', icon: <TeamOutlined /> },
  };

  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人设置',
      onClick: () => navigate('/admin/profile'),
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
  const isPending = user?.approval_status === 0;

  // Show pending approval screen for unapproved members
  if (isPending && !canManage) {
    return (
      <div style={{ minHeight: '100vh', display: 'flex', justifyContent: 'center', alignItems: 'center', background: '#f5f5f5' }}>
        <Result
          status="info"
          title="账号待审批"
          subTitle={
            <div>
              <Text>你的账号 <strong>{user?.username}</strong> 已提交，正在等待管理员审批。</Text>
              <br />
              <Text type="secondary">审批通过后即可使用全部功能。你可以先退出等待通知。</Text>
            </div>
          }
          extra={
            <Button type="primary" onClick={() => { logout(); navigate('/login'); }}>
              退出登录
            </Button>
          }
        />
      </div>
    );
  }

  return (
    <ProLayout
      title="宏曦优惠券管理平台"
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
          { path: '/admin', name: '仪表盘', icon: <DashboardOutlined /> },
        ];

        if (canManage || isIssuer) {
          items.push({ path: '/admin/stores', name: '门店管理', icon: <ShopOutlined /> });
        }
        if (canManage) {
          items.push({ path: '/admin/templates', name: '模板管理', icon: <ProjectOutlined /> });
        }
        if (canManage || isIssuer) {
          items.push({ path: '/admin/browse', name: '模板浏览', icon: <EyeOutlined /> });
        }

        const couponChildren: Array<{ path: string; name: string; icon: React.ReactNode }> = [];
        if (isIssuer || canManage) {
          couponChildren.push({ path: '/admin/coupons/issue', name: '发券', icon: <SendOutlined /> });
          couponChildren.push({ path: '/admin/coupons/records', name: '发券记录', icon: <FileTextOutlined /> });
        }
        if (isConsumer || canManage) {
          couponChildren.push({ path: '/admin/coupons/consume', name: '核销', icon: <CheckCircleOutlined /> });
        }
        items.push({ name: '优惠券', icon: <GiftOutlined />, children: couponChildren });

        if (isMember) {
          items.push({ path: '/admin/apps', name: '我的应用', icon: <AppstoreOutlined /> });
        }

        items.push({ path: '/admin/reports', name: '数据报表', icon: <BarChartOutlined /> });

        if (isSuperAdmin) {
          items.push({ path: '/admin/users', name: '用户管理', icon: <TeamOutlined /> });
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
