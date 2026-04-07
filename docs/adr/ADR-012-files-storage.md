# ADR-012: Files/Storage Bounded Context

## Status

Accepted

## Context

Most modern applications work with files:
- User avatars and profile images
- Document uploads (PDF, DOCX)
- Media files (images, videos, audio)
- Exports and reports
- Attachments (emails, messages)

Need to:
- Store files in cloud storage (S3, GCS) or locally
- Generate temporary URLs for upload/download
- Process images (resize, crop, compress)
- Validate files (type, size)
- Manage file access
- Clean up old files

Currently, there is no mechanism for working with files.

## Decision

Create a separate **Files** bounded context with abstraction over storage providers.

### 1. Domain Layer

#### Aggregates

**File** - main aggregate:
```go
type File struct {
    id              FileID
    ownerID        *domain.UserID    // optional owner
    filename        string            // original filename
    storedName      string            // generated unique name
    mimeType        string            // MIME type
    size            int64             // bytes
    path            string            // storage path
    storageProvider StorageProvider   // s3, gcs, local
    checksum        string            // SHA-256 hash
    metadata        FileMetadata      // additional metadata
    accessLevel     AccessLevel       // public, private, restricted
    uploadedAt      time.Time
    expiresAt       *time.Time        // optional expiration
    processedAt     *time.Time        // for processed files
    createdAt       time.Time
    updatedAt       time.Time
}

type FileID string

type StorageProvider string
const (
    StorageS3    StorageProvider = "s3"
    StorageGCS   StorageProvider = "gcs"
    StorageLocal StorageProvider = "local"
)

type AccessLevel string
const (
    AccessPublic      AccessLevel = "public"       // anyone can access
    AccessPrivate     AccessLevel = "private"      // only owner
    AccessRestricted  AccessLevel = "restricted"   // specific permissions
)

type FileMetadata struct {
    Width       *int      // image width
    Height      *int      // image height
    Duration    *int      // video/audio duration (seconds)
    Pages       *int      // document pages
    Thumbnail   *string   // thumbnail file ID
    OriginalID  *FileID   // if this is processed version
    Custom      map[string]string
}

// Business methods
func NewFile(
    ownerID *domain.UserID,
    filename string,
    mimeType string,
    size int64,
    provider StorageProvider,
    accessLevel AccessLevel,
) (*File, error)

func (f *File) GeneratePath() string
func (f *File) GenerateStoredName() string
func (f *File) SetMetadata(metadata FileMetadata) error
func (f *File) SetExpiration(expiresAt time.Time) error
func (f *File) IsExpired() bool
func (f *File) IsImage() bool
func (f *File) IsVideo() bool
func (f *File) IsDocument() bool
func (f *File) CanAccess(userID *domain.UserID) bool
```

**FileUpload** - for file uploads:
```go
type FileUpload struct {
    id              UploadID
    file            *File
    uploadURL       string            // presigned upload URL
    expiresAt       time.Time         // URL expiration
    fields          map[string]string // additional form fields
    status          UploadStatus
    createdAt       time.Time
}

type UploadID string

type UploadStatus string
const (
    UploadPending   UploadStatus = "pending"
    UploadCompleted UploadStatus = "completed"
    UploadFailed    UploadStatus = "failed"
    UploadExpired   UploadStatus = "expired"
)

func NewFileUpload(file *File, ttl time.Duration) (*FileUpload, error)
func (u *FileUpload) GeneratePresignedURL(provider StorageProvider) (string, error)
```

**FileProcessing** - file processing status:
```go
type FileProcessing struct {
    id              ProcessingID
    fileID          FileID
    operation       ProcessingOperation
    status          ProcessingStatus
    result          *FileID           // result file ID
    error           *string
    startedAt       *time.Time
    completedAt     *time.Time
    createdAt       time.Time
}

type ProcessingOperation string
const (
    OperationResize        ProcessingOperation = "resize"
    OperationCrop          ProcessingOperation = "crop"
    OperationCompress      ProcessingOperation = "compress"
    OperationConvert       ProcessingOperation = "convert"
    OperationThumbnail     ProcessingOperation = "thumbnail"
    OperationWatermark     ProcessingOperation = "watermark"
)

type ProcessingStatus string
const (
    ProcessingPending   ProcessingStatus = "pending"
    ProcessingRunning    ProcessingStatus = "running"
    ProcessingCompleted  ProcessingStatus = "completed"
    ProcessingFailed     ProcessingStatus = "failed"
)
```

