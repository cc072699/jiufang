import { useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { Card, Typography, Descriptions, Button, Divider, Modal, Form, Input, Upload, message, Avatar, Spin } from 'antd';
import { LogoutOutlined, LockOutlined, UserOutlined, UploadOutlined } from '@ant-design/icons';
import { useQueryClient, useQuery } from '@tanstack/react-query';
import { useAuthStore } from '../shared/auth/store';
import { logout as logoutApi, changePassword, uploadAvatar, getProfile } from '../shared/api';
import type { UploadProps } from 'antd';

const { Title, Text } = Typography;

export default function ProfilePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { user, clearAuth } = useAuthStore();
  const { data: profile, isLoading: profileLoading } = useQuery({
    queryKey: ['profile'],
    queryFn: getProfile,
  });
  const [pwdModalOpen, setPwdModalOpen] = useState(false);
  const [pwdLoading, setPwdLoading] = useState(false);
  const [pwdForm] = Form.useForm();
  const [avatarUrl, setAvatarUrl] = useState('');
  const avatarUrlRef = useRef<string>('');

  const revokeAvatarUrl = () => {
    if (avatarUrlRef.current) {
      URL.revokeObjectURL(avatarUrlRef.current);
      avatarUrlRef.current = '';
    }
  };

  const handleLogout = () => {
    Modal.confirm({
      title: '确认退出',
      content: '确定要退出登录吗？',
      okText: '确定',
      cancelText: '取消',
      onOk: async () => {
        try {
          await logoutApi();
        } catch {
          // 无论成功失败都清理本地态
        }
        queryClient.clear();
        clearAuth();
        navigate('/login', { replace: true });
      },
    });
  };

  const handleChangePassword = async () => {
    try {
      const values = await pwdForm.validateFields();
      setPwdLoading(true);
      await changePassword({
        old_password: values.oldPassword,
        new_password: values.newPassword,
        confirm_password: values.confirmPassword,
      });
      message.success('密码修改成功，请重新登录');
      setPwdModalOpen(false);
      pwdForm.resetFields();
      clearAuth();
      navigate('/login', { replace: true });
    } catch (err: unknown) {
      if (err && typeof err === 'object' && 'errorFields' in err) {
        // Form validation error, do nothing (AntD shows messages)
        return;
      }
      const msg = err instanceof Error ? err.message : '密码修改失败，请稍后重试';
      message.error(msg);
    } finally {
      setPwdLoading(false);
    }
  };

  const handleUpload: UploadProps['customRequest'] = async (options) => {
    const { file, onSuccess, onError } = options;
    try {
      revokeAvatarUrl();
      const localUrl = URL.createObjectURL(file as File);
      avatarUrlRef.current = localUrl;
      setAvatarUrl(localUrl);
      await uploadAvatar(file as File);
      onSuccess?.(null);
      message.success('头像上传成功');
    } catch {
      onError?.(new Error('上传失败'));
      message.error('头像上传失败，请稍后重试');
    }
  };

  const roleLabels: Record<string, string> = {
    admin: '管理员',
    manager: '部门经理',
    executive: '高管',
  };

  return (
    <div>
      <Title level={4}>个人中心</Title>
      <Spin spinning={profileLoading}>
      <Card style={{ maxWidth: 560 }}>
        <div style={{ textAlign: 'center', marginBottom: 24 }}>
          <Upload
            showUploadList={false}
            customRequest={handleUpload}
            accept=".jpg,.jpeg,.png,.gif"
            beforeUpload={(file) => {
              const isImage = ['image/jpeg', 'image/png', 'image/gif'].includes(file.type);
              if (!isImage) {
                message.error('图片格式仅支持JPG/PNG/GIF');
                return Upload.LIST_IGNORE;
              }
              const isLt2M = file.size / 1024 / 1024 < 2;
              if (!isLt2M) {
                message.error('图片大小不超过2MB');
                return Upload.LIST_IGNORE;
              }
              return true;
            }}
          >
            {avatarUrl ? (
              <Avatar src={avatarUrl} size={72} style={{ cursor: 'pointer' }} />
            ) : (
              <div style={{ cursor: 'pointer', position: 'relative', display: 'inline-block' }}>
                <UserOutlined style={{ fontSize: 64, color: '#1677ff' }} />
                <div style={{
                  position: 'absolute', bottom: -4, right: -8,
                  background: '#1677ff', borderRadius: '50%',
                  width: 24, height: 24, display: 'flex', alignItems: 'center', justifyContent: 'center',
                }}>
                  <UploadOutlined style={{ color: '#fff', fontSize: 12 }} />
                </div>
              </div>
            )}
          </Upload>
          <Title level={4} style={{ marginTop: 8, marginBottom: 0 }}>{user?.username}</Title>
          <Text type="secondary">{roleLabels[user?.role || ''] || user?.role}</Text>
        </div>

        <Descriptions column={1} bordered size="small">
          <Descriptions.Item label="用户名">{user?.username}</Descriptions.Item>
          <Descriptions.Item label="角色">{roleLabels[user?.role || ''] || user?.role}</Descriptions.Item>
          <Descriptions.Item label="邮箱">{profile?.email || user?.email || '-'}</Descriptions.Item>
          <Descriptions.Item label="用户组">{user?.groups?.join(', ') || '-'}</Descriptions.Item>
        </Descriptions>

        <Divider />

        <Button
          block
          icon={<LockOutlined />}
          style={{ marginBottom: 12 }}
          onClick={() => setPwdModalOpen(true)}
        >
          修改密码
        </Button>

        <Button block danger icon={<LogoutOutlined />} onClick={handleLogout}>
          退出登录
        </Button>
      </Card>
      </Spin>

      {/* Change password modal */}
      <Modal
        title="修改密码"
        open={pwdModalOpen}
        onOk={handleChangePassword}
        onCancel={() => { setPwdModalOpen(false); pwdForm.resetFields(); }}
        confirmLoading={pwdLoading}
        okText="确认"
        cancelText="取消"
        destroyOnClose
      >
        <Form
          form={pwdForm}
          layout="vertical"
          style={{ marginTop: 16 }}
        >
          <Form.Item
            name="oldPassword"
            label="当前密码"
            rules={[{ required: true, message: '请输入当前密码' }]}
          >
            <Input.Password placeholder="请输入当前密码" />
          </Form.Item>
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码长度须在6~20位之间' },
              { max: 20, message: '密码长度须在6~20位之间' },
            ]}
          >
            <Input.Password placeholder="6-20位，无字符复杂度限制" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认新密码"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请再次输入新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'));
                },
              }),
            ]}
          >
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
}