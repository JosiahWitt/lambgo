package zipper_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensuring"
	"github.com/JosiahWitt/lambgo/internal/zipper"
)

const sampleFile = "**************************************************************" +
	"**************************************************************" +
	"**************************************************************" +
	"**************************************************************" +
	"**************************************************************"

func TestZipFile(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("when successfully zipping file", func(ensure ensuring.E) {
		dir := ensure.T().TempDir()

		const fileName = "test-file.txt"
		const zippedFileName = "test-file-zipped.txt"
		path := filepath.Join(dir, fileName)
		outDir := filepath.Join(dir, "out")
		zipPath := path + ".zip"

		// Write file
		err := os.WriteFile(path, []byte(sampleFile), 0o655)
		ensure(err).IsNotError()

		// Zip file
		z := zipper.Zip{}
		err = z.ZipFile(path, zippedFileName)
		ensure(err).IsNotError()

		// Ensure zip file was compressed
		zipFileInfo, err := os.Stat(zipPath)
		ensure(err).IsNotError()
		ensure(zipFileInfo.Size() < int64(len([]byte(sampleFile)))-50).IsTrue()

		// Unzip file
		cmd := exec.Command("unzip", zipPath, "-d", outDir)
		err = cmd.Run()
		ensure(err).IsNotError()

		// Ensure unzipped file equals original
		outPath := filepath.Join(outDir, zippedFileName)
		data, err := os.ReadFile(outPath)
		ensure(err).IsNotError()
		ensure(string(data)).Equals(sampleFile)

		// Ensure unzipped file has correct modified time
		fileInfo, err := os.Stat(outPath)
		ensure(err).IsNotError()
		ensure(fileInfo.ModTime()).Equals(time.Date(2009, 11, 10, 0, 0, 0, 0, time.UTC))
	})

	ensure.Run("when zip file cannot be created", func(ensure ensuring.E) {
		dir := ensure.T().TempDir()

		// Try to create zip in an invalid location (some-dir/file-name.zip), since some-dir doesn't exist
		invalidPath := filepath.Join(dir, "some-dir", "file-name")

		z := zipper.Zip{}
		err := z.ZipFile(invalidPath, "output.zip")
		ensure(err).IsNotNil()
	})

	ensure.Run("when path does not exist", func(ensure ensuring.E) {
		dir := ensure.T().TempDir()

		invalidPath := filepath.Join(dir, "file-name")

		z := zipper.Zip{}
		err := z.ZipFile(invalidPath, "output.zip")
		ensure(err).IsNotNil()
	})

	// TODO: More tests
}
