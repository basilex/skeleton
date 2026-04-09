package domain

import (
	"errors"
	"fmt"
	"time"
)

type VersionNumber int

func NewVersionNumber(n int) (VersionNumber, error) {
	if n < 1 {
		return 0, errors.New("version number must be positive")
	}
	return VersionNumber(n), nil
}

func (v VersionNumber) Int() int {
	return int(v)
}

func (v VersionNumber) String() string {
	return fmt.Sprintf("v%d", v)
}

func (v VersionNumber) Next() VersionNumber {
	return v + 1
}

type ChangeType string

const (
	ChangeTypeCreate    ChangeType = "create"
	ChangeTypeUpdate    ChangeType = "update"
	ChangeTypeSign      ChangeType = "sign"
	ChangeTypeArchive   ChangeType = "archive"
	ChangeTypeUnarchive ChangeType = "unarchive"
)

func (c ChangeType) String() string {
	return string(c)
}

type DocumentVersion struct {
	version     VersionNumber
	changeType  ChangeType
	changedBy   string
	changedAt   time.Time
	description string
	checksum    string
	fileID      string
}

func NewDocumentVersion(
	version VersionNumber,
	changeType ChangeType,
	changedBy string,
	description string,
	checksum string,
	fileID string,
) (*DocumentVersion, error) {
	if version < 1 {
		return nil, errors.New("version number must be positive")
	}
	if changedBy == "" {
		return nil, errors.New("changed by is required")
	}

	return &DocumentVersion{
		version:     version,
		changeType:  changeType,
		changedBy:   changedBy,
		changedAt:   time.Now().UTC(),
		description: description,
		checksum:    checksum,
		fileID:      fileID,
	}, nil
}

func (v *DocumentVersion) GetVersion() VersionNumber {
	return v.version
}

func (v *DocumentVersion) GetChangeType() ChangeType {
	return v.changeType
}

func (v *DocumentVersion) GetChangedBy() string {
	return v.changedBy
}

func (v *DocumentVersion) GetChangedAt() time.Time {
	return v.changedAt
}

func (v *DocumentVersion) GetDescription() string {
	return v.description
}

func (v *DocumentVersion) GetChecksum() string {
	return v.checksum
}

func (v *DocumentVersion) GetFileID() string {
	return v.fileID
}

func (v *DocumentVersion) String() string {
	return fmt.Sprintf("DocumentVersion{version=%s, changeType=%s, by=%s}",
		v.version, v.changeType, v.changedBy)
}
