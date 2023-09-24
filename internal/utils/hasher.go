package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

type Hash struct {
	hexHash    []byte
	stringHash string
}

func (hash *Hash) CalcHexHash(data []byte) []byte {
	hexHash := make([]byte, 64)
	h := sha256.New()
	h.Write(data)
	hex.Encode(hexHash, h.Sum(nil))
	hash.hexHash = hexHash
	hash.stringHash = string(hexHash)
	return hexHash
}

func (hash *Hash) String() string {
	return hash.stringHash
}

func (hash *Hash) Hex() []byte {
	return hash.hexHash
}

func NewHasher() *Hash {
	return &Hash{
		hexHash:    []byte{},
		stringHash: "",
	}
}
