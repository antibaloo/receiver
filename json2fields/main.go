package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Record struct {
	Id         int64     `db:"id"`
	Recvtime   time.Time `db:"recvtime"`
	Termtime   time.Time `db:"termtime"`
	Termnumber int64     `db:"termnumber"`
	Recnumber  int64     `db:"recnumber"`
	Record     string    `db:"record"`
	Export     bool      `db:"export"`
	Sent       bool      `db:"sent"`
}
type RecordJson struct {
	HwVer               uint8     `json:"hwVer"`
	SwVer               uint8     `json:"swVer"`
	IMEI                string    `json:"IMEI"`
	Delay               int64     `json:"delay"`
	Latitude            float64   `json:"latitude"`
	Longitude           float64   `json:"longitude"`
	Speed               uint16    `json:"speed"`
	Height              int       `json:"height"`
	Pdop                uint16    `json:"pdop"`
	Hdop                uint16    `json:"hdop"`
	Vdop                uint16    `json:"vdop"`
	Nsat                uint8     `json:"nsat"`
	Ns                  uint16    `json:"ns"`
	Course              uint8     `json:"course"`
	TerminalStatus      string    `json:"terminalstatus"`
	VoltagePower        uint16    `json:"voltagepower"`
	VoltageBattery      uint16    `json:"voltagebattery"`
	TerminalTemperature int8      `json:"terminaltemperature"`
	OutputStatus        string    `json:"outputstatus"`
	InputStatus         string    `json:"inputstatus"`
	Can_a0              uint32    `json:"can_a0"`
	FuelLevel           uint8     `json:"fuellevel"`
	Cooltemp            int8      `json:"cooltemp"`
	Rpm                 uint16    `json:"rpm"`
	Can_b0              uint32    `json:"can_b0"`
	Can8bitr0           uint8     `json:"can8bitr0"`
	Can8bitr1           uint8     `json:"can8bitr1"`
	Can8bitr2           uint8     `json:"can8bitr2"`
	Can8bitr3           uint8     `json:"can8bitr3"`
	Can8bitr4           uint8     `json:"can8bitr4"`
	Can8bitr5           uint8     `json:"can8bitr5"`
	Can8bitr6           uint8     `json:"can8bitr6"`
	Can8bitr7           int8      `json:"can8bitr7"`
	Can8bitr8           uint8     `json:"can8bitr8"`
	Can8bitr9           uint8     `json:"can8bitr9"`
	Can8bitr10          uint8     `json:"can8bitr10"`
	Can8bitr11          uint8     `json:"can8bitr11"`
	Can8bitr12          uint8     `json:"can8bitr12"`
	Can8bitr13          uint8     `json:"can8bitr13"`
	Can8bitr14          string    `json:"can8bitr14"`
	Can8bitr15          uint8     `json:"can8bitr15"`
	Can8bitr27          string    `json:"can8bitr27"`
	Can8bitr28          string    `json:"can8bitr28"`
	Can8bitr29          string    `json:"can8bitr29"`
	Can8bitr30          string    `json:"can8bitr30"`
	Can16bitr0          uint16    `json:"can16bitr0"`
	Can16bitr1          uint16    `json:"can16bitr1"`
	Can16bitr2          uint16    `json:"can16bitr2"`
	Can16bitr3          uint16    `json:"can16bitr3"`
	Can16bitr4          uint16    `json:"can16bitr4"`
	Can16bitr5          float32   `json:"can16bitr5"`
	Can16bitr6          float32   `json:"can16bitr6"`
	Can16bitr7          uint16    `json:"can16bitr7"`
	Can16bitr8          uint16    `json:"can16bitr8"`
	Can16bitr9          uint16    `json:"can16bitr9"`
	Can16bitr10         uint16    `json:"can16bitr10"`
	Can16bitr11         float32   `json:"can16bitr11"`
	Can16bitr12         uint16    `json:"can16bitr12"`
	Can32bitr0          uint32    `json:"can32bitr0"`
	Can32bitr1          uint32    `json:"can32bitr1"`
	UserTag             [8]uint32 `json:"usertag"`
}

var schema = `
CREATE TABLE records4lens (
    pid bigserial primary key,
	recvtime timestamptz,
	termtime timestamptz,
	termnumber bigserial,
	recnumber bigserial,
    latitude real,
	longitude real,
    speed smallserial,
	height integer,
	pdop smallserial,
	hdop smallserial,
	vdop smallserial,
	nsat smallserial,
	ns smallserial,
	course smallserial,
	terminalstatus text,
	voltagepower smallserial,
	voltagebattery smallserial,
	terminaltemperature smallint,
	outputstatus text,
	inputstatus text,
	cana0 serial,
	fuellevel smallserial,
	cooltemp smallint,
	rpm smallserial,
	canb0 serial,
	can8bitr0 smallserial,
	can8bitr1 smallserial,
	can8bitr2 smallserial,
	can8bitr3 smallserial,
	can8bitr4 smallserial,
	can8bitr5 smallserial,
	can8bitr6 smallserial,
	can8bitr7 smallint,
	can8bitr8 smallserial,
	can8bitr9 smallserial,
	can8bitr10 smallserial,
	can8bitr11 smallserial,
	can8bitr12 smallserial,
	can8bitr13 smallserial,
	can8bitr14 text,
	can8bitr15 smallserial,
	can8bitr27 text,
	can8bitr28 text,
	can8bitr29 text,
	can8bitr30 text,
	can16bitr0 smallserial,
	can16bitr1 smallserial,
	can16bitr2 smallserial,
	can16bitr3 smallserial,
	can16bitr4 smallserial,
	can16bitr5 real,
	can16bitr6 real,
	can16bitr7 smallserial,
	can16bitr8 smallserial,
	can16bitr9 smallserial,
	can16bitr10 smallserial,
	can16bitr11 real,
	can16bitr12 smallserial,
	can32bitr0 serial,
	can32bitr1 serial
);`

