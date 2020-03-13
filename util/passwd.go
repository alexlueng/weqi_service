package util

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

// md5
func GenMD5Password(passwd string) string {
	digest := md5.Sum([]byte(passwd))
	return hex.EncodeToString(digest[:])
}

func GenSHA256Password(passwd string) string {
	digest := sha256.Sum256([]byte(passwd))
	return hex.EncodeToString(digest[:])
}
