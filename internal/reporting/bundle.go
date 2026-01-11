package reporting

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// BuildReportBundle creates a deterministic zip bundle from provided entries.
// entries map keys are bundle paths, values are source file paths.
func BuildReportBundle(bundlePath string, entries map[string]string) error {
	if bundlePath == "" {
		return errors.New("bundle path is required")
	}
	if len(entries) == 0 {
		return errors.New("bundle entries are required")
	}
	if err := os.MkdirAll(filepath.Dir(bundlePath), 0o755); err != nil {
		return err
	}
	file, err := os.Create(bundlePath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := zip.NewWriter(file)
	defer writer.Close()

	keys := make([]string, 0, len(entries))
	for key := range entries {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	zeroTime := time.Unix(0, 0).UTC()
	for _, name := range keys {
		source := entries[name]
		info, err := os.Stat(source)
		if err != nil || info.IsDir() {
			continue
		}
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(name)
		header.Method = zip.Deflate
		header.SetModTime(zeroTime)

		entryWriter, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if err := copyFile(entryWriter, source); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(writer io.Writer, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(writer, file)
	return err
}
