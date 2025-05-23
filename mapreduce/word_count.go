package mapreduce

import (
	"strconv"
	"strings"
	"unicode"
)

// The mapping function is called once for each piece of the input.
// In this framework, the value is the contents of the file being
// processed. The return value should be a slice of key/value pairs,
// each represented by a mapreduce.KeyValue.
// A COMPLETER
func MapWordCount(value string) (res []KeyValue) {
	// Split by non-letter characters
	splitFunc := func(r rune) bool {
		return !unicode.IsLetter(r)
	}

	words := strings.FieldsFunc(value, splitFunc)
	wordCounts := make(map[string]int)

	for _, word := range words {
		normalized := strings.ToLower(word)
		wordCounts[normalized]++
	}

	for word, count := range wordCounts {
		res = append(res, KeyValue{Key: word, Value: strconv.Itoa(count)})
	}
	return
}

// The reduce function is called once for each key generated by Map,
// with a list of that key's string value (merged across all
// inputs). The return value should be a single output value for that
// key.
// A COMPLETER
func ReduceWordCount(key string, values []string) string {
	sum := 0
	for _, val := range values {
		n, _ := strconv.Atoi(val)
		sum += n
	}
	return strconv.Itoa(sum)
}
