package utils

import (
	"strings"
)

func GetKeysId(keywords ...string) int {
	if len(keywords) == 0 {
		return 0
	}

	var id int
	if len(keywords) == 1 {
		r := []byte(keywords[0])
		for _, v := range r {
			id += int(v)
		}
	} else if len(keywords) > 1 {
		for _, keyword := range keywords {
			r := []byte(keyword)
			for _, v := range r {
				id += int(v)
			}
		}
	}

	return id
}

func GetWordsInfo(word string) (wsi []string) {
	wsi = strings.Split(word, "/")
	if len(wsi) != 2 {
		return nil
	}

	return wsi
}