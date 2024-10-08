package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/spetr/go-zabbix-sender"
)

type galileoParsePacket struct {
	Sent2Zabbix         bool      `json:"-"`
	TerminalNumber      uint32    `json:"-"`
	HwVer               uint8     `json:"hwVer"`
	SwVer               uint8     `json:"swVer"`
	IMEI                string    `json:"IMEI"`
	PacketID            uint32    `json:"-"`
	NavigationTimestamp int64     `json:"-"`
	Milliseconds        uint16    `json:"-"`
	ReceivedTimestamp   int64     `json:"-"`
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
	Can8bitr16          uint8     `json:"can8bitr16"`
	Can8bitr17          uint8     `json:"can8bitr17"`
	Can8bitr26          string    `json:"can8bitr26"`
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
	Can32bitr2          uint32    `json:"can32bitr2"`
	Can32bitr3          uint32    `json:"can32bitr3"`
	Can32bitr4          uint32    `json:"can32bitr4"`
	Can32bitr5          uint32    `json:"can32bitr5"`
	Can32bitr6          uint32    `json:"can32bitr6"`
	UserTag             [8]uint32 `json:"usertag"`
}

func (g galileoParsePacket) Send(zabbixHost string) error {
	logger.Info("Начинаем отправку в zabbix")
	result, err := json.Marshal(g)
	if err != nil {
		return fmt.Errorf("ошибка парсинга данных: %v", err)
	}
	logger.Info("Формируем метрики")

	metric := []*zabbix.Metric{zabbix.NewMetric(fmt.Sprint(g.TerminalNumber), "galileoRecord", string(result), true, g.NavigationTimestamp)}
	logger.Info("инициализируем сендер")

	z := zabbix.NewSender(zabbixHost)

	logger.Info("отправляем метрики")

	resActive, errActive, _, _ := z.SendMetrics(metric)
	logger.Info("Читаем результат")
	logger.Infof("результат %v", resActive)
	logger.Infof("ошибка: %s", errActive)
	if errActive != nil {
		return errActive
	}
	sent, _ := resActive.GetInfo()
	logger.Infof("Отправка в zabbix - %s записи %d. Всего отправлено: %d, обработано: %d. Ошибка: %v", zabbixHost, g.PacketID, sent.Total, sent.Processed, errActive)
	if sent.Processed != sent.Total {
		return fmt.Errorf("запись не отправлена")
	}
	return nil
}
func (g galileoParsePacket) Save(db *sqlx.DB) (int64, error) {

	//result, err := json.MarshalIndent(g, " ", " ")
	result, err := json.Marshal(g)
	if err != nil {
		return 0, fmt.Errorf("ошибка парсинга данных: %v", err)
	}

	// check db
	err = db.Ping()
	if err != nil {
		logger.Fatalf("Соединение с БД не прошло проверку: ", err)
		return 0, err
	}
	var id int64
	if err = db.QueryRow(`INSERT INTO "records" ("recvtime", "termtime", "termnumber", "recnumber", "record", "sent") VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`, time.Unix(g.ReceivedTimestamp, 0), time.Unix(g.NavigationTimestamp, 0), g.TerminalNumber, g.PacketID, string(result), g.Sent2Zabbix).Scan(&id); err != nil {
		return 0, err
	} else {
		logger.Infof("Архивная запись %d с терминала %d записана в лог.", g.PacketID, g.TerminalNumber)
		return id, nil
	}
}

