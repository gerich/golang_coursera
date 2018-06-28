package main

import (
	"sort"
)

// SingleHash single hash
func SingleHash(in, out chan) job {
	return DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
}

// MultiHash multi hash
func MultiHash() job {
	return data
}

func CombineResults(data []string) job {
	sort.Strings(data)
	for _, str := range data {
		res += str + "_"
	}
	return res[:len(res)-1]
}

//ExecutePipeline execute pipeline
func ExecutePipeline(jobs ...job) {

}
