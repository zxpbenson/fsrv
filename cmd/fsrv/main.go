package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"fsrv/internal/config"
	"fsrv/internal/handler"
	"fsrv/internal/service"
	"fsrv/internal/util"
	"fsrv/web"
)

func main() {
	// Parse configuration
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to parse configuration: %v", err)
	}

	// Create store directory
	if err := util.CheckAndCreateDir(cfg.Store); err != nil {
		log.Fatalf("Failed to create store directory: %v", err)
	}

	// Prepare temporary directory
	tmpDir, err := util.PrepareTmpDir()
	if err != nil {
		log.Fatalf("Failed to prepare temporary directory: %v", err)
	}

	// Set temporary directory environment variable
	// This is necessary because the default temporary directory (/tmp) may have limited space
	// and may not be sufficient for large file uploads (>2GB)
	os.Setenv("TMPDIR", tmpDir)

	// Create service layer
	svc := service.New(cfg)

	// Create template filesystem
	templates, err := fs.Sub(web.TemplatesFS, "templates")
	if err != nil {
		log.Fatalf("Failed to create template filesystem: %v", err)
	}

	// Create HTTP handler
	h, err := handler.New(svc, templates)
	if err != nil {
		log.Fatalf("Failed to create handler: %v", err)
	}

	// Register routes
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	log.Printf("Store directory: %s", cfg.Store)
	log.Printf("Temporary directory: %s", tmpDir)
	log.Printf("Max upload size: %s", svc.GetMaxUploadSizeHuman())
	log.Printf("Delete enabled: %t", svc.IsDeleteEnabled())

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
