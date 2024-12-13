package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"galileo"
	"io"
	"net"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

const headerLen = 3

func recoverAll() {
	if err := recover(); err != nil {
		logger.Info(err)
	}
}

func handleRecvPkg(conn net.Conn, ttl time.Duration) {
	var (
		recvPacket []byte
	)
	defer conn.Close()

	// Записывает в лог панику, если она приключится в результате работы метода
	defer recoverAll()

	//packet.GalileoPaket
	logger.Warnf("Установлено соединение с %s", conn.RemoteAddr())

	logger.Infof("Получена строка соединения с сервером БД: %s", config.getConnectionString())
	// open database
	db, err := sqlx.Open("postgres", config.getConnectionString())
	if err != nil {
		logger.Fatalf("Не удалось установить соединение с БД: ", err)
	}
	// close database
	defer db.Close()
	// check db
	err = db.Ping()
	if err != nil {
		logger.Fatalf("Соединение с БД не прошло проверку: ", err)
		return
	}
	logger.Infof("Соединение с БД установлено!")

	outPkg := galileoParsePacket{}
	for {
		connTimer := time.NewTimer(ttl)

		// считываем заголовок пакета
		headerBuf := make([]byte, headerLen)
		_, err := conn.Read(headerBuf)

		switch err {
		case nil:
			connTimer.Reset(ttl)

			// если пакет не егтс формата закрываем соединение
			if headerBuf[0] != 0x01 {
				logger.Warnf("Пакет не соответствует формату Galileo. Закрыто соединение %s", conn.RemoteAddr())
				return
			}

			// вычисляем длину пакета, 2 байта после тега
			pkgLen := binary.LittleEndian.Uint16(headerBuf[1:])
			pkgLen <<= 1
			pkgLen >>= 1
			pkgLen += 2

			// получаем концовку пакета
			buf := make([]byte, pkgLen)
			if _, err := io.ReadFull(conn, buf); err != nil {
				logger.Errorf("Ошибка при получении тела пакета: %v", err)
				return
			}

			// формируем полный пакет
			recvPacket = append(headerBuf, buf...)
			//!
			logger.Infof("Получен пакет: %x", recvPacket)
		case io.EOF:
			<-connTimer.C
			_ = conn.Close()
			logger.Warnf("Соединение %s закрыто по таймауту", conn.RemoteAddr())
			return
		default:
			logger.Errorf("Ошибка при получении: %v", err)
			return
		}

		pkg := galileo.Packet{}
		if err := pkg.Decode(recvPacket); err != nil {
			logger.Warn("Ошибка расшифровки пакета")
			logger.Error(err)
			return
		}

		receivedTime := time.Now().UTC().Unix()
		crc := make([]byte, 2)
		binary.LittleEndian.PutUint16(crc, pkg.Crc16)
		//!
		logger.Infof("Контрольная сумма пакета: %x", crc)
		resp := append([]byte{0x02}, crc...)

		if _, err = conn.Write(resp); err != nil {
			logger.Errorf("Ошибка отправки пакета подтверждения: %v", err)
		}
		//!
		logger.Infof("Пакет подтверждение: %x", resp)

		if config.LogRecvPacket {
			_, err = db.Exec(`INSERT INTO "recvpacket" ("recvtime", "termaddress", "packet") VALUES ($1, $2, $3)`, time.Unix(receivedTime, 0), conn.RemoteAddr().String(), fmt.Sprintf("%X", recvPacket))
			if err != nil {
				logger.Errorf("Ошибка записи принятого пакета в лог: %v", err)
				return
			} else {
				logger.Info("Принятый пакет записан в лог.")
			}
		}

		if len(pkg.Tags) < 1 {
			//!
			logger.Info("нулевая длина пакета")
			continue
		}

		if config.LogDecPacket {
			b, _ := json.Marshal(pkg)
			_, err = db.Exec(`INSERT INTO "decpacket" ("recvtime", "termaddress", "packet") VALUES ($1, $2, $3)`, time.Unix(receivedTime, 0), conn.RemoteAddr().String(), string(b))
			if err != nil {
				logger.Errorf("Ошибка записи дешифрованного пакета в лог: %v", err)
				return
			} else {
				logger.Info("Дешифрованный пакет пакет записан в лог.")
			}
		}
		outPkg.ReceivedTimestamp = receivedTime
		prevTag := uint8(0)
		for _, curTag := range pkg.Tags {
			if prevTag > curTag.Tag {
				if err := outPkg.Send(config.getZabbixHost()); err != nil {
					logger.Errorf("Ошибка отправки архивной записи в zabbix по адресу %s: %v", config.getZabbixHost(), err)
				} else {
					outPkg.Sent2Zabbix = true
					logger.Infof("Архивная запись %d с терминала %d отправлена в zabbix по адресу %s", outPkg.PacketID, outPkg.TerminalNumber, config.getZabbixHost())
				}

				var id int64
				if id, err = outPkg.Save(db); err != nil {
					logger.Errorf("Ошибка записи в records архивной записи: %v", err)
				} else if err := outPkg.noJSONSave(db, id); err != nil {
					logger.Errorf("Ошибка записи в records4lens архивной записи: %v", err)
				}

				TerminalNumber := outPkg.TerminalNumber
				outPkg = galileoParsePacket{
					TerminalNumber:    TerminalNumber,
					ReceivedTimestamp: receivedTime,
				}
			}

			switch curTag.Tag {
			case 0x01:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.HwVer = uint8(val.Val)
			case 0x02:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.SwVer = uint8(val.Val)
			case 0x03:
				val := curTag.Value.(*galileo.StringTag)
				outPkg.IMEI = string(val.Val)
			case 0x04:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.TerminalNumber = uint32(val.Val)
			case 0x10:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.PacketID = uint32(val.Val)
			case 0x20:
				val := curTag.Value.(*galileo.TimeTag)
				outPkg.NavigationTimestamp = val.Val.Unix()
				outPkg.Delay = outPkg.ReceivedTimestamp - outPkg.NavigationTimestamp
			case 0x21:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Milliseconds = uint16(val.Val)
			case 0x30:
				val := curTag.Value.(*galileo.CoordTag)
				outPkg.Nsat = val.Nsat
				outPkg.Latitude = val.Latitude
				outPkg.Longitude = val.Longitude
			case 0x33:
				val := curTag.Value.(*galileo.SpeedTag)
				outPkg.Course = uint8(val.Course)
				outPkg.Speed = uint16(val.Speed)
			case 0x34:
				val := curTag.Value.(*galileo.IntTag)
				outPkg.Height = int(val.Val)
			case 0x35:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Hdop = uint16(val.Val)
			case 0x40:
				val := curTag.Value.(*galileo.BitsTag)
				outPkg.TerminalStatus = string(val.Val)
			case 0x41:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.VoltagePower = uint16(val.Val)
			case 0x42:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.VoltageBattery = uint16(val.Val)
			case 0x43:
				val := curTag.Value.(*galileo.IntTag)
				outPkg.TerminalTemperature = int8(val.Val)
			case 0x45:
				val := curTag.Value.(*galileo.BitsTag)
				outPkg.OutputStatus = string(val.Val)
			case 0x46:
				val := curTag.Value.(*galileo.BitsTag)
				outPkg.InputStatus = string(val.Val)
			case 0xc0:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can_a0 = uint32(val.Val) / 2
			case 0xc1:
				val := curTag.Value.(*galileo.CanA1Tag)
				outPkg.FuelLevel = val.FuelLevel
				outPkg.Cooltemp = val.Cooltemp
				outPkg.Rpm = val.Rpm
			case 0xc2:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can_b0 = uint32(val.Val) * 5
			case 0xc4:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr0 = uint8(val.Val)
			case 0xc5:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr1 = uint8(val.Val)
			case 0xc6:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr2 = uint8(val.Val)
			case 0xc7:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr3 = uint8(val.Val)
			case 0xc8:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr4 = uint8(val.Val)
			case 0xc9:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr5 = uint8(val.Val)
			case 0xca:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr6 = uint8(val.Val)
			case 0xcb:
				val := curTag.Value.(*galileo.IntTag)
				outPkg.Can8bitr7 = int8(val.Val)
			case 0xcc:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr8 = uint8(val.Val)
			case 0xcd:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr9 = uint8(val.Val)
			case 0xce:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr10 = uint8(val.Val)
			case 0xcf:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr11 = uint8(val.Val)
			case 0xd0:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr12 = uint8(val.Val)
			case 0xd1:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr13 = uint8(val.Val)
			case 0xd2:
				val := curTag.Value.(*galileo.BitsTag)
				outPkg.Can8bitr14 = string(val.Val)
			case 0xa0:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr15 = uint8(val.Val)
			case 0xa1:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr16 = uint8(val.Val)
			case 0xa2:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr17 = uint8(val.Val)
			case 0xa9:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr24 = uint8(val.Val)
			case 0xaa:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can8bitr25 = uint8(val.Val)
			case 0xab:
				val := curTag.Value.(*galileo.BitsTag)
				outPkg.Can8bitr26 = string(val.Val)
			case 0xac:
				val := curTag.Value.(*galileo.BitsTag)
				outPkg.Can8bitr27 = string(val.Val)
			case 0xad:
				val := curTag.Value.(*galileo.BitsTag)
				outPkg.Can8bitr28 = string(val.Val)
			case 0xae:
				val := curTag.Value.(*galileo.BitsTag)
				outPkg.Can8bitr29 = string(val.Val)
			case 0xaf:
				val := curTag.Value.(*galileo.BitsTag)
				outPkg.Can8bitr30 = string(val.Val)
			case 0xd6:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr0 = uint16(val.Val)
			case 0xd7:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr1 = uint16(val.Val)
			case 0xd8:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr2 = uint16(val.Val)
			case 0xd9:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr3 = uint16(val.Val)
			case 0xda:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr4 = uint16(val.Val)
			case 0xb0:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr5 = float32(val.Val/256)/100 + float32(val.Val%256) // в старшем байте дробная часть, в младшем байте целая
			case 0xb1:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr6 = float32(val.Val/256)/100 + float32(val.Val%256) // в старшем байте дробная часть, в младшем байте целая
			case 0xb2:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr7 = uint16(val.Val)
			case 0xb3:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr8 = uint16(val.Val)
			case 0xb4:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr9 = uint16(val.Val)
			case 0xb5:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr10 = uint16(val.Val)
			case 0xb6:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr11 = float32(val.Val/256)/100 + float32(val.Val%256) // в старшем байте дробная часть, в младшем байте целая
			case 0xb7:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can16bitr12 = uint16(val.Val)
			case 0xdb:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr0 = uint32(val.Val)
			case 0xdc:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr1 = uint32(val.Val)
			case 0xdd:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr2 = uint32(val.Val)
			case 0xde:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr3 = uint32(val.Val)
			case 0xdf:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr4 = uint32(val.Val)
			case 0xf0:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr5 = uint32(val.Val)
			case 0xf1:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr6 = uint32(val.Val)
			case 0xf2:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr7 = uint32(val.Val)
			case 0xf3:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr8 = uint32(val.Val)
			case 0xf4:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.Can32bitr9 = uint32(val.Val)
			case 0xe2:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.UserTag[0] = uint32(val.Val)
			case 0xe3:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.UserTag[1] = uint32(val.Val)
			case 0xe4:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.UserTag[2] = uint32(val.Val)
			case 0xe5:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.UserTag[3] = uint32(val.Val)
			case 0xe6:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.UserTag[4] = uint32(val.Val)
			case 0xe7:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.UserTag[5] = uint32(val.Val)
			case 0xe8:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.UserTag[6] = uint32(val.Val)
			case 0xe9:
				val := curTag.Value.(*galileo.UintTag)
				outPkg.UserTag[7] = uint32(val.Val)
			default:
				logger.Infof("отсутствует обработчик для регистра %x, переданное значение %v. Теминал %d", curTag.Tag, curTag.Value, outPkg.TerminalNumber)
			}
			prevTag = curTag.Tag
		}

		if err := outPkg.Send(config.getZabbixHost()); err != nil {
			logger.Errorf("Ошибка отправки архивной записи в zabbix по адресу %s: %v", config.getZabbixHost(), err)
		} else {
			outPkg.Sent2Zabbix = true
			logger.Infof("Архивная запись %d с терминала %d отправлена в zabbix по адресу %s", outPkg.PacketID, outPkg.TerminalNumber, config.getZabbixHost())
		}
		var id int64
		if id, err = outPkg.Save(db); err != nil {
			logger.Errorf("Ошибка записи в records архивной записи: %v", err)
		} else if err := outPkg.noJSONSave(db, id); err != nil {
			logger.Errorf("Ошибка записи в records4lens архивной записи: %v", err)
		}
		/*
			if outPkg.TerminalNumber < 11000 { // ТОлько если Номер терминала меньше 11000, т.е. установка s15
				if err := outPkg.checkModes(db); err != nil {
					logger.Errorf("Ошибка записи в БД режима работы установки: %v", err)
				}
			}*/

	}
}
