package main

import (
	"net"
	"os"
	"time"

	"github.com/labstack/gommon/log"
)

// СТруктура для хранения данных о режиме работы установки и двигателя
type Mode struct {
	Begin, End  time.Time
	Mode, Count int
}

var (
	config     settings
	logger     *log.Logger
	EngineMode map[uint32]*Mode
	DrillMode  map[uint32]*Mode
)

func main() {
	EngineMode = make(map[uint32]*Mode)
	DrillMode = make(map[uint32]*Mode)
	logger = log.New("-")
	logger.SetHeader("${time_rfc3339_nano} ${short_file}:${line} ${level} -${message}")
	f, err := os.OpenFile("/usr/local/smartDrillServer/receiver/receiver.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatalf("error opening file: %v", err)
	}
	defer f.Close()
	logger.SetOutput(f)

	if len(os.Args) == 2 {
		if err := config.Load(os.Args[1]); err != nil {
			logger.Fatalf("Ошибка парсинга конфига: %v", err)
		}
	} else {
		logger.Fatalf("Не задан путь до конфига")
	}
	logger.SetLevel(config.getLogLevel())
	runServer(config.getListenAddress(), config.getEmptyConnTTL())
}

func runServer(srvAddress string, conTTL time.Duration) {
	l, err := net.Listen("tcp", srvAddress)
	if err != nil {
		logger.Fatalf("Не удалось открыть соединение: %v", err)
	}
	defer l.Close()

	logger.Infof("Запущен сервер %s...", srvAddress)
	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Errorf("Ошибка соединения: %v", err)
		} else {
			go handleRecvPkg(conn, conTTL)
		}
	}
}
