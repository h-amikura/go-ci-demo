package main

import "fmt"
// これはテスト用の変更です
func Hello() string {
	return "Hello, CI!"
}

func main() {
	fmt.Println(Hello())
}
