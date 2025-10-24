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

		// 定义服务器上的存储目录
		// 为了安全和可移植性，我们保存在程序运行目录下的 "uploads" 文件夹
		uploadDir := "./uploads"
		if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create storage directory: " + err.Error()})
			return
		}

		// 将文件保存到服务器的目标路径
		destPath := filepath.Join(uploadDir, file.Filename)
		if err := c.SaveUploadedFile(file, destPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
			return
		}

		//将元数据存入数据库
		_, err = s.DB.Exec("INSERT INTO drivelist (name, capacity) VALUES ($1, $2)", meta.Name, meta.Capacity)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save metadata: " + err.Error()})
			return
		}

		// 打印日志并返回成功响应
		fmt.Printf("File '%s' received and saved to '%s'. Meta: %+v\n", file.Filename, destPath, meta)
		c.JSON(http.StatusOK, gin.H{
			"message":  "File uploaded successfully",
			"filename": file.Filename,
			"path":     destPath,
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
