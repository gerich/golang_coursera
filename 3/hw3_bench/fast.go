package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	strAndroid = "Android"
	strMSIE    = "MSIE"
)

//easyjson:json
type User struct {
	Browsers []string `json:"browsers,[]string"`
	Name     string   `json:"name,string"`
	Email    string   `json:"email,string"`
}

// вам надо написать более быструю оптимальную этой функции
func FastSearch(out io.Writer) {
	foundUsers := make([]string, 0, 100)
	notSeenBefore := true
	var isAndroid bool
	var isMSIE bool
	seenBrowsers := make(map[string]bool)

	file, err := os.Open(filePath)
	defer file.Close()
	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)
	// scanner.Split(bufio.ScanLines)
	// dec := json.NewDecoder(file)
	i := -1
	u := User{}
	var l []byte
	for scanner.Scan() {
		// err := dec.Decode(&u)
		l = scanner.Bytes()
		err := u.UnmarshalJSON(l)
		if err != nil {
			panic(err)
		}

		i++

		if len(u.Browsers) == 0 {
			// log.Println("cant cast browsers")
			continue
		}

		isAndroid = false
		isMSIE = false

		for _, browser := range u.Browsers {
			if strings.Contains(browser, strAndroid) {
				isAndroid = true
				_, notSeenBefore = seenBrowsers[browser]
				if !notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[browser] = true
				}
			}

			if strings.Contains(browser, strMSIE) {
				isMSIE = true
				_, notSeenBefore = seenBrowsers[browser]
				if !notSeenBefore {
					// log.Printf("SLOW New browser: %s, first seen: %s", browser, user["name"])
					seenBrowsers[browser] = true
				}
			}
		}

		if !(isAndroid && isMSIE) {
			continue
		}

		// log.Println("Android and MSIE user:", user["name"], user["email"])
		foundUsers = append(foundUsers, fmt.Sprintf("[%d] %s <%s>\n", i, u.Name, strings.Replace(u.Email, "@", " [at] ", 1)))
	}

	fmt.Fprintln(out, "found users:\n"+strings.Join(foundUsers, ""))
	fmt.Fprintln(out, "Total unique browsers", len(seenBrowsers))
}