#### Domain Events

```go
// File lifecycle events
type FileUploaded struct {
    FileID    FileID
    OwnerID   *domain.UserID
    Filename  string
    MimeType  string
    Size      int64
    UploadedAt time.Time
}

type FileDeleted struct {
    FileID    FileID
    Path      string
    DeletedAt time.Time
}

type FileAccessed struct {
    FileID    FileID
    UserID    *domain.UserID
    AccessType string  // view, download
    AccessedAt time.Time
}

type FileProcessingRequested struct {
    ProcessingID ProcessingID
    FileID       FileID
    Operation    ProcessingOperation
}

type FileProcessingCompleted struct {
    ProcessingID ProcessingID
    FileID       FileID
    ResultFileID FileID
    CompletedAt  time.Time
}

type FileProcessingFailed struct {
    ProcessingID ProcessingID
    FileID       FileID
    Error        string
    FailedAt     time.Time
}
```

#### Repository Interfaces

```go
type FileRepository interface {
    Create(ctx context.Context, file *File) error
    Update(ctx context.Context, file *File) error
    GetByID(ctx context.Context, id FileID) (*File, error)
    GetByOwner(ctx context.Context, ownerID domain.UserID, limit int) ([]*File, error)
    GetByPath(ctx context.Context, path string) (*File, error)
    GetExpiredFiles(ctx context.Context, before time.Time) ([]*File, error)
    Delete(ctx context.Context, id FileID) error
    
    // Batch operations
    CreateBatch(ctx context.Context, files []*File) error
    DeleteBatch(ctx context.Context, ids []FileID) error
}

type UploadRepository interface {
    Create(ctx context.Context, upload *FileUpload) error
    GetByID(ctx context.Context, id UploadID) (*FileUpload, error)
    UpdateStatus(ctx context.Context, id UploadID, status UploadStatus) error
    GetExpiredUploads(ctx context.Context, before time.Time) ([]*FileUpload, error)
    Delete(ctx context.Context, id UploadID) error
}

type ProcessingRepository interface {
    Create(ctx context.Context, processing *FileProcessing) error
    Update(ctx context.Context, processing *FileProcessing) error
    GetByID(ctx context.Context, id ProcessingID) (*FileProcessing, error)
    GetByFile(ctx context.Context, fileID FileID) ([]*FileProcessing, error)
    GetPending(ctx context.Context, limit int) ([]*FileProcessing, error)
}
```

#### Storage Interface

```go
type StorageProvider interface {
    // Upload file
    Upload(ctx context.Context, path string, reader io.Reader, contentType string) error
    UploadFromBytes(ctx context.Context, path string, data []byte, contentType string) error
    
    // Download file
    Download(ctx context.Context, path string) (io.ReadCloser, error)
    DownloadToBytes(ctx context.Context, path string) ([]byte, error)
    
    // URL generation
    GetPublicURL(path string) string
    GetPresignedUploadURL(ctx context.Context, path string, expiresIn time.Duration) (string, map[string]string, error)
    GetPresignedDownloadURL(ctx context.Context, path string, expiresIn time.Duration) (string, error)
    
    // File operations
    Delete(ctx context.Context, path string) error
    Exists(ctx context.Context, path string) (bool, error)
    GetMetadata(ctx context.Context, path string) (*StorageMetadata, error)
    Copy(ctx context.Context, src, dst string) error
    
    // Storage info
    Name() StorageProvider
}

type StorageMetadata struct {
    Size         int64
    ContentType  string
    LastModified time.Time
    ETag         string
}
```

### 2. Application Layer

