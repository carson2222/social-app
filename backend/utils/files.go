package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

func GenerateFileName() (string, error) {
	timestamp := time.Now().UnixNano()

	// Generate random bytes
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	input := fmt.Sprintf("%d%s", timestamp, hex.EncodeToString(randomBytes))

	// Encrypt name with SHA256
	hash := sha256.New()
	hash.Write([]byte(input))
	hashedName := hex.EncodeToString(hash.Sum(nil))

	return hashedName, nil
}

func ValidateFileType(file multipart.File) (string, error) {
	// Read a small portion of the file to detect its MIME type
	buffer := make([]byte, 512) // 512 bytes are enough to sniff the content type
	_, err := file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Reset the file read pointer after reading
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	// Detect the content type (MIME type)
	mimeType := http.DetectContentType(buffer)

	// Check if it's a valid PNG or JPG/JPEG
	if mimeType == "image/png" {
		return "png", nil
	}
	if mimeType == "image/jpeg" {
		return "jpg", nil
	}

	return "", nil
}

func UploadProfilePicture(r *http.Request) (string, error) {
	// Parse file
	r.ParseMultipartForm(10 << 20) // Ograniczenie do 10MB
	file, _, err := r.FormFile("profile_picture")
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Validate if it's a PNG or JPG file
	imgType, err := ValidateFileType(file)
	if err != nil {
		return "", err
	}

	if imgType != "png" && imgType != "jpg" {
		return "", errors.New("only pnh and jpg files are allowed")
	}

	// Save image to uploads folder
	fileName, err := GenerateFileName()
	if err != nil {
		return "", err
	}

	filePath := fmt.Sprintf("./uploads/pfp/%s", fileName+"."+imgType)
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		return "", err
	}

	return filePath, nil
}
