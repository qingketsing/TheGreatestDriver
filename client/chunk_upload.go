package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	serverURL = "http://localhost:8000"
	chunkSize = 1024 * 1024 // 1MB per chunk
)

// QuickUploadResponse 秒传响应
type QuickUploadResponse struct {
	Message    string `json:"message"`
	ExistingID int64  `json:"existing_id"`
	NeedUpload bool   `json:"needUpload"`
	UploadID   string `json:"uploadId"`
	UploadURL  string `json:"uploadUrl"`
}

// ProgressResponse 进度响应
type ProgressResponse struct {
	UploadID       string  `json:"uploadId"`
	Status         string  `json:"status"`
	ReceivedChunks int     `json:"receivedChunks"`
	TotalChunks    int     `json:"totalChunks"`
	ReceivedBytes  int64   `json:"receivedBytes"`
	TotalBytes     int64   `json:"totalBytes"`
	Percent        float64 `json:"percent"`
	Path           string  `json:"path"`
	FileName       string  `json:"fileName"`
}

// calculateFileHash 计算文件的 SHA256 哈希
func calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// tryQuickUpload 尝试秒传
func tryQuickUpload(fileName, fileHash, targetPath string, fileSize int64) (*QuickUploadResponse, error) {
	data := map[string]string{
		"fileName":  fileName,
		"fileHash":  fileHash,
		"path":      targetPath,
		"totalSize": fmt.Sprintf("%d", fileSize),
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	for key, val := range data {
		_ = writer.WriteField(key, val)
	}
	writer.Close()

	resp, err := http.Post(serverURL+"/upload/quick", writer.FormDataContentType(), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result QuickUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// uploadChunk 上传单个分片
func uploadChunk(uploadID, fileName, fileHash, targetPath string, chunkIndex, totalChunks int, totalSize int64, chunkData []byte) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加表单字段
	_ = writer.WriteField("uploadId", uploadID)
	_ = writer.WriteField("fileName", fileName)
	_ = writer.WriteField("fileHash", fileHash)
	_ = writer.WriteField("totalChunks", fmt.Sprintf("%d", totalChunks))
	_ = writer.WriteField("chunkIndex", fmt.Sprintf("%d", chunkIndex))
	_ = writer.WriteField("totalSize", fmt.Sprintf("%d", totalSize))
	_ = writer.WriteField("path", targetPath)

	// 添加分片文件
	part, err := writer.CreateFormFile("chunk", fmt.Sprintf("chunk_%d", chunkIndex))
	if err != nil {
		return err
	}
	if _, err := part.Write(chunkData); err != nil {
		return err
	}
	writer.Close()

	resp, err := http.Post(serverURL+"/upload/chunk", writer.FormDataContentType(), body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload chunk failed: %s", string(bodyBytes))
	}

	return nil
}

// getUploadProgress 获取上传进度
func getUploadProgress(uploadID string) (*ProgressResponse, error) {
	resp, err := http.Get(fmt.Sprintf("%s/upload/progress/%s", serverURL, uploadID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 检查状态码，404表示会话不存在（已完成或不存在）
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("upload session not found (completed or expired)")
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get progress failed: %s", string(bodyBytes))
	}

	var result ProgressResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// uploadFileInChunks 分块上传文件
func uploadFileInChunks(filePath, targetPath string) error {
	// 1. 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat file: %v", err)
	}

	fileName := filepath.Base(filePath)
	fileSize := fileInfo.Size()

	// 2. 计算文件哈希
	fmt.Printf("计算文件哈希...\n")
	fileHash, err := calculateFileHash(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate hash: %v", err)
	}
	fmt.Printf("文件哈希: %s\n", fileHash)

	// 3. 尝试秒传
	fmt.Printf("尝试秒传...\n")
	quickResp, err := tryQuickUpload(fileName, fileHash, targetPath, fileSize)
	if err != nil {
		return fmt.Errorf("quick upload request failed: %v", err)
	}

	if !quickResp.NeedUpload {
		fmt.Printf("✓ 秒传成功！文件已存在 (ID: %d)\n", quickResp.ExistingID)
		return nil
	}

	uploadID := quickResp.UploadID
	fmt.Printf("需要上传，uploadId: %s\n", uploadID)

	// 4. 计算分片数量
	totalChunks := int((fileSize + chunkSize - 1) / chunkSize)
	fmt.Printf("文件大小: %d bytes, 分片数量: %d\n", fileSize, totalChunks)

	// 5. 逐个上传分片
	file.Seek(0, 0) // 重置文件指针
	for i := 1; i <= totalChunks; i++ {
		// 读取分片数据
		chunkData := make([]byte, chunkSize)
		n, err := file.Read(chunkData)
		if err != nil && err != io.EOF {
			return fmt.Errorf("failed to read chunk %d: %v", i, err)
		}
		chunkData = chunkData[:n]

		// 上传分片
		fmt.Printf("上传分片 %d/%d (%d bytes)...\n", i, totalChunks, n)
		if err := uploadChunk(uploadID, fileName, fileHash, targetPath, i, totalChunks, fileSize, chunkData); err != nil {
			return fmt.Errorf("failed to upload chunk %d: %v", i, err)
		}

		// 查询进度
		progress, err := getUploadProgress(uploadID)
		if err == nil {
			fmt.Printf("  进度: %.2f%% (%d/%d chunks)\n", progress.Percent, progress.ReceivedChunks, progress.TotalChunks)
		}
	}

	// 6. 等待合并完成
	fmt.Printf("等待服务器合并分片...\n")
	for {
		progress, err := getUploadProgress(uploadID)
		if err != nil {
			// 会话已删除，说明合并完成
			break
		}

		fmt.Printf("  状态: %s, 进度: %.2f%%\n", progress.Status, progress.Percent)

		if progress.Status == "done" {
			fmt.Printf("✓ 上传完成！\n")
			break
		} else if progress.Status == "error" {
			return fmt.Errorf("upload failed with error status")
		}

		time.Sleep(500 * time.Millisecond)
	}

	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("用法: go run chunk_upload_test.go <文件路径> [目标路径]")
		fmt.Println("示例: go run chunk_upload_test.go test.txt")
		fmt.Println("示例: go run chunk_upload_test.go test.txt test/data")
		os.Exit(1)
	}

	filePath := os.Args[1]
	targetPath := ""
	if len(os.Args) > 2 {
		targetPath = os.Args[2]
	}

	fmt.Printf("========== 分块上传测试 ==========\n")
	fmt.Printf("文件路径: %s\n", filePath)
	fmt.Printf("目标路径: %s\n", targetPath)
	fmt.Printf("服务器地址: %s\n", serverURL)
	fmt.Printf("==================================\n\n")

	if err := uploadFileInChunks(filePath, targetPath); err != nil {
		fmt.Printf("✗ 上传失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ 测试完成！\n")
}
