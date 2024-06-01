package main

import (
	"fmt"
)

func ohno(err error) {
	if err != nil {
		panic(fmt.Sprintf("%v", err))
	}
}
