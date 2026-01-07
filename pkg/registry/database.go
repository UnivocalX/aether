package registry

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

func GetAssetRecord(db *gorm.DB, sha256 string) (*Asset, error) {
	slog.Debug("Getting asset", "checksum", sha256)
	normalizedSha256 := NormalizeString(sha256)

	var asset Asset
	if err := db.Where("checksum = ?", normalizedSha256).First(&asset).Error; err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %w", sha256, err)
	}

	return &asset, nil
}

func GetAssetRecordTags(db *gorm.DB, sha256 string) ([]*Tag, error) {
	slog.Debug("Getting asset tags", "checksum", sha256)

	asset, err := GetAssetRecord(db, sha256)
	if err != nil {
		return nil, err
	}

	var tags []*Tag
	if err := db.Model(asset).Association("Tags").Find(&tags); err != nil {
		return nil, fmt.Errorf("failed to get asset %s tags: %w", sha256, err)
	}

	return tags, nil
}

func (engine *Engine) CreateAssetRecord(
	db *gorm.DB,
	sha256 string,
	display string,
	extra map[string]any,
) (*Asset, error) {
	slog.Debug("Creating asset record", "display", display, "checksum", sha256)

	asset := &Asset{
		Checksum: sha256,
		Display:  display,
	}

	// check for extra values
	if extra != nil {
		if err := asset.SetExtra(extra); err != nil {
			return nil, fmt.Errorf("failed setting new asset %s extra field: %w", sha256, err)
		}
	}

	if err := db.Create(asset).Error; err != nil {
		return nil, fmt.Errorf("failed creating new asset %s: %w", sha256, err)
	}

	return asset, nil
}

func (engine *Engine) UpdateAssetRecord(db *gorm.DB, asset *Asset) error {
	if asset.Checksum == "" {
		return fmt.Errorf("asset checksum is required for update")
	}

	result := db.Model(asset).Updates(asset)
	if result.Error != nil {
		return fmt.Errorf("failed to update asset %s: %w", asset.Checksum, result.Error)
	}

	// Reload to get fresh timestamps, but reuse the same pointer
	return db.First(asset, asset.ID).Error
}

func (engine *Engine) AttachTags(db *gorm.DB, asset *Asset, tags []*Tag) error {
	slog.Debug("Attempting to attach tags to asset", "AssetID", asset.ID, "tagCount", len(tags))

	if len(tags) == 0 {
		return nil
	}

	// Append all tags at once
	if err := db.Model(asset).Association("Tags").Append(tags); err != nil {
		return fmt.Errorf("failed to attach tags to asset %s: %w", asset.Checksum, err)
	}

	// Single reload for all tags
	if err := db.Preload("Tags").First(asset, asset.ID).Error; err != nil {
		return err
	}

	return nil
}

func (engine *Engine) DetachTags(db *gorm.DB, asset *Asset, tags []*Tag) error {
	slog.Debug("Attempting to detach tags from asset", "AssetID", asset.ID, "tagCount", len(tags))

	if len(tags) == 0 {
		return nil
	}

	// Delete all tags at once
	if err := db.Model(asset).Association("Tags").Delete(tags); err != nil {
		return fmt.Errorf("failed to detach tags from asset %s: %w", asset.Checksum, err)
	}

	// Single reload for all tags
	if err := db.Preload("Tags").First(asset, asset.ID).Error; err != nil {
		return err
	}

	return nil
}

func (engine *Engine) CreateTagRecord(db *gorm.DB, name string) (*Tag, error) {
	slog.Debug(fmt.Sprintf("Creating tag: %s", name))

	tag := &Tag{Name: name}
	if err := db.Create(tag).Error; err != nil {
		return nil, fmt.Errorf("failed to create tag %s: %w", name, err)
	}

	return tag, nil
}

func (engine *Engine) GetTagRecord(db *gorm.DB, name string) (*Tag, error) {
	normalizedName := NormalizeString(name)
	slog.Debug("Getting tag", "name", name)

	var tag Tag
	if err := db.Where("name = ?", normalizedName).First(&tag).Error; err != nil {
		return nil, fmt.Errorf("failed to get tag %s: %w", name, err)
	}

	return &tag, nil
}

func (engine *Engine) GetTagRecordAssets(db *gorm.DB, name string, limit int, offset int) ([]*Asset, error) {
	tag, err := GetTagRecord(db, name)
	if err != nil {
		return nil, err
	}

	var assets []*Asset
	if err := db.Limit(limit).Offset(offset).Model(tag).Association("Assets").Find(&assets); err != nil {
		return nil, fmt.Errorf("failed to get tag %s assets: %w", name, err)
	}

	return assets, nil
}

