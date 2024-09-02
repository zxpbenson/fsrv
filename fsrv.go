package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

var port *string
var delable *bool
var hostname *string
var store *string

func uploadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

		fmt.Fprintf(w, `<html><head><title>FSrv</title></head><body><h1>Upload File</h1>`)

		// 添加一个跳转到文件列表页面的链接
		fmt.Fprintf(w, `<p><a href="/files">Go to File List Page</a></p>`)

		fmt.Fprintf(w, `<p>U can upload file by curl : </p><p>curl -F 'file=@/path/file' http://%s:%s/upload</p><p>or:</p>`, *hostname, *port)

		fmt.Fprintf(w, `
<form id="uploadForm" action="/upload" method="post" enctype="multipart/form-data">
    <input type="file" name="file" id="fileInput" text="SelectFile">
    <input type="submit" value="Upload" text="Upload">
</form>

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
</html>`)
		return
	}

	if r.Method == "POST" {
		file, header, err := r.FormFile("file")
		if err != nil {
			fmt.Fprintf(w, "No file selected for upload")
			return
		}
		defer file.Close()

		filename := filepath.Base(header.Filename)
		fullPath := filepath.Join(*store, filename)

		if _, err := os.Stat(fullPath); err == nil {
			fmt.Fprintf(w, `<html><body>
                <h1>File uploaded faild!</h1>
                <p>File already exists: %s</p>
                <p><a href="/files">Go to File List</a></p>
                </body></html>`, filename)
			return
		}

		dst, err := os.Create(fullPath)
		if err != nil {
			fmt.Fprintf(w, `<html><body>
                <h1>File uploaded faild!</h1>
                <p>Failed to save file: %v</p>
                <p><a href="/files">Go to File List</a></p>
                </body></html>`, err)

			return
		}
		defer dst.Close()

		size, err := io.Copy(dst, file)
		if err != nil {
			fmt.Fprintf(w, `<html><body>
                <h1>File uploaded faild!</h1>
                <p>Failed to save file: %v</p>
                <p><a href="/files">Go to File List</a></p>
                </body></html>`, err)
			return
		}

		currentTime := time.Now().Format("2006-01-02 15:04:05")
		fmt.Printf("Uploaded file successfully : %s, Size: %d bytes, Time: %s\n", filename, size, currentTime)

		// 上传成功后的页面内容
		fmt.Fprintf(w, `<html><body>
            <h1>Uploaded file successfully !</h1>
            <p>Uploaded file: %s, </p><p>Size: %d bytes, </p><p>Time: %s</p>
            <p><a href="/files">Go to File List</a></p>
            </body></html>\n`, filename, size, currentTime)

	}
}

func listFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {

		dir, err := os.Open(*store)
		if err != nil {
			fmt.Fprintf(w, "Failed to open directory: %v", err)
			return
		}
		defer dir.Close()

		files, err := dir.Readdir(-1)
		if err != nil {
			fmt.Fprintf(w, "Failed to read directory: %v", err)
			return
		}

		fmt.Fprintf(w, `<html><head><title>FSrv</title><script>
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

</script></head><body><h1>File List</h1>`)

		// 添加一个跳转到上传页面的链接
		fmt.Fprintf(w, `<p><a href="/upload">Go to Upload Page</a></p>`)

		fmt.Fprintf(w, `<table border="1px">
            <tr><td>Download Link</td>
            <td>CURL</td>
            <!--td>Copy</td-->`) //如果Server端没证书未开启Https，不允许通过JS操作剪贴板

		if *delable {
			fmt.Fprintf(w, `<td>Delete</td>`)
		}

		fmt.Fprintf(w, `</tr>`)

		empty := true
		for _, file := range files {
			if !file.IsDir() {
				empty = false
				fileName := file.Name()
				downloadURL := fmt.Sprintf("%s/download?file=%s", getURLRoot(), fileName)
				fmt.Fprintf(w, `<tr>
                    <td><a href="%s">%s</a></td>
                    <td><code>curl -L -o '%s' '%s'</code></td>
                    <!--td><button onclick="copyToClipboard('curl -L -o \'%s\' \'%s\'')">Copy</button></td-->`, downloadURL, fileName, fileName, downloadURL, fileName, downloadURL)
				if *delable {
					fmt.Fprintf(w, `<td><button onclick="delFile('%s')">Delete</button></td>`, fileName)
				}
				fmt.Fprintf(w, `</tr>`)
			}
		}

		if empty {
			if *delable {
				fmt.Fprintf(w, `<tr><td colspan=3>This file store is empty, you can upload something now.</td></tr>`)
			} else {
				fmt.Fprintf(w, `<tr><td colspan=2>This file store is empty, you can upload something now.</td></tr>`)
			}
		}

		fmt.Fprintf(w, `</table></body></html>`)
	}
}

