import { useEffect, useState } from 'react';
import { useNavigate, useParams, useLocation } from 'react-router-dom';
import { Form, Input, Select, Typography, Button, Space, message, Card, Checkbox, Table, Tag } from 'antd';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { createAlert, updateAlert, getAlertList, getUserList, getGroupList } from '../shared/api';
import type { CreateAlertRequest, UpdateAlertRequest, AlertRecord, UserRecord } from '../shared/api/types';
import { ACTIVE_STATUS_SELECT_OPTIONS, ROLE_LABELS, PUSH_CHANNEL_OPTIONS } from '../shared/constants';

const { Title, Text } = Typography;
const { TextArea } = Input;

export default function AlertFormPage() {
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const navState = location.state as { record?: AlertRecord } | null;

  const { data } = useQuery({
    queryKey: ['alerts'],
    queryFn: () => getAlertList({ page: 1, page_size: 100 }),
    enabled: isEdit && !navState?.record,
  });

  const { data: userData } = useQuery({
    queryKey: ['users-options'],
    queryFn: () => getUserList({ page: 1, page_size: 100 }),
  });

  const { data: groupData } = useQuery({
    queryKey: ['groups-options'],
    queryFn: () => getGroupList({ page: 1, page_size: 100 }),
  });

  const userOptions = (userData?.users ?? []).map((u) => ({
    label: u.username,
    value: u.id,
  }));

  const groupNameMap: Record<string, string> = Object.fromEntries(
    (groupData?.groups ?? []).map((g) => [g.id, g.name]),
  );

  useEffect(() => {
    if (!isEdit) return;
    const record = navState?.record || data?.alerts?.find((a) => a.id === id);
    if (record) {
      form.setFieldsValue({
        name: record.name,
        condition: record.condition,
        status: record.status,
        recipients: record.recipients,
        push_channel: record.push_channel ? [record.push_channel] : ['email'],
      });
    }
  }, [isEdit, navState, data, id, form]);

  const handleFinish = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      const recipients = (values.recipients as string[]) || [];
      const channels = (values.push_channel as string[]) || ['email'];
      if (isEdit) {
        const payload: UpdateAlertRequest = {
          name: values.name as string,
          sql: values.sql as string,
          condition: values.condition as string,
          recipients,
          push_channel: channels[0] as 'wechat' | 'email',
          description: values.description as string,
          status: values.status as 'active' | 'inactive',
        };
        await updateAlert(id!, payload);
        message.success('更新成功');
      } else {
        const payload: CreateAlertRequest = {
          name: values.name as string,
          sql: values.sql as string,
          condition: values.condition as string,
          recipients,
          push_channel: channels[0] as 'wechat' | 'email',
          description: values.description as string,
        };
        await createAlert(payload);
        message.success('创建成功');
      }
      queryClient.invalidateQueries({ queryKey: ['alerts'] });
      navigate('/alerts');
    } catch (err: unknown) {
      message.error(err instanceof Error ? err.message : '操作失败');
    } finally {
      setLoading(false);
    }
  };

  const selectedRecipients = Form.useWatch('recipients', form) as string[] | undefined;
  const selectedUsers: UserRecord[] = (userData?.users ?? []).filter((u) =>
    (selectedRecipients ?? []).includes(u.id),
  );

  const selectedUserColumns = [
    { title: '用户名', dataIndex: 'username', key: 'username' },
    {
      title: '角色',
      dataIndex: 'role',
      key: 'role',
      render: (role: string) => ROLE_LABELS[role] || role,
    },
    {
      title: '所属用户组',
      dataIndex: 'groups',
      key: 'groups',
      render: (groups?: string[]) =>
        groups && groups.length > 0
          ? groups.map((gid) => <Tag key={gid}>{groupNameMap[gid] || gid}</Tag>)
          : '-',
    },
  ];

  return (
    <div>
      <Title level={4}>{isEdit ? '编辑预警规则' : '新建预警规则'}</Title>
      <Card style={{ maxWidth: 720 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleFinish}
          initialValues={{ status: 'active', push_channel: ['email'] }}
        >
          <Form.Item
            name="name"
            label="规则名称"
            rules={[
              { required: true, message: '请输入名称' },
              { max: 100, message: '名称不超过 100 个字符' },
            ]}
          >
            <Input placeholder="例：库存低于安全线" />
          </Form.Item>
          <Form.Item
            name="sql"
            label="监测 SQL"
            rules={[
              { required: !isEdit, message: '请输入SQL' },
              { max: 5000, message: 'SQL 长度不超过 5000 个字符' },
            ]}
          >
            <TextArea rows={3} placeholder="SELECT ..." />
          </Form.Item>
          <Form.Item
            name="condition"
            label="触发条件"
            rules={[
              { required: true, message: '请输入触发条件' },
              { max: 200, message: '条件不超过 200 个字符' },
            ]}
          >
            <Input placeholder="例：stock < safety_stock" />
          </Form.Item>
          <Form.Item name="recipients" label="接收人" rules={[{ required: !isEdit, message: '请选择接收人' }]}>
            <Select
              mode="multiple"
              placeholder="搜索并选择接收人"
              options={userOptions}
              showSearch
              filterOption={(input, option) =>
                (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
              }
            />
          </Form.Item>
          {selectedUsers.length > 0 && (
            <div style={{ marginBottom: 24 }}>
              <Text strong style={{ marginBottom: 8, display: 'block' }}>已选接收人</Text>
              <Table
                dataSource={selectedUsers}
                columns={selectedUserColumns}
                rowKey="id"
                size="small"
                pagination={false}
              />
            </div>
          )}
          <Form.Item name="push_channel" label="推送渠道">
            <Checkbox.Group options={PUSH_CHANNEL_OPTIONS} />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
            rules={[{ max: 200, message: '描述不超过 200 个字符' }]}
          >
            <TextArea rows={2} placeholder="规则说明（可选）" />
          </Form.Item>
          {isEdit && (
            <Form.Item name="status" label="状态">
              <Select options={ACTIVE_STATUS_SELECT_OPTIONS} />
            </Form.Item>
          )}
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                {isEdit ? '保存' : '创建'}
              </Button>
              <Button onClick={() => navigate('/alerts')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
