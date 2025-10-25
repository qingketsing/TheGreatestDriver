package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"single_drive/shared"
)

// Client 客户端结构体，管理文件和元数据
type Client struct {
	Files   []shared.FileObject // 本地文件对象缓存
	Metas   []shared.MetaData   // 服务器端元数据缓存
	BaseURL string              // 服务器地址
}

// NewClient 创建新的客户端实例
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = os.Getenv("UPLOAD_URL")
	}
	if baseURL == "" {
		baseURL = "http://139.196.15.66:8000"
	}
	return &Client{
		BaseURL: baseURL,
		Files:   make([]shared.FileObject, 0),
		Metas:   make([]shared.MetaData, 0),
	}
}

// StoreFileObject 将文件对象存储到本地目录
func (c *Client) StoreFileObject(fo *shared.FileObject) error {
	storageDir := `D:\drivetest`
	absDir, err := filepath.Abs(storageDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(absDir, os.ModePerm); err != nil {
		return err
	}
	destPath := filepath.Join(absDir, fo.Name)

	// write the file content to destination
	if err := os.WriteFile(destPath, fo.Content, 0644); err != nil {
		return err
	}

	// 添加到本地缓存
	c.Files = append(c.Files, *fo)
	return nil
}

// UploadFileObject 上传文件对象到服务器
func (c *Client) UploadFileObject(fo *shared.FileObject, meta *shared.MetaData) error {
	uploadURL := c.BaseURL + "/upload"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fo.Name)
	if err != nil {
		return err
	}

	if _, err = part.Write(fo.Content); err != nil {
		writer.Close()
		return err
	}

	// 序列化 meta 为 JSON 并作为字段上传
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		writer.Close()
		return err
	}
	_ = writer.WriteField("meta", string(metaJSON))
	_ = writer.WriteField("path", "driver_test")

	if err = writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("上传失败，状态码: %d", resp.StatusCode)
	}

	fmt.Printf("文件 %s 已上传到服务器\n", fo.Name)

	// 上传成功后自动刷新服务器端元数据缓存
	if err := c.RefreshMetaList(); err != nil {
		fmt.Printf("warning: 刷新元数据列表失败: %v\n", err)
	}

	return nil
}

// RefreshMetaList 从服务器获取最新的元数据列表并更新缓存
func (c *Client) RefreshMetaList() error {
	resp, err := http.Get(c.BaseURL + "/list")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	var metas []shared.MetaData
	if err := json.NewDecoder(resp.Body).Decode(&metas); err != nil {
		return err
	}

	c.Metas = metas
	fmt.Printf("已刷新元数据列表，共 %d 项\n", len(c.Metas))
	return nil
}

// DeleteFile 删除服务器上的文件
func (c *Client) DeleteFile(filename string) error {
	delURL := fmt.Sprintf("%s/delete?name=%s", c.BaseURL, url.QueryEscape(filename))
	req, err := http.NewRequest(http.MethodDelete, delURL, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete failed: server returned %d", resp.StatusCode)
	}

	// 从本地文件缓存中移除
	var newFiles []shared.FileObject
	for _, f := range c.Files {
		if f.Name != filename {
			newFiles = append(newFiles, f)
		}
	}
	c.Files = newFiles

	fmt.Printf("文件 %s 已从服务器删除\n", filename)

	// 删除成功后自动刷新服务器端元数据缓存
	if err := c.RefreshMetaList(); err != nil {
		fmt.Printf("warning: 刷新元数据列表失败: %v\n", err)
	}

	return nil
}

func main() {
	// 创建客户端实例
	client := NewClient("")
	// 读取测试文件
	p, _ := filepath.Abs("test/app.js")
	fo, meta, err := shared.NewFileObject(p)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("文件对象创建成功: %+v\n", meta)

	// 存储文件到本地
	if err := client.StoreFileObject(fo); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("文件存储成功")

	// 上传文件到服务器（会自动刷新元数据列表）
	if err := client.UploadFileObject(fo, meta); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("文件上传成功")

	// 显示服务器上的所有文件
	fmt.Printf("\n服务器文件列表: %+v\n", client.Metas)

	// 删除文件示例
	if err := client.DeleteFile("app.js"); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("文件删除成功")

	fmt.Printf("\n服务器文件列表: %+v\n", client.Metas)

}
