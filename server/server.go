package server

import (
	"archive/zip"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"single_drive/shared"
	"strings"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

type Server struct {
	uploadDir string
	host      string
	DB        *sql.DB
	Metalist  []shared.MetaData
	Ge        *gin.Engine
}

func (s *Server) SetupDefaultSql() {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=329426 dbname=tododb sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		log.Fatal(err)
	}
	log.Println("数据库连接成功")

	// 确保 drivelist 表存在，若不存在则自动创建
	createTbl := `CREATE TABLE IF NOT EXISTS drivelist (
		id SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		capacity BIGINT NOT NULL,
		created_at TIMESTAMPTZ DEFAULT now()
	)`
	if _, err := db.Exec(createTbl); err != nil {
		db.Close()
		log.Fatalf("failed to create drivelist table: %v", err)
	}
	log.Println("确保表 drivelist 存在")

	// 创建闭包表（Closure Table）用于存储文件树结构
	createClosureTbl := `CREATE TABLE IF NOT EXISTS drivelist_closure (
    ancestor INTEGER NOT NULL,
    descendant INTEGER NOT NULL,
    depth INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (ancestor, descendant),
    FOREIGN KEY (ancestor) REFERENCES drivelist(id) ON DELETE CASCADE,
    FOREIGN KEY (descendant) REFERENCES drivelist(id) ON DELETE CASCADE
	)`
	if _, err := db.Exec(createClosureTbl); err != nil {
		db.Close()
		log.Fatalf("failed to create drivelist_closure table: %v", err)
	}
	log.Println("确保表 drivelist_closure 存在")

	// 创建索引以优化闭包表查询性能
	createClosureIndexes := `
		CREATE INDEX IF NOT EXISTS idx_closure_ancestor ON drivelist_closure(ancestor);
		CREATE INDEX IF NOT EXISTS idx_closure_descendant ON drivelist_closure(descendant);
		CREATE INDEX IF NOT EXISTS idx_closure_depth ON drivelist_closure(depth);
	`
	if _, err := db.Exec(createClosureIndexes); err != nil {
		log.Printf("warning: failed to create some closure indexes: %v", err)
	} else {
		log.Println("确保 drivelist_closure 索引存在")
	}

	s.DB = db
	s.Metalist = s.ReadItemsFromDB(db)
}

func (s *Server) ReadItemsFromDB(db *sql.DB) []shared.MetaData {
	rows, err := db.Query("SELECT name , capacity FROM drivelist")
	if err != nil {
		log.Fatal(err)
	}
	var metalist []shared.MetaData
	for rows.Next() {
		var name string
		var capacity int64
		if err := rows.Scan(&name, &capacity); err != nil {
			log.Fatal(err)
		}
		metalist = append(metalist, shared.MetaData{
			Name:     name,
			Capacity: capacity,
		})
	}
	s.Metalist = metalist
	return metalist
}

// 路由处理器方法

func (s *Server) handleIndex(c *gin.Context) {
	c.String(http.StatusOK, "Hello! This is the Single Drive server.")
}

