import axios, { AxiosProgressEvent } from 'axios';
import type {
  FileMetadata,
  FileTreeResponse,
  FileInfo,
  ApiResponse,
  UploadRequest,
} from '@/types';

// 创建 axios 实例
const api = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 响应拦截器
api.interceptors.response.use(
  (response) => response,
  (error) => {
    console.error('API Error:', error);
    return Promise.reject(error);
  }
);

// API 服务类
class ApiService {
  // 获取文件列表（树形结构）
  async getFileList(): Promise<FileTreeResponse> {
    const response = await api.get<FileTreeResponse>('/list');
    return response.data;
  }

  // 获取简单文件列表
  async getSimpleFileList(): Promise<FileMetadata[]> {
    const response = await api.get<FileMetadata[]>('/list?format=simple');
    return response.data;
  }

  // 上传文件
  async uploadFile({ file, path = '', onProgress }: UploadRequest): Promise<ApiResponse> {
    const formData = new FormData();
    formData.append('file', file);
    
    const meta: FileMetadata = {
      name: path ? `${path}/${file.name}` : file.name,
      capacity: file.size,
    };
    formData.append('meta', JSON.stringify(meta));
    
    if (path) {
      formData.append('path', path);
    }

    const response = await api.post<ApiResponse>('/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
      onUploadProgress: (progressEvent: AxiosProgressEvent) => {
        if (progressEvent.total && onProgress) {
          const percent = Math.round((progressEvent.loaded * 100) / progressEvent.total);
          onProgress(percent);
        }
      },
    });
    return response.data;
  }

  // 下载文件
  async downloadFile(name: string): Promise<void> {
    const response = await api.get('/download', {
      params: { name },
      responseType: 'blob',
    });
    
    // 创建下载链接
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', name.split('/').pop() || 'download');
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  }

  // 下载文件夹（zip）
  async downloadFolder(dirname: string): Promise<void> {
    const response = await api.get('/downloaddir', {
      params: { dirname },
      responseType: 'blob',
    });
    
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `${dirname.split('/').pop()}.zip`);
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  }

  // 删除文件
  async deleteFile(name: string): Promise<ApiResponse> {
    const response = await api.delete<ApiResponse>('/delete', {
      params: { name },
    });
    return response.data;
  }

  // 删除文件夹
  async deleteFolder(dirname: string): Promise<ApiResponse> {
    const response = await api.delete<ApiResponse>('/deletedir', {
      params: { dirname },
    });
    return response.data;
  }

  // 创建文件夹
  async createFolder(path: string): Promise<ApiResponse> {
    const response = await api.post<ApiResponse>('/createdir', null, {
      params: { path },
    });
    return response.data;
  }

  // 重命名文件/文件夹
  async rename(oldName: string, newName: string): Promise<ApiResponse> {
    const response = await api.put<ApiResponse>('/rename', null, {
      params: { oldName, newName },
    });
    return response.data;
  }

  // 获取文件信息
  async getFileInfo(name: string): Promise<FileInfo> {
    const response = await api.get<FileInfo>('/info', {
      params: { name },
    });
    return response.data;
  }

  // 批量删除
  async batchDelete(names: string[]): Promise<ApiResponse> {
    const response = await api.delete<ApiResponse>('/batch-delete', {
      data: { names },
    });
    return response.data;
  }

  // 批量下载
  async batchDownload(names: string[]): Promise<void> {
    const response = await api.post('/batch-download', 
      { names },
      { responseType: 'blob' }
    );
    
    const url = window.URL.createObjectURL(new Blob([response.data]));
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', 'batch-download.zip');
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.URL.revokeObjectURL(url);
  }

  // 搜索文件
  async searchFiles(keyword: string): Promise<FileMetadata[]> {
    const response = await api.get<FileMetadata[]>('/search', {
      params: { keyword },
    });
    return response.data;
  }

  // 按类型过滤
  async filterByType(type: string): Promise<FileMetadata[]> {
    const response = await api.get<FileMetadata[]>('/filter/type', {
      params: { type },
    });
    return response.data;
  }

  // 按日期过滤
  async filterByDate(startDate: string, endDate: string): Promise<FileMetadata[]> {
    const response = await api.get<FileMetadata[]>('/filter/date', {
      params: { start_date: startDate, end_date: endDate },
    });
    return response.data;
  }

  // 按大小过滤
  async filterBySize(minSize: number, maxSize: number): Promise<FileMetadata[]> {
    const response = await api.get<FileMetadata[]>('/filter/size', {
      params: { min_size: minSize, max_size: maxSize },
    });
    return response.data;
  }
}

export default new ApiService();
