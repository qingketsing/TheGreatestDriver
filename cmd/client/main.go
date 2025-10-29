package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
// relativePath 参数指定文件在服务器上的相对路径（相对于 uploads 目录）
func (c *Client) UploadFileObject(fo *shared.FileObject, meta *shared.MetaData, relativePath string) error {
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

	// 使用传入的相对路径
	if relativePath != "" {
		_ = writer.WriteField("path", relativePath)
	}

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

func (c *Client) UploadFileTree(ft *shared.FileTree, basePath string) error {
	// 构造当前节点的完整路径（使用正斜杠以保证跨平台兼容性）
	currentPath := ft.Name
	if basePath != "" {
		currentPath = basePath + "/" + ft.Name
	}

	if ft.IsDir {
		// 目录：先创建目录节点
		createDirURL := fmt.Sprintf("%s/createdir?path=%s", c.BaseURL, url.QueryEscape(currentPath))
		resp, err := http.Post(createDirURL, "application/json", nil)
		if err != nil {
			return fmt.Errorf("创建目录 %s 失败: %v", currentPath, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("创建目录 %s 失败，状态码: %d, 响应: %s", currentPath, resp.StatusCode, string(bodyBytes))
		}

		fmt.Printf("✓ 创建目录: %s\n", currentPath)

		// 递归上传子节点
		for _, child := range ft.Children {
			if err := c.UploadFileTree(&child, currentPath); err != nil {
				return err
			}
		}
	} else {
		// 文件：使用 FileObject 上传
		if ft.Fileobj == nil {
			return fmt.Errorf("文件 %s 的 Fileobj 为空", currentPath)
		}

		meta := &shared.MetaData{
			Name:     ft.Fileobj.Name,
			Capacity: ft.Fileobj.Capacity,
		}

		// 传递文件所在的目录路径（不包括文件名本身）
		if err := c.UploadFileObject(ft.Fileobj, meta, basePath); err != nil {
			return fmt.Errorf("上传文件 %s 失败: %v", currentPath, err)
		}
		fmt.Printf("✓ 上传文件: %s (%d 字节)\n", currentPath, ft.Capacity)
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

func (c *Client) DownloadFileTree(dirName string) error {
	downloadURL := fmt.Sprintf("%s/downloaddir?dirname=%s", c.BaseURL, url.QueryEscape(dirName))
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
	// 下载下来的文件是zip，所以需要解压
	zipPath := "./download/" + dirName + ".zip"
	if err := os.WriteFile(zipPath, data, 0644); err != nil {
		return err
	}
	fmt.Printf("目录 %s 下载成功，已保存为 %s\n", dirName, zipPath)
	// 解压zip文件
	if err := shared.Unzip(zipPath, "./download/"+dirName); err != nil {
		return err
	}
	defer os.Remove(zipPath) // 删除zip文件
	fmt.Printf("目录 %s 已解压到 ./download/%s\n", dirName, dirName)

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
	// 使用空字符串表示上传到根目录
	if err := client.UploadFileObject(fo, meta, ""); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("文件上传成功")

	// 显示服务器上的所有文件
	fmt.Printf("\n服务器文件列表: %+v\n", client.Metas)

	// 下载文件示例
	if err := client.downloadFileObject("app.js"); err != nil {
		fmt.Print(err)
		return
	}

	// 删除文件示例
	if err := client.DeleteFile("app.js"); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("文件删除成功")

	fmt.Printf("\n服务器文件列表: %+v\n", client.Metas)

	// 递归上传目录示例
	ft, err := shared.ReadFileTree("test")
	if err != nil {
		log.Fatal(err)
	}
	err = client.UploadFileTree(ft, "") // 递归上传 test/ 目录
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("目录树上传成功")

	//下载文件夹示例
	err = client.DownloadFileTree("test")
	if err != nil {
		log.Fatal(err)
	}

	// 删除目录示例
	err = client.DeleteFile("test")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("目录删除成功")

	fmt.Printf("\n服务器文件列表: %+v\n", client.Metas)

}