func (s *Server) handleUpload(c *gin.Context) {
	// 从表单中获取 "meta" 字段
	metaJSON := c.PostForm("meta")
	var meta shared.MetaData
	if err := json.Unmarshal([]byte(metaJSON), &meta); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid meta data: " + err.Error()})
		return
	}

	// 从表单中获取文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not provided: " + err.Error()})
		return
	}

	// 获取可选的路径字段（如 "test/data"）
	userPath := c.PostForm("path")
	// 基本安全检查：不允许绝对路径或上级引用
	if userPath != "" {
		if strings.Contains(userPath, "..") || strings.HasPrefix(userPath, "/") || strings.HasPrefix(userPath, "\\") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
			return
		}
		// 清理路径（去除多余分隔符）
		userPath = filepath.Clean(userPath)
	}

	// 定义服务器上的存储目录（uploads/<userPath>）
	destDir := s.uploadDir
	if userPath != "" && userPath != "." {
		destDir = filepath.Join(s.uploadDir, userPath)
	}
	// 确保目标目录存在
	if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage directory: " + err.Error()})
		return
	}

	// 将文件保存到服务器的目标路径
	destPath := filepath.Join(destDir, file.Filename)
	if err := c.SaveUploadedFile(file, destPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	// 将元数据存入数据库：存在则更新 capacity，否则插入新记录
	var existingID int
	err = s.DB.QueryRow("SELECT id FROM drivelist WHERE name=$1", meta.Name).Scan(&existingID)
	if err != nil {
		if err == sql.ErrNoRows {
			// 不存在，插入新记录
			var newID int64
			err = s.DB.QueryRow("INSERT INTO drivelist (name, capacity) VALUES ($1, $2) RETURNING id",
				meta.Name, meta.Capacity).Scan(&newID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert metadata: " + err.Error()})
				return
			}

			// 插入闭包表记录：自己到自己 (depth=0)
			_, err = s.DB.Exec("INSERT INTO drivelist_closure (ancestor, descendant, depth) VALUES ($1, $1, 0)", newID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert closure for file: " + err.Error()})
				return
			}

			// 如果文件在子目录中，建立与父目录的闭包关系
			if userPath != "" && userPath != "." {
				var parentID int64
				err = s.DB.QueryRow("SELECT id FROM drivelist WHERE name=$1", userPath).Scan(&parentID)
				if err == nil {
					// 复制父节点的所有祖先关系
					_, err = s.DB.Exec(`
						INSERT INTO drivelist_closure (ancestor, descendant, depth)
						SELECT ancestor, $1, depth + 1
						FROM drivelist_closure
						WHERE descendant = $2
					`, newID, parentID)
					if err != nil {
						log.Printf("Warning: failed to link file to parent closure: %v", err)
					}
				}
			}
		} else {
			// 查询出错
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check metadata: " + err.Error()})
			return
		}
	} else {
		// 记录已存在，更新容量（capacity）字段
		_, err = s.DB.Exec("UPDATE drivelist SET capacity=$1 WHERE id=$2", meta.Capacity, existingID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update metadata: " + err.Error()})
			return
		}
	}

	// 打印日志并返回成功响应
	fmt.Printf("File '%s' received and saved to '%s'. Meta: %+v\n", file.Filename, destPath, meta)
	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": file.Filename,
		"path":     destPath,
	})
}

func (s *Server) handleList(c *gin.Context) {
	// 检查是否请求简单列表格式（用于向后兼容）
	format := c.Query("format")
	if format == "simple" || format == "flat" {
		// 返回简单的数组格式
		items := s.ReadItemsFromDB(s.DB)
		c.JSON(http.StatusOK, items)
		return
	}

	// 默认返回树形结构
	tree, err := s.buildFileTree()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, tree)
}

// TreeNode 表示文件树的一个节点
type TreeNode struct {
	ID       int64       `json:"id"`
	Name     string      `json:"name"`
	Capacity int64       `json:"capacity"`
	IsDir    bool        `json:"is_dir"`
	Path     string      `json:"path"`
	Children []*TreeNode `json:"children,omitempty"`
}

