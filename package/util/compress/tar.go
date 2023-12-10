package compress

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func TarFiles(srcPath string, destPath string) error {
	tarFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	gzipWriter := gzip.NewWriter(tarFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(srcPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name, _ = filepath.Rel(srcPath, filePath)

		if info.Mode()&os.ModeSymlink == os.ModeSymlink {
			link, err := os.Readlink(filePath)
			if err != nil {
				return err
			}
			header.Linkname = link
			header.Typeflag = tar.TypeSymlink
		} else if info.Mode()&os.ModeType == os.ModeDir {
			header.Typeflag = tar.TypeDir
		} else if info.Mode()&os.ModeType == 0 {
			header.Typeflag = tar.TypeReg
		} else if info.Mode()&os.ModeType == os.ModeNamedPipe {
			header.Typeflag = tar.TypeChar
		} else if info.Mode()&os.ModeType == os.ModeSocket {
			header.Typeflag = tar.TypeBlock
		} else if info.Mode()&os.ModeType == os.ModeDevice {
			header.Typeflag = tar.TypeBlock
		} else {
			return fmt.Errorf("unsupported file type: %s", filePath)
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			file, err := os.Open(filePath)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tarWriter, file); err != nil {
				return err
			}
		}

		return nil
	})
}

func UntarFiles(tarPath string, destPath string) error {
	tarFile, err := os.Open(tarPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	gzipReader, err := gzip.NewReader(tarFile)
	if err != nil {
		return err
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)

	for {
		header, err := tarReader.Next()

		if err == io.EOF {
			break // end of archive
		}

		if err != nil {
			return err
		}

		targetPath := filepath.Join(destPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			file, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return err
			}

			if err := os.Chmod(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported tar file type: %c", header.Typeflag)
		}
	}

	return nil
}

func test() {
	sourcePath := "/path/to/source"
	tarFilePath := "/path/to/output/archive.tar.gz"
	destinationPath := "/path/to/extracted"

	// Create a tar archive
	if err := TarFiles(sourcePath, tarFilePath); err != nil {
		fmt.Println("Error creating tar archive:", err)
		return
	}

	fmt.Println("Tar archive created successfully:", tarFilePath)

	// Extract files from the tar archive
	if err := UntarFiles(tarFilePath, destinationPath); err != nil {
		fmt.Println("Error extracting files from tar archive:", err)
		return
	}

	fmt.Println("Files extracted successfully to:", destinationPath)
}
