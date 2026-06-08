import { Modal, Typography, Input, Space, Button, Alert } from 'antd';
import { CopyOutlined } from '@ant-design/icons';

const { Text, Paragraph } = Typography;

interface Props {
  open: boolean;
  credentials: { app_key: string; app_secret: string } | null;
  onClose: () => void;
}

export default function CredentialsModal({ open, credentials, onClose }: Props) {
  if (!credentials) return null;

  const copyKey = () => navigator.clipboard.writeText(credentials.app_key);
  const copySecret = () => navigator.clipboard.writeText(credentials.app_secret);

  return (
    <Modal
      open={open}
      title="API 密钥"
      onCancel={onClose}
      footer={[
        <Button key="close" type="primary" onClick={onClose}>关闭</Button>,
      ]}
    >
      <Alert
        type="warning"
        message="请立即保存密钥！关闭后将无法再次查看 App Secret。"
        style={{ marginBottom: 16 }}
        showIcon
      />
      <div style={{ marginBottom: 16 }}>
        <Text strong>App Key</Text>
        <Input.Group compact style={{ marginTop: 4 }}>
          <Input value={credentials.app_key} readOnly style={{ width: 'calc(100% - 40px)' }} />
          <Button icon={<CopyOutlined />} onClick={copyKey} />
        </Input.Group>
      </div>
      <div>
        <Text strong>App Secret</Text>
        <Input.Group compact style={{ marginTop: 4 }}>
          <Input.Password value={credentials.app_secret} readOnly style={{ width: 'calc(100% - 40px)' }} />
          <Button icon={<CopyOutlined />} onClick={copySecret} />
        </Input.Group>
      </div>
    </Modal>
  );
}
