package main

import (
	"fmt"
)

func main() {
	foo := &[]string{"foo", "bar"}
	bar := *foo

	a, b, c := bar[0], bar[1], (*foo)[2:]

	fmt.Println(a, b, c)
}
