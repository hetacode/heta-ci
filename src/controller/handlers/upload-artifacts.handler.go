package handlers

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

func (h *Handlers) UploadArtifactsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildID := vars["buildId"]
	jobID := vars["jobId"]
	file, handler, _ := r.FormFile("file")
	defer file.Close()

	reader, _ := zip.NewReader(file, handler.Size)
	for _, file := range reader.File {
		fr, _ := file.Open()
		b, _ := io.ReadAll(fr)
		fmt.Println(string(b))
	}
	fmt.Printf("upladed artifacts for job: %s | build: %s", jobID, buildID)

	// TODO:
	// artifacts should be save in some temporary directory for given build pipeline
}
