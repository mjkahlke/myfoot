package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/go-sql-driver/mysql"
)

var db *sql.DB

type Table struct {
	TableSchema string
	TableName string
	TableRows int64
	DataLength int64
	DataFree int64
	IndexLength int64
	Size int64
	SizeFS int64
}

type Footprint struct {
	fs int64	// file system
	db int64	// database
	dl int64	// data length
	il int64	// index length
	df int64	// data free
}

func getDataDir() (string, error) {
	dontcare, datadir := "", ""

	row := db.QueryRow("SHOW GLOBAL VARIABLES LIKE 'datadir'")
	if err := row.Scan(&dontcare, &datadir); err != nil {
		if err == sql.ErrNoRows {
			return datadir, fmt.Errorf("getDataDir datadir: no such global variable")
		}
		return datadir, fmt.Errorf("getDataDir datadir: %v", err)
	}
	return datadir, nil
}

func getTablesMetaData() ([]Table, error) {
	var tables []Table

	rows, err := db.Query("SELECT TABLE_SCHEMA,TABLE_NAME,TABLE_ROWS,DATA_LENGTH,DATA_FREE,INDEX_LENGTH FROM information_schema.TABLES WHERE ENGINE LIKE 'InnoDB' AND TABLE_TYPE LIKE 'BASE TABLE' AND TABLE_ROWS IS NOT NULL AND TABLE_SCHEMA NOT RLIKE '_schema$' AND TABLE_SCHEMA NOT RLIKE '^mysql$' ORDER BY TABLE_SCHEMA,TABLE_NAME")
	if err != nil {
		return nil, fmt.Errorf("getTablesMetaData: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var table Table
		if err := rows.Scan(&table.TableSchema, &table.TableName, &table.TableRows, &table.DataLength, &table.DataFree, &table.IndexLength); err != nil {
			return nil, fmt.Errorf("getTablesMetaData %q: %v", table, err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getTablesMetaData: %v", err)
	}
	return tables, nil
}

func getFsFootprint(datadir string, schema string, table string) int64 {
	files, err := ioutil.ReadDir(datadir + string(os.PathSeparator) + schema)
	if err != nil {
		log.Fatal(err)
	}

	var size int64
	for _, f := range files {
		match, _ := filepath.Match(table + "*.ibd", f.Name())
		if match {
			size += f.Size()
		}
	}
	return size;
}

func displayable(sizeInBytes int64) string {
	if sizeInBytes >= 1024*1024*1024 {
		return strconv.FormatInt(sizeInBytes/(1024*1024*1024), 10) + "GB"
	} else if sizeInBytes >= 1024*1024 {
		return strconv.FormatInt(sizeInBytes/(1024*1024), 10) + "MB"
	} else if sizeInBytes >= 1024 {
		return strconv.FormatInt(sizeInBytes/(1024), 10) + "KB"
	} else {
		return strconv.FormatInt(sizeInBytes, 10)
	}
}

func main() {
	// Capture connection properties.
	cfg := mysql.Config{
		User:	os.Getenv("DBUSER"),
		Passwd:	os.Getenv("DBPASS"),
		Net:	"tcp",
		Addr:	"127.0.0.1:3306",
		DBName: "information_schema",
	}

	// Get a database handle.
	var err error
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}

	// Get location of database files on the file system
	datadir, err := getDataDir()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Printf("datadir: %v\n", datadir)

	//var sumFS, sumDB, sumDL, sumIL, sumDF int64
	var curSchema string
	tabmap := map[string][]Table{}
	fpmap := map[string]*Footprint{}

	tables, err := getTablesMetaData()
	if err != nil {
		log.Fatal(err)
	}
	for _, t := range tables {
		t.SizeFS = getFsFootprint(datadir, t.TableSchema, t.TableName)
		t.Size = t.DataLength + t.IndexLength + t.DataFree
		if curSchema != t.TableSchema {
			curSchema = t.TableSchema
			tabmap[curSchema] = []Table{t}
			fpmap[curSchema] = new(Footprint)
		} else {
			tabmap[curSchema] = append(tabmap[curSchema], t)
		}
		fpmap[curSchema].fs += t.SizeFS
		fpmap[curSchema].db += t.Size
		fpmap[curSchema].dl += t.DataLength
		fpmap[curSchema].il += t.IndexLength
		fpmap[curSchema].df += t.DataFree
	}
	keys := make([]string, 0, len(tabmap))
	for k := range tabmap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("======== Database: %-23s  szDB: %5s  szFS: %5s  eff: %2d%%  ovhd: %2d%%\n",
			k, displayable(fpmap[k].db), displayable(fpmap[k].fs), 100*fpmap[k].db/fpmap[k].fs, 100*fpmap[k].il/fpmap[k].db)
		for _, t := range tabmap[k] {
			fmt.Printf("  table: %-16s  #rows: %8d  szDB: %5s  szFS: %5s  eff: %2d%%  ovhd: %2d%%\n",
				t.TableName, t.TableRows, displayable(t.Size), displayable(t.SizeFS), 100*t.Size/t.SizeFS, 100*t.IndexLength/(t.Size))
		}
	}
}

