package common

import (
	"strconv"
	"strings"
	"unicode"
)

func MapF(filename string, content string) []KeyValue {
	words := strings.FieldsFunc(content, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	kva := []KeyValue{}
	for _, w := range words {
		kva = append(kva, KeyValue{Key: w, Value: "1"})
	}
	return kva
}

func ReduceF(key string, values []string) string {
	return strconv.Itoa(len(values))
}
