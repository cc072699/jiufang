package upload

import (
	"mime/multipart"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFile_ValidJPG(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     1 * 1024 * 1024,
	}

	err := validateFile(file)

	assert.NoError(t, err)
}

func TestValidateFile_ValidPNG(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "test.png",
		Size:     1 * 1024 * 1024,
	}

	err := validateFile(file)

	assert.NoError(t, err)
}

func TestValidateFile_ValidGIF(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "test.gif",
		Size:     1 * 1024 * 1024,
	}

	err := validateFile(file)

	assert.NoError(t, err)
}

func TestValidateFile_ValidJPEG(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "test.jpeg",
		Size:     1 * 1024 * 1024,
	}

	err := validateFile(file)

	assert.NoError(t, err)
}

func TestValidateFile_InvalidFileType(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "test.txt",
		Size:     1 * 1024 * 1024,
	}

	err := validateFile(file)

	assert.Error(t, err)
	assert.Equal(t, ErrInvalidFileType, err)
}

func TestValidateFile_FileTooLarge(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     3 * 1024 * 1024,
	}

	err := validateFile(file)

	assert.Error(t, err)
	assert.Equal(t, ErrFileTooLarge, err)
}

func TestValidateFile_EdgeCase_ExactMaxSize(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     MaxFileSize,
	}

	err := validateFile(file)

	assert.NoError(t, err)
}

func TestValidateFile_EdgeCase_OneByteOverMaxSize(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "test.jpg",
		Size:     MaxFileSize + 1,
	}

	err := validateFile(file)

	assert.Error(t, err)
	assert.Equal(t, ErrFileTooLarge, err)
}

func TestValidateFile_CaseInsensitive(t *testing.T) {
	file := &multipart.FileHeader{
		Filename: "test.JPG",
		Size:     1 * 1024 * 1024,
	}

	err := validateFile(file)

	assert.NoError(t, err)
}