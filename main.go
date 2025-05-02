package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Azure App Service から渡されるポート番号を取得（なければ80）
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
