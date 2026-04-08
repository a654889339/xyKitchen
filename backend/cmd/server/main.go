package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"xykitchen/backend/internal/bootstrap"
	"xykitchen/backend/internal/config"
	"xykitchen/backend/internal/db"
	"xykitchen/backend/internal/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()
	if err := db.Connect(cfg); err != nil {
		log.Fatalf("[xyKitchen] DB connect: %v", err)
	}
	if err := db.AutoMigrate(); err != nil {
		log.Printf("[xyKitchen] AutoMigrate: %v", err)
	}
	if err := bootstrap.Run(); err != nil {
		log.Fatalf("[xyKitchen] bootstrap: %v", err)
	}

	if cfg.NodeEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())
	engine.Use(cors.Default())
	engine.MaxMultipartMemory = 10 << 20

	handlers.RegisterRoutes(engine, cfg)

	uploadsDir := filepath.Join("public", "uploads")
	if err := os.MkdirAll(uploadsDir, 0o755); err != nil {
		log.Printf("[xyKitchen] mkdir uploads: %v", err)
	}
	engine.Static("/uploads", uploadsDir)
	engine.StaticFile("/", filepath.Join("static", "admin.html"))

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	log.Printf("[xyKitchen] listening on http://%s", addr)
	if err := engine.Run(addr); err != nil {
		log.Fatal(err)
	}
}
