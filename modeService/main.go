package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type myError struct {
	ErrorMessage string
}

type Coord struct {
	Latitude  float32 `json:"latitude"`
	Longitude float32 `json:"longitude"`
}

type coordsHandler struct {
	Coords []Coord
	Avg    Coord
}

func (csh coordsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	responseError := myError{}
	params := r.URL.Query()
	from := params.Get("from")
	till := params.Get("till")
	termnumber := params.Get("termnumber")
	timeZone := params.Get("timeZone")

	if len(termnumber) == 0 {
		responseError.ErrorMessage = "no termnumber in request"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	if len(from) == 0 {
		responseError.ErrorMessage = "parameter from is required"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	if len(till) == 0 {
		responseError.ErrorMessage = "parameter till is required"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	if len(timeZone) == 0 {
		responseError.ErrorMessage = "parameter timeZone is required"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	layout := "2006-01-02 15:04"

	temp, err := time.Parse(layout, from)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprintf("Error while parsing parameter 'from': %v", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprintf("Error while parsing parameter 'timeZone': %v", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	begin := time.Date(temp.Year(), temp.Month(), temp.Day(), temp.Hour(), temp.Minute(), temp.Second(), 0, location)

	temp, err = time.Parse(layout, till)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprintf("Error while parsing parameter 'till': %v", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	end := time.Date(temp.Year(), temp.Month(), temp.Day(), temp.Hour(), temp.Minute(), temp.Second(), 0, location)

	//	query := fmt.Sprintf("select distinct round(cast(latitude as numeric), 3), round(cast(longitude as numeric),3) from records4lens where recnumber = 0 AND termnumber = %s AND termtime at time zone '%s' BETWEEN '%s' AND '%s'", termnumber, timeZone, from, till)
	//	log.Println(query)
	csh.Coords = make([]Coord, 0)

	// Получение координат установки за период
	rows, err := db.Query("select distinct round(cast(latitude as numeric), 3), round(cast(longitude as numeric),3) from records4lens where recnumber = 0 AND termnumber = $1 AND termtime BETWEEN $2 AND $3", termnumber, begin, end)
	//rows, err := db.Query(query)
	if err != nil {
		responseError.ErrorMessage = err.Error()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}
	defer rows.Close()
	coord := Coord{}
	for rows.Next() {
		err := rows.Scan(&coord.Latitude, &coord.Longitude)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(coord)
		if coord.Latitude != 0 && coord.Longitude != 0 {
			csh.Coords = append(csh.Coords, coord)
		}

	}
	var avLat, avLon float32
	for _, point := range csh.Coords {
		avLat += point.Latitude
		avLon += point.Longitude
	}
	csh.Avg.Latitude = avLat / float32(len(csh.Coords))
	csh.Avg.Longitude = avLon / float32(len(csh.Coords))
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.Encode(&csh)
}

type coordHandler struct {
	Latitude  float32
	Longitude float32
}

func (ch coordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")
	responseError := myError{}
	params := r.URL.Query()
	termnumber := params.Get("termnumber")
	termnumberInt, _ := strconv.Atoi(termnumber)
	if termnumberInt < 1 {
		responseError.ErrorMessage = "no termnumber parameter"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}
	row := db.QueryRow("select latitude, longitude from records4lens WHERE termnumber = $1 AND recnumber = 0 ORDER BY termtime DESC LIMIT 1", termnumber)
	err := row.Scan(&ch.Latitude, &ch.Longitude)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprint(err)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.Encode(&ch)
}

type analiticHandler struct {
}

func (ah analiticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept")

	type Trip struct {
		Begin         time.Time `json:"begin"`
		End           time.Time `json:"end"`
		DurationTrip  float64   `json:"durationtrip"`
		DurationDrill float64   `json:"durationdrill"`
		TripSpeed     float64   `json:"tripspeed"`
	}
	type Mode struct {
		Beginmod time.Time `json:"beginmode"`
		Endmode  time.Time `json:"endmode"`
		Mode     int       `json:"mode"`
		Count    int       `json:"count"`
	}
	type ModeResponse struct {
		Termnumber    int       `json:"termnumber"`
		Begin         time.Time `json:"begin"`
		End           time.Time `json:"end"`
		AllCount      int       `json:"allcount"`
		StopCount     int       `json:"stopcount"`
		StopWEnCount  int       `json:"stopwencount"`
		EnOnCount     int       `json:"enoncount"`
		DrillCount    int       `json:"drillcount"`
		WashCount     int       `json:"washcount"`
		UpDownCount   int       `json:"updowncount"`
		OvershotCount int       `json:"overshotcount"`
		Topping       float64   `json:"topping"`
		Consumption   float64   `json:"consumption"`
		Alarms        int       `json:"alarms"`
		AvgTripSpeed  float64   `json:"avgtripspeed"`
		TripCount     int       `json:"tripcount"`
		Trips         []Trip    `json:"trips"`
		EnModes       []Mode    `json:"enmodes"`
		DrModes       []Mode    `json:"drmodes"`
	}

	params := r.URL.Query()
	termnumber := params.Get("termnumber")
	from := params.Get("from")
	till := params.Get("till")
	timeZone := params.Get("timeZone")
	responseError := myError{}

	if len(termnumber) == 0 {
		responseError.ErrorMessage = "no termnumber in request"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	if len(from) == 0 {
		responseError.ErrorMessage = "parameter from is required"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	if len(till) == 0 {
		responseError.ErrorMessage = "parameter till is required"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	if len(timeZone) == 0 {
		responseError.ErrorMessage = "parameter timeZone is required"
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}
	layout := "2006-01-02 15:04"

	temp, err := time.Parse(layout, from)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprintf("Error while parsing parameter 'from': %v", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	location, err := time.LoadLocation(timeZone)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprintf("Error while parsing parameter 'timeZone': %v", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	begin := time.Date(temp.Year(), temp.Month(), temp.Day(), temp.Hour(), temp.Minute(), temp.Second(), 0, location)

	temp, err = time.Parse(layout, till)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprintf("Error while parsing parameter 'till': %v", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	end := time.Date(temp.Year(), temp.Month(), temp.Day(), temp.Hour(), temp.Minute(), temp.Second(), 0, location)

	termnumberInt, _ := strconv.Atoi(termnumber)

	// Инициализация ответной структуры
	response := ModeResponse{
		Termnumber:  termnumberInt,
		Begin:       begin,
		End:         end,
		TripCount:   0,
		Alarms:      0,
		Topping:     0,
		Consumption: 0,
		EnModes:     make([]Mode, 0),
		DrModes:     make([]Mode, 0),
	}
	response.AllCount = int(time.Duration(response.End.Sub(response.Begin)).Seconds())

	// Запрос данных по уровню топлива за период от терминала
	//query := fmt.Sprintf("select can8bitr5 from records4lens WHERE termnumber = %s AND termtime at time zone '%s' BETWEEN '%s' AND '%s'", termnumber, timeZone, from, till)
	rows, err := db.Query("select can8bitr5 from records4lens WHERE termnumber = $1 AND termtime BETWEEN $2 AND $3 ORDER BY termtime ASC", termnumber, begin, end)
	//rows, err := db.Query(query)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprintf("Error while executing sql query for fuel levels in period: %v", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}
	defer rows.Close()
	levels := []int{}
	for rows.Next() {
		var level int
		err := rows.Scan(&level)
		if err != nil {
			fmt.Println(err)
			continue
		}
		levels = append(levels, level)
	}
	// Подсчет приходных/расходных операций
	if len(levels) > 0 {
		lastLevel := levels[0]
		for _, level := range levels {
			if level == 0 {
				continue
			} // Слишком много нулей в ряду
			if math.Abs(float64(lastLevel)-float64(level)) > 1 {
				if lastLevel-level > 1 {
					response.Consumption += float64(lastLevel - level)
				} else {
					response.Topping += float64(level - lastLevel)
				}
				lastLevel = level
			}
		}
	}
	response.Consumption *= 2.56
	response.Topping *= 2.56
	// Подсчет овершотов в заданном периоде
	// query = fmt.Sprintf("select count(*) from drillmods WHERE termnumber = %s AND endmod at time zone '%s' >= '%s' AND beginmod at time zone '%s' <= '%s' AND mode = 4", termnumber, timeZone, from, timeZone, till)
	row := db.QueryRow("select count(*) from drillmods WHERE termnumber = $1 AND endmod >= $2 AND beginmod  <= $3 AND mode = 4", termnumber, begin, end)
	//row := db.QueryRow(query)
	err = row.Scan(&response.TripCount)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprintf("Error while executing sql query for overshot counts in period: %v", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	// Подсчет аварийных подъемов в заданном периоде
	//query = fmt.Sprintf("select count(*) from drillmods WHERE termnumber = %s AND endmod at time zone '%s' >= '%s' AND beginmod at time zone '%s' <= '%s' AND mode = 5", termnumber, timeZone, from, timeZone, till)
	row = db.QueryRow("select count(*) from drillmods WHERE termnumber = $1 AND endmod >= $2 AND beginmod <= $3 AND mode = 5", termnumber, begin, end)
	//row = db.QueryRow(query)
	err = row.Scan(&response.Alarms)
	if err != nil {
		responseError.ErrorMessage = fmt.Sprintf("Error while executing sql query for alarms counts in period: %v", err.Error())
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		encoder := json.NewEncoder(w)
		encoder.Encode(&responseError)
		return
	}

	// Формирование массива рейсов
	if response.TripCount > 0 {
		response.Trips = make([]Trip, response.TripCount)
		response.Trips[0].Begin = begin
		//query = fmt.Sprintf("select endmod from drillmods WHERE termnumber = %s AND endmod at time zone '%s' >= '%s' AND beginmod at time zone '%s' <= '%s' AND mode = 4", termnumber, timeZone, from, timeZone, till)
		rows, err = db.Query("select endmod from drillmods WHERE termnumber = $1 AND endmod >= $2 AND beginmod <= $3 AND mode = 4", termnumber, begin, end)
		//log.Println(query)
		//rows, err = db.Query(query)
		if err != nil {
			responseError.ErrorMessage = fmt.Sprintf("Error while executing sql query for trips array  in period: %v", err.Error())
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			encoder := json.NewEncoder(w)
			encoder.Encode(&responseError)
			return
		}
		defer rows.Close()
		counter := 0
		for rows.Next() {
			err := rows.Scan(&response.Trips[counter].End)
			if err != nil {
				log.Println(err)
				//log.Printf("При формировании массива рейсов на установке %v за период с по %v по %v, сообщение об ошибке: %v", termnumber, time.Unix(fromUT, 0), time.Unix(tillUT, 0), err)
				continue
			}
			response.Trips[counter].End = response.Trips[counter].End.In(location)
			if counter < response.TripCount-1 {
				response.Trips[counter+1].Begin = response.Trips[counter].End
			}
			response.Trips[counter].DurationTrip = response.Trips[counter].End.Sub(response.Trips[counter].Begin).Seconds()
			counter++
		}

		for i, trip := range response.Trips {
			if i == 0 { // первый рейс считаем неполным и он не участвует в расчете средней рейсовой скорости
				continue
			}
			row = db.QueryRow("select sum(count) from drillmods WHERE termnumber = $1 AND endmod >= $2 AND beginmod <=$3 and mode = 1", termnumber, trip.Begin, trip.End)
			err = row.Scan(&response.Trips[i].DurationDrill)
			if err != nil {
				response.Trips[i].DurationDrill = 0
				log.Printf("При расчете времени бурения на установке %v в рамках рейса, за период с по %v по %v, сообщение об ошибке: %v", termnumber, trip.Begin, trip.End, err)
				log.Println(err)
			}
			response.Trips[i].DurationDrill *= 4 //Превращаем циклы в секунды
			response.Trips[i].TripSpeed = response.Trips[i].DurationDrill / response.Trips[i].DurationTrip
			response.AvgTripSpeed += response.Trips[i].TripSpeed
		}
		response.AvgTripSpeed /= float64(response.TripCount - 1)
	}

	// Получение вектора режимов работы установки
	//query = fmt.Sprintf("select beginmod, endmod, mode, count from drillmods WHERE termnumber = %s AND endmod  at time zone '%s' >= '%s' AND beginmod  at time zone '%s' <= '%s' ORDER BY beginmod ASC", termnumber, timeZone, from, timeZone, till)
	rows, err = db.Query("select beginmod, endmod, mode from drillmods WHERE termnumber = $1 AND endmod >= $2 AND beginmod <= $3 ORDER BY beginmod ASC", termnumber, begin, end)
	//rows, err = db.Query(query)
	if err != nil {
		log.Print(err)
	}
	defer rows.Close()
	mode := Mode{}
	for rows.Next() {
		err := rows.Scan(&mode.Beginmod, &mode.Endmode, &mode.Mode)
		if err != nil {
			fmt.Println(err)
			continue
		}
		mode.Beginmod = mode.Beginmod.In(location)
		mode.Endmode = mode.Endmode.In(location)
		response.DrModes = append(response.DrModes, mode)
	}

	// Приводим начало первого режима и конец последнего к началу и концу общего периода запроса
	response.DrModes[0].Beginmod = begin
	response.DrModes[len(response.DrModes)-1].Endmode = end

	// Расчитываем длительности периодов режимов их их границ, а не тактов
	for i, m := range response.DrModes {
		response.DrModes[i].Count = int(time.Duration(m.Endmode.Sub(m.Beginmod)).Seconds())
		switch m.Mode {
		case 1:
			response.DrillCount += response.DrModes[i].Count
		case 2:
			response.WashCount += response.DrModes[i].Count
		case 3:
			response.UpDownCount += response.DrModes[i].Count
		case 4:
			response.OvershotCount += response.DrModes[i].Count
		case 5:
			response.Alarms += response.DrModes[i].Count
		}
	}

	// Получение вектора режимов работы двигателя
	//query = fmt.Sprintf("select beginmod, endmod, mode, count from enginemods WHERE termnumber = %s AND endmod at time zone '%s' >= '%s' AND beginmod  at time zone '%s' <= '%s' ORDER BY beginmod ASC", termnumber, timeZone, from, timeZone, till)
	rows, err = db.Query("select beginmod, endmod, mode from enginemods WHERE termnumber = $1 AND endmod >= $2 AND beginmod <= $3 ORDER BY beginmod ASC", termnumber, begin, end)
	//rows, err = db.Query(query)
	if err != nil {
		log.Print(err)
	}
	defer rows.Close()
	mode = Mode{}
	for rows.Next() {
		err := rows.Scan(&mode.Beginmod, &mode.Endmode, &mode.Mode)
		if err != nil {
			fmt.Println(err)
			continue
		}
		mode.Beginmod = mode.Beginmod.In(location)
		mode.Endmode = mode.Endmode.In(location)

		response.EnModes = append(response.EnModes, mode)
	}

	// Приводим начало первого режима и конец последнего к началу и концу общего периода запроса
	response.EnModes[0].Beginmod = begin
	response.EnModes[len(response.EnModes)-1].Endmode = end

	// Расчитываем длительности периодов режимов их их границ, а не тактов
	for i, m := range response.EnModes {
		response.EnModes[i].Count = int(time.Duration(m.Endmode.Sub(m.Beginmod)).Seconds())
		switch m.Mode {
		case 4, 5:
			response.EnOnCount += response.EnModes[i].Count
		}
	}

	// Получение времени простоя с заглушенным двигателем
	// как разницу между длиной периода и временем работы двигателя
	response.StopCount = response.AllCount - response.EnOnCount
	// Получение времени простоя с заведенным двигателем, как разницу мед-жду временем работы с заведенным двигателем и
	// общим временем работы во всех режимах работы установки
	response.StopWEnCount = response.EnOnCount - response.DrillCount - response.WashCount - response.UpDownCount - response.OvershotCount

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.Encode(&response)
}

var (
	db  *sql.DB
	err error
)

func init() {
	// loads values from .env into the system
	if err := godotenv.Load("/usr/local/smartDrillServer/receiver/modeService/.env"); err != nil {
		log.Fatalln("No .env file found")
	}
}

func main() {
	pass, exists := os.LookupEnv("POSTGRES_PASS")
	if !exists {
		log.Fatalln("Variable POSTGRES_PASS not found!")
	}

	connStr := fmt.Sprintf("user=postgres password=%s dbname=galileodb sslmode=disable", pass)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Print(err)
	}
	defer db.Close()

	mux := http.NewServeMux()
	ah := analiticHandler{}
	mux.Handle("/analitic", ah)

	ch := coordHandler{}
	mux.Handle("/coord", ch)
	csh := coordsHandler{}
	mux.Handle("/coords", csh)
	log.Print("Listening...")
	err := http.ListenAndServeTLS("api.sd-one.ru:3000", "/etc/letsencrypt/live/api.sd-one.ru/fullchain.pem", "/etc/letsencrypt/live/api.sd-one.ru/privkey.pem", mux)
	if err != nil {
		fmt.Println(err)
	}
}
