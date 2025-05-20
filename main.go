package main

import (
	"mapreduce/common"
)

func main() {
	common.DoMap("job1", 0, "input.txt", 3, common.MapF)
}
