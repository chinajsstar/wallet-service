package utils

import (
	"fmt"
	"math/big"
)

func string_has_prefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[0:len(prefix)] == prefix
}

func string_cat_prefix(s string, prefix string) (string, error) {
	if string_has_prefix(s, prefix) {
		return s[len(prefix):], nil
	}
	return "", fmt.Errorf("string:%s, has no prfix:%s", s, prefix)
}

func Hex_string_to_big_int(s string) (*big.Int, error) {
	var (
		err        error
		bigint     *big.Int
		string_hex string
	)
	string_hex, err = string_cat_prefix(s, "0x")
	if nil != err {
		return nil, err
	}

	bigint, isok := new(big.Int).SetString(string_hex, 16)
	if !isok {
		err = fmt.Errorf("can not convert hex string:%s, to big int", string_hex)
	}

	return bigint, err
}
