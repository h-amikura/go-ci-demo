package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Goアプリがポート80で起動中です！")
	})

	log.Println(">>> ポート80でサーバーを起動します...")
	log.Fatal(http.ListenAndServe(":80", nil))
}
