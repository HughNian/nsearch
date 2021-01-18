package utils

import (
	"strings"
	"crypto/md5"
	"encoding/binary"
)


func GetKeysId(keywords ...string) (id uint64) {
	if len(keywords) == 0 {
		return 0
	}

	if len(keywords) == 1 {
		r := []byte(keywords[0])
		Md5Inst := md5.New()
		Md5Inst.Write(r)
		uret := Md5Inst.Sum([]byte(""))
		id = binary.BigEndian.Uint64(uret)
	} else if len(keywords) > 1 {
		for _, keyword := range keywords {
			r := []byte(keyword)
			Md5Inst := md5.New()
			Md5Inst.Write(r)
			uret := Md5Inst.Sum([]byte(""))
			id = binary.BigEndian.Uint64(uret)
		}
	}

	return
}

func GetWordsInfo(word string) (wsi []string) {
	wsi = strings.Split(word, "/")
	if len(wsi) != 2 {
		return nil
	}

	return wsi
}