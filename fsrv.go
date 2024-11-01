package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type FsrvCfg struct {
	Port     *string
	DelAble  *bool
	Hostname *string
	Store    *string
	Max      *int64
}

func (cfg *FsrvCfg) parseArgs() bool {
	hn, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return false
	}

	fsrvCfg.Port = flag.String("p", "8080", "Specify the port to listen on")
	fsrvCfg.DelAble = flag.Bool("d", false, "Enable delete file by UI") //golang处理bool参数的方式是穿了就是true，没传就是false
	fsrvCfg.Store = flag.String("s", "./store", "Specify the directory to store files")
	//hostname = &hn
	fsrvCfg.Hostname = flag.String("n", hn, "Specify the server name, default hostname")
	fsrvCfg.Max = flag.Int64("m", 32, "Max file size to upload, default 32(1<<32=4GB)")

	flag.Parse()
	fmt.Printf("delable : %t\n", *fsrvCfg.DelAble)
	fmt.Printf("store : %s\n", *fsrvCfg.Store)
	fmt.Printf("port : %s\n", *fsrvCfg.Port)
	fmt.Printf("host : %s\n", *fsrvCfg.Hostname)
	fmt.Printf("max : %d -> %s\n", *fsrvCfg.Max, humanReadableSize(1<<*fsrvCfg.Max))
	return true
}

var fsrvCfg *FsrvCfg = &FsrvCfg{}

type UploadFile struct {
	DownloadLink string
	Filename     string
	Size         string
	ModifyTime   string
	Curl         string
}

type FsrvPageParam struct {
	Title   string
	Msgs    []string
	Param1  string
	Param2  string
	Param3  string
	Param4  string
	Param5  string
	Files   []UploadFile
	Empty   bool
	DelAble bool
}

func NewPageParam() *FsrvPageParam {
	return &FsrvPageParam{Title: `FSrv`, Files: make([]UploadFile, 0)}
}

type HTML string

