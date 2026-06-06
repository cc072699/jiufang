import { useEffect, useState } from 'react';
import { useNavigate, useParams, useLocation } from 'react-router-dom';
import { Form, Input, Select, Typography, Button, Space, message, Card } from 'antd';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { createUser, updateUser, getUserList, getGroupList } from '../shared/api';
import type { CreateUserRequest, UpdateUserRequest } from '../shared/api/types';
import { ROLE_SELECT_OPTIONS, USER_STATUS_SELECT_OPTIONS } from '../shared/constants';

const { Title } = Typography;

export default function UserFormPage() {
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  // Try to get user data from navigation state first
  const navState = location.state as { record?: Record<string, unknown> } | null;

  // Fallback: fetch from list and find by id
  const { data } = useQuery({
    queryKey: ['users'],
    queryFn: () => getUserList({ page: 1, page_size: 100 }),
    enabled: isEdit && !navState?.record,
  });

  // Fetch groups for multi-select
  const { data: groupData } = useQuery({
    queryKey: ['groups-options'],
    queryFn: () => getGroupList({ page: 1, page_size: 100 }),
  });
  const groupOptions = (groupData?.groups ?? []).map((g) => ({ label: g.name, value: g.id }));

  useEffect(() => {
    if (!isEdit) return;
    const record = navState?.record || data?.users?.find((u) => u.id === id);
    if (record) {
      // Delay to ensure Select fields are registered with the form
      requestAnimationFrame(() => {
        form.setFieldsValue({
          username: record.username,
          email: record.email,
          role: record.role,
          status: record.status,
          groups: (record as Record<string, unknown>).groups as string[] ?? [],
        });
      });
    }
  }, [isEdit, navState, data, id, form]);

  const handleFinish = async (values: CreateUserRequest & { status?: number }) => {
    setLoading(true);
    try {
      if (isEdit) {
        const payload: UpdateUserRequest = {
          email: values.email,
          role: values.role,
          status: values.status,
          groups: values.groups,
        };
        // 仅在填写了新密码时传递 password
        if (values.password) {
          payload.password = values.password;
        }
        await updateUser(id!, payload);
        message.success('更新成功');
        queryClient.invalidateQueries({ queryKey: ['users'] });
        navigate('/users');
      } else {
        const payload: CreateUserRequest = {
          username: values.username,
          password: values.password,
          email: values.email,
          role: values.role,
          groups: values.groups,
        };
        await createUser(payload);
        message.success('创建成功');
        queryClient.invalidateQueries({ queryKey: ['users'] });
        navigate('/users');
      }
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : '操作失败';
      message.error(msg);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div>
      <Title level={4}>{isEdit ? '编辑人员' : '新增人员'}</Title>
      <Card style={{ maxWidth: 560 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleFinish}
          initialValues={{ role: 'executive', status: 1 }}
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, max: 50, message: '3-50个字符' },
            ]}
          >
            <Input disabled={isEdit} placeholder="3-50个字符" />
          </Form.Item>

          <Form.Item
            name="email"
            label="邮箱"
            rules={[
              { required: true, message: '请输入邮箱' },
              { type: 'email', message: '邮箱格式不正确' },
              { max: 100, message: '邮箱长度不超过 100 个字符' },
            ]}
          >
            <Input placeholder="user@company.com" />
          </Form.Item>

          <Form.Item
            name="password"
            label={isEdit ? '新密码（留空则不修改）' : '密码'}
            rules={
              isEdit
                ? [{ min: 8, max: 100, message: '密码长度 8~100 个字符' }]
                : [
                    { required: true, message: '请输入密码' },
                    { min: 8, max: 100, message: '密码长度 8~100 个字符' },
                  ]
            }
          >
            <Input.Password placeholder="8~100个字符" />
          </Form.Item>

          <Form.Item name="role" label="角色" rules={[{ required: true }]}>
            <Select
              options={ROLE_SELECT_OPTIONS}
            />
          </Form.Item>

          <Form.Item name="groups" label="用户组">
            <Select
              mode="multiple"
              placeholder="选择用户组（可选）"
              options={groupOptions}
              allowClear
            />
          </Form.Item>

          {isEdit && (
            <Form.Item name="status" label="状态">
              <Select
                options={USER_STATUS_SELECT_OPTIONS}
              />
            </Form.Item>
          )}

          {/* TODO: 企业微信推送地址字段 - 接口未定义 */}
          {/* TODO: 用户组选择 - 接口未定义成员管理 */}

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {isEdit ? '保存' : '创建'}
              </Button>
              <Button onClick={() => navigate('/users')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
