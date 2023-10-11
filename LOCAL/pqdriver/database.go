package pqdriver

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"
)

var globalPool *pgxpool.Pool

func init() {
	globalPool = InitDB()
}

func InitDB() *pgxpool.Pool {
	settings, err := ioutil.ReadFile("settings.json")
	if err != nil {
		log.Fatalf("Failed to read settings.json: %v", err)
	}

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable",
		gjson.GetBytes(settings, "pgUser").String(),
		gjson.GetBytes(settings, "pgPassword").String(),
		gjson.GetBytes(settings, "pgServer").String(),
		gjson.GetBytes(settings, "pgPort").String(),
		gjson.GetBytes(settings, "pgDataBase").String())

	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	return dbpool
}

func DataProc(a string, q string, w string, t []string, e string) {
	qColumns := strings.Split(strings.Trim(q, "()"), ", ")
	whereClause := fmt.Sprintf("%s = %s", qColumns[0], strings.Split(strings.Trim(w, "()"), ", ")[0])

	sqlSelect := fmt.Sprintf("SELECT %s FROM %s WHERE %s;", strings.Join(qColumns, ", "), a, whereClause)

	rows, err := globalPool.Query(context.Background(), sqlSelect)
	if err != nil {
		log.Println(err)
		return
	}
	defer rows.Close()

	wValues := strings.Split(strings.Trim(w, "()"), ", ")
	edited := make([]bool, len(qColumns))
	newStr := make([]interface{}, len(qColumns))

	for rows.Next() {
		values, _ := rows.Values()
		for i, v := range values {
			if t[i] == "int" {
				parsedInt, err := strconv.ParseInt(wValues[i], 10, 64)
				if err != nil {
					log.Printf("Error parsing string to int for value '%s': %v", wValues[i], err)
					continue
				}

				switch vTyped := v.(type) {
				case int16:
					if int64(vTyped) != parsedInt {
						edited[i] = true
						newStr[i] = v
					}
				case int64:
					if vTyped != parsedInt {
						edited[i] = true
						newStr[i] = v
					}
				default:
					log.Printf("Unexpected type for value: %T", v)
				}

			} else if t[i] == "str" {
				if v.(string) != wValues[i] {
					edited[i] = true
					newStr[i] = v
				}
			}
		}

	}
	for i, isEdited := range edited {
		if isEdited {
			var sqlUpdate, sqlLog string
			if t[i] == "int" {
				sqlUpdate = fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s;", a, qColumns[i], whereClause)
				sqlLog = fmt.Sprintf("INSERT INTO editLog (columnEdit, oldDataColumn, newDataColumn, timeEdited) VALUES ('%s', %s, %v, %d);", qColumns[i], wValues[i], newStr[i], time.Now().Unix())

			} else if t[i] == "str" {
				sqlUpdate = fmt.Sprintf("UPDATE %s SET %s = $1 WHERE %s;", a, qColumns[i], whereClause)
				sqlLog = fmt.Sprintf("INSERT INTO editLog (columnEdit, oldDataColumn, newDataColumn, timeEdited) VALUES ('%s', %s, '%v', %d);", qColumns[i], wValues[i], newStr[i], time.Now().Unix())
			}
			_, err := globalPool.Exec(context.Background(), sqlUpdate, newStr[i])
			if err != nil {
				log.Println(err)
			}
			_, err = globalPool.Exec(context.Background(), sqlLog)
			if err != nil {
				log.Println(err)
			}
		}
	}

	if !rows.Next() && len(edited) == 0 {
		_, err := globalPool.Exec(context.Background(), e)
		if err != nil {
			log.Println(err)
		}
	}
}