func (html HTML) renderHtml(param *FsrvPageParam, w http.ResponseWriter) {
	// 解析并渲染模板
	t, err := template.New(param.Title).Parse(string(html))
	if err != nil {
		log.Println("Error parsing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, param)
	if err != nil {
		log.Println("Error executing template:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

const filesHtml HTML = `
<html>
  <head>
    <title>{{.Title}}</title>
  <script>
    function copyToClipboard(text) {
      navigator.clipboard.writeText(text).then(function() {
        alert('Copied to clipboard!');
      }, function(err) {
        alert('Failed to copy: ' + err);
      });
    }
    function delFile(file) {
      window.location.href='/del?file='+file;
    }
  </script>
  </head>
  <body>
    <h1>File List</h1>
    <p><a href="/toUpload">Go to Upload Page</a></p>
    <table border="1px">
      <thead>
        <tr>
          <td>Download Link</td>
          <td>Size</td>
          <td>ModifyTime</td>
          <td>CURL</td>
          <!--td>Copy</td-->
          {{if .DelAble}}
          <td>Delete</td>
          {{end}}
        </tr>
      </thead>
      <tbody>
        {{range .Files}}
        <tr>
          <td><a href="{{.DownloadLink}}">{{.Filename}}</a></td>
          <td>{{.Size}}</td>
          <td>{{.ModifyTime}}</td>
          <td><code>curl -L -o '{{.Filename}}' '{{.DownloadLink}}'</code></td>
          <!--td><button onclick="copyToClipboard('curl -L -o \'{{.Filename}}\' \'{{.DownloadLink}}\'')">Copy</button></td-->
          {{if $.DelAble}}
          <td><button onclick="delFile('{{.Filename}}')">Delete</button></td>
          {{end}}
        </tr>
        {{end}}
        {{if .Empty}}
        {{if .DelAble}}
        <tr><td colspan=5>This file store is empty, you can upload something now.</td></tr>
        {{else}}
        <tr><td colspan=4>This file store is empty, you can upload something now.</td></tr>
        {{end}}
        {{end}}
      </tbody>
    </table>
  </body>
</html>
`

const infoHtml HTML = `
<html>
  <head>
    <title>{{.Title}}</title>
  </head>
  <body>
    <h1>Attention !</h1>
    {{range .Msgs}}
    <p>{{.}}</p>
    {{end}}
    <p><a href="/files">Go to File List Page</a></p>
    <p><a href="/toUpload">Go to Upload Page</a></p>
  </body>
</html>
`

const uploadHtml HTML = `
<html>
  <head>
    <title>{{.Title}}</title>
  </head>
  <body>
    <h1>Upload File</h1>
    <p><a href="/files">Go to File List Page</a></p>
    <p>U can upload file by curl : </p><p>curl -F 'file=@/path/file' http://{{.Param1}}:{{.Param2}}/upload</p><p>or:</p>
    <form id="uploadForm" action="/upload" method="post" enctype="multipart/form-data">
      <input type="file" name="file" id="fileInput" text="SelectFile">
      <input type="submit" value="Upload" text="Upload">
    </form>
    <p>Attention : Max upload file size is {{.Param3}}.</p>
    <script>
    document.getElementById("uploadForm").onsubmit = function() {
      var fileInput = document.getElementById("fileInput");
      if (fileInput.files.length === 0) {
        alert("Please select a file to upload.");
        return false; // 阻止表单提交
      }
      return true; // 允许表单提交
    }
    </script>
  </body>
</html>
`

func uploadPage(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, "GET") {
		return
	}
	maxSize := humanReadableSize(1 << (*fsrvCfg.Max))
	param := NewPageParam()
	param.Param1 = *fsrvCfg.Hostname
	param.Param2 = *fsrvCfg.Port
	param.Param3 = maxSize
	uploadHtml.renderHtml(param, w)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, "POST") {
		return
	}

	// 限制文件大小为 4GB
	r.Body = http.MaxBytesReader(w, r.Body, 1<<(*fsrvCfg.Max)) // 1<<32 表示 4GB

	file, header, err := r.FormFile("file")
	if err != nil {
		htmlInfo(w, `No file selected for upload or file is too large`)
		fmt.Println("Failed to upload file : ", err)
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	fullPath := filepath.Join(*fsrvCfg.Store, filename)

	if _, err := os.Stat(fullPath); err == nil {
		htmlInfo(w, `File upload failed!`,
			fmt.Sprintf(`File already exists : '%s'`, filename))
		return
	}

	dst, err := os.Create(fullPath)
	if err != nil {
		htmlInfo(w, `File upload failed!`,
			fmt.Sprintf(`Failed to save file : '%v'`, err))
		return
	}

	defer dst.Close()
	// 使用缓冲区分块读取和写入，避免一次性加载大文件
	buffer := make([]byte, 1024*1024) // 1MB 缓冲区
	size, err := io.CopyBuffer(dst, file, buffer)
	if err != nil {
		htmlInfo(w, `File upload failed!`,
			fmt.Sprintf(`Failed to save file : '%v'`, err))
		return
	}

	humanSize := humanReadableSize(size)
	//记个日志
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	// 上传成功后的页面内容
	htmlInfo(w,
		`Uploaded file successfully !`,
		fmt.Sprintf(`Uploaded file : %s`, filename),
		fmt.Sprintf(`Size : %s`, humanSize),
		fmt.Sprintf(`Time : %s`, currentTime))
}

func (param *FsrvPageParam) loadFileInfo(files []os.FileInfo) {
	empty := true
	for _, file := range files {
		if !file.IsDir() {
			empty = false
			uploadFile := UploadFile{}
			fileName := file.Name()
			downloadURL := fmt.Sprintf("%s/download?file=%s", getURLRoot(), fileName)
			uploadFile.DownloadLink = downloadURL
			uploadFile.Filename = fileName
			uploadFile.Size = humanReadableSize(file.Size())
			uploadFile.ModifyTime = file.ModTime().Format("2006-01-02 15:04:05")
			param.Files = append(param.Files, uploadFile)
		}
	}
	param.Empty = empty
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, "GET") {
		return
	}

	dir, err := os.Open(*fsrvCfg.Store)
	if err != nil {
		htmlInfo(w, fmt.Sprintf(`Failed to open directory : '%v'`, err))
		return
	}

	defer dir.Close()
	files, err := dir.Readdir(-1)
	if err != nil {
		htmlInfo(w, fmt.Sprintf(`Failed to read directory : '%v'`, err))
		return
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	param := NewPageParam()
	param.loadFileInfo(files)
	param.DelAble = *fsrvCfg.DelAble
	filesHtml.renderHtml(param, w)
}

func deleteFile(w http.ResponseWriter, r *http.Request) {

	if !checkMethod(w, r, "GET") {
		return
	}

	filename := r.URL.Query().Get("file")
	filename = filepath.Base(filename)
	filePath := filepath.Join(*fsrvCfg.Store, filename)

	//校验文件是否存在
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		htmlInfo(w, fmt.Sprintf(`File does not exist : '%s'`, filename))
		return
	}
	//删除的不能是目录
	if info.IsDir() {
		htmlInfo(w, fmt.Sprintf(`Cannot delete directory : '%s'`, filename))
		return
	}

	dir, err := os.Open(*fsrvCfg.Store)
	if err != nil {
		htmlInfo(w, fmt.Sprintf(`Failed to open directory : '%v'`, err))
		return
	}

	defer dir.Close()
	files, err := dir.Readdir(-1)
	if err != nil {
		htmlInfo(w, fmt.Sprintf(`Failed to read directory : '%v'`, err))
		return
	}

	for _, file := range files {
		if filename == file.Name() {
			err = os.Remove(filePath)
			if err != nil {
				htmlInfo(w, fmt.Sprintf(`Failed to delete file : '%s', error : %v`, filename, err))
				return
			} else {
				fmt.Printf("Deleted file successfully : %s\n", filePath)
				htmlInfo(w, fmt.Sprintf(`Deleted file successfully : '%s'`, filename))
				return
			}
		}
	}
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, "GET") {
		return
	}

	filename := r.URL.Query().Get("file")
	filename = filepath.Base(filename)
	filePath := filepath.Join(*fsrvCfg.Store, filename)

	//校验文件是否存在
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		htmlInfo(w, fmt.Sprintf(`File does not exist : '%s'`, filename))
		return
	}

	//下载的不能是目录
	if info.IsDir() {
		htmlInfo(w, fmt.Sprintf(`Cannot download directory : '%s'`, filename))
		return
	}

	// 设置响应头，提示浏览器下载文件
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, filePath)

	//下载成功记日志
	fmt.Printf("Download file successfully : %s\n", filename)
}

