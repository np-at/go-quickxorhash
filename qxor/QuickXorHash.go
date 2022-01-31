package qxor

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"hash"
	"io"
	"log"
)

const BitsInLastCell = 32
const Shift byte = 11
const Threshold = 600
const WidthInBits byte = 160

// Size size of hash in bytes
const Size = 20

// QuickXorHash Implements hash.Hash interface.
type QuickXorHash struct {
	_data        []uint64
	_lengthSoFar uint64
	_shiftSoFar  int
}

// ComputeQuickXorHash convenience function that when given an io.Reader, returns the calculated base64 encoded string representation of the QuickXor hash
func ComputeQuickXorHash(rdr io.Reader) (string, error) {
	q := New()
	_,err := io.Copy(q,rdr)
	if err != nil {
		return "", err
	}
	sum := q.Sum(nil)
	return base64.StdEncoding.EncodeToString(sum), nil
}

func (q *QuickXorHash) Write(p []byte) (n int, err error) {
	q.hashCore(p, 0, len(p))
	return len(p), nil
}

func (q *QuickXorHash) Sum(b []byte) []byte {
	hf := q.hashFinal()
	return append(b, hf[:]...)
}

func (q *QuickXorHash) Reset() {
	q._lengthSoFar = 0
	q._data = make([]uint64, uint64(WidthInBits-1)/64+1)
	q._shiftSoFar = 0
}

func (q *QuickXorHash) Size() int {
	return Size
}

func (q *QuickXorHash) BlockSize() int {
	return int(WidthInBits-1)/64 + 1
}

func New() hash.Hash {
	return &QuickXorHash{
		_data:        make([]uint64, uint64(WidthInBits-1)/64+1),
		_lengthSoFar: 0,
		_shiftSoFar:  0,
	}
}

func (q *QuickXorHash) hashFinal() []byte {
	// Create a byte array big enough to hold all our data
	rgb := make([]byte, (WidthInBits-1)/8+1)
	buf := new(bytes.Buffer)

	for i := 0; i < len(q._data)-1; i++ {
		err := binary.Write(buf, binary.LittleEndian, q._data[i])
		if err != nil {
			log.Println("error encountered while attempting to compute final hash")
			log.Fatalln(err)
		}
		_, err = blockCopy(buf.Bytes(), 0, rgb, i*8, 8)
		if err != nil {
			log.Fatalln(err)
		}
		buf.Reset()
	}
	err := binary.Write(buf, binary.LittleEndian, q._data[len(q._data)-1])
	if err != nil {
		log.Fatalln(err)
	}
	_, err = blockCopy(buf.Bytes(), 0, rgb, (len(q._data)-1)*8, len(rgb)-(len(q._data)-1)*8)
	if err != nil {
		log.Fatalln(err)
	}
	buf.Reset()
	// XOR the file length with the least significant bits
	// Note that GetBytes is architecture-dependent, so care should
	// be taken with porting. The expected value is 8-bytes in length in little-endian format
	err = binary.Write(buf, binary.LittleEndian, q._lengthSoFar)
	if err != nil {
		log.Fatalln(err)
	}

	lengthBytes := buf.Bytes()
	for i := 0; i < len(lengthBytes); i++ {
		rgb[(int(WidthInBits/8) - len(lengthBytes) + i)] ^= lengthBytes[i]
	}
	return rgb
}


func (q *QuickXorHash) hashCore(array []byte, ibStart int, cbSize int) {
	currentShift := q._shiftSoFar
	// The bitvector where we'll start xoring
	vectorArrayIndex := currentShift / 64

	// The position within the bit vector at which we begin xoring
	vectorOffset := currentShift % 64
	iterations := min(cbSize, int(WidthInBits))
	for i := 0; i < iterations; i++ {
		isLastCell := vectorArrayIndex == (len(q._data) - 1)
		var bitsInVectorCell int
		if isLastCell {
			bitsInVectorCell = BitsInLastCell
		} else {
			bitsInVectorCell = 64
		}
		// There's at least 2 bitvectors before we reach the end of the array
		if vectorOffset <= bitsInVectorCell-8 {
			for j := ibStart + i; j < cbSize+ibStart; j += int(WidthInBits) {
				q._data[vectorArrayIndex] ^= uint64(array[j]) << vectorOffset
			}
		} else {
			index1 := vectorArrayIndex
			var index2 int
			if isLastCell {
				index2 = 0
			} else {
				index2 = vectorArrayIndex + 1
			}
			low := (byte)(bitsInVectorCell - vectorOffset)

			xoredByte := byte(0)
			for j := ibStart + i; j < cbSize+ibStart; j += int(WidthInBits) {
				xoredByte ^= array[j]
			}
			q._data[index1] ^= (uint64)(xoredByte) << vectorOffset
			q._data[index2] ^= (uint64)(xoredByte) >> low
		}
		vectorOffset += int(Shift)
		for vectorOffset >= bitsInVectorCell {
			if isLastCell {
				vectorArrayIndex = 0
			} else {
				vectorArrayIndex += 1
			}
			vectorOffset -= bitsInVectorCell
		}
	}
	// Update the starting position in a circular shift pattern
	q._shiftSoFar = (q._shiftSoFar + int(Shift)*(cbSize%int(WidthInBits))) % int(WidthInBits)
	q._lengthSoFar += uint64(cbSize)
}