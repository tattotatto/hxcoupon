import { Navigate, useLocation } from 'react-router-dom';
import { Result, Button } from 'antd';
import { useAuthStore } from '../stores/authStore';

export function AuthGuard({ children }: { children: React.ReactNode }) {
  const isLoggedIn = useAuthStore((s) => s.isLoggedIn);
  const location = useLocation();

  if (!isLoggedIn) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }
  return <>{children}</>;
}

export function ApprovalGuard({ children }: { children: React.ReactNode }) {
  const user = useAuthStore((s) => s.user);
  const role = user?.role;
  const approvalStatus = user?.approval_status;

  // super_admin and admin bypass approval check
  if (role === 'super_admin' || role === 'admin') {
    return <>{children}</>;
  }

  if (approvalStatus !== 1) {
    return (
      <Result
        status="warning"
        title={
          approvalStatus === 0
            ? '账号待审批'
            : approvalStatus === 2
              ? '账号已被拒绝'
              : '账号已被停用'
        }
        subTitle={
          approvalStatus === 0
            ? '您的账号正在等待管理员审批，审批通过后方可使用此功能。'
            : '请联系管理员处理。'
        }
      />
    );
  }

  return <>{children}</>;
}

export function RoleGuard({ roles, children }: { roles: string[]; children: React.ReactNode }) {
  const user = useAuthStore((s) => s.user);

  if (!user || !roles.includes(user.role)) {
    return (
      <Result
        status="403"
        title="403"
        subTitle="抱歉，您没有权限访问此页面。"
        extra={
          <Button type="primary" onClick={() => window.history.back()}>
            返回
          </Button>
        }
      />
    );
  }

  return <>{children}</>;
}