func main() {
	start := time.Now()
	records := []Record{}
	connectionstring := "postgres://postgres:gKwRprQmD9iqLi85r@localhost:5432/galileodb?sslmode=disable"
	// open database
	db, err := sqlx.Open("postgres", connectionstring)
	if err != nil {
		fmt.Println("Не удалось установить соединение с БД:", err)
		return
	}
	// close database
	defer db.Close()
	// check db
	err = db.Ping()
	if err != nil {
		fmt.Println("Соединение с БД не прошло проверку:", err)
		return
	}
	fmt.Println("Соединение с БД установлено!")
	//db.MustExec(schema)
	err = db.Select(&records, "SELECT * FROM records WHERE termnumber = 10001 AND export = false ORDER BY termtime DESC LIMIT 10000")

	if err != nil {
		fmt.Println(err)
		return
	}

	//fmt.Println(records)

	data := RecordJson{}
	for _, record := range records {
		json.Unmarshal([]byte(record.Record), &data)
		//fmt.Println(data)
		_, err = db.Exec(
			`INSERT INTO records4lens (
				"recvtime", "termtime", "termnumber", "recnumber","latitude","longitude","speed","height","pdop",
				"hdop","vdop","nsat","ns","course","terminalstatus", "voltagepower","voltagebattery","terminaltemperature",
				"outputstatus", "inputstatus", "cana0", "fuellevel", "cooltemp", "rpm", "canb0", "can8bitr0", "can8bitr1",
				"can8bitr2", "can8bitr3", "can8bitr4", "can8bitr5", "can8bitr6", "can8bitr7", "can8bitr8", "can8bitr9",
				"can8bitr10", "can8bitr11", "can8bitr12", "can8bitr13", "can8bitr14", "can8bitr15", "can8bitr27", "can8bitr28",
				"can8bitr29", "can8bitr30", "can16bitr0", "can16bitr1", "can16bitr2", "can16bitr3", "can16bitr4", "can16bitr5",
				"can16bitr6", "can16bitr7", "can16bitr8", "can16bitr9", "can16bitr10", "can16bitr11", "can16bitr12", 
				"can32bitr0", "can32bitr1"
			)
			VALUES (
				$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23,
				$24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45,
				$46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60
			)`,
			record.Recvtime,
			record.Termtime,
			record.Termnumber,
			record.Recnumber,
			data.Latitude,
			data.Longitude,
			data.Speed,
			data.Height,
			data.Pdop,
			data.Hdop,
			data.Vdop,
			data.Nsat,
			data.Ns,
			data.Course,
			data.TerminalStatus,
			data.VoltagePower,
			data.VoltageBattery,
			data.TerminalTemperature,
			data.OutputStatus,
			data.InputStatus,
			data.Can_a0,
			data.FuelLevel,
			data.Cooltemp,
			data.Rpm,
			data.Can_b0,
			data.Can8bitr0,
			data.Can8bitr1,
			data.Can8bitr2,
			data.Can8bitr3,
			data.Can8bitr4,
			data.Can8bitr5,
			data.Can8bitr6,
			data.Can8bitr7,
			data.Can8bitr8,
			data.Can8bitr9,
			data.Can8bitr10,
			data.Can8bitr11,
			data.Can8bitr12,
			data.Can8bitr13,
			data.Can8bitr14,
			data.Can8bitr15,
			data.Can8bitr27,
			data.Can8bitr28,
			data.Can8bitr29,
			data.Can8bitr30,
			data.Can16bitr0,
			data.Can16bitr1,
			data.Can16bitr2,
			data.Can16bitr3,
			data.Can16bitr4,
			data.Can16bitr5,
			data.Can16bitr6,
			data.Can16bitr7,
			data.Can16bitr8,
			data.Can16bitr9,
			data.Can16bitr10,
			data.Can16bitr11,
			data.Can16bitr12,
			data.Can32bitr0,
			data.Can32bitr1,
		)
		if err != nil {
			fmt.Println(err)
		} else {
			_, errUpdate := db.Exec(`UPDATE records SET export = true WHERE id = $1`, record.Id)
			if errUpdate != nil {
				fmt.Println(errUpdate)
			}
		}
	}
	stop := time.Now()
	fmt.Println("Отработано за ", stop.Sub(start))
}
