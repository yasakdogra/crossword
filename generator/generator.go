package main

import (
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func main() {
	f, err := os.ReadFile("/usr/share/dict/words")

	if err != nil {
		log.Fatal(err)
	}

	words := string(f)

	for i := 2; i < 6; i++ {
		re := regexp.MustCompile(`(?m)^[a-zA-Z]{` + strconv.Itoa(i) + `}$`)
		tmp := re.FindAllString(words, -1)

		w, err := os.Create("words-" + strconv.Itoa(i) + ".txt")
		if err != nil {
			log.Fatal(err)
		}

		w.WriteString(strings.ToUpper(strings.Join(tmp, "\n")))
		w.Close()
	}
}
