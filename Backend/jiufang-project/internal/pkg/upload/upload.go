package upload

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	ErrInvalidFileType = errors.New("invalid file type, only JPG/PNG/GIF are allowed")
	ErrFileTooLarge    = errors.New("file too large, maximum size is 2MB")
	ErrCreateDir       = errors.New("failed to create upload directory")
	ErrSaveFile        = errors.New("failed to save file")
)

const (
	MaxFileSize = 2 * 1024 * 1024
	UploadDir   = "uploads/avatars"
)

var allowedTypes = map[string]bool{
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"gif":  true,
}

type UploadResult struct {
	URL      string
	FileName string
}

func UploadAvatar(file *multipart.FileHeader, userID uint) (*UploadResult, error) {
	if err := validateFile(file); err != nil {
		return nil, err
	}

	if err := ensureUploadDir(); err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext == ".jpeg" {
		ext = ".jpg"
	}

	fileName := fmt.Sprintf("%d_%d%s", userID, time.Now().Unix(), ext)
	filePath := filepath.Join(UploadDir, fileName)

	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSaveFile, err)
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSaveFile, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrSaveFile, err)
	}

	return &UploadResult{
		URL:      "/" + filePath,
		FileName: fileName,
	}, nil
}

func validateFile(file *multipart.FileHeader) error {
	if file.Size > MaxFileSize {
		return ErrFileTooLarge
	}

	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
	if !allowedTypes[ext] {
		return ErrInvalidFileType
	}

	return nil
}

func ensureUploadDir() error {
	if err := os.MkdirAll(UploadDir, 0755); err != nil {
		return fmt.Errorf("%w: %v", ErrCreateDir, err)
	}
	return nil
}