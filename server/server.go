package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"single_drive/shared"

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
		ancestor BIGINT NOT NULL,
		descendant BIGINT NOT NULL,
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

		// 获取可选的 parent_id 和 is_dir 参数
		parentIDStr := c.DefaultPostForm("parent_id", "0")
		isDirStr := c.DefaultPostForm("is_dir", "false")

		var parentID int64
		if _, err := fmt.Sscanf(parentIDStr, "%d", &parentID); err != nil {
			parentID = 0
		}

		isDir := isDirStr == "true"

		// 定义服务器上的存储目录
		uploadDir := "./uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage directory: " + err.Error()})
			return
		}

		var destPath string
		var newID int64

		// 开始事务
		tx, err := s.DB.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction: " + err.Error()})
			return
		}
		defer tx.Rollback()

		if !isDir {
			// 处理文件上传
			file, err := c.FormFile("file")
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "File not provided: " + err.Error()})
				return
			}

			// 保存文件到磁盘
			destPath = filepath.Join(uploadDir, file.Filename)
			if err := c.SaveUploadedFile(file, destPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
				return
			}
		}

		// 插入或更新 drivelist 记录
		var existingID int64
		err = tx.QueryRow("SELECT id FROM drivelist WHERE name=$1", meta.Name).Scan(&existingID)
		if err != nil {
			if err == sql.ErrNoRows {
				// 不存在，插入新记录
				err = tx.QueryRow("INSERT INTO drivelist (name, capacity) VALUES ($1, $2) RETURNING id",
					meta.Name, meta.Capacity).Scan(&newID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert metadata: " + err.Error()})
					return
				}
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check metadata: " + err.Error()})
				return
			}
		} else {
			// 记录已存在，更新
			_, err = tx.Exec("UPDATE drivelist SET capacity=$1 WHERE id=$2", meta.Capacity, existingID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update metadata: " + err.Error()})
				return
			}
			newID = existingID
		}

		// 维护闭包表关系
		// 1. 插入自己到自己的记录（depth=0）
		_, err = tx.Exec("INSERT INTO drivelist_closure (ancestor, descendant, depth) VALUES ($1, $1, 0) ON CONFLICT DO NOTHING", newID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert closure self-reference: " + err.Error()})
			return
		}

		// 2. 如果有父节点，复制父节点的所有祖先关系
		if parentID > 0 {
			_, err = tx.Exec(`
				INSERT INTO drivelist_closure (ancestor, descendant, depth)
				SELECT ancestor, $1, depth + 1
				FROM drivelist_closure
				WHERE descendant = $2
				ON CONFLICT DO NOTHING
			`, newID, parentID)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert closure parent relations: " + err.Error()})
				return
			}
		}

		// 提交事务
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
			return
		}

		// 打印日志并返回成功响应
		if isDir {
			fmt.Printf("Directory '%s' created with ID %d (parent: %d)\n", meta.Name, newID, parentID)
		} else {
			fmt.Printf("File '%s' uploaded to '%s' with ID %d (parent: %d). Meta: %+v\n", meta.Name, destPath, newID, parentID, meta)
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "Upload successful",
			"id":       newID,
			"filename": meta.Name,
			"path":     destPath,
			"is_dir":   isDir,
		})
	})

	// 新增：列出 drivelist 中的记录，方便检查数据库中是否有数据
	r.GET("/list", func(c *gin.Context) {
		items := s.ReadItemsFromDB(s.DB)
		c.JSON(http.StatusOK, items)
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

	s.Ge = r
}

func InitServer() *Server {
	s := &Server{}
	s.SetupDefaultSql()
	s.SetupDefaultRouter()
	return s
}
