package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var dbStatus = "❌ DB接続に失敗しました"

func connectToDB() *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("❌ DB接続失敗: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		log.Printf("❌ DB疎通エラー: %v", err)
		return nil
	}

	log.Println("✅ PostgreSQL に接続成功しました！")
	dbStatus = "✅ PostgreSQL に接続できています"
	return db
}

func main() {
	db := connectToDB()
	if db != nil {
		defer db.Close()
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	log.Printf(">>> ポート%sでサーバーを起動します...\n", port)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := fmt.Sprintf(`
			<!DOCTYPE html>
			<html lang="ja">
			<head><meta charset="UTF-8"><title>Goサーバー</title></head>
			<body>
				<h1>Goアプリがポート%sで起動中です！</h1>
				<p><strong>DB接続状態:</strong> %s</p>
				<p><a href="/env">▶ 環境変数を確認する</a></p>
			</body>
			</html>
		`, port, dbStatus)
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/env", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := `
			<!DOCTYPE html>
			<html lang="ja">
			<head><meta charset="UTF-8"><title>環境変数一覧</title></head>
			<body>
				<h1>環境変数の確認</h1>
				<table border="1" cellpadding="5" cellspacing="0">
					<tr><th>変数名</th><th>値</th></tr>
		`

		keys := []string{"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME", "PORT"}
		for _, key := range keys {
			val := os.Getenv(key)
			html += fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>", key, val)
		}

		html += `
				</table>
			</body>
			</html>
		`
		fmt.Fprint(w, html)
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("サーバー起動に失敗: %v", err)
	}
}

// Hello関数（テスト用）
func Hello() string {
	return "Hello, CI!"
}
