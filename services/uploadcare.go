package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"socialmedia/config"
	"socialmedia/models"
	"strings"
)

type UploadcareService struct {
	PublicKey string
	SecretKey string
	BaseURL   string
}

type UploadcareResponse struct {
	FileID   string `json:"uuid"`
	URL      string `json:"original_file_url"`
	MimeType string `json:"mime_type"`
}

var uploadcareService *UploadcareService

func GetUploadcareService() *UploadcareService {
	if uploadcareService == nil {
		uploadcareService = NewUploadcareService(
			config.UploadcarePublicKey,
			config.UploadcareSecretKey,
		)
	}
	return uploadcareService
}

func NewUploadcareService(publicKey, secretKey string) *UploadcareService {
	return &UploadcareService{
		PublicKey: publicKey,
		SecretKey: secretKey,
		BaseURL:   "https://upload.uploadcare.com",
	}
}

func (s *UploadcareService) UploadFile(file *multipart.FileHeader) (*UploadcareResponse, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Create a new multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add the MIME file
	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return nil, err
	}

	// Copy the uploaded file to the form field
	_, err = io.Copy(part, src)
	if err != nil {
		return nil, err
	}

	// Add public key to form
	writer.WriteField("UPLOADCARE_PUB_KEY", s.PublicKey)
	writer.WriteField("UPLOADCARE_STORE", "1") // Auto store the file

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	// Create the request
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/base/", s.BaseURL), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Parse the response
	var result UploadcareResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func DetermineMediaType(mimeType string) models.MediaType {
	switch {
	case strings.HasPrefix(mimeType, "image/gif"):
		return models.GifType
	case strings.HasPrefix(mimeType, "image/"):
		return models.ImageType
	case strings.HasPrefix(mimeType, "video/"):
		return models.VideoType
	default:
		return models.ImageType
	}
}
