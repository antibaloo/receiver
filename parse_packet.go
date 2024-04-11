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
func (g galileoParsePacket) Save(db *sqlx.DB) error {

	//result, err := json.MarshalIndent(g, " ", " ")
	result, err := json.Marshal(g)
	if err != nil {
		return fmt.Errorf("ошибка парсинга данных: %v", err)
	}

	// check db
	err = db.Ping()
	if err != nil {
		logger.Fatalf("Соединение с БД не прошло проверку: ", err)
		return err
	}

	_, err = db.Exec(`INSERT INTO "records" ("recvtime", "termtime", "termnumber", "recnumber", "record", "sent") VALUES ($1, $2, $3, $4, $5, $6)`, time.Unix(g.ReceivedTimestamp, 0), time.Unix(g.NavigationTimestamp, 0), g.TerminalNumber, g.PacketID, string(result), g.Sent2Zabbix)
	if err != nil {
		return err
	} else {
		logger.Infof("Архивная запись %d с терминала %d записана в лог.", g.PacketID, g.TerminalNumber)
	}
	//logger.Infof("Соединение с БД подтверждено!")
	//fmt.Println(string(result))
	return err
}
