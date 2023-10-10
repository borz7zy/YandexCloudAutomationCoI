package pqdriver

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type PgxId struct {
	id *pgx.Conn
}

type ConnectPool struct {
	dbpool *pgxpool.Pool
}

/*
db := InitDB()
// code
// ...
// end code
defer db.dbpool.Close()
*/
func InitDB() *ConnectPool {
	settings, _ := ioutil.ReadFile("settings.json")
	user := fmt.Sprintf("%s", gjson.Get(string(settings), "pgUser"))
	password := fmt.Sprintf("%s", gjson.Get(string(settings), "pgPassword"))
	port := fmt.Sprintf("%s", gjson.Get(string(settings), "pgPort"))
	server := fmt.Sprintf("%s", gjson.Get(string(settings), "pgServer"))
	database := fmt.Sprintf("%s", gjson.Get(string(settings), "pgDataBase"))

	dbpool, err := pgxpool.New(context.Background(), fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", user, password, server, port, database))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}

	return &ConnectPool{dbpool: dbpool}
}

func DataProc(a string, q string, w string, t []string, e string) {
	db := InitDB()

	qSplit := strings.Split(q, ", ")
	qSplit[0] = strings.ReplaceAll(qSplit[0], "(", "")
	wSplit := strings.Split(w, ", ")
	wSplit[0] = strings.ReplaceAll(wSplit[0], "(", "")
	var sqlOne string
	for i := 0; i < len(qSplit); i++ {
		str := strings.ReplaceAll(qSplit[i], "(", "")
		str = strings.ReplaceAll(str, "'", "")
		str = strings.ReplaceAll(str, ")", "")
		if i >= 0 && i < len(qSplit)-1 {
			sqlOne = sqlOne + str + ", "
		} else {
			sqlOne = sqlOne + str
		}
	}
	sqlExec := "SELECT " + sqlOne + " FROM " + a + " WHERE " + qSplit[0] + " = " + wSplit[0] + ";"
	rows, _ := db.dbpool.Query(context.Background(), sqlExec)
	defer rows.Close()

	columnsNameSplit := strings.Split(sqlOne, ", ")

	wSplitClear := []string{}
	for i := 0; i < len(wSplit); i++ {
		str := strings.ReplaceAll(wSplit[i], "(", "")
		str = strings.ReplaceAll(str, "'", "")
		str = strings.ReplaceAll(str, ")", "")
		wSplitClear = append(wSplitClear, str)
	}

	edited := []bool{}
	newStr := []any{}
	for rows.Next() {
		varValues, _ := rows.Values()
		for i := 0; i < len(varValues); i++ {
			if t[i] == "int" {
				nt, _ := strconv.ParseInt(wSplitClear[i], 10, 64)
				if varValues[i] != nt {
					edited = append(edited, true)
					newStr = append(newStr, varValues[i])
				} else {
					edited = append(edited, false)
					newStr = append(newStr, 0)
				}
			} else if t[i] == "str" {
				if varValues[i] != wSplitClear[i] {
					edited = append(edited, true)
					newStr = append(newStr, varValues[i])
				} else {
					edited = append(edited, false)
					newStr = append(newStr, 0)
				}
			}
		}
	}
	var ii int
	for i := 0; i < len(edited); i++ { // update column and add log edit
		if edited[i] != false && newStr[i] != 0 {
			//sqlUpdate := "UPDATE " + a + " SET " + columnsNameSplit[i] + " = " + newStr[i] //UPDATE table_name SET column_name = $1 WHERE condition = $2
			var sqlUpdate string
			if t[i] == "int" {
				sqlUpdate = fmt.Sprintf("UPDATE %s SET %s = %d WHERE %s = %s;", a, columnsNameSplit[i], newStr[i], qSplit[0], wSplit[0])
			} else if t[i] == "str" {
				sqlUpdate = fmt.Sprintf("UPDATE %s SET %s = '%s' WHERE %s = %s;", a, columnsNameSplit[i], newStr[i], qSplit[0], wSplit[0])
			}
			_, err := db.dbpool.Exec(context.Background(), sqlUpdate)
			if err != nil {
				log.Println(err)
			}
			if t[i] == "int" {
				nt, _ := strconv.ParseInt(wSplitClear[i], 10, 64)
				sqlUpdate = fmt.Sprintf("INSERT INTO editLog (columnEdit, oldDataColumn, newDataColumn, timeEdited) VALUES ('%s', %d, %d, %d);", columnsNameSplit[i], nt, newStr[i], time.Now().Unix())
			} else if t[i] == "str" {
				sqlUpdate = fmt.Sprintf("INSERT INTO editLog (columnEdit, oldDataColumn, newDataColumn, timeEdited) VALUES ('%s', '%s', '%s', %d);", columnsNameSplit[i], wSplitClear[i], newStr[i], time.Now().Unix())
			}
			_, err = db.dbpool.Exec(context.Background(), sqlUpdate)
			if err != nil {
				log.Println(err)
			}
			ii++
		}
	}

	if rows.Next() == false && ii == 0 { //insert
		_, err := db.dbpool.Exec(context.Background(), e)
		if err != nil {
			log.Println(err)
		}
	}
	//always from below
	defer db.dbpool.Close()
}
