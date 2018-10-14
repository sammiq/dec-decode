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

type partitionInfo struct {
	PartitionType      uint
	PartitionOffset    uint64
	PartitionEndOffset uint64
	DataOffset         uint64
	DataSize           uint64
	PartitionKey       []byte
}

func decodeWii(r *os.File, outPath string, sectorSize uint) {
	startSector := make([]byte, sectorSize)
	_, err := io.ReadFull(r, startSector)
	errorExit(err)

	startBuffer := bytes.NewBuffer(startSector)

	sig := startBuffer.Next(4)
	id := startBuffer.Next(4)
	hashValue := startBuffer.Next(16)
	numPartitions := binary.LittleEndian.Uint32(startBuffer.Next(4))
	var partitions = make([]partitionInfo, numPartitions)

	fmt.Println("Wii Disc")
	fmt.Printf("Signature: %s\n", string(sig))
	fmt.Printf("ID: %s\n", string(id))
	fmt.Printf("MD5: %x\n", hashValue)

	for i := uint32(0); i < numPartitions; i++ {
		partitions[i].DataOffset = uint64(binary.LittleEndian.Uint32(startBuffer.Next(4))) << 2
		partitions[i].DataSize = uint64(binary.LittleEndian.Uint32(startBuffer.Next(4))) << 2
		partitions[i].PartitionOffset = uint64(binary.LittleEndian.Uint32(startBuffer.Next(4))) << 2
		partitions[i].PartitionEndOffset = uint64(binary.LittleEndian.Uint32(startBuffer.Next(4))) << 2
		partitions[i].PartitionKey = startBuffer.Next(16)

		fmt.Printf("Partition %d of %d\n", i+1, numPartitions)
		fmt.Printf("--------------------\n")
		fmt.Printf("Data Offset:      0x%x\n", partitions[i].DataOffset)
		fmt.Printf("Data Size:        0x%x\n", partitions[i].DataSize)
		fmt.Printf("Partition Offset: 0x%x\n", partitions[i].PartitionOffset)
		fmt.Printf("Partition End:    0x%x\n", partitions[i].PartitionEndOffset)
		fmt.Printf("Partition Key:    0x%x\n", partitions[i].PartitionKey)
		fmt.Printf("====================\n")
	}

	w, err := os.Create(outPath)
	errorExit(err)
	defer w.Close()

	bytesWritten := uint64(0)

	transfer := make([]byte, 1024)

	hash := md5.New()
	fmt.Print("Writing Disc Header...")
	for i := uint64(0); i < partitions[0].PartitionOffset; i += 1024 {
		readNextOffset(r, startBuffer) // we don't need this
		bytesWritten += uint64(blockTransferWithHash(r, w, transfer, hash))
		vLog("\nWRITE: offset: %x\n", bytesWritten)
	}
	fmt.Println("Done")

	for j := uint32(0); j < numPartitions; j++ {
		fmt.Printf("Writing Partition %d Header...", j)
		for i := uint64(0); i < partitions[j].DataOffset; i += 1024 {
			setNextOffset(r, startBuffer)
			bytesWritten += uint64(blockTransferWithHash(r, w, transfer, hash))
			vLog("\nWRITE: offset: %x\n", bytesWritten)
		}
		fmt.Println("Done")

		padBlock := uint32(0)
		padOffset := uint64(0)
		dataSize := uint64(0)
		padding := make([]byte, 0x40000)

		fmt.Printf("Writing Partition %d Data.....", j)
		for dataSize < partitions[j].DataSize {
			setNextOffset(r, startBuffer)

			wrote := uint64(blockTransferWithHash(r, w, transfer, hash))
			bytesWritten += wrote
			dataSize += wrote
			vLog("\nWRITE: offset: %x\n", bytesWritten)

			writeBuffer := NewFixedRecord(0x7C00)
			transfer2 := make([]byte, 1024)

			for k := 0; k < 31; k++ {
				if (padOffset & 0x3FFFF) == 0 {
					vLog("\nPADDING: block: %d id: %s\n", padBlock, id)
					padding = generatePaddingBlock(padBlock, id, 0)
					padBlock++
					padOffset = 0
				}

				rawOffset := binary.LittleEndian.Uint32(startBuffer.Next(4))
				if rawOffset != 0xffffffff {
					offset := rawOffset << 8
					checkOffset(r, offset)
					r.Seek(int64(offset), io.SeekStart)

					padOffset += uint64(blockTransfer(r, writeBuffer, transfer2))
					vLog("\ntfr %d poffset: %x buffer: %x\n", k, padOffset, writeBuffer.Len())
				} else {
					slice := padding[padOffset : padOffset+1024]
					_, err = writeBuffer.Write(slice)
					errorExit(err)
					padOffset += 1024
					vLog("\npad %d poffset: %x buffer: %x\n", k, padOffset, writeBuffer.Len())
				}
			}
			iv := getIV(transfer)
			output := writeBuffer.Record()
			encodeAES(output, partitions[j].PartitionKey, iv)
			io.Copy(hash, bytes.NewBuffer(output))
			_, err = w.Write(output)
			errorExit(err)
			bytesWritten += 0x7C00
			vLog("\nWRITE: offset: %x\n", bytesWritten)

			dataSize += 0x7C00
			vLog("\nDATA SIZE: %d / %d\n", dataSize, partitions[j].DataSize)
		}
		fmt.Println("Done")

		fmt.Printf("Writing Partition %d Fill.....", j)
		padOffset = 0
		padBlock = uint32(bytesWritten / 0x40000)
		for bytesWritten != partitions[j].PartitionEndOffset {
			if (padOffset & 0x3FFFF) == 0 {
				vLog("\ngenerate padding: %d %s\n", padBlock, id)
				padding = generatePaddingBlock(padBlock, id, 0)
				padBlock++
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
				vLog("\ntfr WRITE: offset: %x\n", bytesWritten)
			} else {
				slice := padding[padOffset : padOffset+1024]
				io.Copy(hash, bytes.NewBuffer(slice))
				_, err = w.Write(slice)
				errorExit(err)
				bytesWritten += 1024
				padOffset += 1024
				vLog("\npad WRITE: offset: %x\n", bytesWritten)
			}
		}
		fmt.Println("Done")
	}
	calcValue := hash.Sum(nil)
	if reflect.DeepEqual(hashValue, calcValue) {
		fmt.Printf("Decode OK: %x\n", hashValue)
	} else {
		fmt.Printf("Decode FAIL: expected: %x calculated: %x\n", hashValue, calcValue)
	}
}
