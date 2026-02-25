package handler

import (
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"

	"fsrv/internal/service"
	"fsrv/internal/util"
)

// PageParam holds parameters for HTML templates
type PageParam struct {
	Title   string
	Msgs    []string
	Param1  string
	Param2  string
	Param3  string
	Param4  string
	Param5  string
	Files   []service.File
	Empty   bool
	DelAble bool
}

// Handler handles HTTP requests
type Handler struct {
	svc       *service.Service
	templates *template.Template
}

// New creates a new HTTP handler
func New(svc *service.Service, templateFS fs.FS) (*Handler, error) {
	templates, err := template.ParseFS(templateFS, "*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &Handler{
		svc:       svc,
		templates: templates,
	}, nil
}

// renderTemplate renders an HTML template with the given parameters
func (h *Handler) renderTemplate(w http.ResponseWriter, name string, param *PageParam) {
	if err := h.templates.ExecuteTemplate(w, name, param); err != nil {
		log.Printf("Error executing template %s: %v", name, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// renderInfo renders an info page with messages
func (h *Handler) renderInfo(w http.ResponseWriter, msgs ...string) {
	param := &PageParam{
		Title: "FSrv Info",
		Msgs:  msgs,
	}
	h.renderTemplate(w, "info.html", param)
}

// checkMethod checks if the request method matches the expected method
func (h *Handler) checkMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		h.renderInfo(w, fmt.Sprintf("HTTP Method should be '%s'", method))
		return false
	}
	return true
}

// UploadPage renders the upload page
func (h *Handler) UploadPage(w http.ResponseWriter, r *http.Request) {
	if !h.checkMethod(w, r, "GET") {
		return
	}

	hostname, port, maxSize := h.svc.GetServerInfo()
	param := &PageParam{
		Title:  "FSrv Upload",
		Param1: hostname,
		Param2: port,
		Param3: maxSize,
	}
	h.renderTemplate(w, "upload.html", param)
}

// UploadFile handles file upload
func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	if !h.checkMethod(w, r, "POST") {
		return
	}

	// Limit file size
	r.Body = http.MaxBytesReader(w, r.Body, h.svc.GetMaxUploadSize())

	file, header, err := r.FormFile("file")
	if err != nil {
		h.renderInfo(w, "No file selected for upload or file is too large")
		log.Printf("Failed to upload file: %v", err)
		return
	}
	defer file.Close()

	size, err := h.svc.UploadFile(header.Filename, file)
	if err != nil {
		h.renderInfo(w, "File upload failed!", err.Error())
		log.Printf("Failed to upload file: %v", err)
		return
	}

	humanSize := util.HumanReadableSize(size)
	currentTime := service.GetCurrentTime()
	h.renderInfo(w,
		"Uploaded file successfully!",
		fmt.Sprintf("Uploaded file: %s", header.Filename),
		fmt.Sprintf("Size: %s", humanSize),
		fmt.Sprintf("Time: %s", currentTime))
}

// ListFiles renders the file list page
func (h *Handler) ListFiles(w http.ResponseWriter, r *http.Request) {
	if !h.checkMethod(w, r, "GET") {
		return
	}

	files, err := h.svc.ListFiles()
	if err != nil {
		h.renderInfo(w, fmt.Sprintf("Failed to list files: %v", err))
		return
	}

	param := &PageParam{
		Title:   "FSrv Files",
		Files:   files,
		Empty:   len(files) == 0,
		DelAble: h.svc.IsDeleteEnabled(),
	}
	h.renderTemplate(w, "files.html", param)
}

// DeleteFile handles file deletion
func (h *Handler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if !h.checkMethod(w, r, "GET") {
		return
	}

	filename := r.URL.Query().Get("file")
	if err := h.svc.DeleteFile(filename); err != nil {
		h.renderInfo(w, err.Error())
		return
	}

	log.Printf("Deleted file successfully: %s", filename)
	h.renderInfo(w, fmt.Sprintf("Deleted file successfully: '%s'", filename))
}

// DownloadFile handles file download
func (h *Handler) DownloadFile(w http.ResponseWriter, r *http.Request) {
	if !h.checkMethod(w, r, "GET") {
		return
	}

	filename := r.URL.Query().Get("file")
	filePath, err := h.svc.GetFilePath(filename)
	if err != nil {
		h.renderInfo(w, err.Error())
		return
	}

	// Set response headers for file download
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Serve the file
	http.ServeFile(w, r, filePath)

	log.Printf("Downloaded file successfully: %s", filename)
}

// RegisterRoutes registers all HTTP routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/toUpload", h.UploadPage)
	mux.HandleFunc("/upload", h.UploadFile)
	mux.HandleFunc("/files", h.ListFiles)
	mux.HandleFunc("/download", h.DownloadFile)
	mux.HandleFunc("/del", h.DeleteFile)
	mux.HandleFunc("/", h.ListFiles)
}
