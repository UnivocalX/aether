package registry

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// Ping verifies S3 connection
func (engine *Engine) PingStorage(ctx context.Context) error {
	_, err := engine.S3Client.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(engine.bucket),
	})
	if err != nil {
		return fmt.Errorf("bucket access: %w", err)
	}
	return nil
}

// IngressKey generates the ingress S3 key path
func (engine *Engine) IngressKey(sha256 string) string {
	return path.Join(engine.prefix, "ingress", sha256)
}

// CuratedKey generates the curated S3 key path
func (engine *Engine) CuratedKey(sha256 string) string {
	return path.Join(engine.prefix, "curated", sha256)
}

// IngressURL generates a presigned URL for upload (default expiry)
func (engine *Engine) IngressURL(ctx context.Context, sha256 string) (*PresignedUrl, error) {
	return engine.IngressUrlExpire(ctx, sha256, DEFAULT_PRESIGN_TTL)
}

func (engine *Engine) IngressUrlExpire(ctx context.Context, sha256 string, expire time.Duration) (*PresignedUrl, error) {
	key := engine.IngressKey(sha256)

	input := &s3.PutObjectInput{
		Bucket:            aws.String(engine.bucket),
		Key:               aws.String(key),
		ChecksumAlgorithm: types.ChecksumAlgorithmSha256,
		ChecksumSHA256:    aws.String(sha256),
	}

	res, err := engine.PresignClient.PresignPutObject(ctx, input,
		s3.WithPresignExpires(expire),
	)

	if err != nil {
		return nil, fmt.Errorf("failed to generate presign put url: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(expire)

	presignUrl := &PresignedUrl{
		URL:       Secret(res.URL),
		ExpiresAt: expiresAt,
		ExpiresIn: expire,
		Checksum:  sha256,
		Key:       key,
		Operation: "put",
		Bucket:    engine.bucket,
	}

	slog.Debug("Generated PUT URL with checksum validation",
		"sha256", sha256,
		"expire", expire,
		"expiresAt", expiresAt)

	return presignUrl, nil
}

// CuratedUrl generates a presigned URL for download (default expiry)
func (engine *Engine) CuratedUrl(ctx context.Context, sha256 string) (*PresignedUrl, error) {
	return engine.CuratedUrlExpire(ctx, sha256, DEFAULT_PRESIGN_TTL)
}

// CuratedUrlExpire generates a presigned URL for download with custom expiry
func (engine *Engine) CuratedUrlExpire(ctx context.Context, sha256 string, expire time.Duration) (*PresignedUrl, error) {
	key := engine.CuratedKey(sha256)

	input := &s3.GetObjectInput{
		Bucket: aws.String(engine.bucket),
		Key:    aws.String(key),
	}

	res, err := engine.PresignClient.PresignGetObject(ctx, input,
		s3.WithPresignExpires(expire),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presign get url: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(expire)

	presignUrl := &PresignedUrl{
		URL:       Secret(res.URL),
		ExpiresAt: expiresAt,
		ExpiresIn: expire,
		Checksum:  sha256,
		Key:       key,
		Operation: "get",
		Bucket:    engine.bucket,
	}

	slog.Debug("Generated GET URL",
		"sha256", sha256,
		"expire", expire,
		"expiresAt", expiresAt)

	return presignUrl, nil
}
