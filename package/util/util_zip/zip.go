package util_zip

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func RecursiveZip(dirPath string, target string) error {
	destinationFile, err := os.Create(target)
	defer destinationFile.Close()

	if err != nil {
		return err
	}
	myZip := zip.NewWriter(destinationFile)
	defer myZip.Close()

	err = filepath.Walk(dirPath, func(filePath string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(filePath, filepath.Dir(dirPath))
		zipFile, err := myZip.Create(relPath)
		if err != nil {
			log.Print("Create err: ", err)
			return err
		}
		fsFile, err := os.Open(filePath)
		defer fsFile.Close()
		if err != nil {
			log.Print("Open err: ", err)
			return err
		}
		_, err = io.Copy(zipFile, fsFile)
		if err != nil {
			log.Print("Copy err: ", err)
			return err
		}

		return nil
	} /* end of anomymouse func */)

	if err != nil {
		return err
	}

	return nil
}

func RecursiveUnzip(zipPath string, dirPath string) error {
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer archive.Close()
	for _, file := range archive.Reader.File {
		reader, err := file.Open()
		if err != nil {
			return err
		}
		defer reader.Close()
		path := filepath.Join(dirPath, file.Name)
		// Remove file if it already exists; no problem if it doesn't; other cases can error out below
		_ = os.Remove(path)
		// Create a directory at path, including parents
		err = os.MkdirAll(path, os.ModePerm)
		if err != nil {
			return err
		}
		// If file is _supposed_ to be a directory, we're done
		if file.FileInfo().IsDir() {
			continue
		}
		// otherwise, remove that directory (_not_ including parents)
		err = os.Remove(path)
		if err != nil {
			return err
		}
		// and create the actual file.  This ensures that the parent directories exist!
		// An archive may have a single file with a nested path, rather than a file for each parent dir
		writer, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return err
		}
		defer writer.Close()
		_, err = io.Copy(writer, reader)
		if err != nil {
			return err
		}
	}

	return nil
}