func prepareTmpDir() (string, error) {
	executablePath, err := os.Executable()
	if err != nil {
		//fmt.Println("Error:", err)
		return "", err
	}
	fmt.Println("Current executable path:", executablePath)
	tmpDir := filepath.Dir(executablePath) + "/tmp"
	fmt.Println("Current tmp dir path:", tmpDir)
	return tmpDir, checkAndCreateDir(tmpDir)
}

func main() {

	if !fsrvCfg.parseArgs() {
		return
	}

	err := checkAndCreateDir(*fsrvCfg.Store)
	if err != nil {
		fmt.Println("Error creating store dir:", err)
		return
	}
	tmpDir, err := prepareTmpDir()
	if err != nil {
		fmt.Println("Error creating tmp dir:", err)
		return
	}
	//因为默认临时文件目录是 /tmp/xxx/xxx 大小有限，当文件超过2G一般临时空间就不足了，所以显式指定临时目录
	os.Setenv("TMPDIR", tmpDir)

	http.HandleFunc("/toUpload", uploadPage)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/files", listFiles)
	http.HandleFunc("/", listFiles)
	http.HandleFunc("/download", downloadFile)
	http.HandleFunc("/del", deleteFile)

	addr := fmt.Sprintf(":%s", *fsrvCfg.Port)
	fmt.Printf("Server started on %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}

func htmlInfo(w http.ResponseWriter, msg ...string) {
	param := NewPageParam()
	param.Msgs = msg
	infoHtml.renderHtml(param, w)
}

func checkMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		htmlInfo(w, fmt.Sprintf(`Http Method should be '%s'`, method))
		return false
	}
	return true
}

func getURLRoot() string {
	return fmt.Sprintf("http://%s:%s", *fsrvCfg.Hostname, *fsrvCfg.Port)
}

// checkAndCreateDir 检查指定目录是否存在，如果不存在则创建
func checkAndCreateDir(dir string) error {
	// 检查目录是否存在
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		// 目录不存在，创建它
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %v", err)
		}
		fmt.Printf("Directory created: %s\n", dir)
	} else if err != nil {
		// 其他错误
		return fmt.Errorf("failed to check directory: %v", err)
	} else {
		fmt.Printf("Directory already exists: %s\n", dir)
	}
	return nil
}

// humanReadableSize 将字节大小转换为更友好的格式
func humanReadableSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
