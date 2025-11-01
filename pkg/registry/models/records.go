package models

import (
	"gorm.io/gorm"

	"gorm.io/datatypes"
)

type Asset struct {
	gorm.Model
	MimeType        string
	SizeBytes       int64			
	Extra           datatypes.JSON   `gorm:"type:jsonb"`
	Checksum        string           `gorm:"uniqueIndex;not null;size:64"`
	Display         string           `gorm:"size:500"`
	State           Status           `gorm:"type:status;not null;default:'pending'"`
	Tags            []Tag            `gorm:"many2many:asset_tags;"`
	DatasetVersions []DatasetVersion `gorm:"many2many:asset_dataset_versions;"`
	Peers           []Peer           `gorm:"many2many:asset_peers;"`
}

type Tag struct {
	gorm.Model
	Name   string  `gorm:"uniqueIndex;not null;size:100"`
	Assets []Asset `gorm:"many2many:asset_tags;"`
}

type Dataset struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null;size:100"`
	Description string
	Versions    []DatasetVersion
}

type DatasetVersion struct {
	gorm.Model
	Display     string `gorm:"size:100"`
	Number      int    `gorm:"not null;uniqueIndex:idx_dataset_number"`
	Description string
	DatasetID   uint `gorm:"not null;uniqueIndex:idx_dataset_number;index"` // Add regular index too
	Dataset     Dataset
	Assets      []Asset `gorm:"many2many:dataset_version_assets;"`
}

type Peer struct {
	gorm.Model
	Name    string   `gorm:"uniqueIndex;not null;size:200"`
	Display string   `gorm:"size:200;not null"`
	Type    PeerType `gorm:"type:peer_type;not null;default:'default'"`
	Assets  []Asset  `gorm:"many2many:asset_peers;"`
}
