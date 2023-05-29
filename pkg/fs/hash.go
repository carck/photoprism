package fs

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

const (
	hashSize = 16 * 1024
)

// Hash returns the SHA1 hash of a file as string.
func Hash(fileName string) string {
	if bytes, err := readHashBytes(fileName); err != nil {
		return ""
	} else {
		hash := sha1.New()
		if _, hErr := hash.Write(bytes); hErr != nil {
			return ""
		}
		return hex.EncodeToString(hash.Sum(nil))
	}
}

func readHashBytes(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	firstBytes := make([]byte, hashSize/2)
	if _, e := file.ReadAt(firstBytes, 0); e != nil {
		return nil, fmt.Errorf("couldn't read first few bytes: %+v", e)
	}
	middleBytes := make([]byte, hashSize/4)
	fileInfo, _ := file.Stat()
	if _, e := file.ReadAt(middleBytes, fileInfo.Size()/2); e != nil {
		return nil, fmt.Errorf("couldn't read middle bytes: %+v", e)
	}
	lastBytes := make([]byte, hashSize/4)
	if _, e := file.ReadAt(lastBytes, fileInfo.Size()-hashSize/4); e != nil {
		return nil, fmt.Errorf("couldn't read end bytes: %+v", e)
	}
	bytes := append(append(firstBytes, middleBytes...), lastBytes...)
	return bytes, nil
}

// Checksum returns the CRC32 checksum of a file as string.
func Checksum(fileName string) string {
	var result []byte

	file, err := os.Open(fileName)

	if err != nil {
		return ""
	}

	defer file.Close()

	hash := crc32.New(crc32.MakeTable(crc32.Castagnoli))

	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(result))
}

// IsHash tests if a string looks like a hash.
func IsHash(s string) bool {
	if s == "" {
		return false
	}

	for _, r := range s {
		if (r < 48 || r > 57) && (r < 97 || r > 102) && (r < 65 || r > 70) {
			return false
		}
	}

	switch len(s) {
	case 8, 16, 32, 40, 56, 64, 80, 128, 256:
		return true
	}

	return false
}
