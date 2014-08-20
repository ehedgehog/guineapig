package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// linkify turns a stream of URIs into HTML links.
// The input has one URI per line and that URI is
// converted into an <a href="URI"> element with body
// the text from the last / to the first ? (if any)
// or end-of-URI. The text is lower-cased.
func main() {
	count := 0
	fmt.Println("line thingy")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		count += 1
		line := scanner.Text()
		content := prune(line)
		fmt.Printf(`<div><a href="%v">%v</a></div>%v`, line, content, "\n")
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

// prune strips out the unwanted text. There's a better way than
// doing two scans, and we'd care if this was intended for heavy
// use.
func prune(s string) string {
	q := strings.Index(s, "?")
	if q > -1 {
		s = s[:q]
	}
	x := strings.LastIndex(s, "/")
	s = s[x+1:]
	return strings.ToLower(strings.Replace(s, "-", " ", -1))
}