// buildFileTree 从数据库构建文件树
func (s *Server) buildFileTree() (map[string]interface{}, error) {
	// 1. 获取所有节点
	rows, err := s.DB.Query("SELECT id, name, capacity FROM drivelist ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodeMap := make(map[int64]*TreeNode)
	var allNodes []*TreeNode

	for rows.Next() {
		var id int64
		var name string
		var capacity int64
		if err := rows.Scan(&id, &name, &capacity); err != nil {
			return nil, err
		}

		node := &TreeNode{
			ID:       id,
			Name:     name,
			Capacity: capacity,
			IsDir:    capacity == 0, // 容量为0表示目录
			Path:     name,
			Children: []*TreeNode{},
		}
		nodeMap[id] = node
		allNodes = append(allNodes, node)
	}

	// 2. 获取父子关系 (depth=1 表示直接父子关系)
	rows, err = s.DB.Query(`
		SELECT ancestor, descendant 
		FROM drivelist_closure 
		WHERE depth = 1
		ORDER BY ancestor, descendant
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	parentChildMap := make(map[int64][]int64)
	for rows.Next() {
		var ancestor, descendant int64
		if err := rows.Scan(&ancestor, &descendant); err != nil {
			return nil, err
		}
		parentChildMap[ancestor] = append(parentChildMap[ancestor], descendant)
	}

	// 3. 构建树结构
	var rootNodes []*TreeNode
	for _, node := range allNodes {
		children := parentChildMap[node.ID]
		if len(children) > 0 {
			// 有子节点
			for _, childID := range children {
				if childNode, ok := nodeMap[childID]; ok {
					node.Children = append(node.Children, childNode)
				}
			}
		}

		// 查找该节点是否有父节点
		hasParent := false
		for _, childIDs := range parentChildMap {
			for _, childID := range childIDs {
				if childID == node.ID {
					hasParent = true
					break
				}
			}
			if hasParent {
				break
			}
		}

		// 如果没有父节点，则为根节点
		if !hasParent {
			rootNodes = append(rootNodes, node)
		}
	}

	// 4. 返回树结构
	result := map[string]interface{}{
		"total": len(allNodes),
		"roots": rootNodes,
	}

	return result, nil
}

func (s *Server) handleDebugDrivelist(c *gin.Context) {
	rows, err := s.DB.Query("SELECT id, name, capacity, created_at FROM drivelist ORDER BY id")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var id int64
		var name string
		var capacity int64
		var createdAt string
		if err := rows.Scan(&id, &name, &capacity, &createdAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, map[string]interface{}{
			"id":         id,
			"name":       name,
			"capacity":   capacity,
			"created_at": createdAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"count": len(items), "items": items})
}

func (s *Server) handleDebugClosure(c *gin.Context) {
	rows, err := s.DB.Query(`
		SELECT 
			c.ancestor,
			c.descendant,
			c.depth,
			d1.name as ancestor_name,
			d2.name as descendant_name,
			d2.capacity as descendant_capacity
		FROM drivelist_closure c
		JOIN drivelist d1 ON c.ancestor = d1.id
		JOIN drivelist d2 ON c.descendant = d2.id
		ORDER BY c.ancestor, c.depth, c.descendant
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var ancestor, descendant int64
		var depth int
		var ancestorName, descendantName string
		var descendantCapacity int64
		if err := rows.Scan(&ancestor, &descendant, &depth, &ancestorName, &descendantName, &descendantCapacity); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, map[string]interface{}{
			"ancestor":            ancestor,
			"descendant":          descendant,
			"depth":               depth,
			"ancestor_name":       ancestorName,
			"descendant_name":     descendantName,
			"descendant_capacity": descendantCapacity,
		})
	}
	c.JSON(http.StatusOK, gin.H{"count": len(items), "items": items})
}

