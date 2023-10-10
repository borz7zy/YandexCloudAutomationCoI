package LOCAL

import (
	"fmt"
	"github.com/tidwall/gjson"
	"strconv"
	"strings"
	"yandex-cloud-data/LOCAL/pqdriver"
)

func ControlData(data []byte) {
	/*

		Тут думаю надо объяснить как работает автоматизация.
		Заполняются три слайса, в первом название строки пришедшей в JSON и название таблицы в БД.
		Второй слайс содержит в себе путь до поля в JSON и название поля в БД куда это будет сохранено.
		Третий указывает тип данных полей в БД.

		В каждом слайсе есть разделитель "|", который разделяет названия различных полей.
		В слайсе нейма полей так же есть разделитель "*", который разделяет два поля - первое это путь до поля в JSON,
		а второе же поле название в таблице базы данных.
		Так же есть "-", которое просто определяет прерывание выполнения цикла если данный тип данных не был заполнен.

	*/
	sliceTypeNames := []string{
		"disks|disksInfo",                 //1
		"instances|vmInfo",                //2
		"serviceAccounts|serviceAccsInfo", //3
	}
	sliceDataNames := []string{ //1  (1*1|2*2|3*3)
		"id*diskId|id*machineId|id*serviceAccId",
		"folderId*folderId|folderId*folderId|folderId*folderId",
		"typeId*typeId|description*description|name*nameSAcc",
		"zoneId*zoneId|zoneId*zoneId|description*description",
		"size*sizeDisk|bootDisk.diskId*bootDiskId|-*-",
		"-*-|serviceAccountId*serviceAccountId|-*-",
		"-*-|resources.memory*ram|-*-",
		"-*-|resources.cores*coresCount|-*-",
		"-*-|name*nameMachine|-*-",
		"-*-|platformId*platformId|-*-",
	}
	sliceDataTypes := []string{ // ( 1 | 2 | 3 )
		"str|str|str",
		"str|str|str",
		"str|str|str",
		"str|str|str",
		"int|str|-",
		"-|str|-",
		"-|int|-",
		"-|int|-",
		"-|str|-",
		"-|str|-",
	}

	for i := 0; i < len(sliceTypeNames); i++ {
		splitTypes := strings.Split(sliceTypeNames[i], "|")
		dataType := fmt.Sprintf("%s", gjson.Get(string(data), splitTypes[0]))
		if dataType != "" {
			typeGetArray := splitTypes[0] + ".#"
			arrayCountGet := fmt.Sprintf("%s", gjson.Get(string(data), typeGetArray))
			convToInt, _ := strconv.Atoi(arrayCountGet)
			var dataGet string //для записи строки
			for c := 0; c < convToInt; c++ {
				for d := 0; d < len(sliceDataNames); d++ {
					sliceDataNamesSplit := strings.Split(sliceDataNames[d], "|")
					twoDataNamesSplit := strings.Split(sliceDataNamesSplit[i], "*")
					getDataFromJSON := fmt.Sprintf("%s.%d.%s", splitTypes[0], c, twoDataNamesSplit[0])
					if twoDataNamesSplit[0] != "-" {
						if d == 0 {
							dataGet = getDataFromJSON
						} else {
							dataGet = dataGet + "|" + getDataFromJSON
						}
					} else {
						break
					}
				}
				//generating a database query
				sliceDataGet := strings.Split(dataGet, "|")

				var sqlColumns string
				var sqlData string
				types := []string{}
				for c := 0; c < len(sliceDataGet); c++ {
					sliceDataNamesSplit := strings.Split(sliceDataNames[c], "|")
					twoDataNamesSplit := strings.Split(sliceDataNamesSplit[i], "*")
					sliceDataTypesSplit := strings.Split(sliceDataTypes[c], "|")
					types = append(types, sliceDataTypesSplit[i])

					if c == 0 {
						sqlColumns = "(" + twoDataNamesSplit[1]
						if sliceDataTypesSplit[i] == "str" {
							sqlData = "('" + fmt.Sprintf("%s", gjson.Get(string(data), sliceDataGet[c])) + "'"
						} else if sliceDataTypesSplit[i] == "int" {
							sqlData = "(" + fmt.Sprintf("%s", gjson.Get(string(data), sliceDataGet[c]))
						} else if sliceDataTypesSplit[i] == "-" {
							continue
						}
					} else {
						sqlColumns = sqlColumns + ", " + twoDataNamesSplit[1]
						if sliceDataTypesSplit[i] == "str" {
							sqlData = sqlData + ", '" + fmt.Sprintf("%s", gjson.Get(string(data), sliceDataGet[c])) + "'"
						} else if sliceDataTypesSplit[i] == "int" {
							sqlData = sqlData + ", " + fmt.Sprintf("%s", gjson.Get(string(data), sliceDataGet[c]))
						} else if sliceDataTypesSplit[i] == "-" {
							continue
						}
					}
				}
				sqlColumns = fmt.Sprintf("%s)", sqlColumns)
				sqlData = fmt.Sprintf("%s)", sqlData)
				sqlExec := fmt.Sprintf("INSERT INTO %s %s VALUES %s;", splitTypes[1], sqlColumns, sqlData)
				pqdriver.DataProc(splitTypes[1], sqlColumns, sqlData, types, sqlExec)
			}

			break
		} else {
			continue
		}
	}
}
