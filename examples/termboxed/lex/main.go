package main

import "fmt"
import "os"

func main() {
	for _, arg := range os.Args[1:] {
		fmt.Println("-->", arg)
		analyse(arg)
	}
}

const (
	INITIAL = iota
	ID
	NUM
	STRING
	CHAR
	MULTISTRING
	OPERATOR
)

func analyse(x string) {
	state := INITIAL
	start := 0
	for here, rune := range x + "\000" {
		switch state {
		case INITIAL:
			start = here
			switch {
			case 'a' <= rune && rune <= 'z':
				state = ID
			case 'A' <= rune && rune <= 'Z':
				state = ID
			case '0' <= rune && rune <= '9':
				state = NUM
			case rune == '_':
				state = ID
			case rune == '"':
				state = STRING
			case rune == '\'':
				state = CHAR
			case rune == '`':
				state = MULTISTRING

			case rune == '+' || rune == '-':
				state = OPERATOR

			case rune == 0:
				fmt.Println("EOF")
				return

			case rune == ' ' || rune == '\t':

			default:
				// stay in initial state
			}

		case OPERATOR:
			if rune == '+' || rune == '-' {
				// continuing
			} else {
				state = INITIAL
				fmt.Println("OP(S)", x[start:here])
			}

		case STRING:

		case NUM:
			if '0' <= rune && rune <= '9' {
				// continuing ...
			} else {
				fmt.Println("NUM", x[start:here])
				state = INITIAL
			}

		case ID:
			if 'a' <= rune && rune <= 'z' || 'A' <= rune && rune <= 'Z' || '0' <= rune && rune <= '9' || rune == '_' {
				// continuing ...
			} else {
				fmt.Println("ID", x[start:here])
				state = INITIAL
			}

		case CHAR:
		case MULTISTRING:

		default:
			// bad state, should never happen
			panic("cannot handle state")
		}
	}
}