#### Commands

```go
// Request upload URL
type RequestUploadURLCommand struct {
    OwnerID     *domain.UserID
    Filename    string
    MimeType    string
    Size        int64
    AccessLevel AccessLevel
    ExpiresIn   time.Duration
}

type RequestUploadURLHandler func(ctx context.Context, cmd RequestUploadURLCommand) (*UploadURLResult, error)

type UploadURLResult struct {
    UploadID    string
    UploadURL   string
    Fields      map[string]string
    ExpiresAt   time.Time
}

// Confirm upload
type ConfirmUploadCommand struct {
    UploadID UploadID
    Checksum string  // SHA-256 hash from client
}

type ConfirmUploadHandler func(ctx context.Context, cmd ConfirmUploadCommand) (FileID, error)

// Upload file directly (small files)
type UploadFileCommand struct {
    OwnerID     *domain.UserID
    Filename    string
    Content     []byte
    MimeType    string
    AccessLevel AccessLevel
    ExpiresAt   *time.Time
}

type UploadFileHandler func(ctx context.Context, cmd UploadFileCommand) (FileID, error)

// Delete file
type DeleteFileCommand struct {
    FileID FileID
}

type DeleteFileHandler func(ctx context.Context, cmd DeleteFileCommand) error

// Request processing
type RequestProcessingCommand struct {
    FileID    FileID
    Operation ProcessingOperation
    Options   ProcessingOptions
}

type ProcessingOptions struct {
    Width     *int
    Height    *int
    Quality   *int
    Format    *string
    Thumbnail *bool
}

type RequestProcessingHandler func(ctx context.Context, cmd RequestProcessingCommand) (ProcessingID, error)
```

#### Queries

```go
// Get file details
type GetFileQuery struct {
    FileID FileID
}

type GetFileHandler func(ctx context.Context, query GetFileQuery) (*File, error)

// Get download URL
type GetDownloadURLQuery struct {
    FileID    FileID
    ExpiresIn time.Duration
}

type GetDownloadURLHandler func(ctx context.Context, query GetDownloadURLQuery) (string, error)

// List files
type ListFilesQuery struct {
    OwnerID  *domain.UserID
    MimeType *string
    FromDate *time.Time
    ToDate   *time.Time
    Limit    int
    Cursor   *string
}

type ListFilesHandler func(ctx context.Context, query ListFilesQuery) (*FileList, error)

// Get processing status
type GetProcessingStatusQuery struct {
    ProcessingID ProcessingID
}

type GetProcessingStatusHandler func(ctx context.Context, query GetProcessingStatusQuery) (*FileProcessing, error)
```

### 3. Infrastructure Layer

#### Storage Implementations

##### Amazon S3

```go
type S3Storage struct {
    client    *s3.Client
    bucket    string
    region    string
    cdnURL    string  // optional CDN
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
    client := s3.New(s3.Options{
        Region: cfg.Region,
        Credentials: aws.NewCredentialsCache(
            credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
        ),
    })
    return &S3Storage{
        client: client,
        bucket: cfg.Bucket,
        region: cfg.Region,
        cdnURL: cfg.CDNURL,
    }, nil
}

func (s *S3Storage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket:      aws.String(s.bucket),
        Key:         aws.String(path),
        Body:        reader,
        ContentType: aws.String(contentType),
    })
    return err
}

func (s *S3Storage) GetPresignedUploadURL(ctx context.Context, path string, expiresIn time.Duration) (string, map[string]string, error) {
    presigner := s3.NewPresignClient(s.client)
    
    req, err := presigner.PresignPutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    }, func(opts *s3.PresignOptions) {
        opts.Expires = expiresIn
    })
    
    if err != nil {
        return "", nil, err
    }
    
    // Additional fields for direct browser upload
    fields := map[string]string{
        "bucket":      s.bucket,
        "key":         path,
        "Content-Type": "image/jpeg", // or dynamic
    }
    
    return req.URL, fields, nil
}

func (s *S3Storage) GetPresignedDownloadURL(ctx context.Context, path string, expiresIn time.Duration) (string, error) {
    presigner := s3.NewPresignClient(s.client)
    
    req, err := presigner.PresignGetObject(ctx, &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(path),
    }, func(opts *s3.PresignOptions) {
        opts.Expires = expiresIn
    })
    
    if err != nil {
        return "", err
    }
    
    return req.URL, nil
}

// ... other methods
```

