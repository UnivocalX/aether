package registry

import (
	"encoding/json"
	"fmt"

	"gorm.io/gorm"

	"gorm.io/datatypes"
)

type Asset struct {
	gorm.Model
	Checksum string         `gorm:"uniqueIndex;not null;size:64;check:checksum <> ''"`
	Display  string         `gorm:"size:120"`
	Extra    datatypes.JSON `gorm:"type:jsonb"`

	MimeType  string
	SizeBytes int64
	State     Status `gorm:"type:status;not null;default:'pending'"`

	Tags            []Tag            `gorm:"many2many:asset_tags;"`
	DatasetVersions []DatasetVersion `gorm:"many2many:asset_dataset_versions;"`
	Peers           []Peer           `gorm:"many2many:asset_peers;"`
}

func (a *Asset) SetExtra(extra map[string]any) error {
	if len(extra) == 0 {
		return fmt.Errorf("cannot set empty extra value")
	}

	// Marshal back
	data, err := json.Marshal(extra)
	if err != nil {
		return fmt.Errorf("failed to marshal extra data: %w", err)
	}

	a.Extra = datatypes.JSON(data)
	return nil
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

func (d *Dataset) LatestVersion(tx *gorm.DB) (*DatasetVersion, error) {
	var latestVersion DatasetVersion
	err := tx.Where("dataset_id = ?", d.ID).
		Order("version_number DESC").
		First(&latestVersion).Error

	if err != nil {
		return nil, err
	}
	
	return &latestVersion, nil
}

type DatasetVersion struct {
	gorm.Model
	Number      int    `gorm:"not null;uniqueIndex:idx_dataset_number"`
	Display     string `gorm:"size:100"`
	Description string
	DatasetID   uint `gorm:"not null;uniqueIndex:idx_dataset_number;index"`
	Dataset     Dataset
	Assets      []Asset `gorm:"many2many:dataset_version_assets;"`
}

type Peer struct {
	gorm.Model
	Name    string  `gorm:"uniqueIndex;not null;size:200"`
	Display string  `gorm:"size:200;not null"`
	Type    string  `gorm:"not null;default:'default'"`
	Assets  []Asset `gorm:"many2many:asset_peers;"`
}
