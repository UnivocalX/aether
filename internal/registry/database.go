package registry

import (
	"fmt"
	"log/slog"

	"gorm.io/gorm"
)

func (e *Engine) WithTx(tx *gorm.DB) *Engine {
	if tx == nil {
		return e
	}

	ne := *e
	ne.DatabaseClient = tx

	return &ne
}

func (engine *Engine) GetAssetRecord(sha256 string) (*Asset, error) {
	slog.Debug("Getting asset", "checksum", sha256)
	normalizedSha256 := NormalizeString(sha256)

	var asset Asset
	if err := engine.DatabaseClient.Where("checksum = ?", normalizedSha256).First(&asset).Error; err != nil {
		return nil, fmt.Errorf("get asset %q: %w", sha256, err)
	}

	return &asset, nil
}

func (engine *Engine) GetAssetRecordTags(sha256 string) ([]*Tag, error) {
	slog.Debug("Getting asset tags", "checksum", sha256)

	asset, err := engine.GetAssetRecord(sha256)
	if err != nil {
		return nil, err
	}

	var tags []*Tag
	if err := engine.DatabaseClient.Model(asset).Association("Tags").Find(&tags); err != nil {
		return nil, fmt.Errorf("get asset %q tags: %w", sha256, err)
	}

	return tags, nil
}

func (engine *Engine) CreateAssetRecord(asset *Asset) error {
	slog.Debug("Creating asset record", "display", asset.Display, "checksum", asset.Checksum)

	if err := engine.DatabaseClient.Create(asset).Error; err != nil {
		return fmt.Errorf("create asset %q: %w", asset.Checksum, err)
	}

	return nil
}

func (engine *Engine) AttachTags(asset *Asset, tags []*Tag) error {
	slog.Debug("Attempting to attach tags to asset", "AssetID", asset.ID, "tagCount", len(tags))

	if len(tags) == 0 {
		return nil
	}

	// Append all tags at once
	if err := engine.DatabaseClient.Model(asset).Association("Tags").Append(tags); err != nil {
		return fmt.Errorf("attach tags %q: %w", asset.Checksum, err)
	}

	// Single reload for all tags
	if err := engine.DatabaseClient.Preload("Tags").First(asset, asset.ID).Error; err != nil {
		return err
	}

	return nil
}

func (engine *Engine) DetachTags(asset *Asset, tags []*Tag) error {
	slog.Debug("Attempting to detach tags from asset", "AssetID", asset.ID, "tagCount", len(tags))

	if len(tags) == 0 {
		return nil
	}

	// Delete all tags at once
	if err := engine.DatabaseClient.Model(asset).Association("Tags").Delete(tags); err != nil {
		return fmt.Errorf("detach tags %q: %w", asset.Checksum, err)
	}

	// Single reload for all tags
	if err := engine.DatabaseClient.Preload("Tags").First(asset, asset.ID).Error; err != nil {
		return err
	}

	return nil
}

func (engine *Engine) CreateTagRecord(name string) (*Tag, error) {
	slog.Debug("creating a new tag", "name", name)

	tag := &Tag{Name: name}
	if err := engine.DatabaseClient.Create(tag).Error; err != nil {
		return nil, fmt.Errorf("create tag %q: %w", name, err)
	}

	return tag, nil
}

func (engine *Engine) GetTagRecord(name string) (*Tag, error) {
	normalizedName := NormalizeString(name)
	slog.Debug("Getting tag", "name", name)

	var tag Tag
	if err := engine.DatabaseClient.Where("name = ?", normalizedName).First(&tag).Error; err != nil {
		return nil, fmt.Errorf("get tag %q: %w", name, err)
	}

	return &tag, nil
}

func (engine *Engine) GetTagRecordAssets(name string, limit int, offset int) ([]*Asset, error) {
	tag, err := engine.GetTagRecord(name)
	if err != nil {
		return nil, err
	}

	var assets []*Asset
	if err := engine.DatabaseClient.Limit(limit).Offset(offset).Model(tag).Association("Assets").Find(&assets); err != nil {
		return nil, fmt.Errorf("get tag %q assets: %w", name, err)
	}

	return assets, nil
}

func (engine *Engine) GetTagRecordById(id uint) (*Tag, error) {
	slog.Debug("Getting tag", "id", id)

	var tag Tag
	if err := engine.DatabaseClient.First(&tag, id).Error; err != nil {
		return nil, fmt.Errorf("get tag by ID %d: %w", id, err)
	}

	return &tag, nil
}

