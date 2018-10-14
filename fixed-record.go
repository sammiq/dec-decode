package main

import (
	"errors"
	"unicode/utf8"
)

//FixedRecord provides a fixed storage Buffer that implements io.Writer
type FixedRecord interface {
	Record() []byte
	Size() int
	Bytes() []byte
	Len() int
	Write(p []byte) (n int, err error)
	WriteString(s string) (n int, err error)
	WriteByte(c byte) error
	WriteRune(r rune) (n int, err error)
}

type internalFixedRecord struct {
	buf []byte
	off int
}

func (b *internalFixedRecord) Record() []byte { return b.buf }

func (b *internalFixedRecord) Size() int { return len(b.buf) }

func (b *internalFixedRecord) Bytes() []byte { return b.buf[0:b.off] }

func (b *internalFixedRecord) Len() int { return b.off }

func (b *internalFixedRecord) Write(p []byte) (n int, err error) {
	if b.off+len(p) > b.Size() {
		return 0, errors.New("too much data to write")
	}
	n = copy(b.buf[b.off:], p)
	b.off += n
	return n, nil
}

func (b *internalFixedRecord) WriteString(s string) (n int, err error) {
	if b.off+len(s) > b.Size() {
		return 0, errors.New("too much data to write")
	}
	n = copy(b.buf[b.off:], s)
	b.off += n
	return n, nil
}

func (b *internalFixedRecord) WriteByte(c byte) error {
	if b.off+1 > b.Size() {
		return errors.New("too much data to write")
	}
	b.buf[b.off] = c
	b.off++
	return nil
}

func (b *internalFixedRecord) WriteRune(r rune) (n int, err error) {
	if b.off+utf8.RuneLen(r) > b.Size() {
		return 0, errors.New("too much data to write")
	}
	n = utf8.EncodeRune(b.buf[b.off:b.off+utf8.UTFMax], r)
	b.off += n
	return n, nil
}

//NewFixedRecord creates a write buffer with a given size
func NewFixedRecord(size int) FixedRecord {
	return &internalFixedRecord{buf: make([]byte, size)}
}
