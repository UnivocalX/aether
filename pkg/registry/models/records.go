package models

import (
	"gorm.io/gorm"
)

type Status string

const (
	StatusPending Status = "pending"
	StatusReady   Status = "ready"
	StatusDeleted Status = "deleted"
)

// IsValid checks if the status is a valid AssetStatus
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusReady, StatusDeleted:
		return true
	}
	return false
}

type Asset struct {
	gorm.Model
	MimeType        string
	SizeBytes       int64
	Checksum        string           `gorm:"uniqueIndex;not null;size:64"`
	Key             string           `gorm:"uniqueIndex;not null"`
	State           Status           `gorm:"type:status;not null;default:'pending'"`
	Tags            []Tag            `gorm:"many2many:asset_tags;"`
	Peers           []Peer           `gorm:"many2many:asset_peers;"`
	DatasetVersions []DatasetVersion `gorm:"many2many:asset_dataset_versions;"`
}

type Tag struct {
	gorm.Model
	Name   string  `gorm:"uniqueIndex;not null;size:100;check:name ~ '^[a-z0-9/.:_-]+$'"`
	Assets []Asset `gorm:"many2many:asset_tags;"`
}

type Dataset struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null;size:100;check:name ~ '^[a-z0-9/.:_-]+$'"`
	Description string
	Versions    []DatasetVersion
}

type DatasetVersion struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null;size:100;check:name ~ '^[a-z0-9/.:_-]+$'"`
	Description string
	DatasetID   uint `gorm:"not null"`
	Dataset     Dataset
	Assets      []Asset `gorm:"many2many:dataset_version_assets;"`
}

type Peer struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null;size:100;check:name ~ '^[a-z0-9/.:_-]+$'"`
	Description string
	Assets      []Asset `gorm:"many2many:peer_assets;"`
}
