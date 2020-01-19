package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var (
	db        *sql.DB
	options   *Options
	indexHTML []byte
)

// Options Model option file
type Options struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Author      string      `json:"author"`
	Language    string      `json:"language"`
	Version     string      `json:"version"`
	ClassesGame []ClassGame `json:"classes-game"`
}

// OperationGet http header filter operation
func OperationGet(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(
		`select * from operations where 
			type LIKE "%" || $1 || "%" AND
			mission_name LIKE "%" || $2 || "%" AND
			date <= $3 AND
			date >= $4`,
		r.FormValue("type"),
		r.FormValue("name"),
		r.FormValue("older"),
		r.FormValue("newer"),
	)
	if err != nil {
		fmt.Println("error:", err)
	}
	ops := Operation{}.GetAll(rows)
	b, err := json.Marshal(ops)
	if err != nil {
		fmt.Println("error:", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

// OperationAdd http header add operation only for server
func OperationAdd(w http.ResponseWriter, r *http.Request) {

}

// StaticHandler write index.html (buffer) or send static file
func StaticHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// if it is index.html
		if path == "/" {
			w.Write(indexHTML)
			return
		}
		// json data already compressed
		if strings.HasPrefix(path, "/data/") {
			r.URL.Path += ".gz"
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Content-Type", "application/x-gzip")
		}
		next.ServeHTTP(w, r)
	})
}

// CreatePage creates a page by templates. Subsequently does not generate it again
func CreatePage(path string, param interface{}) ([]byte, error) {
	var (
		tmpl *template.Template
		err  error
		buf  bytes.Buffer
	)
	if tmpl, err = template.ParseFiles("template/" + path); err != nil {
		return nil, err
	}
	if err = tmpl.Execute(&buf, param); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func init() {
	// Parsing option file
	var (
		blob []byte
		err  error
	)
	if blob, err = ioutil.ReadFile("option.json"); err != nil {
		panic(err)
	}
	if err = json.Unmarshal(blob, &options); err != nil {
		panic(err)
	}
}

func main() {
	var err error

	// Create home page
	if indexHTML, err = CreatePage("index.html", options); err != nil {
		panic(err)
	}

	if db, err = sql.Open("sqlite3", "data.db"); err != nil {
		panic(err)
	}
	defer db.Close()

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", StaticHandler(fs))
	http.HandleFunc("/api/v1/operations/add", OperationAdd)
	http.HandleFunc("/api/v1/operations/get", OperationGet)

	fmt.Println("Listen:", http.ListenAndServe(":5000", nil))
}
