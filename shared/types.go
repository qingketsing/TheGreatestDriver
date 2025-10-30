package shared

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// 导出类型，包外可见
type MetaData struct {
	id       int64  // 私有字段，包外不可见
	Name     string `json:"name"`
	Capacity int64  `json:"capacity"`
}

type FileObject struct {
	Name     string `json:"name"`
	Capacity int64  `json:"capacity"`
	Content  []byte `json:"content"`
}

type FileTree struct {
	Name     string      `json:"name"`
	Capacity int64       `json:"capacity"`
	Fileobj  *FileObject `json:"fileobj,omitempty"` // if is a directory, Fileobj is nil
	Children []FileTree  `json:"children,omitempty"`
	IsDir    bool        `json:"is_dir"`
}

// ReadFileTree 读取目录并返回文件树结构
func ReadFileTree(rootPath string) (*FileTree, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, err
	}
	node := &FileTree{
		Name:     info.Name(),
		Capacity: info.Size(),
		IsDir:    info.IsDir(),
	}

	// 如果是目录，递归读取子节点
	if info.IsDir() {
		entries, err := os.ReadDir(rootPath)
		if err != nil {
			return nil, err
		}
		for _, entry := range entries {
			childPath := rootPath + string(os.PathSeparator) + entry.Name()
			childNode, err := ReadFileTree(childPath)
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, *childNode)
		}
		return node, nil
	} else {
		// 如果是文件，读取文件内容
		node.Fileobj, _, err = NewFileObject(rootPath)
		if err != nil {
			return nil, err
		}
		node.Capacity = node.Fileobj.Capacity
		return node, nil
	}
}

func Unzip(zipPath string, destDir string) error {
	// 打开 zip 文件
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %v", err)
	}
	defer reader.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create destination directory: %v", err)
	}

	// 遍历 zip 中的所有文件
	for _, file := range reader.File {
		// 构建目标路径
		destPath := filepath.Join(destDir, file.Name)

		// 安全检查：防止 zip slip 攻击(路径穿越)
		if !strings.HasPrefix(destPath, filepath.Clean(destDir)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", file.Name)
		}

		if file.FileInfo().IsDir() {
			// 创建目录
			if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", destPath, err)
			}
		} else {
			// 创建文件的父目录
			if err := os.MkdirAll(filepath.Dir(destPath), os.ModePerm); err != nil {
				return fmt.Errorf("failed to create parent directory for %s: %v", destPath, err)
			}

			// 解压文件
			if err := extractFile(file, destPath); err != nil {
				return fmt.Errorf("failed to extract file %s: %v", file.Name, err)
			}
		}
	}

	return nil
}

// extractFile 解压单个文件
func extractFile(file *zip.File, destPath string) error {
	// 打开 zip 中的文件
	rc, err := file.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	// 创建目标文件
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 复制内容
	if _, err := io.Copy(destFile, rc); err != nil {
		return err
	}

	// 设置文件权限
	if err := os.Chmod(destPath, file.Mode()); err != nil {
		return err
	}

	return nil
}

// NewFileObject 从路径读取文件并返回共享的类型
func NewFileObject(path string) (*FileObject, *MetaData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, nil, err
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, nil, err
	}
	fo := &FileObject{
		Name:     info.Name(),
		Capacity: info.Size(),
		Content:  data,
	}
	md := &MetaData{
		Name:     info.Name(),
		Capacity: info.Size(),
	}
	return fo, md, nil
}
