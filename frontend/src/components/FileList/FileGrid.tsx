import React, { useState } from 'react';
import { Card, Checkbox, Dropdown, Modal, Input, Row, Col } from 'antd';
import type { MenuProps } from 'antd';
import {
  FileOutlined,
  FolderOutlined,
  DownloadOutlined,
  DeleteOutlined,
  EditOutlined,
  InfoCircleOutlined,
  MoreOutlined,
  FilePdfOutlined,
  FileImageOutlined,
  FileZipOutlined,
  FileTextOutlined,
  FileExcelOutlined,
  FileWordOutlined,
  VideoCameraOutlined,
} from '@ant-design/icons';
import { formatFileSize, getFileExtension } from '../../utils/helpers';
import type { FileItem } from '../../types';
import './FileGrid.css';

interface FileGridProps {
  files: FileItem[];
  onDelete: (file: FileItem) => void;
  onDownload: (file: FileItem) => void;
  onRename: (file: FileItem, newName: string) => void;
  onOpen: (file: FileItem) => void;
  selectedFiles: string[];
  onSelectChange: (keys: string[]) => void;
}

const FileGrid: React.FC<FileGridProps> = ({
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
      return <FolderOutlined className="file-icon folder-icon" />;
    }

    const ext = getFileExtension(file.name);
    const iconProps = { className: 'file-icon' };

    const iconMap: Record<string, JSX.Element> = {
      pdf: <FilePdfOutlined {...iconProps} style={{ color: '#f5222d' }} />,
      jpg: <FileImageOutlined {...iconProps} style={{ color: '#52c41a' }} />,
      jpeg: <FileImageOutlined {...iconProps} style={{ color: '#52c41a' }} />,
      png: <FileImageOutlined {...iconProps} style={{ color: '#52c41a' }} />,
      gif: <FileImageOutlined {...iconProps} style={{ color: '#52c41a' }} />,
      zip: <FileZipOutlined {...iconProps} style={{ color: '#faad14' }} />,
      rar: <FileZipOutlined {...iconProps} style={{ color: '#faad14' }} />,
      txt: <FileTextOutlined {...iconProps} style={{ color: '#1890ff' }} />,
      doc: <FileWordOutlined {...iconProps} style={{ color: '#2f54eb' }} />,
      docx: <FileWordOutlined {...iconProps} style={{ color: '#2f54eb' }} />,
      xls: <FileExcelOutlined {...iconProps} style={{ color: '#52c41a' }} />,
      xlsx: <FileExcelOutlined {...iconProps} style={{ color: '#52c41a' }} />,
      mp4: <VideoCameraOutlined {...iconProps} style={{ color: '#722ed1' }} />,
      avi: <VideoCameraOutlined {...iconProps} style={{ color: '#722ed1' }} />,
    };

    return iconMap[ext] || <FileOutlined {...iconProps} style={{ color: '#8c8c8c' }} />;
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

  // 处理选择
  const handleSelect = (key: string, checked: boolean) => {
    if (checked) {
      onSelectChange([...selectedFiles, key]);
    } else {
      onSelectChange(selectedFiles.filter((k) => k !== key));
    }
  };

  return (
    <>
      <Row gutter={[16, 16]} className="file-grid">
        {files.map((file) => (
          <Col key={file.key} xs={12} sm={8} md={6} lg={4} xl={3}>
            <Card
              hoverable
              className={`file-card ${selectedFiles.includes(file.key) ? 'selected' : ''}`}
              bodyStyle={{ padding: 16 }}
            >
              <div className="file-card-content">
                <Checkbox
                  className="file-checkbox"
                  checked={selectedFiles.includes(file.key)}
                  onChange={(e) => handleSelect(file.key, e.target.checked)}
                />
                <Dropdown
                  menu={getActionMenu(file)}
                  trigger={['click']}
                  placement="bottomRight"
                >
                  <MoreOutlined className="file-more" onClick={(e) => e.stopPropagation()} />
                </Dropdown>
                <div
                  className="file-icon-wrapper"
                  onClick={() => onOpen(file)}
                  style={{ cursor: file.is_dir ? 'pointer' : 'default' }}
                >
                  {getFileIcon(file)}
                </div>
                <div className="file-info">
                  <div className="file-name" title={file.name}>
                    {file.name}
                  </div>
                  <div className="file-size">{formatFileSize(file.capacity)}</div>
                </div>
              </div>
            </Card>
          </Col>
        ))}
      </Row>

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

export default FileGrid;
