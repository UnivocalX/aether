package registry

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

// ValidateString checks if a name matches the allowed pattern
func ValidateString(name string) bool {
	var ValidNamePattern = regexp.MustCompile(`^[a-z0-9/.:_-]+$`)
	return ValidNamePattern.MatchString(name)
}

// NormalizeString converts to lowercase and trims spaces
func NormalizeString(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

func GeneratePeerName(id uint, baseName string, peerType string) string {
	// Clean the base name for use in the peer name
	cleanBaseName := strings.ReplaceAll(baseName, " ", "-")
	cleanBaseName = NormalizeString(cleanBaseName)
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
func (a *Asset) BeforeCreate(tx *gorm.DB) error {
	a.MimeType = NormalizeString(a.MimeType)
	a.Checksum = NormalizeString(a.Checksum)

	// Validate checksum on creation
	if err := ValidateSHA256(a.Checksum); err != nil {
		return err
	}

	// Set initial state
	a.State = StatusPending
	return nil
}

func (a *Asset) BeforeUpdate(tx *gorm.DB) error {
	if a.State == "" {
		return fmt.Errorf("%w: asset status is required", ErrValidation)
	}

	// Prevent checksum modification after creation
	if tx.Statement.Changed("Checksum") {
		return fmt.Errorf("%w: checksum cannot be modified after creation", ErrValidation)
	}
	return nil
}

// Hooks for tag
func (t *Tag) BeforeSave(tx *gorm.DB) error {
	t.Name = NormalizeString(t.Name)
	if !ValidateString(t.Name) {
		return fmt.Errorf("%w: tag name contains invalid characters", ErrValidation)
	}
	return nil
}

// BeforeSave hook to normalize dataset name
func (d *Dataset) BeforeSave(tx *gorm.DB) error {
	d.Name = NormalizeString(d.Name)
	if !ValidateString(d.Name) {
		return fmt.Errorf("%w: dataset name contains invalid characters", ErrValidation)
	}
	return nil
}

// BeforeSave hook for DatasetVersion name normalization
func (dv *DatasetVersion) BeforeSave(tx *gorm.DB) error {
	dv.Display = NormalizeString(dv.Display)
	if !ValidateString(dv.Display) {
		return fmt.Errorf("%w: dataset version name contains invalid characters", ErrValidation)
	}
	return nil
}

// BeforeCreate hook for DatasetVersion - auto-increment version number
func (dv *DatasetVersion) BeforeCreate(tx *gorm.DB) error {
	var maxVersion int
	err := tx.Model(&DatasetVersion{}).
		Where("dataset_id = ?", dv.DatasetID).
		Select("COALESCE(MAX(number), 0)").
		Scan(&maxVersion).Error

	if err != nil {
		return err
	}
	
	dv.Number = maxVersion + 1
	return nil
}

// BeforeCreate hook for Peer
func (p *Peer) BeforeCreate(tx *gorm.DB) error {
	// Set default type if empty
	if p.Type == "" {
		p.Type = "default"
	}

	// Normalize base name
	p.Type = NormalizeString(p.Type)
	if !ValidateString(p.Type) {
		return fmt.Errorf("%w: peer Type contains invalid characters", ErrValidation)
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