// GetTagsByNames fetches multiple tags by their names in a single query
func (engine *Engine) GetTagsByNames(names []string) ([]*Tag, error) {
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
	err := engine.DatabaseClient.Where("name IN ?", normalized).Find(&tags).Error

	if err != nil {
		return nil, err
	}

	return tags, nil
}

func (engine *Engine) ListAssetsRecords(opts ...SearchAssetsOption) ([]*Asset, error) {
	slog.Debug("Listing assets", "totalOptions", len(opts))

	query, err := NewSearchAssetsQuery(opts...)
	if err != nil {
		slog.Debug("failed to build query")
		return nil, err
	}
	slog.Debug("created new query", "query", query)

	// Start query with base filters
	tx := engine.DatabaseClient.Model(&Asset{}).Where(&Asset{
		MimeType: query.MimeType,
		State:    query.State,
	})

	// Filter by checksums
	if len(query.CheckSums) > 0 {
		tx = tx.Where("checksum IN ?", query.CheckSums)
	}

	// IncludedTags: Filter assets that have ALL specified tags (AND logic)
	if len(query.IncludedTags) > 0 {
		subQuery := engine.DatabaseClient.
			Table("assets").
			Select("assets.id").
			Joins("JOIN asset_tags ON asset_tags.asset_id = assets.id").
			Joins("JOIN tags ON tags.id = asset_tags.tag_id").
			Where("tags.name IN ?", query.IncludedTags).
			Group("assets.id").
			Having("COUNT(*) = ?", len(query.IncludedTags))

		tx = tx.Where("id IN (?)", subQuery)
	}

	// ExcludedTags: Filter out assets that have ANY of these tags
	if len(query.ExcludedTags) > 0 {
		subQuery := engine.DatabaseClient.
			Table("assets").
			Select("assets.id").
			Joins("JOIN asset_tags ON asset_tags.asset_id = assets.id").
			Joins("JOIN tags ON tags.id = asset_tags.tag_id").
			Where("tags.name IN ?", query.ExcludedTags)

		tx = tx.Where("id NOT IN (?)", subQuery)
	}

	// Pagination
	if query.Cursor > 0 {
		tx = tx.Where("id > ?", query.Cursor)
	}
	tx = tx.Limit(int(query.Limit))

	// Execute query with preloaded tags
	var assets []*Asset
	if err := tx.Preload("Tags").Order("id ASC").Find(&assets).Error; err != nil {
		return nil, err
	}

	return assets, nil
}

func (engine *Engine) CreateDatasetRecord(name string, description string) (*Dataset, error) {
	slog.Debug("creating a new dataset", "name", name)

	ds := &Dataset{Name: name, Description: description}
	if err := engine.DatabaseClient.Create(ds).Error; err != nil {
		return nil, fmt.Errorf("create dataset %q: %w", name, err)
	}

	return ds, nil
}

func (engine *Engine) GetDatasetRecord(name string) (*Dataset, error) {
	slog.Debug("getting dataset", "dataset", name)
	
	var ds *Dataset
	if err := engine.DatabaseClient.Where(&Dataset{Name: name}).First(ds).Error; err != nil {
		return nil, fmt.Errorf("get dataset %q: %w", name, err)
	}

	return ds, nil
}

func (engine *Engine) CreateDatasetVersionRecord(datasetName string, description string) (*DatasetVersion, error) {
	slog.Debug("creating a new dataset version", "dataset", datasetName)

	// Get dataset
	ds, err := engine.GetDatasetRecord(datasetName)
	if err != nil {
		return nil, err
	}

	// Create version
	dsv := &DatasetVersion{
		DatasetID:   ds.ID,
		Dataset:     *ds,
		Description: description,
	}

	if err := engine.DatabaseClient.Create(dsv).Error; err != nil {
		return nil, fmt.Errorf("create dataset version %q: %w", datasetName, err)
	}

	return dsv, nil
}

func (engine *Engine) CreateAssetRecords(assets ...*Asset) error {
	slog.Debug("creating new assets", "total", len(assets))
	if err := engine.DatabaseClient.Create(assets).Error; err != nil {
		return fmt.Errorf("create assets: %w", err)
	}

	return nil
}
