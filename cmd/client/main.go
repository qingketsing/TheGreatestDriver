package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	return c.uploadFileObjectWithParent(fo, meta, 0)
}

// uploadFileObjectWithParent 上传文件对象到服务器，支持指定父节点ID
func (c *Client) uploadFileObjectWithParent(fo *shared.FileObject, meta *shared.MetaData, parentID int64) error {
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
	_ = writer.WriteField("parent_id", fmt.Sprintf("%d", parentID))

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

// UploadFileTree 递归上传整个文件树
// 返回上传的文件数量和错误
func (c *Client) UploadFileTree(rootPath string) (int, error) {
	// 读取文件树结构
	tree, _, err := shared.ReadFileTree(rootPath)
	if err != nil {
		return 0, fmt.Errorf("读取文件树失败: %w", err)
	}

	// 递归上传，从根节点（parentID=0）开始
	count := 0
	if err := c.uploadTreeNode(tree, 0, &count); err != nil {
		return count, err
	}

	fmt.Printf("\n文件树上传完成，共上传 %d 个文件/目录\n", count)
	return count, nil
}

// uploadTreeNode 递归上传树节点
func (c *Client) uploadTreeNode(node *shared.FileTree, parentID int64, count *int) error {
	// 准备上传当前节点
	uploadURL := c.BaseURL + "/upload"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 构造元数据
	meta := shared.MetaData{
		Name:     node.Name,
		Capacity: node.Capacity,
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	// 如果是文件，上传文件内容
	if !node.IsDir && node.Fileobj != nil {
		part, err := writer.CreateFormFile("file", node.Fileobj.Name)
		if err != nil {
			return err
		}
		if _, err = part.Write(node.Fileobj.Content); err != nil {
			writer.Close()
			return err
		}
	}

	// 添加元数据和父节点ID
	_ = writer.WriteField("meta", string(metaJSON))
	_ = writer.WriteField("parent_id", fmt.Sprintf("%d", parentID))
	_ = writer.WriteField("is_dir", fmt.Sprintf("%t", node.IsDir))

	if err = writer.Close(); err != nil {
		return err
	}

	// 发送请求
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
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("上传 %s 失败，状态码: %d, 响应: %s", node.Name, resp.StatusCode, string(bodyBytes))
	}

	// 解析响应获取新节点的ID
	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("解析响应失败: %w", err)
	}

	*count++
	if node.IsDir {
		fmt.Printf("📁 目录 %s 已创建 (ID: %d)\n", node.Name, result.ID)
	} else {
		fmt.Printf("📄 文件 %s 已上传 (ID: %d)\n", node.Name, result.ID)
	}

	// 如果是目录，递归上传子节点
	if node.IsDir && len(node.Children) > 0 {
		for i := range node.Children {
			if err := c.uploadTreeNode(&node.Children[i], result.ID, count); err != nil {
				return fmt.Errorf("上传子节点 %s 失败: %w", node.Children[i].Name, err)
			}
		}
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

func (c *Client) downloadFileObject(name string) error {
	downloadURL := fmt.Sprintf("%s/download?name=%s", c.BaseURL, url.QueryEscape(name))
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: server returned %d", resp.StatusCode)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Printf("文件 %s 下载成功，大小 %d 字节\n", name, len(data))
	// 保存到本地download文件夹
	storageDir := "./download"
	absDir, err := filepath.Abs(storageDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(absDir, os.ModePerm); err != nil {
		return err
	}

	destPath := filepath.Join(absDir, name)
	if err := os.WriteFile(destPath, data, 0644); err != nil {
		return err
	}
	fmt.Printf("文件 %s 已保存到本地路径 %s\n", name, destPath)
	return nil
}

func main() {
	// 创建客户端实例
	client := NewClient("")

	// 示例1: 上传单个文件
	fmt.Println("=== 示例1: 上传单个文件 ===")
	p, _ := filepath.Abs("test/app.js")

	// 检查是文件还是目录
	info, err := os.Stat(p)
	if err != nil {
		fmt.Printf("无法访问路径: %v\n", err)
		return
	}

	if info.IsDir() {
		// 如果是目录，使用文件树上传
		fmt.Printf("检测到目录: %s，开始递归上传...\n", p)
		count, err := client.UploadFileTree(p)
		if err != nil {
			fmt.Printf("文件树上传失败: %v\n", err)
			return
		}
		fmt.Printf("✓ 成功上传 %d 个文件/目录\n", count)
	} else {
		// 如果是文件，使用单文件上传
		fo, meta, err := shared.NewFileObject(p)
		if err != nil {
			fmt.Printf("创建文件对象失败: %v\n", err)
			return
		}

		fmt.Printf("文件对象创建成功: %+v\n", meta)

		// 存储文件到本地
		if err := client.StoreFileObject(fo); err != nil {
			fmt.Printf("本地存储失败: %v\n", err)
			return
		}
		fmt.Println("✓ 文件存储成功")

		// 上传文件到服务器
		if err := client.UploadFileObject(fo, meta); err != nil {
			fmt.Printf("上传失败: %v\n", err)
			return
		}
		fmt.Println("✓ 文件上传成功")
	}

	// 显示服务器上的所有文件
	fmt.Printf("\n服务器文件列表: %+v\n", client.Metas)

	// 示例2: 上传整个目录树（如果你想测试）
	// 取消下面的注释来测试目录上传
	/*
		fmt.Println("\n=== 示例2: 上传整个目录 ===")
		testDir := "test"  // 修改为你想上传的目录
		count, err := client.UploadFileTree(testDir)
		if err != nil {
			fmt.Printf("目录上传失败: %v\n", err)
			return
		}
		fmt.Printf("✓ 成功上传目录，共 %d 个文件/目录\n", count)
	*/

	// 下载文件示例
	fmt.Println("\n=== 下载文件示例 ===")
	if err := client.downloadFileObject("app.js"); err != nil {
		fmt.Printf("下载失败: %v\n", err)
	}

	// 删除文件示例
	fmt.Println("\n=== 删除文件示例 ===")
	if err := client.DeleteFile("app.js"); err != nil {
		fmt.Printf("删除失败: %v\n", err)
	} else {
		fmt.Println("✓ 文件删除成功")
	}

	fmt.Printf("\n最终服务器文件列表: %+v\n", client.Metas)
}
