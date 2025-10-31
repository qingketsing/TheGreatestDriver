import React, { useState } from 'react';
import { Table, Tag, Dropdown, Modal, Input, Space } from 'antd';
import type { ColumnsType, TableProps } from 'antd/es/table';
import type { MenuProps } from 'antd';
import {
  FileOutlined,
  FolderOutlined,
  DownloadOutlined,
  DeleteOutlined,
  EditOutlined,
  InfoCircleOutlined,
  MoreOutlined,
} from '@ant-design/icons';
import { formatFileSize, formatDate, getFileType } from '../../utils/helpers';
import type { FileItem } from '../../types';
import './FileList.css';

interface FileListProps {
  files: FileItem[];
  onDelete: (file: FileItem) => void;
  onDownload: (file: FileItem) => void;
  onRename: (file: FileItem, newName: string) => void;
  onOpen: (file: FileItem) => void;
  selectedFiles: string[];
  onSelectChange: (keys: string[]) => void;
}

const FileList: React.FC<FileListProps> = ({
  files,
  onDelete,
  onDownload,
  onRename,
  onOpen,
  selectedFiles,
  onSelectChange,
}) => {
  const [renameModalVisible, setRenameModalVisible] = useState(false);
  const [currentFile, setCurrentFile] = useState<FileItem | null>(null);
  const [newName, setNewName] = useState('');

  // 获取文件图标
  const getFileIcon = (file: FileItem) => {
    if (file.is_dir) {
      return <FolderOutlined style={{ fontSize: 20, color: '#faad14' }} />;
    }
    return <FileOutlined style={{ fontSize: 20, color: '#1890ff' }} />;
  };

  // 获取文件类型标签
  const getFileTypeTag = (file: FileItem) => {
    if (file.is_dir) {
      return <Tag color="gold">文件夹</Tag>;
    }
    
    const type = getFileType(file.name);
    const typeMap: Record<string, { color: string; text: string }> = {
      image: { color: 'purple', text: '图片' },
      video: { color: 'red', text: '视频' },
      audio: { color: 'orange', text: '音频' },
      document: { color: 'blue', text: '文档' },
      archive: { color: 'green', text: '压缩包' },
      code: { color: 'geekblue', text: '代码' },
    };
    
    const typeInfo = typeMap[type];
    if (typeInfo) {
      return <Tag color={typeInfo.color}>{typeInfo.text}</Tag>;
    }
    
    return <Tag>其他</Tag>;
  };

  // 操作菜单
  const getActionMenu = (file: FileItem): MenuProps => ({
    items: [
      {
        key: 'download',
        icon: <DownloadOutlined />,
        label: '下载',
        onClick: () => onDownload(file),
      },
      {
        key: 'rename',
        icon: <EditOutlined />,
        label: '重命名',
        onClick: () => {
          setCurrentFile(file);
          setNewName(file.name);
          setRenameModalVisible(true);
        },
      },
      {
        key: 'info',
        icon: <InfoCircleOutlined />,
        label: '详情',
      },
      {
        type: 'divider',
      },
      {
        key: 'delete',
        icon: <DeleteOutlined />,
        label: '删除',
        danger: true,
        onClick: () => onDelete(file),
      },
    ],
  });

  // 处理重命名
  const handleRename = () => {
    if (currentFile && newName) {
      onRename(currentFile, newName);
      setRenameModalVisible(false);
      setCurrentFile(null);
      setNewName('');
    }
  };

  // 表格列定义
  const columns: ColumnsType<FileItem> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: '40%',
      render: (text, record) => (
        <Space
          className="file-name-cell"
          onClick={() => onOpen(record)}
          style={{ cursor: record.is_dir ? 'pointer' : 'default' }}
        >
          {getFileIcon(record)}
          <span className="file-name">{text}</span>
        </Space>
      ),
      sorter: (a, b) => a.name.localeCompare(b.name, 'zh-CN'),
    },
    {
      title: '类型',
      key: 'type',
      width: '15%',
      render: (_, record) => getFileTypeTag(record),
    },
    {
      title: '大小',
      dataIndex: 'capacity',
      key: 'size',
      width: '15%',
      render: (size) => formatFileSize(size),
      sorter: (a, b) => a.capacity - b.capacity,
    },
    {
      title: '修改时间',
      dataIndex: 'modifiedTime',
      key: 'modifiedTime',
      width: '20%',
      render: (time) => time ? formatDate(time) : '-',
    },
    {
      title: '操作',
      key: 'action',
      width: '10%',
      render: (_, record) => (
        <Dropdown menu={getActionMenu(record)} trigger={['click']}>
          <MoreOutlined style={{ fontSize: 20, cursor: 'pointer' }} />
        </Dropdown>
      ),
    },
  ];

  const rowSelection: TableProps<FileItem>['rowSelection'] = {
    selectedRowKeys: selectedFiles,
    onChange: (selectedRowKeys) => {
      onSelectChange(selectedRowKeys as string[]);
    },
  };

  return (
    <>
      <Table
        columns={columns}
        dataSource={files}
        rowSelection={rowSelection}
        pagination={{
          pageSize: 20,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 项`,
        }}
        className="file-list-table"
      />

      <Modal
        title="重命名"
        open={renameModalVisible}
        onOk={handleRename}
        onCancel={() => {
          setRenameModalVisible(false);
          setCurrentFile(null);
          setNewName('');
        }}
        okText="确定"
        cancelText="取消"
      >
        <Input
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          onPressEnter={handleRename}
          placeholder="请输入新名称"
        />
      </Modal>
    </>
  );
};

export default FileList;