func getPort() string {
	return *port
}

func getURLRoot() string {
	return fmt.Sprintf("http://%s:%s", *hostname, *port)
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	filename = filepath.Base(filename)
	filePath := filepath.Join(*store, filename)

	fmt.Fprintf(w, `<html><head><title>FSrv</title></head><body><h1>Delete File</h1>`)

	// 添加一个跳转到上传页面的链接
	fmt.Fprintf(w, `<p><a href="/files">Go to File List Page</a></p>`)

	//校验文件是否存在
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("<p>File does not exist: %s<p></body></html>", filename), http.StatusNotFound)
		return
	}
	//删除的不能是目录
	if info.IsDir() {
		http.Error(w, fmt.Sprintf("<p>Cannot delete directory: %s<p></body></html>", filename), http.StatusNotFound)
		return
	}

	dir, err := os.Open(*store)
	if err != nil {
		http.Error(w, fmt.Sprintf("<p>Failed to open directory: %v<p></body></html>", err), http.StatusInternalServerError)
		return
	}
	defer dir.Close()

	files, err := dir.Readdir(-1)
	if err != nil {
		http.Error(w, fmt.Sprintf("<p>Failed to read directory: %v<p></body></html>", err), http.StatusInternalServerError)
		return
	}

	for _, file := range files {
		if filename == file.Name() {
			err = os.Remove(filePath)
			if err != nil {
				http.Error(w, fmt.Sprintf("<p>Failed to delete file: %s<p></body></html>", filename), http.StatusInternalServerError)
				return
			} else {
				fmt.Fprintf(w, `<p>Deleted file successfully : %s</p></body></html>`, filename)
				fmt.Printf("Deleted file successfully : %s\n", filePath)
				return
			}
		}
	}
}

func downloadFile(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		filename := r.URL.Query().Get("file")
		filename = filepath.Base(filename)
		filePath := filepath.Join(*store, filename)

		//校验文件是否存在
		info, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			http.Error(w, fmt.Sprintf("<html><head><title>FSrv</title></head><body><h1>Download File</h1><p>File does not exist: %s<p></body></html>", filename), http.StatusNotFound)
			return
		}
		//删除的不能是目录
		if info.IsDir() {
			http.Error(w, fmt.Sprintf("<html><head><title>FSrv</title></head><body><h1>Download File</h1><p>Cannot download directory: %s<p></body></html>", filename), http.StatusNotFound)
			return
		}

		// 设置响应头，提示浏览器下载文件
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
		w.Header().Set("Content-Type", "application/octet-stream")

		http.ServeFile(w, r, filePath)
		fmt.Printf("Download file successfully : %s\n", filename)
	}
}

func main() {

	hn, err := os.Hostname()
	if err != nil {
		fmt.Println("Error getting hostname:", err)
		return
	}

	port = flag.String("p", "8080", "Specify the port to listen on")
	delable = flag.Bool("d", false, "Enable delete file by UI") //golang处理bool参数的方式是穿了就是true，没传就是false
	store = flag.String("s", "./store", "Specify the directory to store files")
	//hostname = &hn
	hostname = flag.String("n", hn, "Specify the server name, default hostname")

	flag.Parse()
	fmt.Printf("delable : %t\n", *delable)
	fmt.Printf("store : %s\n", *store)
	fmt.Printf("port : %s\n", *port)
	fmt.Printf("host : %s\n", *hostname)

	err = checkAndCreateDir(*store)
	if err != nil {
		fmt.Println("Error creating store dir:", err)
		return
	}

	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/files", listFiles)
	http.HandleFunc("/", listFiles)
	http.HandleFunc("/download", downloadFile)
	http.HandleFunc("/del", deleteFile)

	addr := fmt.Sprintf(":%s", *port)
	fmt.Printf("Server started on %s\n", addr)

	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
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
