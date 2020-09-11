package indexer

import (
	"testing"
	"strings"
	"fmt"
)

func Test_Intersect(t *testing.T) {
	str1 := `明朝/n|翰林院/n|以/n|永乐/n|朝/n|为/n|界/n`
	str2 := `明朝/n|翰林院/n|内阁/n|大学士/n|张居正/n`
	str3 := `明朝/n|永乐/n|时期/n|内阁/n|首辅/n|谢晋/n`
	str4 := `距今/n|300/n|年/n|的/n|明朝/n|翰林院/n|开启/n|了/n|我国/n|内阁/n|制度/n|，/n|取消/n|了/n|宰相/n|制度/n`

	keys1 := strings.Split(str1, `|`)
	//fmt.Println(keys1)
	keys2 := strings.Split(str2, `|`)
	//fmt.Println(keys2)
	keys3 := strings.Split(str3, `|`)
	//fmt.Println(keys3)
	keys4 := strings.Split(str4, `|`)

	index := &Index{}
	var allDoc [][]*Document
	doc1 := index.CreateDocument(1,2, float32(len(keys1)),keys1,str1)
	doc2 := index.CreateDocument(2,2, float32(len(keys2)),keys2,str2)
	doc3 := index.CreateDocument(3,2, float32(len(keys3)),keys3,str3)
	doc4 := index.CreateDocument(4,2, float32(len(keys4)),keys4,str4)
	allDoc = append(allDoc, []*Document{doc1, doc2, doc4})
	allDoc = append(allDoc, []*Document{doc2, doc3, doc1})
	allDoc = append(allDoc, []*Document{doc3, doc2, doc1, doc4})

	interDoc := getDocIntersect(allDoc)
	for k, d := range interDoc {
		fmt.Println(k, "---", d.DocId)
		fmt.Println(k, "---", d.DocType)
		fmt.Println(k, "---", d.Words)
		fmt.Println(k, "---", d.Content)
	}
}
