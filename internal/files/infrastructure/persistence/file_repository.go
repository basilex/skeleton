package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/files/domain"
	identitydomain "github.com/basilex/skeleton/internal/identity/domain"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FileRepository struct {
	pool *pgxpool.Pool
	psql sq.StatementBuilderType
}

func NewFileRepository(pool *pgxpool.Pool) *FileRepository {
	return &FileRepository{
		pool: pool,
		psql: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

type fileDTO struct {
	ID              string     `db:"id"`
	OwnerID         *string    `db:"owner_id"`
	Filename        string     `db:"filename"`
	StoredName      string     `db:"stored_name"`
	MimeType        string     `db:"mime_type"`
	Size            int64      `db:"size"`
	Path            string     `db:"path"`
	StorageProvider string     `db:"storage_provider"`
	Checksum        string     `db:"checksum"`
	Metadata        []byte     `db:"metadata"`
	AccessLevel     string     `db:"access_level"`
	ScanStatus      string     `db:"scan_status"`
	ThreatInfo      string     `db:"threat_info"`
	ScannedAt       *time.Time `db:"scanned_at"`
	UploadedAt      time.Time  `db:"uploaded_at"`
	ExpiresAt       *time.Time `db:"expires_at"`
	ProcessedAt     *time.Time `db:"processed_at"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

func (r *FileRepository) Create(ctx context.Context, file *domain.File) error {
	metadataJSON, err := json.Marshal(file.Metadata())
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	var ownerID *string
	if file.OwnerID() != nil {
		oid := file.OwnerID().String()
		ownerID = &oid
	}

	query, args, err := r.psql.Insert("files").
		Columns("id", "owner_id", "filename", "stored_name", "mime_type", "size", "path",
			"storage_provider", "checksum", "metadata", "access_level", "scan_status",
			"threat_info", "scanned_at",
			"uploaded_at", "expires_at", "processed_at", "created_at", "updated_at").
		Values(file.ID().String(), ownerID, file.Filename(), file.StoredName(), file.MimeType(),
			file.Size(), file.Path(), string(file.StorageProvider()), file.Checksum(),
			metadataJSON, string(file.AccessLevel()), string(file.ScanStatus()),
			file.ThreatInfo(), file.ScannedAt(),
			file.UploadedAt(), file.ExpiresAt(),
			file.ProcessedAt(), file.CreatedAt(), file.UpdatedAt()).
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("insert file: %w", err)
	}

	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id domain.FileID) (*domain.File, error) {
	var dto fileDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, owner_id, filename, stored_name, mime_type, size, path,
			storage_provider, checksum, metadata, access_level, scan_status,
			threat_info, scanned_at,
			uploaded_at, expires_at, processed_at, created_at, updated_at
		FROM files WHERE id = $1`,
		id.String())
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, domain.ErrFileNotFound
		}
		return nil, fmt.Errorf("get file by id: %w", err)
	}

	return r.dtoToDomain(dto)
}

func (r *FileRepository) Update(ctx context.Context, file *domain.File) error {
	metadataJSON, err := json.Marshal(file.Metadata())
	if err != nil {
		return fmt.Errorf("marshal metadata: %w", err)
	}

	query, args, err := r.psql.Update("files").
		Set("filename", file.Filename()).
		Set("stored_name", file.StoredName()).
		Set("mime_type", file.MimeType()).
		Set("size", file.Size()).
		Set("path", file.Path()).
		Set("storage_provider", string(file.StorageProvider())).
		Set("checksum", file.Checksum()).
		Set("metadata", metadataJSON).
		Set("access_level", string(file.AccessLevel())).
		Set("expires_at", file.ExpiresAt()).
		Set("processed_at", file.ProcessedAt()).
		Set("updated_at", time.Now()).
		Where(sq.Eq{"id": file.ID().String()}).
		ToSql()
	if err != nil {
		return fmt.Errorf("build update query: %w", err)
	}

	result, err := r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("update file: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrFileNotFound
	}

	return nil
}

func (r *FileRepository) Delete(ctx context.Context, id domain.FileID) error {
	result, err := r.pool.Exec(ctx, `DELETE FROM files WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrFileNotFound
	}

	return nil
}

func (r *FileRepository) GetByPath(ctx context.Context, path string) (*domain.File, error) {
	var dto fileDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, owner_id, filename, stored_name, mime_type, size, path,
			storage_provider, checksum, metadata, access_level, scan_status,
			threat_info, scanned_at,
			uploaded_at, expires_at, processed_at, created_at, updated_at
		FROM files WHERE path = $1`,
		path)
	if err != nil {
		if pgxscan.NotFound(err) {
			return nil, domain.ErrFileNotFound
		}
		return nil, fmt.Errorf("get file by path: %w", err)
	}

	return r.dtoToDomain(dto)
}

