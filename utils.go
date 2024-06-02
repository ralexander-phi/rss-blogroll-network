package main

import (
	"fmt"
)

func ohno(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
