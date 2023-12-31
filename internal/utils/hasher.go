// Package utils Пакет вспомогательных функций для агента и сервера
package utils

import (
	"crypto/sha256"
	"encoding/hex"
)

// Hash Структура хешера
type Hash struct {
	stringHash string
	hexHash    []byte
}

// CalcHexHash Функция рассчета хеша по слайсу байт
func (hash *Hash) CalcHexHash(data []byte) []byte {
	hexHash := make([]byte, 64)
	h := sha256.New()
	h.Write(data)
	hex.Encode(hexHash, h.Sum(nil))
	hash.hexHash = hexHash
	hash.stringHash = string(hexHash)
	return hexHash
}

// String Функция возврата хеша в виде строки
func (hash *Hash) String() string {
	return hash.stringHash
}

// Hex Функция возврата хеша в виде массива байт
func (hash *Hash) Hex() []byte {
	return hash.hexHash
}

// NewHasher Функция инициализации новой структуры хеша
func NewHasher() *Hash {
	return &Hash{
		hexHash:    []byte{},
		stringHash: "",
	}
}