func (r *FileRepository) GetByOwner(ctx context.Context, ownerID string, limit, offset int) ([]*domain.File, error) {
	var dtos []fileDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, owner_id, filename, stored_name, mime_type, size, path,
			storage_provider, checksum, metadata, access_level, scan_status,
			threat_info, scanned_at,
			uploaded_at, expires_at, processed_at, created_at, updated_at
		FROM files 
		WHERE owner_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`,
		ownerID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("query files: %w", err)
	}

	return r.dtosToDomains(dtos)
}

func (r *FileRepository) GetExpired(ctx context.Context, before time.Time, limit int) ([]*domain.File, error) {
	var dtos []fileDTO
	err := pgxscan.Select(ctx, r.pool, &dtos,
		`SELECT id, owner_id, filename, stored_name, mime_type, size, path,
			storage_provider, checksum, metadata, access_level, scan_status,
			threat_info, scanned_at,
			uploaded_at, expires_at, processed_at, created_at, updated_at
		FROM files 
		WHERE expires_at IS NOT NULL AND expires_at < $1
		ORDER BY expires_at ASC
		LIMIT $2`,
		before, limit)
	if err != nil {
		return nil, fmt.Errorf("query expired files: %w", err)
	}

	return r.dtosToDomains(dtos)
}

func (r *FileRepository) List(ctx context.Context, filter *domain.FileFilter, limit, offset int) ([]*domain.File, error) {
	q := r.psql.Select("id", "owner_id", "filename", "stored_name", "mime_type", "size", "path",
		"storage_provider", "checksum", "metadata", "access_level", "scan_status",
		"threat_info", "scanned_at",
		"uploaded_at", "expires_at", "processed_at", "created_at", "updated_at").
		From("files")

	if filter != nil {
		if filter.OwnerID != nil {
			q = q.Where(sq.Eq{"owner_id": *filter.OwnerID})
		}
		if filter.MimeType != nil {
			q = q.Where(sq.ILike{"mime_type": *filter.MimeType + "%"})
		}
		if filter.AccessLevel != nil {
			q = q.Where(sq.Eq{"access_level": string(*filter.AccessLevel)})
		}
		if filter.UploadedFrom != nil {
			q = q.Where(sq.GtOrEq{"uploaded_at": *filter.UploadedFrom})
		}
		if filter.UploadedTo != nil {
			q = q.Where(sq.LtOrEq{"uploaded_at": *filter.UploadedTo})
		}
	}

	q = q.OrderBy("created_at DESC").Limit(uint64(limit)).Offset(uint64(offset))

	query, args, err := q.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var dtos []fileDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return nil, fmt.Errorf("query files: %w", err)
	}

	return r.dtosToDomains(dtos)
}

func (r *FileRepository) DeleteBatch(ctx context.Context, ids []domain.FileID) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, id := range ids {
		_, err := tx.Exec(ctx, `DELETE FROM files WHERE id = $1`, id.String())
		if err != nil {
			return fmt.Errorf("delete file %s: %w", id, err)
		}
	}

	return tx.Commit(ctx)
}

func (r *FileRepository) Count(ctx context.Context, filter *domain.FileFilter) (int64, error) {
	q := r.psql.Select("COUNT(*)").From("files")

	if filter != nil {
		if filter.OwnerID != nil {
			q = q.Where(sq.Eq{"owner_id": *filter.OwnerID})
		}
		if filter.MimeType != nil {
			q = q.Where(sq.ILike{"mime_type": *filter.MimeType + "%"})
		}
		if filter.AccessLevel != nil {
			q = q.Where(sq.Eq{"access_level": string(*filter.AccessLevel)})
		}
	}

	query, args, err := q.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	var count int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count files: %w", err)
	}

	return count, nil
}

func (r *FileRepository) dtoToDomain(dto fileDTO) (*domain.File, error) {
	var metadata domain.FileMetadata
	if len(dto.Metadata) > 0 {
		if err := json.Unmarshal(dto.Metadata, &metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}

	var ownerIDPtr *identitydomain.UserID
	if dto.OwnerID != nil && *dto.OwnerID != "" {
		uid, parseErr := identitydomain.ParseUserID(*dto.OwnerID)
		if parseErr != nil {
			return nil, fmt.Errorf("parse user id: %w", parseErr)
		}
		ownerIDPtr = &uid
	}

	fileIDParsed, err := domain.ParseFileID(dto.ID)
	if err != nil {
		return nil, fmt.Errorf("parse file id: %w", err)
	}

	return domain.ReconstituteFile(
		fileIDParsed,
		ownerIDPtr,
		dto.Filename,
		dto.StoredName,
		dto.MimeType,
		dto.Size,
		dto.Path,
		domain.StorageProvider(dto.StorageProvider),
		dto.Checksum,
		metadata,
		domain.AccessLevel(dto.AccessLevel),
		domain.ScanStatus(dto.ScanStatus),
		dto.ThreatInfo,
		dto.ScannedAt,
		dto.UploadedAt,
		dto.ExpiresAt,
		dto.ProcessedAt,
		dto.CreatedAt,
		dto.UpdatedAt,
	), nil
}

func (r *FileRepository) dtosToDomains(dtos []fileDTO) ([]*domain.File, error) {
	files := make([]*domain.File, 0, len(dtos))
	for _, dto := range dtos {
		file, err := r.dtoToDomain(dto)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}
