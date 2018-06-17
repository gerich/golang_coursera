package main

import (
	"fmt"
)

func main() {
	foo := []string{"foo", "bar"}

	a, b, c := foo[0], foo[1], foo[2:]

	fmt.Println(a, b, c)
}
