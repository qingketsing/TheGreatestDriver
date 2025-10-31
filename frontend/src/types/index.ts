// 文件元数据类型
export interface FileMetadata {
  id?: number;
  name: string;
  capacity: number;
  is_dir?: boolean;
  path?: string;
  created_at?: string;
  mod_time?: string;
}

// 树节点类型
export interface TreeNode {
  id: number;
  name: string;
  capacity: number;
  is_dir: boolean;
  path: string;
  children?: TreeNode[];
}

// 文件树响应
export interface FileTreeResponse {
  total: number;
  roots: TreeNode[];
}

// 文件信息响应
export interface FileInfo {
  name: string;
  size: number;
  mode: string;
  mod_time: string;
  is_directory: boolean;
}

// 上传进度
export interface UploadProgress {
  filename: string;
  percent: number;
  status: 'uploading' | 'done' | 'error';
}

// API响应类型
export interface ApiResponse<T = any> {
  message?: string;
  error?: string;
  data?: T;
}

// 文件列表项（用于UI展示）
export interface FileItem extends TreeNode {
  key: string;
  type: 'file' | 'folder';
  size: number;
  modifiedTime?: string;
  extension?: string;
}

// 面包屑导航项
export interface BreadcrumbItem {
  name: string;
  path: string;
}

// 文件上传请求
export interface UploadRequest {
  file: File;
  path?: string;
  onProgress?: (percent: number) => void;
}
