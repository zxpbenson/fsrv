# FSrv - Simple File Server

A simple HTTP file server written in Go that allows you to upload, download, list, and delete files through a web interface.

## Features

- 📤 Upload files via web interface or curl
- 📥 Download files with a single click
- 📋 List all files with size and modification time
- 🗑️ Delete files (optional)
- 📊 Human-readable file sizes
- 🔒 Safe filename handling
- 💾 Support for large file uploads (configurable)

## Project Structure

```
fsrv/
├── cmd/
│   └── fsrv/
│       └── main.go              # Main application entry point
├── internal/
│   ├── config/                  # Configuration management
│   │   └── config.go
│   ├── handler/                 # HTTP request handlers
│   │   └── handler.go
│   ├── service/                 # Business logic layer
│   │   └── service.go
│   └── util/                    # Utility functions
│       └── util.go
├── web/
│   ├── templates/               # HTML templates
│   │   ├── files.html
│   │   ├── info.html
│   │   └── upload.html
│   └── fs.go                    # Embedded filesystem
├── Makefile
├── go.mod
├── go.sum
├── README.md
└── .gitignore
```

## Installation

### Build from source

```bash
# Clone the repository
git clone https://github.com/zxpbenson/fsrv.git
cd fsrv

# Build using Makefile
make build

# Run the application
./build/fsrv
```

### Manual Build

```bash
go build -o build/fsrv ./cmd/fsrv
```

### Using go install

```bash
go install github.com/zxpbenson/fsrv/cmd/fsrv@latest
```

## Usage

### Command Line Options

```bash
./fsrv [options]
```

Options:
- `-p <port>`: Specify the port to listen on (default: 8080)
- `-d`: Enable delete file by UI (default: false)
- `-s <directory>`: Specify the directory to store files (default: ./store)
- `-n <hostname>`: Specify the server name (default: system hostname)
- `-m <size>`: Max file size to upload in bits (default: 32, which means 1<<32 = 4GB)

### Examples

Start server on port 8080:
```bash
./fsrv
```

Start server on port 3000 with delete enabled:
```bash
./fsrv -p 3000 -d
```

Start server with custom store directory:
```bash
./fsrv -s /path/to/store
```

Start server with max file size of 8GB:
```bash
./fsrv -m 33
```

## API Endpoints

- `GET /` or `GET /files`: List all files
- `GET /toUpload`: Show upload page
- `POST /upload`: Upload a file
- `GET /download?file=<filename>`: Download a file
- `GET /del?file=<filename>`: Delete a file (if enabled)

## Upload Files

### Via Web Interface

1. Open `http://localhost:8080/toUpload` in your browser
2. Select a file and click "Upload"

### Via curl

```bash
curl -F 'file=@/path/to/file' http://localhost:8080/upload
```

## Download Files

### Via Web Interface

1. Open `http://localhost:8080/files` in your browser
2. Click on the file you want to download

### Via curl

```bash
curl -L -o 'filename' 'http://localhost:8080/download?file=filename'
```

## Delete Files

### Via Web Interface

1. Open `http://localhost:8080/files` in your browser
2. Click the "Delete" button next to the file (if delete is enabled)

### Via curl

```bash
curl 'http://localhost:8080/del?file=filename'
```

## Development

This project includes a `Makefile` to simplify common development tasks.

### Running tests

```bash
make test
```

### Building

```bash
make build
```

### Running

```bash
# Build and run
make run

# Or run directly from build
./build/fsrv
```

### Cleaning

```bash
make clean
```

## Architecture

The application follows a clean architecture pattern:

- **Config Layer**: Handles configuration parsing and management
- **Service Layer**: Contains business logic for file operations
- **Handler Layer**: Handles HTTP requests and responses
- **Util Layer**: Provides utility functions for common operations

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
