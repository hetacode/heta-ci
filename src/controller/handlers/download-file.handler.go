package handlers

import (
	"archive/zip"
	"bytes"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func (h *Handlers) DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	category := FileCategory(vars["category"])
	buildID := vars["buildId"]

	var b []byte
	switch category {
	case RepoFileCategory:
		b = h.prepareAndGetCodeRepository(buildID)
	case ArtifactsFileCategory:
		b = h.prepareAndGetArtifacts(buildID)
	}

	// TOOD: just for test
	// in real implementation that should create archive for all files for code repository or pipline artifacts
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	zc, _ := zw.Create("test.txt")
	zc.Write(b)
	zw.Close()

	w.Header().Add("Content-Type", "application/zip")
	w.Header().Add("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.zip", category, buildID))
	w.Header().Add("Content-Length", strconv.Itoa(buf.Len()))
	if _, err := w.Write(buf.Bytes()); err != nil {
		http.Error(w, fmt.Sprintf("%s download failed", category), http.StatusInternalServerError)
	}
}

func (h *Handlers) prepareAndGetCodeRepository(buildID string) []byte {
	return []byte("test prepareAndGetCodeRepository")
}

func (h *Handlers) prepareAndGetArtifacts(buildID string) []byte {
	panic("unimplemented prepareAndGetArtifacts")
}
