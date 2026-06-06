import { Modal } from 'antd';
import { ExclamationCircleOutlined } from '@ant-design/icons';

interface ConfirmOptions {
  title: string;
  content: string;
  onOk: () => void | Promise<void>;
  okText?: string;
  okType?: 'primary' | 'danger';
  cancelText?: string;
}

export function confirmAction({
  title,
  content,
  onOk,
  okText = '确认',
  okType = 'danger',
  cancelText = '取消',
}: ConfirmOptions) {
  Modal.confirm({
    title,
    icon: <ExclamationCircleOutlined />,
    content,
    okText,
    okType,
    cancelText,
    onOk,
  });
}
