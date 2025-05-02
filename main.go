
package main

import (
    "database/sql"
    "fmt"
    "log"
    "net/http"
    _ "github.com/lib/pq"
)

func main() {
    dbURL := "postgres://postgres:mysecretpassword@host.docker.internal:5432/postgres?sslmode=disable"
    db, err := sql.Open("postgres", dbURL)
    if err != nil {
        log.Fatal("DB接続エラー:", err)
    }
    defer db.Close()

    err = db.Ping()
    if err != nil {
        log.Fatal("DB応答なし:", err)
    }

    fmt.Println("DB接続成功！")

    // ポート80でHTTPサーバー起動
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "DB接続成功！Goサーバーは正常に動いています。")
    })

    log.Println("サーバーをポート80で起動中...")
    log.Fatal(http.ListenAndServe(":80", nil))
}
