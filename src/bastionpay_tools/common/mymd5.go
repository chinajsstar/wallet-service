package common

import (
	"crypto/md5"
	"encoding/hex"
	"io/ioutil"
	"errors"
)

const(
	// 16 bytes salt important!!!
	salt = "H@e&29_Eo$g3BC_2"
)

// from []byte
func GetSaltMd5Hex(data []byte) (string, error) {
	h := md5.New()
	h.Write(data)
	h.Write([]byte(salt))
	sum := h.Sum(nil)

	return hex.EncodeToString(sum), nil
}

func CompareSaltMd5Hex(data []byte, md5Slat string) error {
	if md5Slat == "" {
		return errors.New("empty md5")
	}

	oldmd5Slat, err := GetSaltMd5Hex(data)
	if err != nil {
		return err
	}
	if oldmd5Slat != md5Slat {
		return errors.New("error md5")
	}

	return nil
}

// from string
func GetSaltMd5HexByText(text string) (string, error) {
	return GetSaltMd5Hex([]byte(text))
}

func CompareSaltMd5HexByText(text, md5Slat string) error {
	return CompareSaltMd5Hex([]byte(text), md5Slat)
}

// from file path
func GetSaltMd5HexByFile(path string) (string, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return GetSaltMd5Hex(data)
}

func CompareSaltMd5HexByFile(path, md5Slat string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	return CompareSaltMd5Hex(data, md5Slat)
}

