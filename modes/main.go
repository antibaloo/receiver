package main

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Record struct {
	Termtime     time.Time `db:"termtime"`
	VoltagePower uint16    `db:"voltagepower"`
	EnBatPower   uint8     `db:"enbatpower"`
	EnRPM        uint16    `db:"enrpm"`
	RotorRPM     uint16    `db:"rotorrpm"`
	FeedForce    uint8     `db:"feedforce"`
	FluidFlow    uint8     `db:"fluidflow"`
	DrillMode    uint8     `db:"drillmode"`
	Depth        uint16    `db:"depth"`
	EnTemp       uint8     `db:"entemp"`
	Recnumber    int64     `db:"recnumber"`
}

type EngineMode struct {
	RecNumBegin, RecNumEnd int64
	Begin, End             time.Time // Время начала и окончания режима
	Mode                   int       // Название режима
}

type DrillMode struct {
	RecNumBegin, RecNumEnd int64
	Begin, End             time.Time // Время начала и окончания режима
	Mode                   int       // Название режима
}

var curEngineMode int
var curDrillMode int
var EngineMode4Save EngineMode
var DrillMode4Save DrillMode
var enModCounter int
var drModCounter int
var modes int

//var schemaEngineMods = `CREATE TABLE engineMods (eid bigserial primary key, beginMod timestamptz, endMod timestamptz, mode serial, count serial);`

//var schemaDrillMods = `CREATE TABLE drillMods (did bigserial primary key, beginMod timestamptz, endMod timestamptz, mode serial, count serial);`

func main() {
	start := time.Now()
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
	//db.MustExec(schemaEngineMods)
	//db.MustExec(schemaDrillMods)
	records := []Record{}
	err = db.Select(&records, "SELECT recnumber, termtime, voltagepower, can8bitr8 AS enbatpower, can16bitr4 AS enrpm, can16bitr0 AS rotorrpm, can8bitr0 AS feedforce, can8bitr3 AS fluidflow, can8bitr12 AS drillmode, can16bitr10 AS depth, can8bitr7 AS entemp FROM records4lens WHERE termnumber = 10001 ORDER BY termtime ASC LIMIT 500000")
	if err != nil {
		fmt.Println(err)
		return
	}
	for _, record := range records {
		if record.Recnumber == 0 {
			continue
		}
		if record.VoltagePower < 10000 {
			curEngineMode = 1
		} else if record.VoltagePower > 20000 && record.EnBatPower == 0 {
			curEngineMode = 2
		} else if record.VoltagePower > 20000 && record.EnBatPower > 0 && record.EnRPM == 0 {
			curEngineMode = 3
		} else if record.EnRPM > 150 && record.EnTemp < 70 {
			curEngineMode = 4
		} else if record.EnRPM > 150 && record.EnTemp >= 70 {
			curEngineMode = 5
		}

		if curEngineMode == EngineMode4Save.Mode {
			EngineMode4Save.End = record.Termtime
			EngineMode4Save.RecNumEnd = record.Recnumber
			enModCounter++
		} else {
			if EngineMode4Save.Mode != 0 {
				if enModCounter > 1 {
					fmt.Println(EngineMode4Save, enModCounter)
					_, err = db.Exec(`INSERT INTO enginemods ("beginmod", "endmod", "mode", "count") VALUES ($1,$2,$3, $4)`,
						EngineMode4Save.Begin,
						EngineMode4Save.End,
						EngineMode4Save.Mode,
						enModCounter,
					)
					if err != nil {
						fmt.Println(err)
					}
					modes++
				}
				enModCounter = 0
			}
			enModCounter++
			EngineMode4Save.Begin = record.Termtime
			EngineMode4Save.End = record.Termtime
			EngineMode4Save.RecNumBegin = record.Recnumber
			EngineMode4Save.RecNumEnd = record.Recnumber
			EngineMode4Save.Mode = curEngineMode
		}
	}

	fmt.Println(EngineMode4Save, enModCounter)
	_, err = db.Exec(`INSERT INTO enginemods ("beginmod", "endmod", "mode", "count") VALUES ($1,$2,$3, $4)`,
		EngineMode4Save.Begin,
		EngineMode4Save.End,
		EngineMode4Save.Mode,
		enModCounter,
	)
	if err != nil {
		fmt.Println(err)
	}

	modes++
	fmt.Println("Mods count", modes)
	modes = 0
	for _, record := range records {
		if record.Recnumber == 0 {
			continue
		}
		if record.RotorRPM > 0 && record.FeedForce > 3 {
			curDrillMode = 1
		} else if record.DrillMode == 6 && record.EnRPM > 150 {
			curDrillMode = 5
		} else if record.Depth > 1 {
			curDrillMode = 4
		} else if record.RotorRPM == 0 && record.FluidFlow > 2 {
			curDrillMode = 2
		} else if record.DrillMode == 1 || record.DrillMode == 2 || record.DrillMode == 4 || record.DrillMode == 5 {
			curDrillMode = 3
		} else {
			curDrillMode = 0
		}
		if curDrillMode == DrillMode4Save.Mode {
			DrillMode4Save.End = record.Termtime
			DrillMode4Save.RecNumEnd = record.Recnumber
			drModCounter++
		} else {
			if DrillMode4Save.Mode != 0 {
				if drModCounter > 1 {
					fmt.Println(DrillMode4Save, drModCounter)
					_, err = db.Exec(`INSERT INTO drillmods ("beginmod", "endmod", "mode", "count") VALUES ($1,$2,$3, $4)`,
						DrillMode4Save.Begin,
						DrillMode4Save.End,
						DrillMode4Save.Mode,
						drModCounter,
					)
					if err != nil {
						fmt.Println(err)
					}
					modes++
				}
				drModCounter = 0
			}
			drModCounter++
			DrillMode4Save.Begin = record.Termtime
			DrillMode4Save.End = record.Termtime
			DrillMode4Save.RecNumBegin = record.Recnumber
			DrillMode4Save.RecNumEnd = record.Recnumber
			DrillMode4Save.Mode = curDrillMode
		}

	}
	fmt.Println(DrillMode4Save, drModCounter)
	_, err = db.Exec(`INSERT INTO drillmods ("beginmod", "endmod", "mode", "count") VALUES ($1,$2,$3, $4)`,
		DrillMode4Save.Begin,
		DrillMode4Save.End,
		DrillMode4Save.Mode,
		drModCounter,
	)
	if err != nil {
		fmt.Println(err)
	}
	modes++
	fmt.Println("Mods count", modes)
	//fmt.Println(records)
	stop := time.Now()
	fmt.Println("Отработано за ", stop.Sub(start))
}
