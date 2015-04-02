package writeton_test

import (
	. "github.com/funny-falcon/go-writeton"
	"io"
	"io/ioutil"
	"testing"
)

type reader struct {
	capa   int
	n      int
	called bool
}

func (rd *reader) Read(b []byte) (n int, err error) {
	n = len(b)
	if n == 0 {
		return
	}
	if n > rd.capa-rd.n {
		n = rd.capa - rd.n
	}
	rd.n += n
	if n == 0 {
		err = io.EOF
	}
	return
}

func (rd *reader) WriteTo(w io.Writer) (n int64, err error) {
	b := [4096]byte{}
	for rd.n < rd.capa && err == nil {
		m := rd.capa - rd.n
		if m > len(b) {
			m = len(b)
		}
		m, err = w.Write(b[:m])
		n += int64(m)
		rd.n += m
	}
	return
}

func (rd *reader) WriteToN(w io.Writer, sz int64) (n int64, err error) {
	rd.called = true
	b := [4096]byte{}
	if sz > int64(rd.capa-rd.n) {
		sz = int64(rd.capa - rd.n)
	}
	lim := rd.n + int(sz)
	for rd.n < lim && err == nil {
		m := lim - rd.n
		if m > len(b) {
			m = len(b)
		}
		m, err = w.Write(b[:m])
		n += int64(m)
		rd.n += m
	}
	return
}

func TestWithWriterToN(t *testing.T) {
	var r reader
	w := &Writer{ioutil.Discard}
	r = reader{capa: 10}
	n, err := io.CopyN(w, &r, 1)
	if n != 1 || err != nil || !r.called {
		t.Error("CopyN:", n, err, r)
	}
	r = reader{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 10), 1)
	if n != 1 || err != nil || !r.called {
		t.Error("CopyN limitReader 10 1:", n, err, r)
	}
	r = reader{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 5), 5)
	if n != 5 || err != nil || !r.called {
		t.Error("CopyN limitReader 5 5:", n, err, r)
	}
	r = reader{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 5), 10)
	if n != 5 || err != io.EOF || !r.called {
		t.Error("CopyN limitReader 5 10:", n, err, r)
	}
	r = reader{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 10), 20)
	if n != 10 || err != io.EOF || !r.called {
		t.Error("CopyN limitReader 10 20:", n, err, r)
	}
	r = reader{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 20), 20)
	if n != 10 || err != io.EOF || !r.called {
		t.Error("CopyN limitReader 20 20:", n, err, r)
	}
}

type simple struct {
	capa int
	n    int
}

func (rd *simple) Read(b []byte) (n int, err error) {
	n = len(b)
	if n == 0 {
		return
	}
	if n > rd.capa-rd.n {
		n = rd.capa - rd.n
	}
	rd.n += n
	if n == 0 {
		err = io.EOF
	}
	return
}

func TestWithSimpleReader(t *testing.T) {
	var r simple
	w := &Writer{ioutil.Discard}
	r = simple{capa: 10}
	n, err := io.CopyN(w, &r, 1)
	if n != 1 || err != nil {
		t.Error("CopyN:", n, err, r)
	}
	r = simple{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 10), 1)
	if n != 1 || err != nil {
		t.Error("CopyN limitReader 10 1:", n, err, r)
	}
	r = simple{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 5), 5)
	if n != 5 || err != nil {
		t.Error("CopyN limitReader 5 5:", n, err, r)
	}
	r = simple{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 5), 10)
	if n != 5 || err != io.EOF {
		t.Error("CopyN limitReader 5 10:", n, err, r)
	}
	r = simple{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 10), 20)
	if n != 10 || err != io.EOF {
		t.Error("CopyN limitReader 10 20:", n, err, r)
	}
	r = simple{capa: 10}
	n, err = io.CopyN(w, io.LimitReader(&r, 20), 20)
	if n != 10 || err != io.EOF {
		t.Error("CopyN limitReader 20 20:", n, err, r)
	}
}
