package qxor_test

import (
	"github.com/np-at/go-quickxorhash/qxor"
	"log"
	"os"
	"testing"
)

const (
	testFilePath              = "test_resources/testBlob"
	testFileQuickXorHashValue = "OlgzOZR3SdAfb/Y/0p0IFcHuZrs="
)

func TestComputeQuickXorHash(t *testing.T) {
	file, err := os.Open(testFilePath)
	if err != nil {
		log.Fatalln(err)
	}
	hashString, err := qxor.ComputeQuickXorHash(file)
	if err != nil {
		file.Close()
		log.Fatalln(err)
	}
	if hashString != testFileQuickXorHashValue {
		t.Errorf("computed hash did not match known value")
	}
	file.Close()
}
func BenchmarkComputeQuickXorHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		file, err := os.Open(testFilePath)
		if err != nil {
			log.Fatalln(err)
		}
		_, err = qxor.ComputeQuickXorHash(file)
		if err != nil {
			file.Close()
			log.Fatalln(err)
		}
		file.Close()
	}
}
