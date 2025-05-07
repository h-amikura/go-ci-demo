package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

// Hello関数（テスト用）
func Hello() string {
	return "Hello, CI!"
}

func main() {
	// PORT 環境変数を取得
	port := os.Getenv("PORT")
	if port == "" {
		port = "80" // 環境変数が設定されていなければ80番で起動
	}

	// 起動ログ
	log.Printf(">>> ポート%sでサーバーを起動します...\n", port)

	// ルートにアクセスしたときの処理
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Goアプリがポート%sで起動中です！", port)
	})

	// サーバー起動（エラーがあれば出力して終了）
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("サーバーの起動に失敗しました: %v", err)
	}
}
