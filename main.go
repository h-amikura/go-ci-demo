package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

var (
	dbStatus   = "❌ DB接続に失敗しました"
	envDetails = ""
	db         *sql.DB
)

func connectToDB() *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	envDetails = fmt.Sprintf(`
		<tr><td>DB_HOST</td><td>%s</td></tr>
		<tr><td>DB_PORT</td><td>%s</td></tr>
		<tr><td>DB_USER</td><td>%s</td></tr>
		<tr><td>DB_PASSWORD</td><td>%s</td></tr>
		<tr><td>DB_NAME</td><td>%s</td></tr>`,
		host, port, user, password, dbname,
	)

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname,
	)

	log.Printf("🔍 DSN: %s", dsn)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("❌ sql.Open 失敗: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		log.Printf("❌ db.Ping 失敗: %v", err)
		return nil
	}

	log.Println("✅ PostgreSQL に接続成功しました！")
	dbStatus = "✅ PostgreSQL に接続できています"
	return db
}

func main() {
	db = connectToDB()
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
		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja">
<head><meta charset="UTF-8"><title>Goサーバー</title></head>
<body>
	<h1>Goアプリがポート%sで起動中です！</h1>
	<p><strong>DB接続状態:</strong> %s</p>
	<p><a href="/env">▶ 環境変数を確認する</a></p>
	<p><a href="/users">▶ users テーブルを表示</a></p>
</body>
</html>`, port, dbStatus)
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/env", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		envHTML := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja">
<head><meta charset="UTF-8"><title>環境変数の確認</title></head>
<body>
	<h1>connectToDB() で使用された環境変数</h1>
	<table border="1" cellpadding="5">
		<tr><th>変数名</th><th>値</th></tr>
		%s
	</table>
	<p><a href="/">← トップに戻る</a></p>
</body>
</html>`, envDetails)
		fmt.Fprint(w, envHTML)
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if db == nil {
			fmt.Fprint(w, "<p>❌ DBが未接続です</p><a href='/'>← 戻る</a>")
			return
		}

		rows, err := db.Query(`SELECT id, "name", email, created_at FROM public.users`)
		if err != nil {
			fmt.Fprintf(w, "<p>❌ クエリ実行失敗: %v</p><a href='/'>← 戻る</a>", err)
			return
		}
		defer rows.Close()

		html := `<h1>users テーブルの一覧</h1><table border="1" cellpadding="5"><tr><th>ID</th><th>名前</th><th>Email</th><th>作成日時</th></tr>`
		for rows.Next() {
			var id int
			var name, email string
			var createdAt string
			if err := rows.Scan(&id, &name, &email, &createdAt); err != nil {
				html += fmt.Sprintf("<tr><td colspan='4'>❌ 読み込みエラー: %v</td></tr>", err)
				continue
			}
			html += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td></tr>", id, name, email, createdAt)
		}
		html += `</table><p><a href="/">← トップに戻る</a></p>`
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
