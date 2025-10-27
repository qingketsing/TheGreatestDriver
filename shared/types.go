package shared

import (
	"io"
	"os"
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
func ReadFileTree(rootPath string) (*FileTree, MetaData, error) {
	info, err := os.Stat(rootPath)
	if err != nil {
		return nil, MetaData{}, err
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
			return nil, MetaData{}, err
		}
		for _, entry := range entries {
			childPath := rootPath + string(os.PathSeparator) + entry.Name()
			childNode, meta, err := ReadFileTree(childPath)
			if err != nil {
				return nil, meta, err
			}
			node.Children = append(node.Children, *childNode)
		}
		return node, MetaData{}, nil
	} else {
		// 如果是文件，读取文件内容
		var meta *MetaData
		node.Fileobj, meta, err = NewFileObject(rootPath)
		if err != nil {
			return nil, MetaData{}, err
		}
		return node, *meta, nil
	}
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
