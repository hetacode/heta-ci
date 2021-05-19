package services

import (
	"bytes"
	"fmt"
	"net/http"
)

type RepositoryService struct {
	controllerBaseURL string
}

func NewRepositoryService(controllerBaseURL string) *RepositoryService {
	s := &RepositoryService{
		controllerBaseURL: controllerBaseURL,
	}

	return s
}

func (s *RepositoryService) DownloadRepositoryPackage(buildID string) ([]byte, error) {
	url := fmt.Sprintf("%s/download/repo/%s", s.controllerBaseURL, buildID)
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
