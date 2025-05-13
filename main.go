package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var (
	db             *sql.DB
	dbStatus       = "❌ DB接続失敗"
	envDetails     = ""
	clientID       = os.Getenv("AZURE_CLIENT_ID")
	clientSecret   = os.Getenv("AZURE_CLIENT_SECRET")
	redirectURL    = os.Getenv("AZURE_REDIRECT_URL")
	tenantID       = os.Getenv("AZURE_TENANT_ID")
	oauth2Config   *oauth2.Config
	verifier       *oidc.IDTokenVerifier
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

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		host, port, user, password, dbname)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("❌ sql.Open エラー: %v", err)
		return nil
	}
	if err := conn.Ping(); err != nil {
		log.Printf("❌ db.Ping エラー: %v", err)
		return nil
	}

	dbStatus = "✅ DB接続成功"
	return conn
}

func main() {
	ctx := context.Background()
	db = connectToDB()
	if db != nil {
		defer db.Close()
	}

	provider, err := oidc.NewProvider(ctx, "https://login.microsoftonline.com/"+tenantID+"/v2.0")
	if err != nil {
		log.Fatalf("❌ OIDCプロバイダ初期化失敗: %v", err)
	}
	oauth2Config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}
	verifier = provider.Verifier(&oidc.Config{ClientID: clientID})

	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/env", handleEnv)
	http.HandleFunc("/add", handleAdd)
	http.HandleFunc("/delete", handleDelete)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/auth/callback", handleCallback)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf(">>> ポート%sでサーバー起動中...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	user := r.URL.Query().Get("user")
	if user == "" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	userRows := ""
	if db != nil {
		rows, err := db.Query("SELECT id, name, email, created_at FROM users ORDER BY id")
		if err != nil {
			log.Printf("❌ ユーザー一覧取得失敗: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var id int
				var name, email string
				var createdAt string
				if err := rows.Scan(&id, &name, &email, &createdAt); err == nil {
					userRows += fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td></tr>", id, name, email, createdAt)
				}
			}
		}
	}
	html := fmt.Sprintf(`<!DOCTYPE html><html lang="ja"><head><meta charset="UTF-8"><title>Goサーバー</title></head><body>
	<h1>Goアプリ稼働中</h1>
	<p><strong>ログインユーザー:</strong> %s</p>
	<p><strong>DB接続状態:</strong> %s</p>
	<p><a href="/env">▶ 環境変数を確認</a></p>
	<h2>ユーザー登録</h2><form action="/add" method="POST">
	<p>名前: <input type="text" name="name" required></p>
	<p>Email: <input type="email" name="email" required></p>
	<button type="submit">登録</button></form>
	<h2>全ユーザー削除</h2><form action="/delete" method="POST">
	<button type="submit" onclick="return confirm('本当に削除？');">削除</button></form>
	<h2>ユーザー一覧</h2><table border="1" cellpadding="5">
	<tr><th>ID</th><th>名前</th><th>Email</th><th>作成日時</th></tr>
	%s</table></body></html>`, user, dbStatus, userRows)
	fmt.Fprint(w, html)
}

func handleEnv(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	html := fmt.Sprintf(`<h2>環境変数</h2><table border="1" cellpadding="5">
	<tr><th>名前</th><th>値</th></tr>%s</table>
	<p><a href="/">← トップに戻る</a></p>`, envDetails)
	fmt.Fprint(w, html)
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
	if r.Method == http.MethodPost && db != nil {
		_, err := db.Exec("DELETE FROM users")
		if err != nil {
			log.Printf("❌ ユーザー削除失敗: %v", err)
		} else {
			log.Println("✅ 全ユーザー削除成功")
		}
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, oauth2Config.AuthCodeURL("state"), http.StatusFound)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	token, err := oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "トークン交換エラー", http.StatusInternalServerError)
		return
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "IDトークン取得失敗", http.StatusInternalServerError)
		return
	}
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		http.Error(w, "IDトークン検証失敗", http.StatusInternalServerError)
		return
	}
	var claims struct {
		Email string `json:"email"`
	}
	if err := idToken.Claims(&claims); err != nil {
		http.Error(w, "クレームパース失敗", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/?user="+claims.Email, http.StatusSeeOther)
}
// Hello関数（テスト用）
func Hello() string {
	return "Hello, CI!"
}