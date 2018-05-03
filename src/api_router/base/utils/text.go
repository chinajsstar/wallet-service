package utils

import (
	"fmt"
	"crypto/md5"
	"encoding/hex"
	"os"
	"io"
)

// Scan input to string with '\n' with end line
func ScanLine() string {
	var c byte
	var err error
	var b []byte
	for ; err == nil; {
		_, err = fmt.Scanf("%c", &c)
		if c != '\n' && c!= '\r' {
			b = append(b, c)
		} else {
			break
		}
	}
	return string(b)
}

func GetMd5Text(text string) string {
	h := md5.New()
	h.Write([]byte(text))
	sum := h.Sum(nil)

	return hex.EncodeToString(sum)
}

func CopyFile(src, dst string)(w int64, err error){
	srcFile,err := os.Open(src)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}
	defer srcFile.Close()

	dstFile,err := os.Create(dst)
	if err!=nil{
		fmt.Println(err.Error())
		return
	}

	defer dstFile.Close()

	return io.Copy(dstFile,srcFile)
}