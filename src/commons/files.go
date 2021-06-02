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

type FileDataIter interface {
	ForEach(func(file *FileData) error) error
}

type FileData struct {
	Reader io.Reader
	Path   string
}

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

		if err := os.MkdirAll(filepath.Dir(p), 0777); err != nil {
			if !os.IsExist(err) {
				return fmt.Errorf("ExtractDirectory - create dir err: %s", err)
			}
		}

		archiveFileReader, err := file.Open()
		if err != nil {
			return fmt.Errorf("Cannot open archive file %s | err: %s", file.Name, err)
		}
		b := bytes.Buffer{}
		if _, err := b.ReadFrom(archiveFileReader); err != nil {
			return fmt.Errorf("Cannot read from archive file %s | err: %s", file.Name, err)
		}
		if err := os.WriteFile(p, b.Bytes(), file.Mode()); err != nil {
			return fmt.Errorf("ExtractDirectory - create file err: %s", err)
		}
	}

	return nil
}

func ArchiveFiles(iter FileDataIter) ([]byte, error) {
	var buf bytes.Buffer

	zipWriter := zip.NewWriter(&buf)
	err := iter.ForEach(func(file *FileData) error {
		zipFile, err := zipWriter.Create(file.Path)
		if err != nil {
			return fmt.Errorf("zipWriter.Create err:  %s", err)
		}
		_, err = io.Copy(zipFile, file.Reader)
		if err != nil {
			return fmt.Errorf("io.Copy zipFile file.Reader err %s", err)
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
