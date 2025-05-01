package main

import "fmt"

// Hello returns a greeting message.
func Hello() string {
	return "Hello, CI!"
}

func main() {
	fmt.Println(Hello())
}
