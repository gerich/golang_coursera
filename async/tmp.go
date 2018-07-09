package main

import (
	"fmt"
)

func tmp(in, out chan string) {
	go func() {
		defer func() {
			fmt.Println("Ended")
		}()
		str := <-in
		str += "-tmp"
		out <- str
	}()
}

const MaxInputDataLen = 10

func main() {
	out := make(chan string, MaxInputDataLen)
	in := make(chan string, MaxInputDataLen)
	// wg := &sync.WaitGroup{}
	// wg.Add(3)
	// go func(in, out chan string, wg *sync.WaitGroup) {
	// 	defer close(out)
	// 	defer wg.Done()
	// 	out <- "foo"
	// }(in, out, wg)
	// go func(in, out chan string, wg *sync.WaitGroup) {
	// 	defer close(out)
	// 	defer wg.Done()
	// 	res := <-in
	// 	out <- (res + "-bar")
	// }(out, in, wg)
	// go func(in, out chan string, wg *sync.WaitGroup) {
	// 	defer close(out)
	// 	defer wg.Done()
	// 	res := <-in
	// 	out <- (res + "-baz")
	// }(in, out, wg)
	// fmt.Println()
	// wg.Wait()

	in <- "foo"
	go tmp(in, out)
	fmt.Println(<-out)
	fmt.Scanln()
	// close(in)
	// close(out)
}
