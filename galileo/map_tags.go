package galileo

type tagDesc struct {
	Len  uint
	Type string
}

var tagsTable = map[byte]tagDesc{
	// версия железа
	0x01: {1, "uint"},
	// версия прошивки
	0x02: {1, "uint"},
	// IMEI
	0x03: {15, "string"},
	// идентификатор устройства
	0x04: {2, "uint"},
	// номер записи в архиве
	0x10: {2, "uint"},
	// Дата и время
	0x20: {4, "time"},
	//  Миллисекунды
	0x21: {2, "uint"},
	// Координаты в градусах, число спутников,
	// признак корректности определения координат и
	// источник координат
	0x30: {9, "coord"},
	// Скорость в км/ч направлене в градусах
	0x33: {4, "speed"},
	// высота, м.
	0x34: {2, "int"},
	// Одно из значений: 1. HDOP (делить на 10) - если истоник координат GPS
	// модуль, 2 погрешность в метрах если источник gsm-сети (умножить на 10)
	0x35: {1, "uint"},
	// Статус устройства
	0x40: {2, "bitstring"},
	// Напряжение питания, мВ
	0x41: {2, "uint"},
	// Напряжение аккумулятора, мВ
	0x42: {2, "uint"},
	/// Температура внутри терминала, °С
	0x43: {1, "int"},
	// Статусы выходов
	0x45: {2, "bitstring"},
	// Статусы входов
	0x46: {2, "bitstring"},
	// EcoDrive и определение стиля вождения
	0x47: {4, "ecodrive"},
	//   Расширенный статус терминала
	0x48: {2, "bitstring"},
	// Значение на входе 0.
	// В зависимости от настроек один из вариантов: напряжение,
	// число импульсов, частота Гц
	0x50: {2, "uint"},
	// Значение на входе 1.
	// В зависимости от настроек один из вариантов: напряжение,
	// число импульсов, частота Гц
	0x51: {2, "uint"},
	// Значение на входе 2.
	// В зависимости от настроек один из вариантов: напряжение,
	// число импульсов, частота Гц
	0x52: {2, "uint"},
	// Значение на входе 3.
	// В зависимости от настроек один из вариантов: напряжение,
	// число импульсов, частота Гц
	0x53: {2, "uint"},
	// Значение на входе 4.
	// В зависимости от настроек один из вариантов: напряжение,
	// число импульсов, частота Гц
	0x54: {2, "uint"},
	// Значение на входе 5.
	// В зависимости от настроек один из вариантов: напряжение,
	// число импульсов, частота Гц
	0x55: {2, "uint"},
	// Значение на входе 6.
	// В зависимости от настроек один из вариантов: напряжение,
	// число импульсов, частота Гц
	0x56: {2, "uint"},
	// Значение на входе 7.
	// В зависимости от настроек один из вариантов: напряжение,
	// число импульсов, частота Гц
	0x57: {2, "uint"},
	//   RS232 0
	0x58: {2, "uint"},
	//   RS232 1
	0x59: {2, "uint"},
	// RS485[0] ДУТ с адресом 0
	0x60: {2, "uint"},
	// RS485[1] ДУТ с адресом 1
	0x61: {2, "uint"},
	// RS485[2] ДУТ с адресом 2
	0x62: {2, "uint"},
	// RS485[3] ДУТ с адресом 3. Относительный уровень топлива и температура.
	0x63: {3, "uint"},
	// RS485[4]. ДУТ с адресом 4. Относительный уровень топлива и температура.
	0x64: {3, "uint"},
	// RS485[4]. ДУТ с адресом 5. Относительный уровень топлива и температура.
	0x65: {3, "uint"},
	// RS485[4]. ДУТ с адресом 6. Относительный уровень топлива и температура.
	0x66: {3, "uint"},
	// RS485[4]. ДУТ с адресом 7. Относительный уровень топлива и температура.
	0x67: {3, "uint"},
	// RS485[4]. ДУТ с адресом 8. Относительный уровень топлива и температура.
	0x68: {3, "uint"},
	// RS485[4]. ДУТ с адресом 9. Относительный уровень топлива и температура.
	0x69: {3, "uint"},
	// RS485[4]. ДУТ с адресом 10. Относительный уровень топлива и температура.
	0x6A: {3, "uint"},
	// RS485[4]. ДУТ с адресом 11. Относительный уровень топлива и температура.
	0x6B: {3, "uint"},
	// RS485[4]. ДУТ с адресом 12. Относительный уровень топлива и температура.
	0x6C: {3, "uint"},
	// RS485[4]. ДУТ с адресом 13. Относительный уровень топлива и температура.
	0x6D: {3, "uint"},
	// RS485[4]. ДУТ с адресом 14. Относительный уровень топлива и температура.
	0x6E: {3, "uint"},
	// RS485[4]. ДУТ с адресом 15. Относительный уровень топлива и температура.
	0x6F: {3, "uint"},
	//  Идентификатор термометра 0 и измеренная температура, °С
	0x70: {2, "uint"},
	//  Идентификатор термометра 1 и измеренная температура, °С
	0x71: {2, "uint"},
	//  Идентификатор термометра 2 и измеренная температура, °С
	0x72: {2, "uint"},
	//  Идентификатор термометра 3 и измеренная температура, °С
	0x73: {2, "uint"},
	//  Идентификатор термометра 4 и измеренная температура, °С
	0x74: {2, "uint"},
	//  Идентификатор термометра 5 и измеренная температура, °С
	0x75: {2, "uint"},
	//  Идентификатор термометра 6 и измеренная температура, °С
	0x76: {2, "uint"},
	//  Идентификатор термометра 7 и измеренная температура, °С
	0x77: {2, "uint"},
	// Значение на входе 8
	0x78: {2, "uint"},
	// Значение на входе 9
	0x79: {2, "uint"},
	// Значение на входе 10
	0x7A: {2, "uint"},
	// Значение на входе 11
	0x7B: {2, "uint"},
	// Значение на входе 12
	0x7C: {2, "uint"},
	// Значение на входе 13
	0x7D: {2, "uint"},
	/*  Расширенные данные RS232[0]
	В зависимости от настройки один из вариантов:
	1. Температура ДУТ, подключенного к нулевому порту RS232, °С.
	2. Вес, полученный от весового индикатора.*/
	0x88: {1, "int"},
	/*  Расширенные данные RS232[0]
	В зависимости от настройки один из вариантов:
	1. Температура ДУТ, подключенного к нулевому порту RS232, °С.
	2. Вес, полученный от весового индикатора.*/
	0x89: {1, "int"},
	// Температура ДУТ с адресом 0, подключенного к порту RS485, °С.
	0x8A: {1, "int"},
	// Температура ДУТ с адресом 1, подключенного к порту RS485, °С.
	0x8B: {1, "int"},
	// Температура ДУТ с адресом 2, подключенного к порту RS485, °С.
	0x8C: {1, "int"},
	// Идентификационный номер первого ключа iButton
	0x90: {4, "uint"},
	// CAN8BITR15
	0xA0: {1, "uint"},
	// CAN8BITR16
	0xA1: {1, "uint"},
	// CAN8BITR17
	0xA2: {1, "uint"},
	// CAN8BITR18
	0xA3: {1, "uint"},
	// CAN8BITR19
	0xA4: {1, "uint"},
	// CAN8BITR20
	0xA5: {1, "uint"},
	// CAN8BITR21
	0xA6: {1, "uint"},
	// CAN8BITR22
	0xA7: {1, "uint"},
	// CAN8BITR23
	0xA8: {1, "uint"},
	// CAN8BITR24
	0xA9: {1, "uint"},
	// CAN8BITR25
	0xAA: {1, "uint"},
	// CAN8BITR26
	0xAB: {1, "bitstring"},
	// CAN8BITR27
	0xAC: {1, "bitstring"},
	// CAN8BITR28
	0xAD: {1, "bitstring"},
	// CAN8BITR29
	0xAE: {1, "bitstring"},
	// CAN8BITR30
	0xAF: {1, "bitstring"},
	/// CAN16BITR5
	0xB0: {2, "uint"},
	// CAN16BITR6
	0xB1: {2, "uint"},
	// CAN16BITR7
	0xB2: {2, "uint"},
	// CAN16BITR8
	0xB3: {2, "uint"},
	// CAN16BITR9
	0xB4: {2, "uint"},
	// CAN16BITR10
	0xB5: {2, "uint"},
	// CAN16BITR11
	0xB6: {2, "uint"},
	// CAN16BITR12
	0xB7: {2, "uint"},
	// CAN16BITR13
	0xB8: {2, "uint"},
	// CAN16BITR14
	0xB9: {2, "uint"},

	// Данные CAN-шины (CAN_A0) или CAN-LOG.
	//Топливо, израсходованное машиной с момента её создания, л
	0xC0: {4, "uint"},
	// Данные CAN-шины (CAN_A1) или CAN-LOG.
	//Уровень топлива, %;
	//температура охлаждающей жидкости, °C;
	//обороты двигателя, об/мин.
	0xC1: {4, "cana1"},
	// Данные CAN-шины (CAN_B0) или CAN-LOG.
	//Пробег автомобиля, м.
	0xC2: {4, "uint"},
	// CAN_B1
	0xC3: {4, "uint"},

	//CAN8BITR0
	0xC4: {1, "uint"},
	//CAN8BITR1
	0xC5: {1, "uint"},
	//CAN8BITR2
	0xC6: {1, "uint"},
	//CAN8BITR3
	0xC7: {1, "uint"},
	//CAN8BITR4
	0xC8: {1, "uint"},
	//CAN8BITR5
	0xC9: {1, "uint"},
	//CAN8BITR6
	0xCA: {1, "uint"},
	//CAN8BITR7
	0xCB: {1, "int"},
	//CAN8BITR8
	0xCC: {1, "uint"},
	//CAN8BITR9
	0xCD: {1, "uint"},
	//CAN8BITR10
	0xCE: {1, "uint"},
	//CAN8BITR11
	0xCF: {1, "uint"},
	//CAN8BITR12
	0xD0: {1, "uint"},
	//CAN8BITR13
	0xD1: {1, "uint"},
	//CAN8BITR14
	0xD2: {1, "bitstring"},

	// Идентификационный номер второго ключа iButton
	0xD3: {4, "uint"},
	// Общий пробег по данным GPS/ГЛОНАСС-модулей, м.
	0xD4: {4, "uint"},
	// Состояние ключей iButton, идентификаторы которых заданы командой iButtons
	0xD5: {1, "bitstring"},

	// CAN16BITR0
	0xD6: {2, "uint"},
	// CAN16BITR1
	0xD7: {2, "uint"},
	// CAN16BITR2
	0xD8: {2, "uint"},
	// CAN16BITR3
	0xD9: {2, "uint"},
	// CAN16BITR4
	0xDA: {2, "uint"},

	// CAN32BITR0
	0xDB: {4, "uint"},
	// CAN32BITR1
	0xDC: {4, "uint"},
	// CAN32BITR2
	0xDD: {4, "uint"},
	// CAN32BITR3
	0xDE: {4, "uint"},
	// CAN32BITR4
	0xDF: {4, "uint"},

	// Данные пользователя 0
	0xE2: {4, "uint"},
	// Данные пользователя 1
	0xE3: {4, "uint"},
	// Данные пользователя 2
	0xE4: {4, "uint"},
	// Данные пользователя 3
	0xE5: {4, "uint"},
	// Данные пользователя 4
	0xE6: {4, "uint"},
	// Данные пользователя 5
	0xE7: {4, "uint"},
	// Данные пользователя 6
	0xE8: {4, "uint"},
	// Данные пользователя 7
	0xE9: {4, "uint"},
	// Массив данных пользователя
	0xEA: {160, "uint"},

	// CAN32BITR5
	0xF0: {4, "uint"},
	// CAN32BITR6
	0xF1: {4, "uint"},
	// CAN32BITR7
	0xF2: {4, "uint"},
	// CAN32BITR8
	0xF3: {4, "uint"},
	// CAN32BITR9
	0xF4: {4, "uint"},
	// CAN32BITR10
	0xF5: {4, "uint"},
	// CAN32BITR11
	0xF6: {4, "uint"},
	// CAN32BITR12
	0xF7: {4, "uint"},
	// CAN32BITR13
	0xF8: {4, "uint"},
	// CAN32BITR14
	0xF9: {4, "uint"},
}
