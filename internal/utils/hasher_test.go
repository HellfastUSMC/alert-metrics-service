package utils

import (
	"fmt"
	"reflect"
	"testing"
)

func TestHash_CalcHexHash(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Calculating hash",
			args: args{data: []byte("123wresdfdryert")},
			want: "8b85bf00e09d27efa619121b06c2c088e9842a9d1e4ff5946107b55258466f4d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{}
			if got := hash.CalcHexHash(tt.args.data); string(got) != tt.want {
				t.Errorf("CalcHexHash() = %X, want %v", got, tt.want)
			}
		})
	}
}

func TestHash_Hex(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Calculating hash",
			args: args{data: []byte("123wresdfdryert")},
			want: "8b85bf00e09d27efa619121b06c2c088e9842a9d1e4ff5946107b55258466f4d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{}
			hash.CalcHexHash(tt.args.data)
			if string(hash.hexHash) != tt.want {
				t.Errorf("hash.Hex() = %X, want %v", hash.hexHash, tt.want)
			}
		})
	}
}

func TestHash_String(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Calculating hash",
			args: args{data: []byte("123wresdfdryert")},
			want: "8b85bf00e09d27efa619121b06c2c088e9842a9d1e4ff5946107b55258466f4d",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := &Hash{}
			hash.CalcHexHash(tt.args.data)
			fmt.Println(hash.String())
			if hash.String() != tt.want {
				t.Errorf("hash.String() = %s, want %v", hash.hexHash, tt.want)
			}
		})
	}
}

func TestNewHasher(t *testing.T) {
	tests := []struct {
		name string
		want *Hash
	}{
		{
			name: "Test new hasher creation",
			want: &Hash{
				hexHash:    []byte{},
				stringHash: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHasher(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHasher() = %v, want %v", got, tt.want)
			}
		})
	}
}
