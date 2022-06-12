package pkg

import "fmt"

func trigger() { // raise: U1000
	fmt.Sprint("trigger") // raise: SA4017, S1039
}