##### Google Cloud Storage

```go
type GCSStorage struct {
    client    *storage.Client
    bucket    string
    cdnURL    string
}

func NewGCSStorage(ctx context.Context, cfg GCSConfig) (*GCSStorage, error) {
    client, err := storage.NewClient(ctx, option.WithCredentialsJSON(cfg.Credentials))
    if err != nil {
        return nil, err
    }
    
    return &GCSStorage{
        client: client,
        bucket: cfg.Bucket,
        cdnURL: cfg.CDNURL,
    }, nil
}

func (g *GCSStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
    wc := g.client.Bucket(g.bucket).Object(path).NewWriter(ctx)
    wc.ContentType = contentType
    _, err := io.Copy(wc, reader)
    return wc.Close()
}

func (g *GCSStorage) GetPresignedUploadURL(ctx context.Context, path string, expiresIn time.Duration) (string, map[string]string, error) {
    // GCS signed URL for upload
    opts := &storage.SignedURLOptions{
        Method:  "PUT",
        Expires: time.Now().Add(expiresIn),
    }
    
    url, err := g.client.Bucket(g.bucket).SignedURL(path, opts)
    if err != nil {
        return "", nil, err
    }
    
    return url, nil, nil
}

// ... other methods
```

##### Local Storage (Development)

```go
type LocalStorage struct {
    basePath string
    baseURL  string
}

func NewLocalStorage(basePath, baseURL string) *LocalStorage {
    return &LocalStorage{
        basePath: basePath,
        baseURL:  baseURL,
    }
}

func (l *LocalStorage) Upload(ctx context.Context, path string, reader io.Reader, contentType string) error {
    fullPath := filepath.Join(l.basePath, path)
    
    // Create directory if not exists
    if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
        return err
    }
    
    // Create file
    file, err := os.Create(fullPath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    _, err = io.Copy(file, reader)
    return err
}

func (l *LocalStorage) GetPublicURL(path string) string {
    return fmt.Sprintf("%s/%s", l.baseURL, path)
}

func (l *LocalStorage) GetPresignedUploadURL(ctx context.Context, path string, expiresIn time.Duration) (string, map[string]string, error) {
    // For local storage, return simple upload endpoint
    return fmt.Sprintf("%s/upload/%s", l.baseURL, path), nil, nil
}

// ... other methods
```

#### Image Processor

```go
type ImageProcessor interface {
    Resize(ctx context.Context, input []byte, width, height int) ([]byte, error)
    Crop(ctx context.Context, input []byte, x, y, width, height int) ([]byte, error)
    Compress(ctx context.Context, input []byte, quality int) ([]byte, error)
    Convert(ctx context.Context, input []byte, format string) ([]byte, error)
    GenerateThumbnail(ctx context.Context, input []byte, width, height int) ([]byte, error)
    GetMetadata(ctx context.Context, input []byte) (*ImageMetadata, error)
}

type ImageMetadata struct {
    Width    int
    Height   int
    Format   string
    Size     int64
}

// Using disintegration/imaging
type ImagingProcessor struct{}

func (p *ImagingProcessor) Resize(ctx context.Context, input []byte, width, height int) ([]byte, error) {
    img, err := imaging.Decode(bytes.NewReader(input))
    if err != nil {
        return nil, err
    }
    
    resized := imaging.Resize(img, width, height, imaging.Lanczos)
    
    var buf bytes.Buffer
    if err := imaging.Encode(&buf, resized, imaging.JPEG); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

#### File Cleaner

```go
type FileCleaner struct {
    fileRepo       FileRepository
    uploadRepo     UploadRepository
    storage        StorageProvider
    cleanInterval  time.Duration
}

