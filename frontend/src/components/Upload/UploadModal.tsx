import React, { useState } from 'react';
import { Modal, Upload, Progress, message, Button } from 'antd';
import { InboxOutlined, DeleteOutlined } from '@ant-design/icons';
import type { UploadProps, UploadFile } from 'antd';
import apiService from '../../services/api';
import './UploadModal.css';

const { Dragger } = Upload;

interface UploadModalProps {
  visible: boolean;
  currentPath: string;
  onClose: () => void;
  onSuccess: () => void;
}

interface FileProgress {
  file: File;
  percent: number;
  status: 'uploading' | 'done' | 'error';
}

const UploadModal: React.FC<UploadModalProps> = ({
  visible,
  currentPath,
  onClose,
  onSuccess,
}) => {
  const [fileList, setFileList] = useState<UploadFile[]>([]);
  const [uploading, setUploading] = useState(false);
  const [uploadProgress, setUploadProgress] = useState<Record<string, number>>({});

  // 自定义上传请求
  const handleUpload = async () => {
    if (fileList.length === 0) {
      message.warning('请先选择文件');
      return;
    }

    setUploading(true);

    try {
      // 依次上传文件
      for (const fileItem of fileList) {
        if (fileItem.originFileObj) {
          await apiService.uploadFile({
            file: fileItem.originFileObj,
            path: currentPath === '/' ? '' : currentPath.replace(/^\//, ''),
            onProgress: (percent) => {
              setUploadProgress((prev) => ({
                ...prev,
                [fileItem.uid]: percent,
              }));
            },
          });
        }
      }

      message.success('所有文件上传成功');
      setFileList([]);
      setUploadProgress({});
      onSuccess();
      onClose();
    } catch (error) {
      message.error('部分文件上传失败');
      console.error('Upload error:', error);
    } finally {
      setUploading(false);
    }
  };

  // 上传属性配置
  const uploadProps: UploadProps = {
    name: 'file',
    multiple: true,
    fileList,
    beforeUpload: (file) => {
      setFileList((prev) => [...prev, file as any]);
      return false; // 阻止自动上传
    },
    onRemove: (file) => {
      setFileList((prev) => prev.filter((item) => item.uid !== file.uid));
      setUploadProgress((prev) => {
        const newProgress = { ...prev };
        delete newProgress[file.uid];
        return newProgress;
      });
    },
    onDrop: (e) => {
      console.log('Dropped files', e.dataTransfer.files);
    },
  };

  const handleCancel = () => {
    if (!uploading) {
      setFileList([]);
      setUploadProgress({});
      onClose();
    }
  };

  return (
    <Modal
      title={`上传文件${currentPath !== '/' ? ` - ${currentPath}` : ''}`}
      open={visible}
      onOk={handleUpload}
      onCancel={handleCancel}
      okText="开始上传"
      cancelText="取消"
      confirmLoading={uploading}
      width={600}
      okButtonProps={{ disabled: fileList.length === 0 }}
    >
      <div className="upload-modal-content">
        <Dragger {...uploadProps}>
          <p className="ant-upload-drag-icon">
            <InboxOutlined />
          </p>
          <p className="ant-upload-text">点击或拖拽文件到这里上传</p>
          <p className="ant-upload-hint">
            支持单个或批量上传。支持拖拽多个文件
          </p>
        </Dragger>

        {fileList.length > 0 && (
          <div className="upload-file-list">
            <div className="upload-file-list-header">
              <span>待上传文件 ({fileList.length})</span>
              {!uploading && (
                <Button
                  type="link"
                  danger
                  size="small"
                  icon={<DeleteOutlined />}
                  onClick={() => {
                    setFileList([]);
                    setUploadProgress({});
                  }}
                >
                  清空列表
                </Button>
              )}
            </div>
            <div className="upload-file-items">
              {fileList.map((file) => (
                <div key={file.uid} className="upload-file-item">
                  <div className="upload-file-info">
                    <span className="upload-file-name">{file.name}</span>
                    <span className="upload-file-size">
                      {((file.size || 0) / 1024 / 1024).toFixed(2)} MB
                    </span>
                  </div>
                  {uploadProgress[file.uid] !== undefined && (
                    <Progress
                      percent={uploadProgress[file.uid]}
                      size="small"
                      status={
                        uploadProgress[file.uid] === 100 ? 'success' : 'active'
                      }
                    />
                  )}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    </Modal>
  );
};

export default UploadModal;
