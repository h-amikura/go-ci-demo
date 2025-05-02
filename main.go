package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// ここに Hello() を定義
func Hello() string {
	return "Hello, CI!"
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Goアプリがポート%sで起動中です！", port)
	})

	log.Printf(">>> ポート%sでサーバーを起動します...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