func (c *FileCleaner) Start(ctx context.Context) error {
    ticker := time.NewTicker(c.cleanInterval)
    defer ticker.Stop()
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
        case <-ticker.C:
            c.cleanup(ctx)
        }
    }
}

func (c *FileCleaner) cleanup(ctx context.Context) {
    // Clean expired files
    expiredFiles, _ := c.fileRepo.GetExpiredFiles(ctx, time.Now())
    for _, file := range expiredFiles {
        c.storage.Delete(ctx, file.Path)
        c.fileRepo.Delete(ctx, file.ID)
    }
    
    // Clean expired uploads
    expiredUploads, _ := c.uploadRepo.GetExpiredUploads(ctx, time.Now())
    for _, upload := range expiredUploads {
        c.uploadRepo.Delete(ctx, upload.ID)
    }
}
```

#### Persistence

SQLite tables:
```sql
CREATE TABLE files (
    id TEXT PRIMARY KEY,
    owner_id TEXT,
    filename TEXT NOT NULL,
    stored_name TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    size INTEGER NOT NULL,
    path TEXT NOT NULL,
    storage_provider TEXT NOT NULL,
    checksum TEXT NOT NULL,
    metadata TEXT, -- JSON
    access_level TEXT NOT NULL,
    uploaded_at TEXT NOT NULL,
    expires_at TEXT,
    processed_at TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX idx_files_owner ON files(owner_id);
CREATE INDEX idx_files_mime_type ON files(mime_type);
CREATE INDEX idx_files_expires ON files(expires_at);
CREATE INDEX idx_files_uploaded ON files(uploaded_at);

CREATE TABLE file_uploads (
    id TEXT PRIMARY KEY,
    file_id TEXT NOT NULL,
    upload_url TEXT NOT NULL,
    expires_at TEXT NOT NULL,
    fields TEXT, -- JSON
    status TEXT NOT NULL,
    created_at TEXT NOT NULL,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE INDEX idx_uploads_status ON file_uploads(status);
CREATE INDEX idx_uploads_expires ON file_uploads(expires_at);

CREATE TABLE file_processings (
    id TEXT PRIMARY KEY,
    file_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    status TEXT NOT NULL,
    result_id TEXT,
    error TEXT,
    started_at TEXT,
    completed_at TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE,
    FOREIGN KEY (result_id) REFERENCES files(id) ON DELETE SET NULL
);

CREATE INDEX idx_processings_file ON file_processings(file_id);
CREATE INDEX idx_processings_status ON file_processings(status);
```

### 4. Ports Layer

#### HTTP Handlers

```go
// File upload (multipart/form-data)
// POST /api/v1/files/upload - upload small file directly
// POST /api/v1/files/upload-url - request presigned upload URL
// POST /api/v1/files/upload/:upload_id/confirm - confirm upload completion

// File management
// GET /api/v1/files/:id - get file metadata
// GET /api/v1/files/:id/download - get download URL
// GET /api/v1/files/:id/view - view file (inline)
// DELETE /api/v1/files/:id - delete file

// File listing
// GET /api/v1/files - list my files
// GET /api/v1/files?owner_id=... - list files by owner (admin)

// File processing
// POST /api/v1/files/:id/process - request processing
// GET /api/v1/files/:id/processings - get processing history
// GET /api/v1/processings/:id - get processing status
```

### 5. Integration with Tasks Context

Files context integrates with Tasks for async processing:

```go
// File processing handler for Tasks
type ProcessFileHandler struct {
    fileRepo        FileRepository
    processingRepo  ProcessingRepository
    storage         StorageProvider
    imageProcessor  ImageProcessor
}

func (h *ProcessFileHandler) Execute(ctx context.Context, payload TaskPayload) (*TaskResult, error) {
    processingID, ok := payload["processing_id"].(string)
    if !ok {
        return nil, fmt.Errorf("missing processing_id")
    }
    
    processing, err := h.processingRepo.GetByID(ctx, ProcessingID(processingID))
    if err != nil {
        return nil, err
    }
    
    file, err := h.fileRepo.GetByID(ctx, processing.FileID)
    if err != nil {
        return nil, err
    }
    
    // Download file
    data, err := h.storage.DownloadToBytes(ctx, file.Path)
    if err != nil {
        return nil, err
    }
    
    // Process file
    var result []byte
    switch processing.Operation {
    case OperationResize:
        result, err = h.imageProcessor.Resize(ctx, data, 
            *processing.Options.Width, 
            *processing.Options.Height)
    case OperationThumbnail:
        result, err = h.imageProcessor.GenerateThumbnail(ctx, data, 200, 200)
    // ... other operations
    }
    
    if err != nil {
        processing.Fail(err.Error())
        h.processingRepo.Update(ctx, processing)
        return nil, err
    }
    
    // Upload processed file
    processedFile := &File{
        OwnerID:   file.OwnerID,
        Filename:  fmt.Sprintf("processed_%s", file.Filename),
        MimeType:  "image/jpeg",
        Metadata: FileMetadata{
            OriginalID: &file.ID,
        },
    }
    
    // ... save and return
}
```

## Workflow Examples

### 1. Direct Upload (Small Files < 5MB)

```go
// Client uploads file directly
POST /api/v1/files/upload
Content-Type: multipart/form-data
file: <binary>

// Server processes and returns file ID
{
    "file_id": "019d6746...",
    "url": "/api/v1/files/019d6746.../download",
    "expires_at": "2024-12-31T23:59:59Z"
}
```

### 2. Presigned URL Upload (Large Files > 5MB)

```go
// Step 1: Request upload URL
POST /api/v1/files/upload-url
{
    "filename": "video.mp4",
    "mime_type": "video/mp4",
    "size": 104857600
}

// Response
{
    "upload_id": "upload_123",
    "upload_url": "https://s3.amazonaws.com/bucket/...",
    "fields": {
        "key": "...",
        "policy": "...",
        "signature": "..."
    },
    "expires_at": "2024-01-01T12:00:00Z"
}

// Step 2: Client uploads directly to S3
PUT <upload_url>
Content-Type: video/mp4
<binary data>

// Step 3: Confirm upload
POST /api/v1/files/upload/upload_123/confirm
{
    "checksum": "sha256:abc123..."
}

// Response
{
    "file_id": "019d6746...",
    "status": "completed"
}
```

### 3. File Processing

```go
// Request image resize
POST /api/v1/files/019d6746.../process
{
    "operation": "resize",
    "options": {
        "width": 800,
        "height": 600,
        "quality": 85
    }
}

// Response
{
    "processing_id": "proc_456",
    "status": "pending"
}

// Check status
GET /api/v1/processings/proc_456

// Response
{
    "processing_id": "proc_456",
    "file_id": "019d6746...",
    "operation": "resize",
    "status": "completed",
    "result_file_id": "019d6747...",
    "completed_at": "2024-01-01T12:05:00Z"
}
```

## Security Considerations

1. **File Validation**: Validate MIME type, extension, and actual content
2. **Size Limits**: Enforce max file size (configurable per type)
3. **Virus Scanning**: Optional integration with ClamAV
4. **Access Control**: Check permissions before download
5. **Presigned URLs**: Time-limited URLs (15-60 minutes)
6. **Checksum Verification**: Verify file integrity
7. **Path Traversal**: Prevent directory traversal attacks
8. **Rate Limiting**: Limit uploads per user per hour

## Supported File Types

### Images
- JPEG, PNG, GIF, WebP, SVG
- Max size: 10MB
- Operations: resize, crop, compress, thumbnail

### Documents
- PDF, DOCX, XLSX, PPTX
- Max size: 50MB
- Operations: thumbnail generation, preview

### Video
- MP4, WebM, MOV
- Max size: 500MB
- Operations: thumbnail, compress (via external service)

### Audio
- MP3, AAC, WAV
- Max size: 50MB
- Operations: compress, convert

## Performance Considerations

1. **CDN Integration**: Serve static files via CDN
2. **Lazy Loading**: Load files on demand
3. **Caching**: Cache presigned URLs
4. **Batch Upload**: Support multiple files in one request
5. **Thumbnail Generation**: Generate thumbnails on upload
6. **Compression**: Compress images before storage
7. **Stream Processing**: Stream large files without loading into memory

## Migration Plan

```sql
-- migrations/011_create_files.up.sql
CREATE TABLE files (...);

-- migrations/012_create_file_uploads.up.sql
CREATE TABLE file_uploads (...);

-- migrations/013_create_file_processings.up.sql
CREATE TABLE file_processings (...);
```

## Testing Strategy

### Unit Tests
- File entity validation
- Storage interface implementations (mock)
- Image processor
- Path generation

### Integration Tests
- Repository operations
- S3/GCS interactions (with test containers)
- Presigned URL generation
- File upload/download flow

### End-to-End Tests
- Full upload workflow
- Processing workflow
- Cleanup workflow

## Deployment Considerations

### Development
- Local filesystem storage
- In-memory processing
- Direct uploads
- No expiration cleanup

### Production
- S3/GCS storage
- Background processing via Tasks
- Presigned URLs for large files
- CDN for file delivery
- Periodic cleanup job
- Monitoring and alerting

### Configuration

```go
type StorageConfig struct {
    Provider    StorageProvider  // s3, gcs, local
    
    // S3 config
    S3Bucket    string
    S3Region    string
    S3AccessKey string
    S3SecretKey string
    S3CDNURL    string
    
    // GCS config
    GCSBucket       string
    GCSCredentials  []byte
    GCSCDNURL       string
    
    // Local config
    LocalPath       string
    LocalURL        string
    
    // Limits
    MaxFileSize     int64
    MaxImageSize    int64
    MaxVideoSize    int64
    AllowedTypes    []string
    PresignedURLTTL time.Duration
    
    // Processing
    EnableThumbnails    bool
    ThumbnailSize       int
    CompressImages      bool
    ImageQuality       int
}
```

## Consequences

### Positive
- ✅ Storage abstraction (easy to switch providers)
- ✅ Presigned URLs for secure large file uploads
- ✅ Async processing via Tasks context
- ✅ File metadata and access control
- ✅ Automatic cleanup of expired files
- ✅ CDN-ready architecture
- ✅ Multiple storage backends (S3, GCS, local)

### Negative
- ❌ Additional complexity (3 aggregates, storage abstraction)
- ❌ External dependencies (S3, GCS)
- ❌ Image processing may require external libraries

### Neutral
- Local storage sufficient for development
- Can use managed services (Cloudinary, imgix) for processing
- Files can be stored with or without owner association

## Alternatives Considered

1. **Direct Client Upload to S3**: Client uploads directly to S3
   - ✅ Less server load
   - ✅ Faster for large files
   - ❌ More complex client-side implementation
   - ❌ Need presigned URLs
   - Decided: Use this approach for large files (> 5MB)

2. **Managed File Services (Cloudinary, Filestack)**: Use external service for everything
   - ✅ Less code to maintain
   - ✅ Built-in processing
   - ❌ Vendor lock-in
   - ❌ Higher cost at scale
   - ❌ Less control

3. **Database BLOB Storage**: Store files in database
   - ✅ Single source of truth
   - ✅ Easy backup
   - ❌ Bad performance
   - ❌ Database bloat
   - ❌ Not suitable for large files

## References

- [ADR-001: Hexagonal Architecture](ADR-001-hexedral-architecture.md)
- [ADR-003: Event Bus](ADR-003-event-bus.md)
- [ADR-011: Tasks Context](ADR-011-tasks-jobs.md)
- [AWS S3 Presigned URLs](https://docs.aws.amazon.com/AmazonS3/latest/userguide/using-presigned-url.html)
- [Google Cloud Storage Signed URLs](https://cloud.google.com/storage/docs/access-control/signed-urls)
- [disintegration/imaging](https://github.com/disintegration/imaging)