func (engine *Engine) GetTagRecordById(db *gorm.DB, id uint) (*Tag, error) {
	slog.Debug("Getting tag", "id", id)

	var tag Tag
	if err := db.First(&tag, id).Error; err != nil {
		return nil, fmt.Errorf("failed to get tag by ID %d: %w", id, err)
	}

	return &tag, nil
}

// GetTagsByNames fetches multiple tags by their names in a single query
func (engine *Engine) GetTagsByNames(db *gorm.DB, names []string) ([]*Tag, error) {
	slog.Debug("Getting tags", "total", len(names))
	if len(names) == 0 {
		return nil, nil
	}

	// Normalize all names
	slog.Debug("normalizing tags names")
	normalized := make([]string, 0, len(names))
	for _, name := range names {
		if n := NormalizeString(name); n != "" {
			normalized = append(normalized, n)
		}
	}

	if len(normalized) == 0 {
		return nil, nil
	}

	var tags []*Tag
	err := db.Where("name IN ?", normalized).Find(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (engine *Engine) ListAssetsRecords(db *gorm.DB, opts ...SearchAssetsOption) ([]*Asset, error) {
	slog.Debug("Listing assets", "totalOptions", len(opts))

	query, err := NewSearchAssetsQuery(opts...)
	if err != nil {
		slog.Debug("failed to build query")
		return nil, err
	}
	slog.Debug("created new query", "query", query)

	db = db.Where(&Asset{MimeType: query.MimeType, State: query.State})

	// Filter by checksums
	if len(query.CheckSums) > 0 {
		db = db.Where("Checksum IN ?", query.CheckSums)
	}

	// IncludedTags: Filter assets that have ALL specified tags (AND logic)
	if len(query.IncludedTags) > 0 {
		db = db.Where("id IN (?)", // Filter main query to only IDs from subquery
			db. // Create fresh DB connection for subquery
				Table("assets").                                             // Query the assets table
				Select("assets.id").                                         // Return only asset IDs
				Joins("JOIN asset_tags ON asset_tags.asset_id = assets.id"). // Connect assets to junction table
				Joins("JOIN tags ON tags.id = asset_tags.tag_id").           // Connect junction table to tags
				Where("tags.name IN ?", query.IncludedTags).                 // Keep only rows with required tag names
				Group("assets.id").                                          // Group rows by asset to count tags
				Having("COUNT(*) = ?", len(query.IncludedTags)),             // Only keep assets with exact tag count (ALL tags present)
		)
	}

	// ExcludedTags: Filter out assets that have ANY of these tags
	if len(query.ExcludedTags) > 0 {
		db = db.Where("id NOT IN (?)", // Exclude IDs that match subquery
			db. // Create fresh DB connection for subquery
				Table("assets").                                             // Query the assets table
				Select("assets.id").                                         // Return only asset IDs
				Joins("JOIN asset_tags ON asset_tags.asset_id = assets.id"). // Connect assets to junction table
				Joins("JOIN tags ON tags.id = asset_tags.tag_id").           // Connect junction table to tags
				Where("tags.name IN ?", query.ExcludedTags),                 // Find assets with any excluded tag
		)
	}

	// Pagination
	db = db.Where("id > ?", query.Cursor)
	db = db.Limit(int(query.Limit))

	// Execute query with preloaded tags
	var assets []*Asset
	if err := db.Preload("Tags").Order("id ASC").Find(&assets).Error; err != nil {
		return nil, err
	}
	return assets, nil
}

func (engine *Engine) CreateDatasetRecord(db *gorm.DB, name string, description string) (*Dataset, error) {
	slog.Debug("creating a new dataset", "name", name)

	ds := &Dataset{Name: name, Description: description}
	if err := db.Create(ds).Error; err != nil {
		return nil, fmt.Errorf("failed to create dataset: %w", err)
	}

	return ds, nil
}

func (engine *Engine) CreateDatasetVersionRecord(db *gorm.DB, datasetName string, versionDisplay string, description string) (*DatasetVersion, error) {
	slog.Debug("creating a new dataset version", "dataset", datasetName, "versionDisplay", versionDisplay)
	
	// Get dataset
	var ds Dataset
	if err := db.Where(&Dataset{Name: datasetName}).First(&ds).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset %s: %w", datasetName, err)
	}

	// Create version
	dsv := &DatasetVersion{
		DatasetID: ds.ID,
		Dataset: ds,
		Display: versionDisplay,
		Description: description,
	}

	if err := db.Create(dsv).Error; err != nil {
		return nil, fmt.Errorf("failed to create a new dataset version for %s: %w", datasetName, err)
	}

	return dsv, nil
}

func (engine *Engine) CreateAssetRecords(db *gorm.DB, assets ...*Asset) error {
	slog.Debug("creating new assets", "total", len(assets))
	if err := db.Create(assets).Error; err != nil {
		return fmt.Errorf("failed creating assets: %w", err)
	}

	return nil
}