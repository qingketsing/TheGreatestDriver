package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"single_drive/shared"
)

func storeFileObject(fo *shared.FileObject) error {
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
	return nil
}

func uploadFileObjectHTTP(fo *shared.FileObject, meta *shared.MetaData) error {
	// 修改为本地测试地址，部署时改为服务器IP
	uploadURL := "http://139.196.15.66:8000/upload"

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
	return nil
}

func GetFileList() ([]shared.MetaData, error) {
	var fileList []shared.MetaData

	// 支持通过环境变量指定服务器地址
	base := os.Getenv("UPLOAD_URL")
	if base == "" {
		base = "http://139.196.15.66:8000"
	}

	resp, err := http.Get(base + "/list")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&fileList); err != nil {
		return nil, err
	}

	return fileList, nil
}

func main() {
	p, _ := filepath.Abs("test/app.js")
	fo, meta, err := shared.NewFileObject(p)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Printf("文件对象创建成功: %+v\n", meta)
	if err := storeFileObject(fo); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("文件存储成功")
	if err := uploadFileObjectHTTP(fo, meta); err != nil {
		fmt.Print(err)
		return
	}
	fmt.Println("文件上传成功")

	items, err := GetFileList()
	if err != nil {
		fmt.Println("GetFileList error:", err)
	} else {
		fmt.Printf("server items: %+v\n", items)
	}
}
