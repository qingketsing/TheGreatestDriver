import React, { useState } from 'react';
import { Outlet, useNavigate } from 'react-router-dom';
import { Layout, Menu, Input, Avatar, Dropdown, Space, Badge } from 'antd';
import {
  CloudOutlined,
  FolderOutlined,
  DeleteOutlined,
  StarOutlined,
  HistoryOutlined,
  SettingOutlined,
  SearchOutlined,
  BellOutlined,
  UserOutlined,
  LogoutOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import './MainLayout.css';

const { Header, Sider, Content } = Layout;

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const [selectedKey, setSelectedKey] = useState('files');

  // 侧边栏菜单项
  const menuItems: MenuProps['items'] = [
    {
      key: 'files',
      icon: <FolderOutlined />,
      label: '我的文件',
    },
    {
      key: 'recent',
      icon: <HistoryOutlined />,
      label: '最近使用',
    },
    {
      key: 'starred',
      icon: <StarOutlined />,
      label: '星标文件',
    },
    {
      key: 'trash',
      icon: <DeleteOutlined />,
      label: '回收站',
    },
    {
      type: 'divider',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
  ];

  // 用户菜单
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
    },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    setSelectedKey(key);
    if (key === 'files') {
      navigate('/');
    } else {
      navigate(`/${key}`);
    }
  };

  return (
    <Layout className="main-layout">
      <Header className="main-header">
        <div className="header-left">
          <CloudOutlined className="logo-icon" />
          <span className="logo-text">Single Drive</span>
        </div>
        <div className="header-center">
          <Input
            placeholder="搜索文件和文件夹"
            prefix={<SearchOutlined />}
            className="search-input"
            size="large"
          />
        </div>
        <div className="header-right">
          <Space size="large">
            <Badge count={3}>
              <BellOutlined className="header-icon" />
            </Badge>
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <Avatar
                size="large"
                icon={<UserOutlined />}
                className="user-avatar"
                style={{ cursor: 'pointer' }}
              />
            </Dropdown>
          </Space>
        </div>
      </Header>
      <Layout>
        <Sider
          width={220}
          className="main-sider"
          theme="light"
        >
          <Menu
            mode="inline"
            selectedKeys={[selectedKey]}
            items={menuItems}
            onClick={handleMenuClick}
            className="side-menu"
          />
          <div className="storage-info">
            <div className="storage-text">
              <span>已使用 0 GB / 100 GB</span>
            </div>
            <div className="storage-bar">
              <div className="storage-used" style={{ width: '0%' }} />
            </div>
          </div>
        </Sider>
        <Layout className="main-content-layout">
          <Content className="main-content">
            <Outlet />
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
