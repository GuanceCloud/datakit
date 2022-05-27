// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

// Package targzutil contains tar.gz file handle functions
package targzutil

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//------------------------------------------------------------------------------

func WriteTarFromMap(data map[string]string, dest string) error {
	tarFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer tarFile.Close() //nolint:errcheck,gosec
	gw := gzip.NewWriter(tarFile)
	defer gw.Close() //nolint:errcheck
	tw := tar.NewWriter(gw)
	defer tw.Close() //nolint:errcheck
	for name, content := range data {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o600,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			return err
		}
	}
	return nil
}

func ReadTarToMap(srcFile string) (map[string]string, error) {
	mRet := make(map[string]string)

	tarFile, err := os.Open(srcFile) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer tarFile.Close() //nolint:errcheck,gosec

	gz, err := gzip.NewReader(tarFile)
	if err != nil {
		return nil, err
	}
	defer gz.Close() //nolint:errcheck

	tr := tar.NewReader(gz)
	// untar each segment
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}

		// determine proper file path info
		finfo := hdr.FileInfo()
		if finfo.Mode().IsDir() {
			continue
		}

		var buf bytes.Buffer
		if _, err := io.Copy(&buf, tr); err != nil { //nolint:gosec
			return nil, err
		}

		mRet[hdr.Name] = buf.String()
	}

	return mRet, nil
}

//------------------------------------------------------------------------------

func UntartarFromMemory(releasedDir string, data []byte) ([]string, error) {
	absPath, err := filepath.Abs(releasedDir)
	if err != nil {
		return nil, err
	}
	tr := tar.NewReader(bytes.NewReader(data))
	if tr == nil {
		return nil, fmt.Errorf("reader nil")
	}

	var releasedAbsPath []string

	// untar each segment
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		// determine proper file path info
		finfo := hdr.FileInfo()
		fileName := hdr.Name
		absFileName := filepath.Join(absPath, fileName) //nolint:gosec
		// if a dir, create it, then go to next segment
		if finfo.Mode().IsDir() {
			if err := os.MkdirAll(absFileName, 0o750); err != nil {
				return nil, err
			}
			continue
		}
		// create new file with original file mode
		file, err := os.OpenFile(absFileName, //nolint:gosec
			os.O_RDWR|os.O_CREATE|os.O_TRUNC,
			finfo.Mode().Perm(),
		)
		if err != nil {
			return nil, err
		}
		releasedAbsPath = append(releasedAbsPath, absFileName)
		n, cpErr := io.Copy(file, tr) //nolint:gosec
		if closeErr := file.Close(); closeErr != nil {
			return nil, err
		}
		if cpErr != nil {
			return nil, cpErr
		}
		if n != finfo.Size() {
			return nil, fmt.Errorf("wrote %d, want %d", n, finfo.Size())
		}
	}
	return releasedAbsPath, nil
}

//------------------------------------------------------------------------------

func CreateTarGz(files []string, dest string) error {
	// Create output file
	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close() //nolint:errcheck,gosec
	// Create the archive and write the output to the "f" Writer
	return createArchive(files, f)
}

func createArchive(files []string, buf io.Writer) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close() //nolint:errcheck
	tw := tar.NewWriter(gw)
	defer tw.Close() //nolint:errcheck

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename) //nolint:gosec
	if err != nil {
		return err
	}
	defer file.Close() //nolint:errcheck,gosec

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

//------------------------------------------------------------------------------
