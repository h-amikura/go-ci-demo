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
        log.Println("DB接続エラー:", err)
    } else {
        defer db.Close()

        if err = db.Ping(); err != nil {
            log.Println("DB応答なし:", err)
        } else {
            fmt.Println("DB接続成功！")
        }
    }

    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Goアプリがポート80で起動中です！")
    })

    log.Println(">>> ポート80でサーバーを起動します...")
    log.Fatal(http.ListenAndServe(":80", nil))
}
