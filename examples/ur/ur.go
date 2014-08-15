package main

import "os"
import "fmt"
import "net/url"
import "strings"

// ur foo=bar URL => URL with &foo=bar added appropriately.
// any existing foo= are discarded.
func main() {
	param := os.Args[1]
	arg := os.Args[2]
	u, err := url.Parse(arg)
	panicUnlessNil(err)
	fmt.Println(adjust(param, u))
}

func adjust(param string, u *url.URL) *url.URL {
	p := strings.Split(param, "=")
	A, B := p[0], p[1]
	v := u.Query()
	v.Set(A, B)
	u.RawQuery = v.Encode()
	return u
}

func panicUnlessNil(err error) {
	if err == nil { return }
	panic(err)
}
