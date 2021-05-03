package services

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/hetacode/heta-ci/agent/utils"
)

type ArtifactsService struct {
	controllerBaseURL string
}

func NewArtifactsService(controllerBaseURL string) *ArtifactsService {
	s := &ArtifactsService{
		controllerBaseURL: controllerBaseURL,
	}

	return s
}

func (s *ArtifactsService) DownloadArtifacts(buildID string) ([]byte, error) {
	url := fmt.Sprintf("%s/download/artifacts/%s", s.controllerBaseURL, buildID)
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(res.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *ArtifactsService) UploadArtifacts(buildID, jobID string, fileBytes []byte) error {

	url := fmt.Sprintf("%s/upload/%s/%s", s.controllerBaseURL, buildID, jobID)
	filename := utils.ArtifactsFileName(buildID, jobID)

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	wr, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return err
	}
	wr.Write(fileBytes)
	writer.Close()

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if _, err = client.Do(req); err != nil {
		return err
	}

	return nil
}
