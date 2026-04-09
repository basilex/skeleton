package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/basilex/skeleton/internal/documents/domain"
	"github.com/basilex/skeleton/pkg/pagination"
	"github.com/georgysavva/scany/v2/pgxscan"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DocumentRepository struct {
	pool *pgxpool.Pool
	psql squirrel.StatementBuilderType
}

func NewDocumentRepository(pool *pgxpool.Pool) *DocumentRepository {
	return &DocumentRepository{
		pool: pool,
		psql: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *DocumentRepository) Save(ctx context.Context, document *domain.Document) error {
	var fileID *string
	if document.GetFileID() != nil {
		fileID = document.GetFileID()
	}

	status := document.GetStatus().String()
	docType := document.GetDocumentType().String()
	metadata := document.GetMetadata()
	if metadata == nil {
		metadata = make(map[string]string)
	}

	query, args, err := r.psql.Insert("documents").
		Columns("id", "document_number", "document_type", "reference_id", "file_id", "status", "metadata", "created_at", "updated_at").
		Values(document.GetID().String(), document.GetDocumentNumber(), docType, document.GetReferenceID(),
			fileID, status, metadata, document.GetCreatedAt(), document.GetUpdatedAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET status = EXCLUDED.status, file_id = EXCLUDED.file_id, " +
			"metadata = EXCLUDED.metadata, updated_at = EXCLUDED.updated_at").
		ToSql()
	if err != nil {
		return fmt.Errorf("build insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save document: %w", err)
	}

	// Save signatures
	for _, sig := range document.GetSignatures() {
		if err := r.saveSignature(ctx, sig); err != nil {
			return fmt.Errorf("save signature: %w", err)
		}
	}

	return nil
}

func (r *DocumentRepository) saveSignature(ctx context.Context, signature *domain.Signature) error {
	status := signature.GetStatus().String()
	var signedAt *time.Time
	if signature.GetSignedAt() != nil {
		signedAt = signature.GetSignedAt()
	}

	query, args, err := r.psql.Insert("document_signatures").
		Columns("id", "document_id", "signer_name", "signer_role", "status", "signed_at", "signature_data", "created_at").
		Values(signature.GetID().String(), signature.GetDocumentID().String(), signature.GetSignerName(),
			signature.GetSignerRole(), status, signedAt, signature.GetSignatureData(), signature.GetCreatedAt()).
		Suffix("ON CONFLICT(id) DO UPDATE SET status = EXCLUDED.status, signed_at = EXCLUDED.signed_at, " +
			"signature_data = EXCLUDED.signature_data").
		ToSql()
	if err != nil {
		return fmt.Errorf("build signature insert query: %w", err)
	}

	_, err = r.pool.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("save signature: %w", err)
	}

	return nil
}

func (r *DocumentRepository) FindByID(ctx context.Context, id domain.DocumentID) (*domain.Document, error) {
	var dto documentDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, document_number, document_type, reference_id, file_id, status, metadata, created_at, updated_at 
		 FROM documents WHERE id = $1`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find document by id: %w", err)
	}

	// Load signatures
	var sigDTOs []signatureDTO
	err = pgxscan.Select(ctx, r.pool, &sigDTOs,
		`SELECT id, document_id, signer_name, signer_role, status, signed_at, signature_data, created_at 
		 FROM document_signatures WHERE document_id = $1 ORDER BY created_at`, id.String())
	if err != nil {
		return nil, fmt.Errorf("find signatures: %w", err)
	}

	sigPtrs := make([]*signatureDTO, len(sigDTOs))
	for i := range sigDTOs {
		sigPtrs[i] = &sigDTOs[i]
	}

	return dto.toDomain(sigPtrs)
}

func (r *DocumentRepository) FindByDocumentNumber(ctx context.Context, documentNumber string) (*domain.Document, error) {
	var dto documentDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, document_number, document_type, reference_id, file_id, status, metadata, created_at, updated_at 
		 FROM documents WHERE document_number = $1`, documentNumber)
	if err != nil {
		return nil, fmt.Errorf("find document by number: %w", err)
	}

	id, _ := domain.ParseDocumentID(dto.ID)
	return r.FindByID(ctx, id)
}

func (r *DocumentRepository) FindByReferenceID(ctx context.Context, referenceID string) (*domain.Document, error) {
	var dto documentDTO
	err := pgxscan.Get(ctx, r.pool, &dto,
		`SELECT id, document_number, document_type, reference_id, file_id, status, metadata, created_at, updated_at 
		 FROM documents WHERE reference_id = $1`, referenceID)
	if err != nil {
		return nil, fmt.Errorf("find document by reference: %w", err)
	}

	id, _ := domain.ParseDocumentID(dto.ID)
	return r.FindByID(ctx, id)
}

func (r *DocumentRepository) FindAll(ctx context.Context, filter domain.DocumentFilter) (pagination.PageResult[*domain.Document], error) {
	q := r.psql.Select("id", "document_number", "document_type", "reference_id", "file_id", "status", "metadata", "created_at", "updated_at").
		From("documents")

	if filter.DocumentType != nil {
		q = q.Where(squirrel.Eq{"document_type": filter.DocumentType.String()})
	}
	if filter.Status != nil {
		q = q.Where(squirrel.Eq{"status": filter.Status.String()})
	}
	if filter.ReferenceID != nil {
		q = q.Where(squirrel.Eq{"reference_id": *filter.ReferenceID})
	}
	if filter.CreatedAfter != nil {
		q = q.Where(squirrel.GtOrEq{"created_at": *filter.CreatedAfter})
	}
	if filter.CreatedBefore != nil {
		q = q.Where(squirrel.LtOrEq{"created_at": *filter.CreatedBefore})
	}
	if filter.Cursor != "" {
		q = q.Where(squirrel.Lt{"id": filter.Cursor})
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = pagination.DefaultLimit
	}

	q = q.OrderBy("id DESC").Limit(uint64(limit + 1))

	query, args, err := q.ToSql()
	if err != nil {
		return pagination.PageResult[*domain.Document]{}, fmt.Errorf("build query: %w", err)
	}

	var dtos []documentDTO
	if err := pgxscan.Select(ctx, r.pool, &dtos, query, args...); err != nil {
		return pagination.PageResult[*domain.Document]{}, fmt.Errorf("select documents: %w", err)
	}

	documents := make([]*domain.Document, 0, len(dtos))
	for _, dto := range dtos {
		id, _ := domain.ParseDocumentID(dto.ID)
		doc, err := r.FindByID(ctx, id)
		if err != nil {
			return pagination.PageResult[*domain.Document]{}, err
		}
		documents = append(documents, doc)
	}

	return pagination.NewPageResult(documents, limit), nil
}

func (r *DocumentRepository) Delete(ctx context.Context, id domain.DocumentID) error {
	// Delete signatures first
	_, err := r.pool.Exec(ctx, `DELETE FROM document_signatures WHERE document_id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete signatures: %w", err)
	}

	// Delete document
	result, err := r.pool.Exec(ctx, `DELETE FROM documents WHERE id = $1`, id.String())
	if err != nil {
		return fmt.Errorf("delete document: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrDocumentNotFound
	}

	return nil
}
