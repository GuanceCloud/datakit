// Package mytargz contains tar.gz file handle functions
package mytargz

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

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
