// ============================================================
// AppLayout - 侧边导航 + 顶部栏 + 内容区
// 导航按角色裁剪，最终授权依赖后端 401/403
// ============================================================

import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { Layout, Menu, Button, Dropdown, Avatar, theme, Modal } from 'antd';
import type { MenuProps } from 'antd';
import {
  SearchOutlined,
  HistoryOutlined,
  StarOutlined,
  HomeOutlined,
  SafetyOutlined,
  FileTextOutlined,
  AlertOutlined,
  UserOutlined,
  TeamOutlined,
  LogoutOutlined,
  SettingOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SendOutlined,
} from '@ant-design/icons';
import { useAuthStore } from '../../shared/auth/store';
import { logout as logoutApi } from '../../shared/api';
import { queryClient } from '../providers';

const { Header, Sider, Content } = Layout;

type MenuItem = Required<MenuProps>['items'][number];

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false);
  const navigate = useNavigate();
  const location = useLocation();
  const { user, isAdmin, clearAuth } = useAuthStore();
  const { token: themeToken } = theme.useToken();

  // 构建菜单项
  const menuItems: MenuItem[] = [
    { key: '/workbench', icon: <HomeOutlined />, label: '工作台' },
    { key: '/query', icon: <SearchOutlined />, label: '对话查询' },
    { key: '/history', icon: <HistoryOutlined />, label: '历史记录' },
    { key: '/favorites', icon: <StarOutlined />, label: '我的收藏' },
  ];

  if (isAdmin) {
    menuItems.push(
      { type: 'divider' },
      { key: 'admin-group', type: 'group', label: '系统管理' },
      { key: '/users', icon: <TeamOutlined />, label: '人员管理' },
      { key: '/permissions', icon: <SafetyOutlined />, label: '权限管理' },
      { key: '/reports', icon: <FileTextOutlined />, label: '定时报告' },
      { key: '/alerts', icon: <AlertOutlined />, label: '预警规则' },
      { key: '/logs', icon: <FileTextOutlined />, label: '操作日志' },
      { key: '/push-records', icon: <SendOutlined />, label: '推送记录' },
    );
  }

  const selectedKey = '/' + location.pathname.split('/')[1];

  // 用户下拉菜单
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

  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <SettingOutlined />,
      label: '个人中心',
      onClick: () => navigate('/profile'),
    },
    { type: 'divider' },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
      onClick: handleLogout,
    },
  ];

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={220}
        style={{
          background: themeToken.colorBgContainer,
          borderRight: `1px solid ${themeToken.colorBorderSecondary}`,
        }}
      >
        <div
          style={{
            height: 64,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
            fontWeight: 700,
            fontSize: collapsed ? 16 : 18,
          }}
        >
          {collapsed ? '久' : '久方查询助手'}
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => navigate(key)}
          style={{ border: 'none' }}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            padding: '0 24px',
            background: themeToken.colorBgContainer,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: `1px solid ${themeToken.colorBorderSecondary}`,
          }}
        >
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
          />
          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <div style={{ cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 8 }}>
              <Avatar size="small" icon={<UserOutlined />} />
              <span>{user?.username || ''}</span>
            </div>
          </Dropdown>
        </Header>
        <Content
          style={{
            margin: 16,
            padding: 24,
            background: themeToken.colorBgContainer,
            borderRadius: themeToken.borderRadiusLG,
            minHeight: 280,
            overflow: 'auto',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  );
}