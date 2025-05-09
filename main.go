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
	envDetails = "" // HTML出力用
	db         *sql.DB
)

func connectToDB() *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// 環境変数の表示用HTML作成
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

	d, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("❌ sql.Open 失敗: %v", err)
		return nil
	}

	if err := d.Ping(); err != nil {
		log.Printf("❌ db.Ping 失敗: %v", err)
		return nil
	}

	log.Println("✅ PostgreSQL に接続成功しました！")
	dbStatus = "✅ PostgreSQL に接続できています"
	return d
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

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/env", handleEnv)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/delete", handleDelete)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("サーバー起動に失敗: %v", err)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja">
<head><meta charset="UTF-8"><title>Goサーバー</title></head>
<body>
	<h1>Goアプリがポート%sで起動中です！</h1>
	<p><strong>DB接続状態:</strong> %s</p>
	<p><a href="/env">▶ 環境変数を確認する</a></p>

	<h2>ユーザー登録</h2>
	<form action="/add" method="POST">
		<p>名前: <input type="text" name="name" required></p>
		<p>Email: <input type="email" name="email" required></p>
		<button type="submit">登録</button>
	</form>

	<h2>全ユーザー削除</h2>
	<form action="/delete" method="POST">
		<button type="submit" onclick="return confirm('本当に全削除してよろしいですか？');">削除</button>
	</form>
</body>
</html>`, os.Getenv("PORT"), dbStatus)
	fmt.Fprint(w, html)
}

func handleEnv(w http.ResponseWriter, r *http.Request) {
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
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		name := r.FormValue("name")
		email := r.FormValue("email")

		if db != nil {
			_, err := db.Exec("INSERT INTO users (name, email) VALUES ($1, $2)", name, email)
			if err != nil {
				log.Printf("❌ ユーザー登録失敗: %v", err)
			} else {
				log.Println("✅ ユーザー登録成功")
			}
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if db != nil {
			_, err := db.Exec("DELETE FROM users")
			if err != nil {
				log.Printf("❌ ユーザー削除失敗: %v", err)
			} else {
				log.Println("✅ 全ユーザー削除成功")
			}
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Hello関数（テスト用）
func Hello() string {
	return "Hello, CI!"
}
