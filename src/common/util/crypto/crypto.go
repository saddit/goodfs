package crypto

import (
	"common/datasize"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"hash"
	"io"
)

// SHA256IO encode data stream by sha256
func SHA256IO(reader io.Reader) string {
	crypto := sha256.New()
	if _, e := io.CopyBuffer(crypto, reader, make([]byte, 4*datasize.MB)); e == nil {
		b := crypto.Sum(make([]byte, 0, crypto.Size()))
		return hex.EncodeToString(b)
	}
	return ""
}

// SHA256 encode bytes by sha256
func SHA256(bt []byte) string {
	return Hash(bt, sha256.New())
}

// MD5 encode bytes by MD5
func MD5(bt []byte) string {
	return Hash(bt, md5.New())
}

// Hash encode bytes
func Hash(bt []byte, exec hash.Hash) string {
	_, _ = exec.Write(bt)
	res := exec.Sum(make([]byte, 0, exec.Size()))
	return hex.EncodeToString(res)
}
