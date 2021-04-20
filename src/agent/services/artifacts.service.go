package services

import (
	"bytes"
	"fmt"
	"net/http"
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

func (s *ArtifactsService) UploadArtifacts(buildID, jobID string, fileBytes []byte) error {
	var buf bytes.Buffer
	_, err := buf.Write(fileBytes)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/upload/%s/%s", s.controllerBaseURL, buildID, jobID)
	req, err := http.NewRequest(http.MethodPost, url, &buf)
	if err != nil {
		return err
	}
	client := &http.Client{}
	if _, err = client.Do(req); err != nil {
		return err
	}

	return nil
}
