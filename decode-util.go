package main

import (
	"bytes"
	"encoding/binary"
	"hash"
	"io"
	"log"
)

func errorExit(err error) {
	if err != nil {
		log.Fatal("ERROR:", err)
	}
}

func vLog(f string, v ...interface{}) {
	if opts.Verbose {
		log.Printf(f, v...)
	}
}

func position(s io.Seeker) (pos int64, err error) {
	return s.Seek(0, io.SeekCurrent)
}

func checkPosition(s io.Seeker) int64 {
	pos, err := position(s)
	errorExit(err)
	return pos
}

func checkOffset(s io.Seeker, offset uint32) {
	pos := checkPosition(s)
	if pos != int64(offset) {
		vLog("READ: position: %x expected: %x", pos, offset)
	}
}

type byteProvider interface {
	Next(n int) []byte
}

func readNextOffset(s io.Seeker, p byteProvider) uint32 {
	offset := binary.LittleEndian.Uint32(p.Next(4)) << 8
	checkOffset(s, offset)
	return offset
}

func setNextOffset(s io.Seeker, p byteProvider) {
	offset := readNextOffset(s, p)
	s.Seek(int64(offset), io.SeekStart)
}

func blockTransfer(r io.Reader, w io.Writer, buffer []byte) int {
	_, err := io.ReadFull(r, buffer)
	errorExit(err)
	wrote, err := w.Write(buffer)
	errorExit(err)
	return wrote
}

func blockTransferWithHash(r io.Reader, w io.Writer, buffer []byte, hash hash.Hash) int {
	wrote := blockTransfer(r, w, buffer)
	io.Copy(hash, bytes.NewBuffer(buffer))
	return wrote
}
