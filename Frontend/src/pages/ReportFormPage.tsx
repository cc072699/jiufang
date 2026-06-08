import { useEffect, useState } from 'react';
import { useNavigate, useParams, useLocation } from 'react-router-dom';
import { Form, Input, Select, DatePicker, Typography, Button, Space, message, Card, Checkbox, Table, Tag } from 'antd';
import type { Dayjs } from 'dayjs';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { createReport, updateReport, getReportList, getUserList, getGroupList } from '../shared/api';
import { SCHEDULE_SELECT_OPTIONS, ACTIVE_STATUS_SELECT_OPTIONS, ROLE_LABELS, PUSH_CHANNEL_OPTIONS } from '../shared/constants';
import type { CreateReportRequest, UpdateReportRequest, ReportRecord, UserRecord } from '../shared/api/types';

const { Title, Text } = Typography;
const { TextArea } = Input;

export default function ReportFormPage() {
  const { id } = useParams<{ id: string }>();
  const isEdit = Boolean(id);
  const navigate = useNavigate();
  const location = useLocation();
  const queryClient = useQueryClient();
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);

  const navState = location.state as { record?: ReportRecord } | null;

  const { data } = useQuery({
    queryKey: ['reports'],
    queryFn: () => getReportList({ page: 1, page_size: 100 }),
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
    const record = navState?.record || data?.reports?.find((r) => r.id === id);
    if (record) {
      form.setFieldsValue({
        name: record.name,
        sql: record.sql,
        schedule_type: record.schedule_type,
        status: record.status,
        recipients: record.recipients,
        push_channel: record.push_channel ? [record.push_channel] : ['email'],
      });
    }
  }, [isEdit, navState, data, id, form]);

  const handleFinish = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      const scheduleTime = (values.schedule_time as Dayjs)?.toISOString();
      const recipients = (values.recipients as string[]) || [];
      const channels = (values.push_channel as string[]) || ['email'];

      if (isEdit) {
        const payload: UpdateReportRequest = {
          name: values.name as string,
          sql: values.sql as string,
          schedule_type: values.schedule_type as 'daily' | 'weekly' | 'monthly',
          schedule_time: scheduleTime,
          recipients,
          push_channel: channels[0] as 'wechat' | 'email',
          description: values.description as string,
          status: values.status as 'active' | 'inactive',
        };
        await updateReport(id!, payload);
        message.success('更新成功');
      } else {
        const payload: CreateReportRequest = {
          name: values.name as string,
          sql: values.sql as string,
          schedule_type: values.schedule_type as 'daily' | 'weekly' | 'monthly',
          schedule_time: scheduleTime || '',
          recipients,
          push_channel: channels[0] as 'wechat' | 'email',
          description: values.description as string,
        };
        await createReport(payload);
        message.success('创建成功');
      }
      queryClient.invalidateQueries({ queryKey: ['reports'] });
      navigate('/reports');
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
      <Title level={4}>{isEdit ? '编辑报告' : '新建报告'}</Title>
      <Card style={{ maxWidth: 720 }}>
        <Form
          form={form}
          layout="vertical"
          onFinish={handleFinish}
          initialValues={{ schedule_type: 'daily', status: 'active', push_channel: ['email'] }}
        >
          <Form.Item
            name="name"
            label="报告名称"
            rules={[
              { required: true, message: '请输入名称' },
              { max: 100, message: '名称不超过 100 个字符' },
            ]}
          >
            <Input placeholder="例：月度销售汇总" />
          </Form.Item>
          <Form.Item
            name="sql"
            label="SQL 查询语句"
            rules={[
              { required: true, message: '请输入 SQL 语句' },
              { min: 1, max: 5000, message: 'SQL 长度 1~5000 个字符' },
            ]}
          >
            <TextArea rows={4} placeholder="SELECT product_name, SUM(amount) FROM sales WHERE ..." />
          </Form.Item>
          <Form.Item name="schedule_type" label="推送频率" rules={[{ required: true }]}>
            <Select options={SCHEDULE_SELECT_OPTIONS} />
          </Form.Item>
          <Form.Item
            name="schedule_time"
            label="推送时间"
            rules={[{ required: true, message: '请选择推送时间' }]}
          >
            <DatePicker showTime format="YYYY-MM-DD HH:mm:ss" style={{ width: '100%' }} />
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
            <TextArea rows={2} placeholder="报告说明（可选）" />
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
              <Button onClick={() => navigate('/reports')}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>
    </div>
  );
}
