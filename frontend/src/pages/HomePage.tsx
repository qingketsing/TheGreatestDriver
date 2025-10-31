import React, { useState, useEffect, useCallback } from 'react';
import {
  Button,
  Space,
  Breadcrumb,
  message,
  Modal,
  Input,
  Spin,
  Empty,
  Dropdown,
  Segmented,
} from 'antd';
import {
  PlusOutlined,
  UploadOutlined,
  ReloadOutlined,
  AppstoreOutlined,
  BarsOutlined,
  FolderOutlined,
  HomeOutlined,
} from '@ant-design/icons';
import type { MenuProps } from 'antd';
import FileList from '../components/FileList/FileList';
import FileGrid from '../components/FileList/FileGrid';
import UploadModal from '../components/Upload/UploadModal';
import apiService from '../services/api';
import { validateFileName } from '../utils/helpers';
import type { FileItem, TreeNode, BreadcrumbItem } from '../types';
import './HomePage.css';

const HomePage: React.FC = () => {
  const [files, setFiles] = useState<FileItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [currentPath, setCurrentPath] = useState<string>('/');
  const [breadcrumbs, setBreadcrumbs] = useState<BreadcrumbItem[]>([
    { name: '我的文件', path: '/' },
  ]);
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');
  const [uploadModalVisible, setUploadModalVisible] = useState(false);
  const [createFolderVisible, setCreateFolderVisible] = useState(false);
  const [newFolderName, setNewFolderName] = useState('');
  const [selectedFiles, setSelectedFiles] = useState<string[]>([]);

  // 加载文件列表
  const loadFiles = useCallback(async () => {
    setLoading(true);
    try {
      const response = await apiService.getFileList();
      
      // 转换树形结构为扁平列表（根据当前路径过滤）
      const fileItems = convertTreeToList(response.roots);
      setFiles(fileItems);
    } catch (error) {
      message.error('加载文件列表失败');
      console.error('Load files error:', error);
    } finally {
      setLoading(false);
    }
  }, []);

  // 将树形结构转换为扁平列表
  const convertTreeToList = (nodes: TreeNode[]): FileItem[] => {
    const result: FileItem[] = [];
    
    const traverse = (node: TreeNode, parentPath: string = '') => {
      const fullPath = parentPath ? `${parentPath}/${node.name}` : node.name;
      
      const item: FileItem = {
        ...node,
        key: `${node.id}`,
        type: node.is_dir ? 'folder' : 'file',
        size: node.capacity,
        path: fullPath,
      };
      
      // 只添加根目录下的文件
      if (!parentPath) {
        result.push(item);
      }
      
      // 递归处理子节点
      if (node.children && node.children.length > 0) {
        node.children.forEach(child => traverse(child, fullPath));
      }
    };
    
    nodes.forEach(node => traverse(node));
    return result;
  };

  useEffect(() => {
    loadFiles();
  }, [loadFiles]);

  // 刷新列表
  const handleRefresh = () => {
    loadFiles();
    message.success('刷新成功');
  };

  // 创建文件夹
  const handleCreateFolder = async () => {
    const validation = validateFileName(newFolderName);
    if (!validation.valid) {
      message.error(validation.message);
      return;
    }

    try {
      const folderPath = currentPath === '/' 
        ? newFolderName 
        : `${currentPath}/${newFolderName}`.replace(/^\//, '');
      
      await apiService.createFolder(folderPath);
      message.success('文件夹创建成功');
      setCreateFolderVisible(false);
      setNewFolderName('');
      loadFiles();
    } catch (error) {
      message.error('创建文件夹失败');
      console.error('Create folder error:', error);
    }
  };

  // 删除文件/文件夹
  const handleDelete = async (item: FileItem) => {
    Modal.confirm({
      title: `确定要删除 ${item.name} 吗？`,
      content: item.is_dir ? '删除文件夹将同时删除其中的所有文件' : '',
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          if (item.is_dir) {
            await apiService.deleteFolder(item.path);
          } else {
            await apiService.deleteFile(item.path);
          }
          message.success('删除成功');
          loadFiles();
        } catch (error) {
          message.error('删除失败');
          console.error('Delete error:', error);
        }
      },
    });
  };

  // 下载文件/文件夹
  const handleDownload = async (item: FileItem) => {
    try {
      message.loading({ content: '正在准备下载...', key: 'download' });
      if (item.is_dir) {
        await apiService.downloadFolder(item.path);
      } else {
        await apiService.downloadFile(item.path);
      }
      message.success({ content: '下载成功', key: 'download' });
    } catch (error) {
      message.error({ content: '下载失败', key: 'download' });
      console.error('Download error:', error);
    }
  };

  // 重命名
  const handleRename = async (item: FileItem, newName: string) => {
    const validation = validateFileName(newName);
    if (!validation.valid) {
      message.error(validation.message);
      return;
    }

    try {
      const newPath = item.path.replace(/[^/]+$/, newName);
      await apiService.rename(item.path, newPath);
      message.success('重命名成功');
      loadFiles();
    } catch (error) {
      message.error('重命名失败');
      console.error('Rename error:', error);
    }
  };

  // 打开文件夹
  const handleOpenFolder = (item: FileItem) => {
    if (item.is_dir) {
      setCurrentPath(item.path);
      const newBreadcrumbs = [
        { name: '我的文件', path: '/' },
        ...item.path.split('/').filter(Boolean).map((part, index, array) => ({
          name: part,
          path: '/' + array.slice(0, index + 1).join('/'),
        })),
      ];
      setBreadcrumbs(newBreadcrumbs);
    }
  };

  // 面包屑导航点击
  const handleBreadcrumbClick = (path: string) => {
    setCurrentPath(path);
    if (path === '/') {
      setBreadcrumbs([{ name: '我的文件', path: '/' }]);
    } else {
      const newBreadcrumbs = [
        { name: '我的文件', path: '/' },
        ...path.split('/').filter(Boolean).map((part, index, array) => ({
          name: part,
          path: '/' + array.slice(0, index + 1).join('/'),
        })),
      ];
      setBreadcrumbs(newBreadcrumbs);
    }
  };

  // 新建菜单
  const createMenuItems: MenuProps['items'] = [
    {
      key: 'folder',
      icon: <FolderOutlined />,
      label: '新建文件夹',
      onClick: () => setCreateFolderVisible(true),
    },
    {
      key: 'upload',
      icon: <UploadOutlined />,
      label: '上传文件',
      onClick: () => setUploadModalVisible(true),
    },
  ];

  return (
    <div className="home-page">
      <div className="page-header">
        <Breadcrumb className="breadcrumb">
          {breadcrumbs.map((item, index) => (
            <Breadcrumb.Item key={item.path}>
              {index === 0 && <HomeOutlined />}
              <span
                onClick={() => handleBreadcrumbClick(item.path)}
                style={{ cursor: 'pointer' }}
              >
                {item.name}
              </span>
            </Breadcrumb.Item>
          ))}
        </Breadcrumb>
        <Space className="toolbar">
          <Dropdown menu={{ items: createMenuItems }} placement="bottomLeft">
            <Button type="primary" icon={<PlusOutlined />}>
              新建
            </Button>
          </Dropdown>
          <Button icon={<UploadOutlined />} onClick={() => setUploadModalVisible(true)}>
            上传
          </Button>
          <Button icon={<ReloadOutlined />} onClick={handleRefresh}>
            刷新
          </Button>
          <Segmented
            value={viewMode}
            onChange={(value) => setViewMode(value as 'list' | 'grid')}
            options={[
              { value: 'list', icon: <BarsOutlined /> },
              { value: 'grid', icon: <AppstoreOutlined /> },
            ]}
          />
        </Space>
      </div>

      <div className="file-content">
        <Spin spinning={loading}>
          {files.length === 0 && !loading ? (
            <Empty
              description="暂无文件"
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              style={{ marginTop: 100 }}
            />
          ) : viewMode === 'list' ? (
            <FileList
              files={files}
              onDelete={handleDelete}
              onDownload={handleDownload}
              onRename={handleRename}
              onOpen={handleOpenFolder}
              selectedFiles={selectedFiles}
              onSelectChange={setSelectedFiles}
            />
          ) : (
            <FileGrid
              files={files}
              onDelete={handleDelete}
              onDownload={handleDownload}
              onRename={handleRename}
              onOpen={handleOpenFolder}
              selectedFiles={selectedFiles}
              onSelectChange={setSelectedFiles}
            />
          )}
        </Spin>
      </div>

      {/* 上传模态框 */}
      <UploadModal
        visible={uploadModalVisible}
        currentPath={currentPath}
        onClose={() => setUploadModalVisible(false)}
        onSuccess={loadFiles}
      />

      {/* 创建文件夹模态框 */}
      <Modal
        title="新建文件夹"
        open={createFolderVisible}
        onOk={handleCreateFolder}
        onCancel={() => {
          setCreateFolderVisible(false);
          setNewFolderName('');
        }}
        okText="创建"
        cancelText="取消"
      >
        <Input
          placeholder="请输入文件夹名称"
          value={newFolderName}
          onChange={(e) => setNewFolderName(e.target.value)}
          onPressEnter={handleCreateFolder}
        />
      </Modal>
    </div>
  );
};

export default HomePage;
