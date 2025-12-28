package registry

import (
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"

	"gorm.io/datatypes"
)

type Asset struct {
	gorm.Model
	Checksum string         `gorm:"uniqueIndex;not null;size:64;check:checksum <> ''"`
	Display  string         `gorm:"size:500"`
	Extra    datatypes.JSON `gorm:"type:jsonb"`

	MimeType        string
	SizeBytes       int64
	State           Status           `gorm:"type:status;not null;default:'pending'"`

	Tags            []Tag            `gorm:"many2many:asset_tags;"`
	DatasetVersions []DatasetVersion `gorm:"many2many:asset_dataset_versions;"`
	Peers           []Peer           `gorm:"many2many:asset_peers;"`
}

func (a *Asset) SetExtra(extra map[string]any) error {
	if len(extra) == 0 {
		return fmt.Errorf("cannot set empty extra value")
	}
	
	// Get existing list
	var extraList []map[string]any
	if len(a.Extra) > 0 {
		json.Unmarshal(a.Extra, &extraList)
	}
	
	// Add date to the new data
	extra["date"] = time.Now()
	
	// Append to list
	extraList = append(extraList, extra)
	
	// Marshal back
	jsonData, err := json.Marshal(extraList)
	if err != nil {
		return fmt.Errorf("failed to marshal extra data: %w", err)
	}

	a.Extra = datatypes.JSON(jsonData)
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
	Name    string  `gorm:"uniqueIndex;not null;size:200"`
	Display string  `gorm:"size:200;not null"`
	Type    string  `gorm:"not null;default:'default'"`
	Assets  []Asset `gorm:"many2many:asset_peers;"`
}
