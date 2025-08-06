package utils

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
)

func CreateTarArchive(srcDir string) (io.ReadCloser, error) {
	pr, pw := io.Pipe()
	tw := tar.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer tw.Close()

		err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Create tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}

			// Set the header name to the relative path
			relPath, err := filepath.Rel(srcDir, path)
			if err != nil {
				return err
			}
			header.Name = relPath

			// Write header
			if err := tw.WriteHeader(header); err != nil {
				return err
			}

			// Write file content if it's a regular file
			if info.Mode().IsRegular() {
				file, err := os.Open(path)
				if err != nil {
					return err
				}
				defer file.Close()
				_, err = io.Copy(tw, file)
				return err
			}

			return nil
		})

		if err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr, nil
}
