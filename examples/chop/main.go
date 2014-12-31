package main

import (
	"encoding/csv"
	"os"
)

import "bufio"

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	sink := csv.NewWriter(os.Stdout)
	for scanner.Scan() {
		line := scanner.Text()
		sink.Write([]string{"TEXT", line})
	}
	sink.Flush()

}
