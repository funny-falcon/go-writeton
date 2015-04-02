// Copyright (c) 2015, Sokolov Yura aka funny_falcon
// This code is in public domain.
// You may either to retain copyright notice or not.
// But I will be a bit happier if you do.
//
// Additional statement about "software provided 'as is'" is in LICENSE file.

// writeton package is workaround io.CopyN and io.LimitedReader limitation
//
// Main io.LimitedReader limitation is that it is not extendable,
// ie there is no way to suggest io.LimitedReader more efficient way
// to copy bytes, so it is always go through Read/Write cycle.
// Only net.TCPConn tries to detect, that io.LimitedReader is wrapper around
// *os.File, and then uses sendfile.
// But there is no way to use sendfile with more complex objects.
//
// Natural way to solve thing should be extending io.LimitedReader, but
// proposal is not accepted
// https://go-review.googlesource.com/#/c/7893/
// https://groups.google.com/forum/#!topic/golang-dev/D7tO5HZTs60
// http://pastebin.com/zCVrzVx6
//
// To work around this limitation, wrap io.Writer with Writer and
// then use io.Copy or io.CopyN.
// If reader is a io.LimitedReader, then wrapped reader will be tested
// - for WriteToN method, and then WriteToN will be called
// - for being other io.LimitedReader, then it recursively dives in.
//
//     w := some_w.(io.Writer)
//     wr := writeton.Writer{w}
//     io.Copy(wr, some_r.(io.Reader))
//     io.CopyN(wr, some_r.(io.Reader), n)
//
//     w := some_w.(io.Writer)
//     writeton.Copy(w, some_r.(io.Reader))
//     writeton.CopyN(w, some_r.(io.Reader), n)
//
//
// NewResponseWriter could be used to wrap http.ResponseWriter with
// Writer
//
//     func handler(rw http.ResponseWriter, rq *http.Request) {
//         var r io.ReadSeeker
//         var m time.Time
//         r, m = SomeComplexFileObject(rq)
//         smartrw := writeton.NewResponseWriter(rw)
//         http.ServeContent(smartrw, req, "", m, r)
//     }
//
// The code remains workable even if proposal accepted.
package writeton

import (
	"io"
	"net/http"
)

// WriterToN is the interface that wraps the WriteToN method.
//
// WriteToN writes at most sz bytes to w or until there's no more data
// to write or when an error occurs. The return value n is the number
// of bytes written. Any error encountered during the write is also returned.
//
// The Writer.ReadFrom function uses WriterToN if available for readers
// wrapped with io.LimitedReader.
type WriterToN interface {
	// WriteToN should copy at most sz bytes from
	WriteToN(w io.Writer, sz int64) (n int64, err error)
}

// Writer is a smart wrapper around io.Writer .
//
// It adds ReadFrom method to detect io.LimitedReader.
type Writer struct {
	W io.Writer
}

var _ io.Writer = &Writer{nil}
var _ io.ReaderFrom = &Writer{nil}

// Write method provides io.Writer
func (w *Writer) Write(b []byte) (n int, err error) {
	return w.W.Write(b)
}

// ReadFrom will detect that reader is a io.LimitedReader
// and will try to call WriteToN method in this case.
// Otherwise it falls back to io.Copy
func (w *Writer) ReadFrom(r io.Reader) (n int64, err error) {
	if lim, ok := r.(*io.LimitedReader); ok {
		switch rd := lim.R.(type) {
		case WriterToN:
			n, err = rd.WriteToN(w, lim.N)
			lim.N -= n
		case *io.LimitedReader:
			if rd.N < lim.N {
				n, err = w.ReadFrom(rd)
			} else {
				n, err = io.CopyN(w, rd.R, lim.N)
				rd.N -= n
			}
			lim.N -= n
		default:
			n, err = io.Copy(w.W, lim)
		}
		return
	} else {
		return io.Copy(w.W, r)
	}
}

// Copy is a simple wrapper around Writer and io.Copy
func Copy(w io.Writer, r io.Reader) (n int64, err error) {
	return io.Copy(&Writer{w}, r)
}

// CopyN is a simple wrapper around Writer and io.CopyN
func CopyN(w io.Writer, r io.Reader, sz int64) (n int64, err error) {
	return io.CopyN(&Writer{w}, r, sz)
}

type responseWriter struct {
	http.ResponseWriter
	w Writer
}

// NewResponseWriter wraps http.ResponseWriter
//
// It returns new http.ResponseWriter which will use Writer to
// workaround io.LimitedReader limitations
func NewResponseWriter(rw http.ResponseWriter) http.ResponseWriter {
	return &responseWriter{
		ResponseWriter: rw,
		w:              Writer{rw},
	}
}

func (rw *responseWriter) ReadFrom(r io.Reader) (n int64, err error) {
	return rw.w.ReadFrom(r)
}
