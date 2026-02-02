package registry


func AssetsToChecksums(assets ...*Asset) []*string {
	checksums := make([]*string, len(assets))

	for i, a := range assets {
		checksums[i] = &a.Checksum
	}

	return checksums
}