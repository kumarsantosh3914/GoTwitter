package services

import (
	env "GoTwitter/config/env"
	db "GoTwitter/db/repositories"
	apperrors "GoTwitter/errors"
	"GoTwitter/models"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MediaService interface {
	CreateUpload(ctx context.Context, userID int64, filename string, contentType string, sizeBytes int64) (*models.MediaUpload, error)
}

type MediaServiceImpl struct {
	mediaRepository db.MediaRepository
	presignClient   *s3.PresignClient
	bucket          string
	publicBaseURL   string
	presignExpiry   time.Duration
}

func NewMediaService(mediaRepository db.MediaRepository) (MediaService, error) {
	region := env.GetString("AWS_REGION", "")
	bucket := env.GetString("AWS_S3_BUCKET", "")
	accessKeyID := env.GetString("AWS_ACCESS_KEY_ID", "")
	secretAccessKey := env.GetString("AWS_SECRET_ACCESS_KEY", "")

	if region == "" || bucket == "" || accessKeyID == "" || secretAccessKey == "" {
		return &MediaServiceImpl{mediaRepository: mediaRepository}, nil
	}

	cfg, err := awsconfig.LoadDefaultConfig(
		context.Background(),
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)
	publicBaseURL := strings.TrimSuffix(env.GetString("AWS_S3_PUBLIC_BASE_URL", fmt.Sprintf("https://%s.s3.%s.amazonaws.com", bucket, region)), "/")

	return &MediaServiceImpl{
		mediaRepository: mediaRepository,
		presignClient:   s3.NewPresignClient(client),
		bucket:          bucket,
		publicBaseURL:   publicBaseURL,
		presignExpiry:   15 * time.Minute,
	}, nil
}

func (m *MediaServiceImpl) CreateUpload(ctx context.Context, userID int64, filename string, contentType string, sizeBytes int64) (*models.MediaUpload, error) {
	if m.presignClient == nil || m.bucket == "" {
		return nil, apperrors.NewAppError("aws s3 media uploads are not configured", http.StatusNotImplemented, nil)
	}
	if strings.TrimSpace(filename) == "" {
		return nil, apperrors.NewAppError("filename is required", http.StatusBadRequest, nil)
	}
	if strings.TrimSpace(contentType) == "" {
		return nil, apperrors.NewAppError("content_type is required", http.StatusBadRequest, nil)
	}
	if sizeBytes < 1 {
		return nil, apperrors.NewAppError("size_bytes must be greater than zero", http.StatusBadRequest, nil)
	}

	s3Key, err := buildMediaKey(userID, filename)
	if err != nil {
		return nil, apperrors.NewAppError("failed to generate media key", http.StatusInternalServerError, err)
	}
	publicURL := m.publicBaseURL + "/" + s3Key

	request, err := m.presignClient.PresignPutObject(
		ctx,
		&s3.PutObjectInput{
			Bucket:      aws.String(m.bucket),
			Key:         aws.String(s3Key),
			ContentType: aws.String(contentType),
		},
		func(options *s3.PresignOptions) {
			options.Expires = m.presignExpiry
		},
	)
	if err != nil {
		return nil, apperrors.NewAppError("failed to generate upload url", http.StatusInternalServerError, err)
	}

	attachment, err := m.mediaRepository.Create(ctx, &models.MediaAttachment{
		UserId:    userID,
		S3Key:     s3Key,
		Url:       publicURL,
		MimeType:  contentType,
		SizeBytes: sizeBytes,
	})
	if err != nil {
		return nil, apperrors.NewAppError("failed to persist media attachment", http.StatusInternalServerError, err)
	}

	headers := map[string]string{}
	for key, values := range request.SignedHeader {
		if len(values) == 0 {
			continue
		}
		headers[key] = strings.Join(values, ",")
	}

	return &models.MediaUpload{
		Attachment: attachment,
		UploadURL:  request.URL,
		Method:     http.MethodPut,
		Headers:    headers,
		ExpiresIn:  int64(m.presignExpiry / time.Second),
	}, nil
}

func buildMediaKey(userID int64, filename string) (string, error) {
	randomBytes := make([]byte, 12)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	extension := strings.ToLower(filepath.Ext(filename))
	extension = strings.TrimSpace(extension)
	if extension == "" {
		extension = ".bin"
	}

	return fmt.Sprintf("users/%d/%d-%s%s", userID, time.Now().Unix(), hex.EncodeToString(randomBytes), extension), nil
}

var _ *v4.PresignedHTTPRequest
