import { createBrowserRouter, Navigate } from 'react-router-dom';
import { AuthGuard, ApprovalGuard, RoleGuard } from './guards';
import AdminLayout from '../layouts/AdminLayout';
import Login from '../pages/Login';
import Register from '../pages/Register';
import Dashboard from '../pages/Dashboard';
import Profile from '../pages/Profile';
import UserList from '../pages/users/UserList';
import StoreList from '../pages/stores/StoreList';
import AppList from '../pages/apps/AppList';
import TemplateList from '../pages/templates/TemplateList';
import TemplateBrowse from '../pages/templates/TemplateBrowse';
import IssueForm from '../pages/coupons/IssueForm';
import CouponRecords from '../pages/coupons/CouponRecords';
import ConsumeForm from '../pages/coupons/ConsumeForm';
import ReportDashboard from '../pages/reports/ReportDashboard';
import ApiDocs from '../pages/ApiDocs';
import Landing from '../pages/Landing';

const router = createBrowserRouter([
  {
    path: '/login',
    element: <Login />,
  },
  {
    path: '/register',
    element: <Register />,
  },
  {
    path: '/',
    element: <Landing />,
  },
  {
    path: '/admin',
    element: (
      <AuthGuard>
        <AdminLayout />
      </AuthGuard>
    ),
    children: [
      { index: true, element: <Dashboard /> },
      { path: 'profile', element: <Profile /> },
      {
        path: 'users',
        element: (
          <RoleGuard roles={['super_admin']}>
            <UserList />
          </RoleGuard>
        ),
      },
      { path: 'stores', element: <StoreList /> },
      {
        path: 'apps',
        element: (
          <ApprovalGuard>
            <AppList />
          </ApprovalGuard>
        ),
      },
      { path: 'templates', element: <TemplateList /> },
      { path: 'browse', element: <TemplateBrowse /> },
      {
        path: 'coupons/issue',
        element: (
          <ApprovalGuard>
            <IssueForm />
          </ApprovalGuard>
        ),
      },
      {
        path: 'coupons/records',
        element: (
          <ApprovalGuard>
            <CouponRecords />
          </ApprovalGuard>
        ),
      },
      {
        path: 'coupons/consume',
        element: (
          <ApprovalGuard>
            <ConsumeForm />
          </ApprovalGuard>
        ),
      },
      {
        path: 'reports',
        element: (
          <ApprovalGuard>
            <ReportDashboard />
          </ApprovalGuard>
        ),
      },
    ],
  },
  {
    path: '/docs',
    element: <ApiDocs />,
  },
  { path: '*', element: <Navigate to="/" replace /> },
]);

export default router;
