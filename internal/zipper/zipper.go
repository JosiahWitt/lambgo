package zipper

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"time"
)

type ZipAPI interface {
	ZipFile(path, zippedFileName string) error
}

type Zip struct{}

var _ ZipAPI = &Zip{}

// ZipFile located at path to <path>.zip.
//
// zippedFileName is the name of the file once it is zipped.
func (z *Zip) ZipFile(path, zippedFileName string) error {
	zipPath := fmt.Sprintf("%s.zip", path)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	binaryFile, err := os.Open(path)
	if err != nil {
		return err
	}

	fileInfo, err := binaryFile.Stat()
	if err != nil {
		return err
	}

	fileHeader, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return err
	}

	// Hardcode the date to keep builds reproducible
	fileHeader.Modified = time.Date(2009, 11, 10, 0, 0, 0, 0, time.UTC)
	fileHeader.Method = zip.Deflate
	fileHeader.Name = zippedFileName

	fileHolder, err := zipWriter.CreateHeader(fileHeader)
	if err != nil {
		return err
	}

	_, err = io.Copy(fileHolder, binaryFile)
	if err != nil {
		return err
	}

	return nil
}
