package qxor_test

import (
	"encoding/base64"
	"github.com/np-at/go-quickxorhash/qxor"
	"github.com/rclone/rclone/backend/onedrive/quickxorhash"
	"io"
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
func BenchmarkOther(b *testing.B) {
	for i := 0; i < b.N; i++ {
		file, err := os.Open(testFilePath)
		if err != nil {
			log.Fatalln(err)
		}
		q := quickxorhash.New()
		_, err = io.Copy(q, file)
		if err != nil {
			log.Fatalln(err)
		}
		sum := q.Sum(nil)
		hashString := base64.StdEncoding.EncodeToString(sum)

		if hashString != testFileQuickXorHashValue {
			log.Println(hashString)
			b.Error("sums don't line up")
		}
		file.Close()
	}
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