func (g galileoParsePacket) noJSONSave(db *sqlx.DB, id int64) error {
	if id == 0 {
		return fmt.Errorf("при записи в архив не определен идентификатор")
	}
	err := db.Ping()
	if err != nil {
		logger.Fatalf("Соединение с БД не прошло проверку: ", err)
		return err
	}
	_, err = db.Exec(`INSERT INTO records4lens (
		"recvtime", "termtime", "termnumber", "recnumber","latitude","longitude",
		"speed","height","pdop", "hdop","vdop","nsat","ns","course","terminalstatus", "voltagepower","voltagebattery",
		"terminaltemperature", "outputstatus", "inputstatus", "cana0", "fuellevel", "cooltemp", "rpm", "canb0", 
		"can8bitr0", "can8bitr1", "can8bitr2", "can8bitr3", "can8bitr4", "can8bitr5", "can8bitr6", "can8bitr7", "can8bitr8",
		"can8bitr9", "can8bitr10", "can8bitr11", "can8bitr12", "can8bitr13", "can8bitr14", "can8bitr15",
		"can8bitr27", "can8bitr28",	"can8bitr29", "can8bitr30", 
		"can16bitr0", "can16bitr1", "can16bitr2", "can16bitr3", "can16bitr4", "can16bitr5",	"can16bitr6", "can16bitr7", 
		"can16bitr8", "can16bitr9", "can16bitr10", "can16bitr11", "can16bitr12", 
		"can32bitr0", "can32bitr1"
	)VALUES (
		$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23,
		$24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45,
		$46, $47, $48, $49, $50, $51, $52, $53, $54, $55, $56, $57, $58, $59, $60
	)`,
		time.Unix(g.ReceivedTimestamp, 0), time.Unix(g.NavigationTimestamp, 0), g.TerminalNumber, g.PacketID, g.Latitude, g.Longitude,
		g.Speed, g.Height, g.Pdop, g.Hdop, g.Vdop, g.Nsat, g.Ns, g.Course, g.TerminalStatus, g.VoltagePower, g.VoltageBattery,
		g.TerminalTemperature, g.OutputStatus, g.InputStatus, g.Can_a0, g.FuelLevel, g.Cooltemp, g.Rpm, g.Can_b0,
		g.Can8bitr0, g.Can8bitr1, g.Can8bitr2, g.Can8bitr3, g.Can8bitr4, g.Can8bitr5, g.Can8bitr6, g.Can8bitr7, g.Can8bitr8,
		g.Can8bitr9, g.Can8bitr10, g.Can8bitr11, g.Can8bitr12, g.Can8bitr13, g.Can8bitr14, g.Can8bitr15,
		g.Can8bitr27, g.Can8bitr28, g.Can8bitr29, g.Can8bitr30,
		g.Can16bitr0, g.Can16bitr1, g.Can16bitr2, g.Can16bitr3, g.Can16bitr4, g.Can16bitr5, g.Can16bitr6, g.Can16bitr7,
		g.Can16bitr8, g.Can16bitr9, g.Can16bitr10, g.Can16bitr11, g.Can16bitr12,
		g.Can32bitr0, g.Can32bitr1,
	)
	if err != nil {
		return err
	} else {
		_, errUpdate := db.Exec(`UPDATE records SET export = true WHERE id = $1`, id)
		if errUpdate != nil {
			fmt.Println(errUpdate)
		}
		logger.Infof("Архивная запись %d с терминала %d без JSON записана в таблицу records4lens.", g.PacketID, g.TerminalNumber)
	}
	return err
}

