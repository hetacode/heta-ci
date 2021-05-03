package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/hetacode/heta-ci/commons"
	"github.com/hetacode/heta-ci/controller/utils"
)

func (h *Handlers) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	category := FileCategory(vars["category"])
	buildID := vars["buildId"]
	build, ok := h.Controller.Builds[buildID]
	if !ok {
		errorReponse(w, fmt.Sprintf("download | build %s doesn't exists", buildID))

		return
	}

	var b []byte
	var err error
	switch category {
	case RepoFileCategory:
		b = h.prepareAndGetCodeRepository(buildID)
	case ArtifactsFileCategory:
		b, err = h.prepareAndGetArtifacts(build)
	}

	if err != nil {
		http.Error(w, fmt.Sprintf("%s download failed | err: %s", category, err), http.StatusInternalServerError)
	}

	w.Header().Add("Content-Type", "application/zip")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.zip", category, buildID))
	w.Header().Add("Content-Length", strconv.Itoa(len(b)))
	if _, err := w.Write(b); err != nil {
		http.Error(w, fmt.Sprintf("%s download failed", category), http.StatusInternalServerError)
	}
}

func (h *Handlers) prepareAndGetCodeRepository(buildID string) []byte {
	return []byte("test prepareAndGetCodeRepository")
}

func (h *Handlers) prepareAndGetArtifacts(build *utils.PipelineBuild) ([]byte, error) {
	artifactsFilePath := build.ArtifactsDir + "/artifacts.zip"
	exists, err := commons.IsFileExists(artifactsFilePath)
	if err != nil {
		return nil, fmt.Errorf("get artifacts file exists failed: %s", err)
	}
	if !exists {
		return make([]byte, 0), nil
	}

	bytes, err := os.ReadFile(artifactsFilePath)
	if err != nil {
		return nil, fmt.Errorf("get artifacts failed: %s", err)
	}

	return bytes, nil
}
