package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

func SingleHash(in, out chan interface{}) {
	runIndex := 0
	var md5Sign string
	wg := &sync.WaitGroup{}
	for data := range in {
		intData, _ := data.(int)
		str := strconv.Itoa(intData)
		md5Sign = DataSignerMd5(str)
		wg.Add(1)
		go func(str, md5Sign string, wrapWg *sync.WaitGroup) {
			defer wrapWg.Done()
			var crc32Sign, crc32md5Sign string
			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				crc32Sign = DataSignerCrc32(str)
			}()
			go func() {
				defer wg.Done()
				crc32md5Sign = DataSignerCrc32(md5Sign)
			}()
			wg.Wait()
			out <- (crc32Sign + "~" + crc32md5Sign)
		}(str, md5Sign, wg)
		fmt.Printf("SingeHash: #%v\n", runIndex)
		runIndex++
	}

	wg.Wait()
}

// MultiHash multi hash
func MultiHash(in, out chan interface{}) {
	runIndex := 0
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go func(data interface{}, wrapWg *sync.WaitGroup) {
			defer wrapWg.Done()

			wg := &sync.WaitGroup{}
			results := make([]string, 6)
			str, _ := data.(string)
			for index := range results {
				wg.Add(1)
				go func(index int) {
					defer wg.Done()
					results[index] = DataSignerCrc32(strconv.Itoa(index) + str)
				}(index)
			}
			wg.Wait()
			out <- strings.Join(results, "")
		}(data, wg)
		fmt.Printf("MultiHash: #%v\n", runIndex)
		runIndex++
	}
	wg.Wait()
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
	out := make(chan interface{}, MaxInputDataLen)
	in := make(chan interface{}, MaxInputDataLen)
	wg := &sync.WaitGroup{}
	for _, curr := range jobs {
		wg.Add(1)
		go func(in, out chan interface{}, curr job, wg *sync.WaitGroup) {
			defer close(out)
			defer wg.Done()
			curr(in, out)
		}(in, out, curr, wg)
		in = out
		out = make(chan interface{}, MaxInputDataLen)
	}
	wg.Wait()
}
