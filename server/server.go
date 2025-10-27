package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"single_drive/shared"
	"strings"

	"database/sql"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

type Server struct {
	DB       *sql.DB
	Metalist []shared.MetaData
	Ge       *gin.Engine
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

func (s *Server) SetupDefaultRouter() {
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello! This is the Single Drive server.")
	})

	r.POST("/upload", func(c *gin.Context) {
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
		uploadDir := "./uploads"
		destDir := uploadDir
		if userPath != "" && userPath != "." {
			destDir = filepath.Join(uploadDir, userPath)
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
				_, err = s.DB.Exec("INSERT INTO drivelist (name, capacity) VALUES ($1, $2)", meta.Name, meta.Capacity)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert metadata: " + err.Error()})
					return
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
	})

	// 新增：列出 drivelist 中的记录，方便检查数据库中是否有数据
	r.GET("/list", func(c *gin.Context) {
		items := s.ReadItemsFromDB(s.DB)
		c.JSON(http.StatusOK, items)
	})

	// 新增：查看 drivelist 表的完整内容（包括 ID）
	r.GET("/debug/drivelist", func(c *gin.Context) {
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
	})

	// 新增：查看 drivelist_closure 表内容（带层级关系）
	r.GET("/debug/closure", func(c *gin.Context) {
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
	})

	// 新增：查看某个节点的子树
	r.GET("/debug/subtree/:id", func(c *gin.Context) {
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
	})

	// 传入的是文件名，通过查询参数 ?name=
	r.DELETE("/delete", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' query parameter"})
			return
		}
		// 删除数据库中的记录
		result, err := s.DB.Exec("DELETE FROM drivelist WHERE name=$1", name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete record: " + err.Error()})
			return
		}
		// 删除文件对象
		uploadDir := "./uploads"
		filePath := filepath.Join(uploadDir, name)
		if err := os.Remove(filePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file: " + err.Error()})
			return
		}
		rowsAffected, _ := result.RowsAffected()
		c.JSON(http.StatusOK, gin.H{
			"message":       "File and record deleted successfully",
			"rows_affected": rowsAffected,
		})
	})
	// 下载, 通过查询参数 ?name=
	r.GET("/download", func(c *gin.Context) {
		name := c.Query("name")
		if name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing 'name' query parameter"})
			return
		}
		uploadDir := "./uploads"
		filePath := filepath.Join(uploadDir, name)
		c.FileAttachment(filePath, name)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

	})

	// 创建目录接口
	r.POST("/createdir", func(c *gin.Context) {
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
		uploadDir := "./uploads"
		fullPath := filepath.Join(uploadDir, path)
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
	})

	s.Ge = r
}

func InitServer() *Server {
	s := &Server{}
	s.SetupDefaultSql()
	s.SetupDefaultRouter()
	return s
}
