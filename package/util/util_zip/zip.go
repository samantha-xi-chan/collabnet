package util_zip

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"github.com/pkg/errors"
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

func tarFile(tarWriter *tar.Writer, fileToTar string, baseDir string) error {
	file, err := os.Open(fileToTar)
	if err != nil {
		return errors.Wrap(err, "os.Open: ")
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return errors.Wrap(err, "file.Stat: ")
	}

	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return errors.Wrap(err, "file.FileInfoHeader: ")
	}

	relPath, err := filepath.Rel(baseDir, fileToTar)
	if err != nil {
		return errors.Wrap(err, "filepath.Rel: ")
	}
	header.Name = relPath

	err = tarWriter.WriteHeader(header)
	if err != nil {
		return errors.Wrap(err, "WriteHeader: ")
	}

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return errors.Wrap(err, "io.Copy: ")
	}
	return nil
}

func TarDirectory(tarWriter *tar.Writer, dirToTar string, baseDir string, ignoreLink bool) error {
	err := filepath.Walk(dirToTar, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.Mode()&os.ModeSymlink != 0 && ignoreLink {
			fmt.Println("WARN: In TarDirectory ", info.Name(), " is symbolicLink")
			return nil
		}

		return tarFile(tarWriter, path, baseDir)
	})

	return err
}

func TarFileOrDir(fileOrDirToTar string, tarFileName string) (e error) {
	tarFileX, err := os.Create(tarFileName)
	if err != nil {
		fmt.Println(err)
		return errors.Wrap(err, "os.Create: ")
	}
	defer tarFileX.Close()

	tarWriter := tar.NewWriter(tarFileX)
	defer tarWriter.Close()

	baseDir, err := filepath.Abs(filepath.Dir(fileOrDirToTar))
	if err != nil {
		fmt.Println(err)
		return errors.Wrap(err, "filepath.Abs: ")
	}

	fileInfo, err := os.Stat(fileOrDirToTar)
	if err != nil {
		fmt.Println(err)
		return errors.Wrap(err, "os.Stat: ")
	}

	if fileInfo.IsDir() {
		err = TarDirectory(tarWriter, fileOrDirToTar, baseDir, true)
	} else {
		err = tarFile(tarWriter, fileOrDirToTar, baseDir)
	}

	if err != nil {
		fmt.Println(err)
		return errors.Wrap(err, "TarDirectory ,tarFile:   ")
	}

	log.Printf("arc ok：%s\n", tarFileName)
	return nil
}

func UnTar(tarFile string, destFolder string) error {
	file, err := os.Open(tarFile)
	if err != nil {
		return errors.Wrap(err, "os.Open:   ")
	}
	defer file.Close()

	tarReader := tar.NewReader(file)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			return errors.Wrap(err, "tarReader.Next():   ")
		}

		filePath := filepath.Join(destFolder, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(filePath, os.ModePerm)
		case tar.TypeReg:
			os.MkdirAll(filepath.Dir(filePath), os.ModePerm)

			destFile, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer destFile.Close()

			_, err = io.Copy(destFile, tarReader)
			if err != nil {
				return errors.Wrap(err, "io.Copy:   ")
			}
		}
	}

	return nil
}
