package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func (h *Handlers) UploadArtifactsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildID := vars["buildId"]
	jobID := vars["jobId"]
	file, _, _ := r.FormFile("file")
	defer file.Close()

	build, ok := h.Controller.Builds[buildID]
	if !ok {
		errorReponse(w, fmt.Sprintf("upload artifacts | build %s doesn't exists", buildID))

		return
	}

	var buf []byte
	buf, err := io.ReadAll(file)
	if err != nil {
		errorReponse(w, fmt.Sprintf("upload artifacts | read file err %s", err))

		return
	}

	filePath := build.ArtifactsDir + "/artifacts.zip"

	os.RemoveAll(build.ArtifactsDir)
	os.MkdirAll(build.ArtifactsDir, 0777)

	if err := os.WriteFile(filePath, buf, 0644); err != nil {
		errorReponse(w, fmt.Sprintf("upload artifacts | save artifacts archive failed err %s", err))
		return
	}

	log.Printf("uploaded artifacts for job: %s | build: %s | path: %s", jobID, buildID, filePath)
}

func errorReponse(w http.ResponseWriter, msg string) error {
	log.Printf(msg)
	w.Write([]byte(msg))
	w.WriteHeader(http.StatusBadRequest)

	return fmt.Errorf(msg)
}
