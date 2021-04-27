package commons

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// TODO: add unit tests

func ExtractDirectory(artifactsBytes []byte, dirPath string) error {
	reader := bytes.NewReader(artifactsBytes)
	zipReader, err := zip.NewReader(reader, int64(len(artifactsBytes)))
	if err != nil {
		return fmt.Errorf("ExtractDirectory - create zip reader err: %s", err)
	}

	for _, file := range zipReader.File {
		p := path.Join(dirPath, file.Name)
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(p, 0777); err != nil {
				return fmt.Errorf("ExtractDirectory - create dir err: %s", err)
			}
			continue
		}

		f, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		defer f.Close()
		if err != nil {
			return fmt.Errorf("ExtractDirectory - create file err: %s", err)
		}

		archiveFileReader, _ := file.Open()
		defer archiveFileReader.Close()
		if _, err := io.Copy(f, archiveFileReader); err != nil {
			return fmt.Errorf("ExtractDirectory - save content to file err: %s", err)
		}
	}

	return nil
}

func ArchiveDirectory(path string) ([]byte, error) {
	var buf bytes.Buffer

	zipWriter := zip.NewWriter(&buf)
	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		p := strings.TrimPrefix(filePath, path)
		zipFile, err := zipWriter.Create(p)
		if err != nil {
			return err
		}
		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if err = zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func IsFileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}

	return err == nil, err
}
