package LOCAL

import (
	"C"
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mackerelio/go-osstat/cpu"
	"github.com/mackerelio/go-osstat/memory"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"os"
	"time"
)

func LoadHW() {
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

	before, _ := cpu.Get()
	time.Sleep(time.Duration(1) * time.Second)
	after, _ := cpu.Get()
	total := float64(after.Total - before.Total)
	//fmt.Printf("cpu load: %f %%\n", 100.0-float64(after.Idle-before.Idle)/total*100)
	memory, _ := memory.Get()
	//fmt.Printf("memory used: %d bytes\n", memory.Used)

	sqlInsert := fmt.Sprintf("INSERT INTO vmAvgLoadHost (cpu, ram) VALUES (%d, %d);", int64(100.0-float64(after.Idle-before.Idle)/total*100), memory.Used)
	_, err = dbpool.Exec(context.Background(), sqlInsert)
	if err != nil {
		log.Println(err)
	}
	//always from below
	defer dbpool.Close()
}