/*
func (g galileoParsePacket) checkModes(db *sqlx.DB) error {
	if _, ok := EngineMode[g.TerminalNumber]; !ok {
		EngineMode[g.TerminalNumber] = &Mode{Mode: 0, Count: 0}
	}
	if _, ok := DrillMode[g.TerminalNumber]; !ok {
		DrillMode[g.TerminalNumber] = &Mode{Mode: 0, Count: 0}
	}
	if g.PacketID == 0 {
		return nil
	}
	err := db.Ping()
	if err != nil {
		logger.Fatalf("Соединение с БД не прошло проверку: ", err)
		return err
	}
	var curEngineMode int
	var curDrillMode int
	// Определение текущего режима двигателя
	if g.VoltagePower < 10000 {
		curEngineMode = 1
	} else if g.VoltagePower > 20000 && g.Can8bitr8 == 0 {
		curEngineMode = 2
	} else if g.VoltagePower > 20000 && g.Can8bitr8 > 0 && g.Can16bitr4 == 0 {
		curEngineMode = 3
	} else if g.Can16bitr4 > 150 && g.Can8bitr7 < 70 {
		curEngineMode = 4
	} else if g.Can16bitr4 > 150 && g.Can8bitr7 >= 70 {
		curEngineMode = 5
	}

	if curEngineMode == EngineMode[g.TerminalNumber].Mode {
		EngineMode[g.TerminalNumber].End = time.Unix(g.NavigationTimestamp, 0)
		EngineMode[g.TerminalNumber].Count++
	} else {
		if EngineMode[g.TerminalNumber].Mode != 0 {
			if EngineMode[g.TerminalNumber].Count > 1 {
				_, err = db.Exec(`INSERT INTO enginemods ("termnumber", "beginmod", "endmod", "mode", "count") VALUES ($1,$2,$3, $4)`,
					g.TerminalNumber,
					EngineMode[g.TerminalNumber].Begin,
					EngineMode[g.TerminalNumber].End,
					EngineMode[g.TerminalNumber].Mode,
					EngineMode[g.TerminalNumber].Count,
				)
			}
		}
		EngineMode[g.TerminalNumber].Count = 0
		EngineMode[g.TerminalNumber].Count++
		EngineMode[g.TerminalNumber].Begin = time.Unix(g.NavigationTimestamp, 0)
		EngineMode[g.TerminalNumber].End = time.Unix(g.NavigationTimestamp, 0)
		EngineMode[g.TerminalNumber].Mode = curEngineMode
	}

	// Определения текущего режима установки
	if g.Can16bitr0 > 0 && g.Can8bitr0 > 3 {
		curDrillMode = 1
	} else if g.Can8bitr12 == 6 && g.Can16bitr4 > 150 {
		curDrillMode = 5
	} else if g.Can16bitr10 > 1 {
		curDrillMode = 4
	} else if g.Can16bitr0 == 0 && g.Can8bitr3 > 2 {
		curDrillMode = 2
	} else if g.Can8bitr12 == 1 || g.Can8bitr12 == 2 || g.Can8bitr12 == 4 || g.Can8bitr12 == 5 {
		curDrillMode = 3
	} else {
		curDrillMode = 0
	}

	if curDrillMode == DrillMode[g.TerminalNumber].Mode {
		DrillMode[g.TerminalNumber].End = time.Unix(g.NavigationTimestamp, 0)
		DrillMode[g.TerminalNumber].Count++
	} else {
		if DrillMode[g.TerminalNumber].Mode != 0 {
			if DrillMode[g.TerminalNumber].Count > 1 {
				_, err = db.Exec(`INSERT INTO drillmods ("termnumber", "beginmod", "endmod", "mode", "count") VALUES ($1,$2,$3, $4)`,
					g.TerminalNumber,
					DrillMode[g.TerminalNumber].Begin,
					DrillMode[g.TerminalNumber].End,
					DrillMode[g.TerminalNumber].Mode,
					DrillMode[g.TerminalNumber].Count,
				)
			}

		}
		DrillMode[g.TerminalNumber].Count = 0
		DrillMode[g.TerminalNumber].Begin = time.Unix(g.NavigationTimestamp, 0)
		DrillMode[g.TerminalNumber].End = time.Unix(g.NavigationTimestamp, 0)
		DrillMode[g.TerminalNumber].Mode = curDrillMode
	}
	logger.Infof("Текущий режим работы двигателя: %v с терминала %d", EngineMode[g.TerminalNumber], g.TerminalNumber)
	logger.Infof("Текущий режим работы установаки: %v с терминала %d", DrillMode[g.TerminalNumber], g.TerminalNumber)
	return err
}*/
