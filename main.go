package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

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
		host, port, user, password, dbname)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja"><head><meta charset="UTF-8"><title>Goサーバー</title></head><body>
	<h1>Goアプリがポート%sで起動中です！</h1>
	<p><strong>DB接続状態:</strong> %s</p>
	<p><a href="/env">▶ 環境変数を確認する</a></p>
	<p><a href="/users">▶ usersテーブルを見る</a></p>
</body></html>`, port, dbStatus)
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/env", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="ja"><head><meta charset="UTF-8"><title>環境変数の確認</title></head><body>
	<h1>connectToDB() で使用された環境変数</h1>
	<table border="1" cellpadding="5">
	<tr><th>変数名</th><th>値</th></tr>%s</table>
	<p><a href="/">← トップに戻る</a></p>
</body></html>`, envDetails)
		fmt.Fprint(w, html)
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "DB接続されていません", http.StatusInternalServerError)
			return
		}

		// 登録処理
		if r.Method == http.MethodPost {
			name := r.FormValue("name")
			email := r.FormValue("email")
			if name != "" && email != "" {
				_, err := db.Exec(`INSERT INTO public.users (name, email, created_at) VALUES ($1, $2, $3)`, name, email, time.Now())
				if err != nil {
					log.Printf("INSERT失敗: %v", err)
				}
			}
			http.Redirect(w, r, "/users", http.StatusSeeOther)
			return
		}

		// 削除処理
		if r.Method == http.MethodGet && r.URL.Query().Has("delete") {
			idStr := r.URL.Query().Get("delete")
			if id, err := strconv.Atoi(idStr); err == nil {
				_, err := db.Exec("DELETE FROM public.users WHERE id = $1", id)
				if err != nil {
					log.Printf("DELETE失敗: %v", err)
				}
			}
			http.Redirect(w, r, "/users", http.StatusSeeOther)
			return
		}

		// 一覧表示
		rows, err := db.Query(`SELECT id, name, email, created_at FROM public.users ORDER BY id`)
		if err != nil {
			http.Error(w, fmt.Sprintf("クエリエラー: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		type User struct {
			ID        int
			Name      string
			Email     string
			CreatedAt string
		}
		var users []User
		for rows.Next() {
			var u User
			if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt); err != nil {
				log.Printf("Scanエラー: %v", err)
				continue
			}
			users = append(users, u)
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<h1>users テーブルの中身</h1>
<form method="POST">
	名前: <input type="text" name="name" required>
	メール: <input type="email" name="email" required>
	<button type="submit">登録</button>
</form><br>
<table border="1"><tr><th>ID</th><th>Name</th><th>Email</th><th>CreatedAt</th><th>操作</th></tr>`)

		for _, u := range users {
			fmt.Fprintf(w,
				"<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td><a href='/users?delete=%d'>削除</a></td></tr>",
				u.ID, u.Name, u.Email, u.CreatedAt, u.ID)
		}

		fmt.Fprint(w, `</table><p><a href="/">← トップに戻る</a></p>`)
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("サーバー起動に失敗: %v", err)
	}
}
