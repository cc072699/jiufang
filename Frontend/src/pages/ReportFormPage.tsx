import { useEffect, useState } from 'react';
import { useNavigate, useParams, useLocation } from 'react-router-dom';
import { Form, Input, Select, DatePicker, Typography, Button, Space, message, Card } from 'antd';
import type { Dayjs } from 'dayjs';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { createReport, updateReport, getReportList } from '../shared/api';
import type { CreateReportRequest, UpdateReportRequest, ReportRecord } from '../shared/api/types';

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

  useEffect(() => {
    if (!isEdit) return;
    const record = navState?.record || data?.reports?.find((r) => r.id === id);
    if (record) {
      form.setFieldsValue({
        name: record.name,
        sql: record.sql,
        schedule_type: record.schedule_type,
        status: record.status,
        // TODO: recipients, description, schedule_time 不在列表返回中，需后端补充详情接口
      });
    }
  }, [isEdit, navState, data, id, form]);

  const handleFinish = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      const scheduleTime = (values.schedule_time as Dayjs)?.toISOString();
      const recipients = ((values.recipients as string) || '').split(',').map((s) => s.trim()).filter(Boolean);

      if (isEdit) {
        const payload: UpdateReportRequest = {
          name: values.name as string,
          sql: values.sql as string,
          schedule_type: values.schedule_type as 'daily' | 'weekly' | 'monthly',
          schedule_time: scheduleTime,
          recipients,
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

  return (
    <div>
      <Title level={4}>{isEdit ? '编辑报告' : '新建报告'}</Title>
      <Card style={{ maxWidth: 640 }}>
        {isEdit && (
          <div style={{ marginBottom: 16 }}>
            <Text type="secondary">提示：编辑页部分字段（接收人、描述、推送时间）需后端补充详情接口后才能完整回填。</Text>
          </div>
        )}
        <Form
          form={form}
          layout="vertical"
          onFinish={handleFinish}
          initialValues={{ schedule_type: 'daily', status: 'active' }}
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
            <Select options={[
              { label: '每日', value: 'daily' },
              { label: '每周', value: 'weekly' },
              { label: '每月', value: 'monthly' },
            ]} />
          </Form.Item>
          <Form.Item
            name="schedule_time"
            label="推送时间"
            rules={[{ required: true, message: '请选择推送时间' }]}
          >
            <DatePicker showTime format="YYYY-MM-DD HH:mm:ss" style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name="recipients" label="接收人（逗号分隔）" rules={[{ required: !isEdit, message: '请输入接收人' }]}>
            <Input placeholder="张三, 李四, 王五" />
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
              <Select options={[
                { label: '启用', value: 'active' },
                { label: '停用', value: 'inactive' },
              ]} />
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
