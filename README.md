# Media Service

A comprehensive media service built with Go, providing media file upload, processing, and management through gRPC APIs. The service supports automatic image optimization, video processing, and multiple storage backends.

## 🚀 Features

* **File Upload & Management**: Support for images, videos, and other media files
* **Image Processing**: Automatic resizing, format conversion (WebP), and thumbnail generation
* **Video Processing**: Video transcoding, thumbnail generation, and multiple resolution support
* **Storage Backends**: Local filesystem and S3-compatible storage
* **gRPC API**: High-performance RPC communication
* **Database Integration**: PostgreSQL with migrations
* **Background Processing**: Asynchronous media processing with Redis queue
* **Optimization**: Advanced image/video optimization techniques

## 🏗️ Architecture

```
media-service/
├── bootstrap/          # Application bootstrap and configuration
├── cmd/               # Application entry point
├── constants/         # Application constants
├── domain/           # Business logic layer
│   ├── entity/       # Domain entities
│   ├── repository/   # Data access interfaces
│   └── usecase/      # Business use cases
├── infrastructure/   # Infrastructure layer
│   ├── grpc_service/ # gRPC server implementations
│   └── repo/         # Repository implementations
└── migrations/       # Database migrations
```

## 📋 Prerequisites

* Go 1.21 or higher
* PostgreSQL 12 or higher
* Redis 6 or higher
* libvips (for image processing)
* FFmpeg (for video processing, optional)

## 🛠️ Installation

1. **Clone the repository**
```bash
git clone <repository-url>
cd media-service
```

2. **Install dependencies**
```bash
make deps
```

3. **Install libvips (for image processing)**
```bash
# Ubuntu/Debian
sudo apt-get install libvips-dev

# macOS
brew install vips

# CentOS/RHEL
sudo yum install vips-devel
```

4. **Set up the database**
```bash
# Create database
make dev-create-db

# Run migrations
make migrate-up
```

5. **Configure environment**
   * Copy `dev.config.yaml` and modify as needed
   * Update database connection string
   * Configure storage settings

## 🚀 Running the Application

### Development Mode

```bash
# Run the application
make run

# Or with live reload (if using air)
air
```

### Build and Run

```bash
# Build the application
make build

# Run the built binary
./bin/media-service
```

## 📊 Database Management

### Migrations

```bash
# Apply migrations
make migrate-up

# Rollback migrations
make migrate-down

# Reset database
make migrate-reset

# Create new migration
make migrate-create NAME=migration_name
```

### Database Operations

```bash
# Create database
make dev-create-db

# Drop database
make dev-drop-db

# Setup development environment
make dev-setup

# Reset development environment
make dev-reset
```

## 🔧 Configuration

The application uses `dev.config.yaml` for configuration. Key settings include:

### Application Settings
```yaml
app:
  name: "media-service"
  mode: "development"
  port_grpc: 8082
  host_grpc: "0.0.0.0"
```

### Storage Configuration
```yaml
storage:
  provider: "local" # local, s3
  local:
    upload_dir: "./uploads"
    public_url: "http://localhost:8080/uploads"
```

### Media Processing Settings
```yaml
media:
  max_file_size: 100MB
  image:
    max_width: 2048
    max_height: 2048
    quality: 85
    thumbnails:
      small: "150x150"
      medium: "300x300"
      large: "600x600"
```

## 🔌 API Endpoints

The service provides the following gRPC endpoints:

### Media Management
* `UploadMedia`: Upload a new media file with streaming
* `GetMedia`: Retrieve media by ID
* `ListMedia`: List media with filters and pagination
* `UpdateMedia`: Update media metadata
* `DeleteMedia`: Delete media file
* `GetMediaVariants`: Get all variants (thumbnails, formats) of a media
* `ProcessMedia`: Manually trigger media processing

## 🖼️ Image Processing Features

* **Automatic WebP Conversion**: Convert images to WebP for better compression
* **Thumbnail Generation**: Create multiple thumbnail sizes
* **Format Optimization**: Automatic format selection based on browser support
* **Compression**: Smart compression with quality optimization
* **Resizing**: Automatic resizing for large images

## 🎥 Video Processing Features

* **Thumbnail Generation**: Extract frames for video thumbnails
* **Multiple Resolutions**: Support for 480p, 720p, 1080p
* **Format Conversion**: Convert to MP4 and WebM
* **Compression**: Video optimization for web delivery

## 🏗️ Project Structure

### Domain Layer
* **Entities**: Media, MediaVariant models
* **Repositories**: Data access interfaces
* **Use Cases**: Business logic implementation

### Infrastructure Layer
* **gRPC Services**: API endpoint implementations
* **Repositories**: Database and storage implementations
* **Processing**: Media processing services

## 🔒 Security Features

* **File Type Validation**: Strict MIME type checking
* **File Size Limits**: Configurable upload size limits
* **Input Sanitization**: Comprehensive request validation
* **Path Security**: Secure file path handling

## 🧪 Development

### Running Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage
```

### Code Quality
```bash
# Format code
make fmt

# Run linter
make lint

# Vet code
make vet
```

### Adding New Features

1. Define domain entities in `domain/entity/`
2. Create repository interfaces in `domain/repository/`
3. Implement business logic in `domain/usecase/`
4. Add gRPC service implementation in `infrastructure/grpc_service/`
5. Update database schema with migrations

## 📝 Dependencies

### Core Dependencies
* `github.com/go-pg/pg/v10`: PostgreSQL ORM
* `google.golang.org/grpc`: gRPC framework
* `github.com/h2non/bimg`: Image processing (libvips)
* `github.com/hibiken/asynq`: Background job processing
* `go.uber.org/zap`: Structured logging
* `github.com/spf13/viper`: Configuration management

### Processing Libraries
* `libvips`: High-performance image processing
* `FFmpeg`: Video processing (optional)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For support and questions, please create an issue in the repository.

## 📚 API Documentation

The service supports gRPC reflection in development mode. You can use tools like:
* [grpcurl](https://github.com/fullstorydev/grpcurl)
* [BloomRPC](https://github.com/bloomrpc/bloomrpc)
* [Evans](https://github.com/ktr0731/evans)

### Example Usage with grpcurl

```bash
# List available services
grpcurl -plaintext localhost:8082 list

# Upload a file
grpcurl -plaintext -d '{"info":{"file_name":"test.jpg","mime_type":"image/jpeg","created_by":"user123"}}' localhost:8082 media.MediaService/UploadMedia

# Get media by ID
grpcurl -plaintext -d '{"id":"media-id"}' localhost:8082 media.MediaService/GetMedia
```
