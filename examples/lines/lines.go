package main

import "fmt"
import "bufio"
import "os"
import "log"

func main() {
	count := 0
	fmt.Println("line thingy")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		count += 1
		fmt.Println(count, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
    log.Fatal(err)
}
}
