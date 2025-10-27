package models

import (
	"fmt"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"
)

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

func generatePeerName() string {
	// Simple timestamp-based name
	return fmt.Sprintf("peer-%d", time.Now().Unix())
}

// BeforeCreate hook to normalize tag name
func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	t.Name = NormalizeName(t.Name)
	if !ValidateName(t.Name) {
		return fmt.Errorf("tag name contains invalid characters")
	}
	return nil
}

// BeforeCreate hook to normalize dataset name
func (d *Dataset) BeforeCreate(tx *gorm.DB) error {
	d.Name = NormalizeName(d.Name)
	if !ValidateName(d.Name) {
		return fmt.Errorf("dataset name contains invalid characters")
	}
	return nil
}

// BeforeCreate hook for DatasetVersion
func (dv *DatasetVersion) BeforeCreate(tx *gorm.DB) error {
	dv.Name = NormalizeName(dv.Name)
	if !ValidateName(dv.Name) {
		return fmt.Errorf("dataset version name contains invalid characters")
	}

	return nil
}

// BeforeCreate hook to generate peer name if not provided
func (p *Peer) BeforeCreate(tx *gorm.DB) error {
	if p.Name == "" {
		p.Name = NormalizeName(generatePeerName())
	} else {
		p.Name = NormalizeName(p.Name)
	}

	if !ValidateName(p.Name) {
		return fmt.Errorf("peer name contains invalid characters")
	}
	return nil
}