func (s *Server) handleDebugSubtree(c *gin.Context) {
	id := c.Param("id")
	rows, err := s.DB.Query(`
		SELECT 
			d.id,
			d.name,
			d.capacity,
			c.depth
		FROM drivelist d
		JOIN drivelist_closure c ON d.id = c.descendant
		WHERE c.ancestor = $1
		ORDER BY c.depth, d.id
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var items []map[string]interface{}
	for rows.Next() {
		var nodeID int64
		var name string
		var capacity int64
		var depth int
		if err := rows.Scan(&nodeID, &name, &capacity, &depth); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, map[string]interface{}{
			"id":       nodeID,
			"name":     name,
			"capacity": capacity,
			"depth":    depth,
		})
	}
	c.JSON(http.StatusOK, gin.H{"root_id": id, "count": len(items), "items": items})
}

func (s *Server) handleDelete(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' query parameter"})
		return
	}

	// 开始事务
	tx, err := s.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction: " + err.Error()})
		return
	}

	// 查询文件的 ID
	var fileID int64
	err = tx.QueryRow("SELECT id FROM drivelist WHERE name=$1", name).Scan(&fileID)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found in database"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query file: " + err.Error()})
		return
	}

	// 删除数据库中的记录（CASCADE 会自动删除 drivelist_closure 中的相关记录）
	result, err := tx.Exec("DELETE FROM drivelist WHERE id=$1", fileID)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record: " + err.Error()})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	// 删除文件对象
	filePath := filepath.Join(s.uploadDir, name)
	if err := os.Remove(filePath); err != nil {
		// 文件系统删除失败，但数据库已删除
		c.JSON(http.StatusOK, gin.H{
			"message": "Database record deleted, but file removal failed: " + err.Error(),
			"warning": true,
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	c.JSON(http.StatusOK, gin.H{
		"message":       "File and record deleted successfully",
		"rows_affected": rowsAffected,
	})
}

func (s *Server) handleDeleteDir(c *gin.Context) {
	dirname := c.Query("dirname")
	if dirname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'dirname' query parameter"})
		return
	}

	// 安全检查：防止路径穿越
	if strings.Contains(dirname, "..") || strings.HasPrefix(dirname, "/") || strings.HasPrefix(dirname, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid dirname"})
		return
	}
	cleanName := filepath.Clean(dirname)

	dirPath := filepath.Join(s.uploadDir, cleanName)

	// 检查目录是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}

	// 开始数据库事务
	tx, err := s.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction: " + err.Error()})
		return
	}

	// 首先获取该目录的 ID
	var dirID int64
	err = tx.QueryRow("SELECT id FROM drivelist WHERE name=$1", cleanName).Scan(&dirID)
	if err != nil {
		tx.Rollback()
		if err == sql.ErrNoRows {
			// 数据库中没有记录，但文件系统有目录，仍然删除文件系统
			if err := os.RemoveAll(dirPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete directory: " + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Directory deleted (no DB record found)"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query directory: " + err.Error()})
		return
	}

	// 删除该目录及其所有后代节点（利用闭包表）
	// CASCADE 会自动删除 drivelist_closure 中的相关记录
	_, err = tx.Exec(`
        DELETE FROM drivelist
        WHERE id IN (
            SELECT descendant FROM drivelist_closure WHERE ancestor = $1
        )
    `, dirID)
	// 删除
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete DB records: " + err.Error()})
		return
	}

	// 提交事务
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	// 删除文件系统中的目录（在数据库操作成功后）
	if err := os.RemoveAll(dirPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB updated but failed to delete directory: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Directory and its contents deleted successfully",
		"path":    cleanName,
	})
}

func (s *Server) DownloadZip(c *gin.Context, dirPath string, zipName string) error {
	// 创建临时 zip 文件
	tmpZipPath := filepath.Join(os.TempDir(), fmt.Sprintf("download_%s.zip", zipName))

	// 创建 zip 文件
	zipFile, err := os.Create(tmpZipPath)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer zipFile.Close()
	defer os.Remove(tmpZipPath) // 发送完后删除临时文件

	zipWriter := zip.NewWriter(zipFile)

	// 递归添加目录中的所有文件到 zip
	err = s.addFilesToZip(zipWriter, dirPath, "")
	if err != nil {
		zipWriter.Close()
		return fmt.Errorf("failed to add files to zip: %v", err)
	}

	// 关闭 zip writer
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close zip: %v", err)
	}

	// 重新打开文件用于发送
	zipFile.Close()

	// 发送 zip 文件
	c.FileAttachment(tmpZipPath, zipName+".zip")
	return nil
}

// addFilesToZip 递归地将目录中的文件添加到 zip
func (s *Server) addFilesToZip(zipWriter *zip.Writer, sourcePath string, baseInZip string) error {
	// 读取目录内容
	entries, err := os.ReadDir(sourcePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(sourcePath, entry.Name())
		zipPath := filepath.Join(baseInZip, entry.Name())

		// 将路径转换为正斜杠(zip 标准)
		zipPath = filepath.ToSlash(zipPath)

		if entry.IsDir() {
			// 递归处理子目录
			if err := s.addFilesToZip(zipWriter, fullPath, zipPath); err != nil {
				return err
			}
		} else {
			// 添加文件到 zip
			if err := s.addFileToZip(zipWriter, fullPath, zipPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// addFileToZip 添加单个文件到 zip
func (s *Server) addFileToZip(zipWriter *zip.Writer, filePath string, nameInZip string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}
	header.Name = nameInZip
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}

func (s *Server) handleDownload(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' query parameter"})
		return
	}
	filePath := filepath.Join(s.uploadDir, name)
	// 使用 ReadFileTree 读取文件树结构并返回
	fileTree, err := shared.ReadFileTree(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file tree: " + err.Error()})
		return
	}
	if fileTree.IsDir {
		// 压缩文件夹，并返回下载
		if err := s.DownloadZip(c, filePath, filepath.Base(name)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download zip: " + err.Error()})
			return
		}
	} else {
		// 返回文件内容
		c.FileAttachment(filePath, filepath.Base(name))
	}
}

func (s *Server) handleDownloadDir(c *gin.Context) {
	dirname := c.Query("dirname")
	if dirname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'dirname' query parameter"})
		return
	}

	dirPath := filepath.Join(s.uploadDir, dirname)

	// 检查目录是否存在
	fileInfo, err := os.Stat(dirPath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Directory not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stat directory: " + err.Error()})
		return
	}

	if !fileInfo.IsDir() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Path is not a directory"})
		return
	}

	// 压缩目录并返回
	if err := s.DownloadZip(c, dirPath, dirname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create zip: " + err.Error()})
		return
	}
}

func (s *Server) handleCreateDir(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'path' query parameter"})
		return
	}

	// 安全检查：防止路径穿越
	if strings.Contains(path, "..") || strings.HasPrefix(path, "/") || strings.HasPrefix(path, "\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
		return
	}

	// 清理路径
	path = filepath.Clean(path)

	// 在文件系统中创建实际目录
	fullPath := filepath.Join(s.uploadDir, path)
	if err := os.MkdirAll(fullPath, os.ModePerm); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory on filesystem: " + err.Error()})
		return
	}

	// 数据库操作：使用完整路径作为名称
	// 1. 先检查路径是否已存在
	var existingID int64
	err := s.DB.QueryRow("SELECT id FROM drivelist WHERE name=$1", path).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{"message": "Directory already exists", "id": existingID})
		return
	}

	// 2. 插入目录节点（容量为0表示目录）
	var newID int64
	err = s.DB.QueryRow("INSERT INTO drivelist (name, capacity) VALUES ($1, 0) RETURNING id",
		path).Scan(&newID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory in database: " + err.Error()})
		return
	}

	// 3. 维护闭包表关系
	// 插入自己到自己 (depth=0)
	_, err = s.DB.Exec("INSERT INTO drivelist_closure (ancestor, descendant, depth) VALUES ($1, $1, 0)", newID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert closure: " + err.Error()})
		return
	}

	// 4. 如果有父路径，建立父子关系
	parentPath := filepath.Dir(path)
	if parentPath != "." && parentPath != "/" && parentPath != "\\" {
		var parentID int64
		err = s.DB.QueryRow("SELECT id FROM drivelist WHERE name=$1", parentPath).Scan(&parentID)
		if err == nil {
			// 复制父节点的所有祖先关系
			_, err = s.DB.Exec(`
				INSERT INTO drivelist_closure (ancestor, descendant, depth)
				SELECT ancestor, $1, depth + 1
				FROM drivelist_closure
				WHERE descendant = $2
			`, newID, parentID)
			if err != nil {
				log.Printf("Warning: failed to link parent closure: %v", err)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Directory created successfully",
		"id":      newID,
		"path":    path,
	})
}

func (s *Server) handleRename(c *gin.Context) {
	// 首先修改文件名字
	oldName := c.Query("oldName")
	newName := c.Query("newName")
	if oldName == "" || newName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'oldName' or 'newName' query parameter"})
		return
	}
	oldPath := filepath.Join(s.uploadDir, oldName)
	newPath := filepath.Join(s.uploadDir, newName)
	if err := os.Rename(oldPath, newPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rename file: " + err.Error()})
		return
	}
	// 然后更新数据库记录
	_, err := s.DB.Exec("UPDATE drivelist SET name=$1 WHERE name=$2", newName, oldName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update database record:" + err.Error()})
		return
	}
	// 对Closure Table不需要额外操作，因为文件ID未变，只有名称变更
	c.JSON(http.StatusOK, gin.H{
		"message":  "File renamed successfully",
		"old_name": oldName,
		"new_name": newName,
	})
}

func (s *Server) handleMove(c *gin.Context) {
	// 把文件从一个目录移动到另一个目录
	// oldpath: 原文件/文件夹路径（如 "folder1/file.txt"）
	// newpath: 新的父目录路径（如 "folder2"）
	oldPath := c.Query("oldpath")
	newParentPath := c.Query("newparent")

	if oldPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'oldpath' query parameter"})
		return
	}

	// 安全检查
	if strings.Contains(oldPath, "..") || strings.Contains(newParentPath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path: contains '..'"})
		return
	}

	// 清理路径
	oldPath = filepath.Clean(oldPath)
	if newParentPath != "" {
		newParentPath = filepath.Clean(newParentPath)
	}

	// 开始数据库事务
	tx, err := s.DB.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction: " + err.Error()})
		return
	}
	defer tx.Rollback() // 如果没有 commit，则回滚

	// 1. 获取要移动的节点 ID
	var nodeID int64
	err = tx.QueryRow("SELECT id FROM drivelist WHERE name=$1", oldPath).Scan(&nodeID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Source file/folder not found in database"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query source: " + err.Error()})
		return
	}

	// 2. 构建新路径（保留文件名）
	fileName := filepath.Base(oldPath)
	var newPath string
	var newParentID int64

	if newParentPath == "" || newParentPath == "." {
		// 移动到根目录
		newPath = fileName
		newParentID = 0 // 标记为根目录
	} else {
		// 移动到指定目录
		newPath = filepath.Join(newParentPath, fileName)

		// 获取新父目录的 ID
		err = tx.QueryRow("SELECT id FROM drivelist WHERE name=$1", newParentPath).Scan(&newParentID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Target parent directory not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query target parent: " + err.Error()})
			return
		}
	}

	// 3. 检查新路径是否已存在
	var existingID int64
	err = tx.QueryRow("SELECT id FROM drivelist WHERE name=$1", newPath).Scan(&existingID)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Target path already exists"})
		return
	} else if err != sql.ErrNoRows {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check target path: " + err.Error()})
		return
	}

	// 4. 移动文件系统中的文件/文件夹
	oldFullPath := filepath.Join(s.uploadDir, oldPath)
	newFullPath := filepath.Join(s.uploadDir, newPath)

	// 确保新父目录存在
	if newParentPath != "" && newParentPath != "." {
		newParentFullPath := filepath.Join(s.uploadDir, newParentPath)
		if err := os.MkdirAll(newParentFullPath, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create target directory: " + err.Error()})
			return
		}
	}

	if err := os.Rename(oldFullPath, newFullPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move file: " + err.Error()})
		return
	}

	// 5. 更新数据库中的路径
	_, err = tx.Exec("UPDATE drivelist SET name=$1 WHERE id=$2", newPath, nodeID)
	if err != nil {
		// 回滚文件系统操作
		os.Rename(newFullPath, oldFullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update database path: " + err.Error()})
		return
	}

	// 6. 更新闭包表关系
	// 6.1 删除与旧父节点的所有祖先关系（保留自己到自己的关系）
	_, err = tx.Exec(`
		DELETE FROM drivelist_closure
		WHERE descendant IN (
			SELECT descendant FROM drivelist_closure WHERE ancestor = $1
		)
		AND ancestor IN (
			SELECT ancestor FROM drivelist_closure 
			WHERE descendant = $1 AND ancestor != descendant
		)
	`, nodeID)
	if err != nil {
		// 回滚文件系统操作
		os.Rename(newFullPath, oldFullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old closure relations: " + err.Error()})
		return
	}

	// 6.2 如果有新父节点，建立新的祖先关系
	if newParentID != 0 {
		// 获取要移动的节点及其所有后代
		_, err = tx.Exec(`
			INSERT INTO drivelist_closure (ancestor, descendant, depth)
			SELECT p.ancestor, c.descendant, p.depth + c.depth + 1
			FROM drivelist_closure p
			CROSS JOIN drivelist_closure c
			WHERE p.descendant = $1
			AND c.ancestor = $2
		`, newParentID, nodeID)
		if err != nil {
			// 回滚文件系统操作
			os.Rename(newFullPath, oldFullPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create new closure relations: " + err.Error()})
			return
		}
	}

	// 7. 如果移动的是目录，需要更新其所有子节点的路径
	// 获取所有后代节点
	rows, err := tx.Query(`
		SELECT d.id, d.name
		FROM drivelist d
		JOIN drivelist_closure c ON d.id = c.descendant
		WHERE c.ancestor = $1 AND c.descendant != $1
		ORDER BY c.depth
	`, nodeID)
	if err != nil {
		os.Rename(newFullPath, oldFullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query descendants: " + err.Error()})
		return
	}
	defer rows.Close()

	// 更新每个子节点的路径
	for rows.Next() {
		var childID int64
		var childOldPath string
		if err := rows.Scan(&childID, &childOldPath); err != nil {
			os.Rename(newFullPath, oldFullPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to scan child: " + err.Error()})
			return
		}

		// 计算新的子路径
		relPath := strings.TrimPrefix(childOldPath, oldPath)
		relPath = strings.TrimPrefix(relPath, string(filepath.Separator))
		childNewPath := filepath.Join(newPath, relPath)

		// 更新子节点路径
		_, err = tx.Exec("UPDATE drivelist SET name=$1 WHERE id=$2", childNewPath, childID)
		if err != nil {
			os.Rename(newFullPath, oldFullPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update child path: " + err.Error()})
			return
		}
	}

	// 8. 提交事务
	if err := tx.Commit(); err != nil {
		// 回滚文件系统操作
		os.Rename(newFullPath, oldFullPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File/folder moved successfully",
		"old_path": oldPath,
		"new_path": newPath,
	})
}

func (s *Server) handleGetInfo(c *gin.Context) {
	filename := c.Query("name")
	filepath := filepath.Join(s.uploadDir, filename)
	info, err := os.Stat(filepath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get file info: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":         info.Name(),
		"size":         info.Size(),
		"mode":         info.Mode().String(),
		"mod_time":     info.ModTime(),
		"is_directory": info.IsDir(),
	})
}

func (s *Server) handleBatchDelete(c *gin.Context) {
	// TODO
}

func (s *Server) handleBatchDownload(c *gin.Context) {
	// TODO
}

func (s *Server) handleSearch(c *gin.Context) {
	// TODO
}

func (s *Server) handleFilterByType(c *gin.Context) {
	// TODO
}

func (s *Server) handleFilterByDate(c *gin.Context) {
	// TODO
}

func (s *Server) handleFilterBySize(c *gin.Context) {
	// TODO
}

func (s *Server) handleChunkUpload(c *gin.Context) {
	// TODO
}

func (s *Server) handleQuickUpload(c *gin.Context) {
	// TODO
}

func (s *Server) handleGetUploadProgress(c *gin.Context) {
	// TODO
}

func (s *Server) SetupDefaultRouter() {
	r := gin.Default()

	// 主要路由
	r.GET("/", s.handleIndex)
	r.POST("/upload", s.handleUpload)
	r.GET("/list", s.handleList)
	r.DELETE("/delete", s.handleDelete)
	r.DELETE("/deletedir", s.handleDeleteDir)
	r.GET("/download", s.handleDownload)
	r.GET("/downloaddir", s.handleDownloadDir)
	r.POST("/createdir", s.handleCreateDir)

	// 重命名文件/目录
	r.PUT("/rename", s.handleRename)
	// 移动文件/目录
	r.PUT("/move", s.handleMove)
	// 获取文件/目录详细信息
	r.GET("/info", s.handleGetInfo)
	// 批量删除
	r.DELETE("/batch-delete", s.handleBatchDelete)
	// 批量下载（打包成zip）
	r.POST("/batch-download", s.handleBatchDownload)

	// 搜索文件
	r.GET("/search", s.handleSearch)
	// 按类型过滤（图片、视频、文档等）
	r.GET("/filter/type", s.handleFilterByType)
	// 按时间范围过滤
	r.GET("/filter/date", s.handleFilterByDate)
	// 按大小过滤
	r.GET("/filter/size", s.handleFilterBySize)

	// 断点续传
	r.POST("/upload/chunk", s.handleChunkUpload)
	// 秒传（文件哈希检查）
	r.POST("/upload/quick", s.handleQuickUpload)
	// 获取上传进度
	r.GET("/upload/progress/:uploadId", s.handleGetUploadProgress)

	// 调试路由
	r.GET("/debug/drivelist", s.handleDebugDrivelist)
	r.GET("/debug/closure", s.handleDebugClosure)
	r.GET("/debug/subtree/:id", s.handleDebugSubtree)

	s.Ge = r
}

func InitServer() *Server {
	s := &Server{}
	s.host = "localhost:8080"
	s.uploadDir = "./uploads"
	// 确保上传目录存在
	if err := os.MkdirAll(s.uploadDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create upload directory: %v", err)
	}
	s.SetupDefaultSql()
	s.SetupDefaultRouter()
	return s
}
