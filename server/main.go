package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"

	"family-tree/config"
	"family-tree/database"
	"family-tree/handler"
	"family-tree/model"
	"family-tree/scheduler"

	"github.com/gin-gonic/gin"
)

func main() {
	checkBirthday := flag.Bool("check-birthday", false, "手动触发一次生日检查并退出")
	flag.Parse()

	cfg := config.Load()

	// 初始化数据库
	db := database.Init(cfg.DBPath)
	defer db.Close()

	// 初始化 repositories
	familyRepo := &model.FamilyRepo{DB: db}
	personRepo := &model.PersonRepo{DB: db}
	relationRepo := &model.RelationRepo{DB: db}

	// 手动触发模式
	if *checkBirthday {
		s := scheduler.NewBirthdayScheduler(personRepo, familyRepo, cfg.WebhookURL, cfg.RemindTime)
		s.TriggerCheck()
		return
	}

	// 启动定时任务
	birthdayScheduler := scheduler.NewBirthdayScheduler(personRepo, familyRepo, cfg.WebhookURL, cfg.RemindTime)
	birthdayScheduler.Start()
	defer birthdayScheduler.Stop()

	// 注册路由
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		familyHandler := &handler.FamilyHandler{Repo: familyRepo}
		personHandler := &handler.PersonHandler{Repo: personRepo}
		relationHandler := &handler.RelationHandler{Repo: relationRepo}

		familyHandler.Register(api)
		personHandler.Register(api)
		relationHandler.Register(api)

		// 手动触发生日检查（调试用）
		api.POST("/debug/check-birthday", func(c *gin.Context) {
			birthdayScheduler.TriggerCheck()
			c.JSON(200, gin.H{"message": "birthday check triggered"})
		})
	}

	// 前端静态文件（如果 web/dist 存在）
	serveStatic(r)

	log.Printf("server starting on %s", cfg.ServerAddr)
	if err := r.Run(cfg.ServerAddr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func serveStatic(r *gin.Engine) {
	distDir := "web/dist"
	if _, err := os.Stat(distDir); os.IsNotExist(err) {
		log.Println("no web/dist directory, skip static serving (use `cd web && npm run build` to build frontend)")
		return
	}

	log.Println("serving frontend from web/dist")

	// 静态资源 + SPA fallback
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API 跳过
		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}

		// 检查文件是否存在
		filePath := distDir + path
		if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
			c.File(filePath)
			return
		}

		// SPA fallback → index.html
		c.File(distDir + "/index.html")
	})
}
