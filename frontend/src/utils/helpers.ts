/**
 * 格式化文件大小
 */
export const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '-';
  
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
};

/**
 * 获取文件扩展名
 */
export const getFileExtension = (filename: string): string => {
  const parts = filename.split('.');
  return parts.length > 1 ? parts.pop()?.toLowerCase() || '' : '';
};

/**
 * 根据文件名获取文件类型
 */
export const getFileType = (filename: string): string => {
  const ext = getFileExtension(filename);
  
  const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'svg', 'webp'];
  const videoExts = ['mp4', 'avi', 'mov', 'wmv', 'flv', 'mkv', 'webm'];
  const audioExts = ['mp3', 'wav', 'flac', 'aac', 'ogg', 'm4a'];
  const documentExts = ['pdf', 'doc', 'docx', 'xls', 'xlsx', 'ppt', 'pptx', 'txt'];
  const archiveExts = ['zip', 'rar', '7z', 'tar', 'gz', 'bz2'];
  const codeExts = ['js', 'ts', 'jsx', 'tsx', 'py', 'java', 'c', 'cpp', 'go', 'rs', 'html', 'css'];
  
  if (imageExts.includes(ext)) return 'image';
  if (videoExts.includes(ext)) return 'video';
  if (audioExts.includes(ext)) return 'audio';
  if (documentExts.includes(ext)) return 'document';
  if (archiveExts.includes(ext)) return 'archive';
  if (codeExts.includes(ext)) return 'code';
  
  return 'other';
};

/**
 * 格式化日期
 */
export const formatDate = (date: string | Date): string => {
  const d = new Date(date);
  const now = new Date();
  const diff = now.getTime() - d.getTime();
  
  // 小于1分钟
  if (diff < 60000) {
    return '刚刚';
  }
  
  // 小于1小时
  if (diff < 3600000) {
    return `${Math.floor(diff / 60000)}分钟前`;
  }
  
  // 小于1天
  if (diff < 86400000) {
    return `${Math.floor(diff / 3600000)}小时前`;
  }
  
  // 小于7天
  if (diff < 604800000) {
    return `${Math.floor(diff / 86400000)}天前`;
  }
  
  // 格式化为日期
  return d.toLocaleDateString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  });
};

/**
 * 解析路径为面包屑导航
 */
export const parseBreadcrumb = (path: string) => {
  if (!path || path === '/') {
    return [{ name: '我的文件', path: '/' }];
  }
  
  const parts = path.split('/').filter(Boolean);
  const breadcrumbs = [{ name: '我的文件', path: '/' }];
  
  let currentPath = '';
  parts.forEach((part) => {
    currentPath += `/${part}`;
    breadcrumbs.push({ name: part, path: currentPath });
  });
  
  return breadcrumbs;
};

/**
 * 验证文件名
 */
export const validateFileName = (name: string): { valid: boolean; message?: string } => {
  if (!name || name.trim() === '') {
    return { valid: false, message: '文件名不能为空' };
  }
  
  const invalidChars = /[<>:"/\\|?*]/;
  if (invalidChars.test(name)) {
    return { valid: false, message: '文件名包含非法字符' };
  }
  
  if (name.includes('..')) {
    return { valid: false, message: '文件名不能包含 ..' };
  }
  
  return { valid: true };
};

/**
 * 排序文件列表
 */
export const sortFiles = <T extends { name: string; capacity: number; is_dir: boolean }>(
  files: T[],
  sortBy: 'name' | 'size' | 'date',
  order: 'asc' | 'desc' = 'asc'
): T[] => {
  const sorted = [...files].sort((a, b) => {
    // 文件夹始终在前面
    if (a.is_dir && !b.is_dir) return -1;
    if (!a.is_dir && b.is_dir) return 1;
    
    let result = 0;
    
    switch (sortBy) {
      case 'name':
        result = a.name.localeCompare(b.name, 'zh-CN');
        break;
      case 'size':
        result = a.capacity - b.capacity;
        break;
      case 'date':
        // 如果有 created_at 字段，按日期排序
        // result = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
        result = a.name.localeCompare(b.name, 'zh-CN');
        break;
      default:
        result = 0;
    }
    
    return order === 'asc' ? result : -result;
  });
  
  return sorted;
};

/**
 * 过滤文件
 */
export const filterFiles = <T extends { name: string }>(
  files: T[],
  keyword: string
): T[] => {
  if (!keyword || keyword.trim() === '') {
    return files;
  }
  
  const lowerKeyword = keyword.toLowerCase();
  return files.filter((file) => file.name.toLowerCase().includes(lowerKeyword));
};

/**
 * 下载文件（通用方法）
 */
export const downloadBlob = (blob: Blob, filename: string) => {
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  window.URL.revokeObjectURL(url);
};
