package main

import (
	"database/sql"
	"flag"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/cn4hub/snippetbox/internal/models"
	_ "github.com/go-sql-driver/mysql"
)

type application struct {
	logger        *slog.Logger
	snippets      *models.SnippetModel
	templateCache map[string]*template.Template
}

func main() {
	addr := flag.String("addr", ":4000", "http service address")
	dsn := flag.String("dsn", "web:webpassword@/snippetbox?parseTime=true", "MySQL data source namw")
	flag.Parse()

	app := &application{
		logger:        slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		snippets:      nil,
		templateCache: nil,
	}

	db, err := openDB(*dsn)
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}
	defer db.Close()

	templateCache, err := newTemplateCache()
	if err != nil {
		app.logger.Error(err.Error())
		os.Exit(1)
	}

	app.snippets = &models.SnippetModel{Db: db}
	app.templateCache = templateCache

	app.logger.Info("starting server", slog.Any("addr", *addr))
	err = http.ListenAndServe(*addr, app.routes())
	app.logger.Error(err.Error())
	os.Exit(1)
}

func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func openDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
