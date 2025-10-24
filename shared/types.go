package shared

import (
	"io"
	"os"
)

// 导出类型，包外可见
type MetaData struct {
	Name     string `json:"name"`
	Capacity int64  `json:"capacity"`
}

type FileObject struct {
	Name     string `json:"name"`
	Capacity int64  `json:"capacity"`
	Content  []byte `json:"content"`
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
