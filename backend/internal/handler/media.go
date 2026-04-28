package handler

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"storefront/backend/internal/middleware"
)

type MediaHandler struct {
	bucketName    string
	client        *s3.Client
	presignClient *s3.PresignClient
	log           *slog.Logger
}

func NewMediaHandler(bucketName, s3API, accessKey, secretKey string, log *slog.Logger) *MediaHandler {
	handler := &MediaHandler{
		bucketName: strings.TrimSpace(bucketName),
		log:        log,
	}

	if handler.bucketName == "" || s3API == "" || accessKey == "" || secretKey == "" {
		return handler
	}

	client := s3.New(s3.Options{
		Region:       "auto",
		BaseEndpoint: aws.String(strings.TrimRight(strings.TrimSpace(s3API), "/")),
		Credentials:  aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
		UsePathStyle: true,
	})

	handler.client = client
	handler.presignClient = s3.NewPresignClient(client)
	return handler
}

// POST /media/upload-url
// Returns a one-time R2 presigned upload URL plus the app-served final URL.
func (h *MediaHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	if h.presignClient == nil || h.bucketName == "" {
		respondErr(w, http.StatusServiceUnavailable, "image upload not configured")
		return
	}

	tenant := middleware.TenantFromCtx(r.Context())
	if tenant == nil {
		respondErr(w, http.StatusUnauthorized, "missing tenant context")
		return
	}

	var req struct {
		Filename    string  `json:"filename" validate:"required"`
		ContentType string  `json:"content_type" validate:"required"`
		ProductID   *string `json:"product_id"`
	}
	if !decodeValid(w, r, &req) {
		return
	}

	ext := strings.ToLower(path.Ext(req.Filename))
	if ext == "" {
		ext = ".bin"
	}

	productSegment := "staged"
	if req.ProductID != nil {
		if parsed, err := uuid.Parse(strings.TrimSpace(*req.ProductID)); err == nil {
			productSegment = parsed.String()
		}
	}

	key := fmt.Sprintf("tenants/%s/products/%s/%s%s", tenant.ID, productSegment, uuid.NewString(), ext)

	presigned, err := h.presignClient.PresignPutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      aws.String(h.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(req.ContentType),
	}, func(options *s3.PresignOptions) {
		options.Expires = 15 * time.Minute
	})
	if err != nil {
		serverErr(w, h.log, r, err)
		return
	}

	respond(w, http.StatusOK, map[string]string{
		"key":        key,
		"upload_url": presigned.URL,
		"public_url": h.objectURL(r, key),
	})
}

func (h *MediaHandler) DeleteObjectByURL(ctx context.Context, objectURL string) error {
	if h.client == nil || h.bucketName == "" {
		return nil
	}

	parsed, err := url.Parse(strings.TrimSpace(objectURL))
	if err != nil {
		return err
	}

	key := strings.TrimSpace(parsed.Query().Get("key"))
	if key == "" {
		return fmt.Errorf("missing object key")
	}

	_, err = h.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(h.bucketName),
		Key:    aws.String(key),
	})
	return err
}

// GET /media/object?key=...
func (h *MediaHandler) GetObject(w http.ResponseWriter, r *http.Request) {
	if h.client == nil || h.bucketName == "" {
		respondErr(w, http.StatusServiceUnavailable, "image delivery not configured")
		return
	}

	key := strings.TrimSpace(r.URL.Query().Get("key"))
	if key == "" {
		respondErr(w, http.StatusBadRequest, "missing object key")
		return
	}

	obj, err := h.client.GetObject(r.Context(), &s3.GetObjectInput{
		Bucket: aws.String(h.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		respondErr(w, http.StatusNotFound, "image not found")
		return
	}
	defer obj.Body.Close()

	if obj.ContentType != nil && *obj.ContentType != "" {
		w.Header().Set("Content-Type", *obj.ContentType)
	}
	if obj.ContentLength != nil && *obj.ContentLength > 0 {
		w.Header().Set("Content-Length", fmt.Sprintf("%d", *obj.ContentLength))
	}
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")

	if _, err := io.Copy(w, obj.Body); err != nil {
		serverErr(w, h.log, r, err)
	}
}

func (h *MediaHandler) objectURL(r *http.Request, key string) string {
	return h.publicBase(r) + "/media/object?key=" + url.QueryEscape(key)
}

func (h *MediaHandler) publicBase(r *http.Request) string {
	scheme := "http"
	if forwardedProto := strings.TrimSpace(r.Header.Get("X-Forwarded-Proto")); forwardedProto != "" {
		scheme = forwardedProto
	} else if r.TLS != nil {
		scheme = "https"
	}

	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = r.Host
	}

	return scheme + "://" + host
}
