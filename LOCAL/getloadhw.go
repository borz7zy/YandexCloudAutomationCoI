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
	"time"
)

func LoadHW() {
	settings, err := ioutil.ReadFile("settings.json")
	if err != nil {
		log.Fatalf("Failed to read settings.json: %v\n", err)
	}

	user := gjson.GetBytes(settings, "pgUser").String()
	password := gjson.GetBytes(settings, "pgPassword").String()
	port := gjson.GetBytes(settings, "pgPort").String()
	server := gjson.GetBytes(settings, "pgServer").String()
	database := gjson.GetBytes(settings, "pgDataBase").String()

	connStr := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", user, password, server, port, database)
	dbpool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v\n", err)
	}
	defer dbpool.Close()

	before, err := cpu.Get()
	if err != nil {
		log.Fatalf("Failed to get initial CPU stats: %v\n", err)
	}

	time.Sleep(1 * time.Second)

	after, err := cpu.Get()
	if err != nil {
		log.Fatalf("Failed to get subsequent CPU stats: %v\n", err)
	}

	ramGet, err := memory.Get()
	if err != nil {
		log.Fatalf("Failed to get ramGet stats: %v\n", err)
	}

	total := float64(after.Total - before.Total)
	cpuUsage := int64(100.0 * (1.0 - (float64(after.Idle-before.Idle) / total)))

	sqlInsert := fmt.Sprintf("INSERT INTO vmAvgLoadHost (cpu, ram) VALUES (%d, %d);", cpuUsage, ramGet.Used)
	_, err = dbpool.Exec(context.Background(), sqlInsert)
	if err != nil {
		log.Println(err)
	}
}
