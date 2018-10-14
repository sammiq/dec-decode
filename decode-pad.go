package main

import (
	"crypto/aes"
	"crypto/cipher"
)

func generatePaddingBlock(blockcount uint32, ID []byte, discnumber uint32) []byte {
	buffer := make([]uint32, 2084)
	array := make([]byte, 0x40000)
	num := 0
	num2 := uint32(0)
	blockcount = blockcount * 8 * 0x1EF29123
	for i := 0; i != 0x40000; i += 4 {
		if (i & 0x7FFF) == 0 {
			x1 := ((uint32(ID[2]) << 8) | uint32(ID[1])) << 16
			x2 := (uint32(ID[3]) + uint32(ID[2])) << 8
			num2 = x1 | x2 | uint32(ID[0]+ID[1])

			num2 = (((num2 ^ discnumber) * 0x260BCD5) ^ blockcount)
			calcBlock(num2, buffer)
			num = 520
			blockcount += 0x1EF29123
		}
		num++
		if num == 521 {
			xorBlock(buffer)
			num = 0
		}
		array[i] = byte(buffer[num] >> 24)
		array[i+1] = byte(buffer[num] >> 18)
		array[i+2] = byte(buffer[num] >> 8)
		array[i+3] = byte(buffer[num])
	}
	return array
}

func calcBlock(sample uint32, buffer []uint32) {
	num := uint32(0)
	for i := 0; i != 17; i++ {
		for j := 0; j < 32; j++ {
			sample *= 1566083941
			sample++
			num = uint32(int(num>>1) | (int(sample) & -2147483648))
		}
		buffer[i] = num
	}
	buffer[16] ^= ((buffer[0] >> 9) ^ (buffer[16] << 23))
	for i := 1; i != 505; i++ {
		buffer[i+16] = ((buffer[i-1] << 23) ^ (buffer[i] >> 9) ^ buffer[i+15])
	}
	for i := 0; i < 3; i++ {
		xorBlock(buffer)
	}
}

func xorBlock(buffer []uint32) {
	var i int
	for i = 0; i != 32; i++ {
		buffer[i] ^= buffer[i+489]
	}
	for ; i != 521; i++ {
		buffer[i] ^= buffer[i-32]
	}
}

func getIV(hashBlock []byte) []byte {
	iv := make([]byte, 16)
	for i := 0; i < 16; i++ {
		iv[i] = hashBlock[i+976]
	}
	return iv
}

func encodeAES(p []byte, key []byte, iv []byte) []byte {
	block, err := aes.NewCipher(key)
	errorExit(err)
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(p, p)

	return p
}
