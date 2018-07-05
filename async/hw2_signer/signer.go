package main

import (
	"sort"
	"strconv"
	"sync"
)

// SingleHash single hash
func SingleHash(in, out chan interface{}) {
	for data := range in {
		intData, _ := data.(int)
		str := strconv.Itoa(intData)
		out <- (DataSignerCrc32(str) + "~" + DataSignerCrc32(DataSignerMd5(str)))
	}
}

// MultiHash multi hash
func MultiHash(in, out chan interface{}) {
	for data := range in {
		res := ""
		str, _ := data.(string)
		for index := 0; index < 6; index++ {
			res += DataSignerCrc32(string(index) + str)
		}
		out <- res
	}
}

// CombineResults combine results
func CombineResults(in, out chan interface{}) {
	store := []string{}

	for data := range in {
		str, _ := data.(string)
		store = append(store, str)
	}

	sort.Strings(store)
	res := ""
	for _, str := range store {
		res += str + "_"
	}
	out <- res[:len(res)-1]
}

//ExecutePipeline execute pipeline
func ExecutePipeline(jobs ...job) {
	out := make(chan interface{})
	in := make(chan interface{})
	wg := &sync.WaitGroup{}
	for _, curr := range jobs {
		wg.Add(1)
		go func(in, out chan interface{}, curr job, wg *sync.WaitGroup) {
			defer wg.Done()
			defer close(out)
			curr(in, out)
		}(in, out, curr, wg)
		in = out
		out = make(chan interface{})
	}
	wg.Wait()
}
