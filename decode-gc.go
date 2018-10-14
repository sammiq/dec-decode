package main

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"reflect"
)

func decodeGameCube(r *os.File, outPath string) {
	startSector := make([]byte, 0x2B8800)
	_, err := io.ReadFull(r, startSector)
	errorExit(err)

	startBuffer := bytes.NewBuffer(startSector)

	sig := startBuffer.Next(4)
	id := startBuffer.Next(4)
	hashValue := startBuffer.Next(16)
	startBuffer.Next(6)
	discNumber, err := startBuffer.ReadByte()
	errorExit(err)

	fmt.Println("GameCube Disc")
	fmt.Printf("Signature: %s\n", string(sig))
	fmt.Printf("ID: %s\n", string(id))
	fmt.Printf("MD5: %x\n", hashValue)

	w, err := os.Create(outPath)
	errorExit(err)
	defer w.Close()

	bytesWritten := uint64(0)

	padBlock := uint32(0)
	padOffset := uint64(0)
	padding := generatePaddingBlock(padBlock, id, uint32(discNumber))

	transfer := make([]byte, 2048)

	hash := md5.New()
	fmt.Printf("Writing Disc Data.....")
	for i := 0; i < 712880; i++ {
		if padOffset == 0x40000 {
			padBlock++
			padding = generatePaddingBlock(padBlock, id, uint32(discNumber))
			padOffset = 0
		}

		rawOffset := binary.LittleEndian.Uint32(startBuffer.Next(4))
		if rawOffset != 0xffffffff {
			offset := rawOffset << 8
			checkOffset(r, offset)
			r.Seek(int64(offset), io.SeekStart)
			wrote := uint64(blockTransferWithHash(r, w, transfer, hash))
			bytesWritten += wrote
			padOffset += wrote
		} else {
			slice := padding[padOffset : padOffset+2048]
			io.Copy(hash, bytes.NewBuffer(slice))
			_, err = w.Write(slice)
			errorExit(err)
			bytesWritten += 2048
			padOffset += 2048
		}
	}
	fmt.Println("Done")

	calcValue := hash.Sum(nil)
	if reflect.DeepEqual(hashValue, calcValue) {
		fmt.Printf("Decode OK: %x\n", hashValue)
	} else {
		fmt.Printf("Decode FAIL: expected: %x calculated: %x\n", hashValue, calcValue)
	}
}
