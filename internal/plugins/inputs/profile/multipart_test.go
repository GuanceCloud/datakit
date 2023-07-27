// Unless explicitly stated otherwise all files in this repository are licensed
// under the MIT License.
// This product includes software developed at Guance Cloud (https://www.guance.com/).
// Copyright 2021-present Guance, Inc.

package profile

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"testing"
	"time"

	"github.com/GuanceCloud/cliutils/testutil"
)

func TestPrefixMultipart(t *testing.T) {
	buf := &bytes.Buffer{}

	m, err := newMultipartPrepend(buf, "457e1c33abda2781595f0c5d78b750fa8105f9239b79b1c47744d8a008a9")
	testutil.Ok(t, err)
	testutil.Ok(t, m.Close())
	testutil.Equals(t, "", buf.String())

	file1, file2 := make([]byte, 1234), make([]byte, 2345)
	timeStart := time.Now().Format(time.RFC3339Nano)
	timeUnix := fmt.Sprintf("%d", time.Now().UnixNano())

	var boundary string
	var all, half []byte

	_, err = rand.Read(file1)
	testutil.Ok(t, err)

	_, err = rand.Read(file2)
	testutil.Ok(t, err)

	// Generate a full multipart/form-data body
	{
		buf.Reset()

		w := multipart.NewWriter(buf)

		boundary = w.Boundary()

		err = w.WriteField("time_start", timeStart)
		testutil.Ok(t, err)

		f, err := w.CreateFormFile("file_1", "file_first")
		testutil.Ok(t, err)

		_, err = f.Write(file1)
		testutil.Ok(t, err)

		err = w.WriteField("time_unix", timeUnix)
		testutil.Ok(t, err)

		f, err = w.CreateFormFile("file_2", "file_second")
		testutil.Ok(t, err)

		_, err = f.Write(file2)
		testutil.Ok(t, err)

		testutil.Ok(t, w.Close())

		all = append([]byte(nil), buf.Bytes()...)
	}

	// Generate right half of multipart/form-data body
	{
		buf.Reset()

		w2 := multipart.NewWriter(buf)
		err = w2.SetBoundary(boundary)
		testutil.Ok(t, err)

		err = w2.WriteField("time_unix", timeUnix)
		testutil.Ok(t, err)

		f, err := w2.CreateFormFile("file_2", "file_second")
		testutil.Ok(t, err)

		_, err = f.Write(file2)
		testutil.Ok(t, err)

		testutil.Ok(t, w2.Close())

		half = append([]byte(nil), buf.Bytes()...)
	}

	// Prepend a few parts to right half body
	{
		buf.Reset()

		mp, err := newMultipartPrepend(buf, boundary)
		testutil.Ok(t, err)

		err = mp.WriteField("time_start", timeStart)
		testutil.Ok(t, err)

		f, err := mp.CreateFormFile("file_1", "file_first")
		testutil.Ok(t, err)

		_, err = f.Write(file1)
		testutil.Ok(t, err)
		testutil.Ok(t, mp.Close())
	}

	_, err = buf.Write(half)
	testutil.Ok(t, err)

	fmt.Print(string(all))

	testutil.Equals(t, true, bytes.Equal(all, buf.Bytes()))
}

func TestGetBoundary(t *testing.T) {
	w := multipart.NewWriter(io.Discard)
	_ = w.Close()

	boundary, err := getBoundary(w.FormDataContentType())

	fmt.Println(w.FormDataContentType())
	fmt.Println(boundary)

	testutil.Ok(t, err)

	testutil.Equals(t, w.Boundary(), boundary)
}
