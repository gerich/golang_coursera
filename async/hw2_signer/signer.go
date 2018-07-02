package main

// SingleHash single hash
func SingleHash(in, out chan interface{}) {
	data := <-in
	str, _ := data.(string)
	out <- DataSignerCrc32(str) + "~" + DataSignerCrc32(DataSignerMd5(str))
}

// MultiHash multi hash
func MultiHash(in, out chan interface{}) {
	res := ""
	data := <-in
	str, _ := data.(string)
	for index := 0; index < 6; index++ {
		res += DataSignerCrc32(string(index) + str)
	}
	out <- res
}

// CombineResults combine results
func CombineResults(in, out chan interface{}) {
	res := ""
	for data := range in {
		str, _ := data.(string)
		res += str + "_"
	}
	out <- res[:len(res)-1]
}

//ExecutePipeline execute pipeline
func ExecutePipeline(jobs ...job) {
	var in, out chan interface{}
	for index, curr := range jobs {
		out = make(chan interface{}, 1)
		if index == 0 {
			curr(in, out)
		} else {
			curr(out, in)
		}
		in = make(chan interface{}, 1)
	}
}
