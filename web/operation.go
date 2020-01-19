package main

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

// ClassGame Model type game in option file
// Example [TvT, tvt]
type ClassGame struct {
	Key  string `json:"key"`
	Name string `json:"name"`
}

// Operation Model opearation table in db
type Operation struct {
	ID              int64   `json:"id"`
	WorldName       string  `json:"world_name"`
	MissionName     string  `json:"mission_name"`
	MissionDuration float64 `json:"mission_duration"`
	Filename        string  `json:"filename"`
	Date            string  `json:"date"`
	Class           string  `json:"type"`
}

// OperationFilter model for filter operations
type OperationFilter struct {
	MissionName string
	DateOlder   string
	DateNewer   string
	Class       string
}

// NewOperation by http request
func NewOperation(r *http.Request) (op Operation, err error) {
	op = Operation{}
	op.WorldName = r.FormValue("worldName")
	op.MissionName = r.FormValue("missionName")
	op.MissionDuration, err = strconv.ParseFloat(r.FormValue("missionDuration"), 64)
	op.Filename = r.FormValue("filename")
	op.Date = time.Now().Format("2006-01-02")
	op.Class = r.FormValue("type")
	return op, err
}

// SaveFileAsGZIP saves the file in compressed form on the server
func (o *Operation) SaveFileAsGZIP(dir string, r io.Reader) (err error) {
	o.Filename = time.Now().Format("2006-01-02_15-04-05") + ".json"

	f, err := os.Create(path.Join(dir, o.Filename+".gz"))
	defer f.Close()
	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	w := gzip.NewWriter(f)
	defer w.Close()
	_, err = w.Write(content)
	if err != nil {
		return err
	}
	return nil
}

// Insert new row in db
func (o *Operation) Insert(db *sql.DB) (sql.Result, error) {
	return db.Exec(
		`insert into operations 
			(world_name, mission_name, mission_duration, filename, date, type)
		values
			($1, $2, $3, $4, $5, $6)
		`,
		o.WorldName, o.MissionName, o.MissionDuration, o.Filename, o.Date, o.Class,
	)
}

// GetByFilter get all operations matching the filter
func (o *OperationFilter) GetByFilter(db *sql.DB) (operations []Operation, err error) {
	rows, err := db.Query(
		`select * from operations where 
			mission_name LIKE "%" || $2 || "%" AND
			date <= $3 AND
			date >= $4 AND
			type LIKE "%" || $1 || "%"`,
		o.MissionName,
		o.DateOlder,
		o.DateNewer,
		o.Class,
	)
	if err != nil {
		return nil, err
	}
	return executeAll(rows), nil
}

// GetAll execute all operation in array
func executeAll(rows *sql.Rows) (operations []Operation) {
	for rows.Next() {
		op := Operation{}
		err := rows.Scan(
			&op.ID,
			&op.WorldName,
			&op.MissionName,
			&op.MissionDuration,
			&op.Filename,
			&op.Date,
			&op.Class,
		)
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		operations = append(operations, op)
	}
	return operations
}
