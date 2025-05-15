package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	_ "github.com/lib/pq"
	"golang.org/x/oauth2"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	db           *sql.DB
	dbStatus     = "❌ DB接続失敗"
	envDetails   = ""
	clientID     = os.Getenv("AZURE_CLIENT_ID")
	clientSecret = os.Getenv("AZURE_CLIENT_SECRET")
	redirectURL  = os.Getenv("AZURE_REDIRECT_URL")
	tenantID     = os.Getenv("AZURE_TENANT_ID")
	oauth2Config *oauth2.Config
	verifier     *oidc.IDTokenVerifier
)

func initTracer() func() {
	ctx := context.Background()
	connStr := os.Getenv("APPLICATIONINSIGHTS_CONNECTION_STRING")
	if connStr == "" {
		log.Println("⚠️ Application Insights接続文字列が未設定")
		return func() {}
	}

	exporter, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint("japaneast-1.in.applicationinsights.azure.com"),
		otlptracehttp.WithHeaders(map[string]string{
			"Authorization": connStr,
		}),
	)
	if err != nil {
		log.Printf("❌ OTLPエクスポーター初期化失敗: %v", err)
		return func() {}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("mywebapp12345"),
		)),
	)
	otel.SetTracerProvider(tp)
	log.Println("✅ OpenTelemetry トレーサー初期化成功")
	return func() {
		_ = tp.Shutdown(ctx)
	}
}

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
	<tr><td>DB_NAME</td><td>%s</td></tr>`, host, port, user, password, dbname)

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

func requireLogin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("id_token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		ctx := context.Background()
		_, err = verifier.Verify(ctx, cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from handleRoot")
}

func handleEnv(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from handleEnv")
}

func handleAdd(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from handleAdd")
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from handleDelete")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from handleLogin")
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from handleCallback")
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello from handleLogout")
}

func getUserEmailFromToken(r *http.Request) string {
	return "test@example.com"
}

func Hello() string {
	return "Hello, CI!"
}

func main() {
	ctx := context.Background()

	shutdown := initTracer()
	defer shutdown()

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

	http.HandleFunc("/", requireLogin(handleRoot))
	http.HandleFunc("/env", requireLogin(handleEnv))
	http.HandleFunc("/add", requireLogin(handleAdd))
	http.HandleFunc("/delete", requireLogin(handleDelete))
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/auth/callback", handleCallback)
	http.HandleFunc("/logout", handleLogout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf(">>> ポート%sでサーバー起動中...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
