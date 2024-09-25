# zipstream
Package zipstream is a stream on the fly extractor/reader for zip archive like Java's `java.util.zip.ZipInputStream`, there is no need to provide `io.ReaderAt` and total archive size parameters, that is, just need only one `io.Reader` parameter.

## Implementation
Most code of this package is copied directly from golang standard library [archive/zip](https://pkg.go.dev/archive/zip), .ZIP archive format specification reference
is [here](https://pkware.cachefly.net/webdocs/casestudies/APPNOTE.TXT)

## Usage
> go get github.com/GuanceCloud/zipstream

## Examples

```go
package main

import (
	"io"
	"log"
	"net/http"

	"github.com/GuanceCloud/zipstream"
)

func main() {

	resp, err := http.Get("https://github.com/golang/go/archive/refs/tags/go1.16.10.zip")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	zr := zipstream.NewReader(resp.Body)

	for {
		e, err := zr.GetNextEntry()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("unable to get next entry: %s", err)
		}

		log.Println("entry name: ", e.Name)
		log.Println("entry comment: ", e.Comment)
		log.Println("entry reader version: ", e.ReaderVersion)
		log.Println("entry modify time: ", e.Modified)
		log.Println("entry compressed size: ", e.CompressedSize64)
		log.Println("entry uncompressed size: ", e.UncompressedSize64)
		log.Println("entry is a dir: ", e.IsDir())

		if !e.IsDir() {
			rc, err := e.Open()
			if err != nil {
				log.Fatalf("unable to open zip file: %s", err)
			}
			content, err := io.ReadAll(rc)
			if err != nil {
				log.Fatalf("read zip file content fail: %s", err)
			}

			log.Println("file length:", len(content))

			if uint64(len(content)) != e.UncompressedSize64 {
				log.Fatalf("read zip file length not equal with UncompressedSize64")
			}
			if err := rc.Close(); err != nil {
				log.Fatalf("close zip entry reader fail: %s", err)
			}
		}
	}
}
```

## Limitation

- Every file in zip archive can read only once for a new Reader, Repeated read is unsupported.
- Some `central directory header` field is not resolved, such as `version made by`, `internal file attributes`, `external file attributes`, `relative offset of local header`, some `central directory header` field may differ from `local file header`, such as `extra field`. 
- Unable to read multi files concurrently.