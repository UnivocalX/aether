package models

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

var (
	ErrValidation = errors.New("validation error")
)

type PeerType string

const (
	DefaultPeerType PeerType = "default"
)

type Status string

const (
	StatusRejected Status = "rejected"
	StatusPending  Status = "pending"
	StatusReady    Status = "ready"
	StatusDeleted  Status = "deleted"
)

// IsValid checks if the status is a valid AssetStatus
func (s Status) IsValid() bool {
	switch s {
	case StatusPending, StatusReady, StatusDeleted:
		return true
	}
	return false
}

// ValidNamePattern defines allowed characters for names
var ValidNamePattern = regexp.MustCompile(`^[a-z0-9/.:_-]+$`)

// ValidateName checks if a name matches the allowed pattern
func ValidateName(name string) bool {
	return ValidNamePattern.MatchString(name)
}

// NormalizeName converts to lowercase and trims spaces
func NormalizeName(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func GeneratePeerName(id uint, baseName string, peerType PeerType) string {
	// Clean the base name for use in the peer name
	cleanBaseName := strings.ReplaceAll(baseName, " ", "-")
	cleanBaseName = NormalizeName(cleanBaseName)
	return fmt.Sprintf("%s-%s-%d", cleanBaseName, peerType, id)
}

func ValidateSHA256(hash string) error {
	h := strings.TrimSpace(hash)
	if len(h) != sha256.Size*2 {
		return fmt.Errorf("invalid SHA256 hash length: %d", len(h))
	}
	if _, err := hex.DecodeString(h); err != nil {
		return fmt.Errorf("invalid SHA256 hex: %w", err)
	}
	return nil
}

// Hooks for Asset
func (a *Asset) BeforeSave(tx *gorm.DB) error {
	a.Display = NormalizeName(a.Display)
	if !ValidateName(a.Display) {
		return fmt.Errorf("%w: asset Display contains invalid characters", ErrValidation)
	}

	// Set default state if empty
	if a.State == "" {
		a.State = StatusPending
	}

	// Validate state
	if !a.State.IsValid() {
		return fmt.Errorf("invalid asset state: %s", a.State)
	}

	if err := ValidateSHA256(a.Checksum); err != nil {
		return err
	}

	return nil
}

func (a *Asset) BeforeUpdate(tx *gorm.DB) error {
	if a.MimeType == "" {
		return fmt.Errorf("%w: asset MimeType cannot be empty", ErrValidation)
	}
	
	if a.SizeBytes <= 0 {
		return fmt.Errorf("%w: asset SizeBytes must be positive", ErrValidation)
	}

	return nil
}

// BeforeSave hook to normalize tag name
func (t *Tag) BeforeSave(tx *gorm.DB) error {
	t.Name = NormalizeName(t.Name)
	if !ValidateName(t.Name) {
		return fmt.Errorf("%w: tag name contains invalid characters", ErrValidation)
	}
	return nil
}

// BeforeSave hook to normalize dataset name
func (d *Dataset) BeforeSave(tx *gorm.DB) error {
	d.Name = NormalizeName(d.Name)
	if !ValidateName(d.Name) {
		return fmt.Errorf("%w: dataset name contains invalid characters", ErrValidation)
	}
	return nil
}

// BeforeSave hook for DatasetVersion name normalization
func (dv *DatasetVersion) BeforeSave(tx *gorm.DB) error {
	dv.Display = NormalizeName(dv.Display)
	if !ValidateName(dv.Display) {
		return fmt.Errorf("%w: dataset version name contains invalid characters", ErrValidation)
	}
	return nil
}

// BeforeCreate hook for DatasetVersion - auto-increment version number
func (dv *DatasetVersion) BeforeCreate(tx *gorm.DB) error {
	if dv.Number == 0 {
		var maxVersion int
		err := tx.Model(&DatasetVersion{}).
			Where("dataset_id = ?", dv.DatasetID).
			Select("COALESCE(MAX(number), 0)").
			Scan(&maxVersion).Error
		if err != nil {
			return err
		}
		dv.Number = maxVersion + 1
	}
	return nil
}

// Utility methods
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

// BeforeCreate hook for Peer
func (p *Peer) BeforeCreate(tx *gorm.DB) error {
	// Set default type if empty
	if p.Type == "" {
		p.Type = DefaultPeerType
	}

	// Normalize base name
	p.Display = NormalizeName(p.Display)
	if !ValidateName(p.Display) {
		return fmt.Errorf("%w: peer base name contains invalid characters", ErrValidation)
	}

	return nil
}

// AfterCreate hook for Peer - generates the name using the assigned ID and user-provided base name
func (p *Peer) AfterCreate(tx *gorm.DB) error {
	if p.Name == "" {
		p.Name = GeneratePeerName(p.ID, p.Display, p.Type)

		// Update the peer with the generated name
		return tx.Model(p).Update("name", p.Name).Error
	}
	return nil
}
