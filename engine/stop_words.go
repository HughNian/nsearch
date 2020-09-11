package engine

import (
	"os"
	"log"
	"bufio"
)

type StopWords struct {
	swords map[string]bool
}

func NewStopWords(stopFile string) *StopWords {
	if stopFile == "" {
		return nil
	}

	swords := make(map[string]bool)

	file, err := os.Open(stopFile)
	if err != nil {
		log.Fatalln(err)
		return nil
	}
	defer file.Close()

	line := bufio.NewScanner(file)
	for line.Scan() {
		text := line.Text()
		if len(text) != 0 {
			swords[text] = true
		}
	}

	return &StopWords {
		swords : swords,
	}
}

func (sw *StopWords) StopWordsExist(words string) bool {
	_, exist := sw.swords[words]

	return exist
}