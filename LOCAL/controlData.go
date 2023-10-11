package LOCAL

import (
	"fmt"
	"github.com/tidwall/gjson"
	"strings"
	"yandex-cloud-data/LOCAL/pqdriver"
)

func ControlData(data []byte) {
	type Field struct {
		DataName   string
		Type       string
		ColumnName string
	}

	type Entity struct {
		Name   string
		Table  string
		Fields []Field
	}

	entities := []Entity{
		{
			Name:  "disks",
			Table: "disksInfo",
			Fields: []Field{
				{"id", "str", "diskId"},
				{"folderId", "str", "folderId"},
				{"typeId", "str", "typeId"},
				{"zoneId", "str", "zoneId"},
				{"size", "int", "sizeDisk"},
			},
		},
		{
			Name:  "instances",
			Table: "vmInfo",
			Fields: []Field{
				{"id", "str", "machineId"},
				{"folderId", "str", "folderId"},
				{"description", "str", "description"},
				{"zoneId", "str", "zoneId"},
				{"bootDisk.diskId", "str", "bootDiskId"},
				{"serviceAccountId", "str", "serviceAccountId"},
				{"resources.memory", "int", "ram"},
				{"resources.cores", "int", "coresCount"},
				{"name", "str", "nameMachine"},
				{"platformId", "str", "platformId"},
			},
		},
		{
			Name:  "serviceAccounts",
			Table: "serviceAccsInfo",
			Fields: []Field{
				{"id", "str", "serviceAccId"},
				{"folderId", "str", "folderId"},
				{"name", "str", "nameSAcc"},
				{"description", "str", "description"},
			},
		},
	}

	for _, entity := range entities {
		dataType := gjson.GetBytes(data, entity.Name).String()
		if dataType == "" {
			continue
		}

		arrayCount := gjson.GetBytes(data, entity.Name+".#").Int()
		for c := 0; c < int(arrayCount); c++ {
			var columns []string
			var values []string
			var types []string

			for _, field := range entity.Fields {
				if field.DataName == "-" {
					continue
				}
				dataPath := fmt.Sprintf("%s.%d.%s", entity.Name, c, field.DataName)
				value := gjson.GetBytes(data, dataPath).String()

				if field.Type == "int" {
					values = append(values, value)
				} else if field.Type == "str" {
					values = append(values, fmt.Sprintf("'%s'", value))
				} else {
					continue
				}
				columns = append(columns, field.ColumnName)
				types = append(types, field.Type)
			}

			sqlColumns := fmt.Sprintf("(%s)", strings.Join(columns, ", "))
			sqlData := fmt.Sprintf("(%s)", strings.Join(values, ", "))
			sqlExec := fmt.Sprintf("INSERT INTO %s %s VALUES %s;", entity.Table, sqlColumns, sqlData)

			pqdriver.DataProc(entity.Table, sqlColumns, sqlData, types, sqlExec)
		}
		break
	}
}
