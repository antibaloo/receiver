package galileo

import (
	"encoding/binary"
	"fmt"
	"time"
)

type tagParser interface {
	Parse(val []byte) error
}

// UintTag тип тэга беззнаковое целое
type UintTag struct {
	Val uint64 `json:"val"`
}

// Parse заполняет значение тэга
func (u *UintTag) Parse(val []byte) error {
	switch size := len(val); {
	case size == 1:
		u.Val = uint64(val[0])
	case size == 2:
		u.Val = uint64(binary.LittleEndian.Uint16(val))
	case size == 3:
		val = append(val, 0)
		u.Val = uint64(binary.LittleEndian.Uint32(val))
	case size == 4:
		u.Val = uint64(binary.LittleEndian.Uint32(val))
	default:
		return fmt.Errorf("входной массив больше 4 байт: %x", val)
	}
	return nil
}

// StringTag тип тэга строка
type StringTag struct {
	Val string `json:"val"`
}

// Parse заполняет значение тэга
func (s *StringTag) Parse(val []byte) error {
	s.Val = string(val)

	return nil
}

// TimeTag тип тэга время
type TimeTag struct {
	Val time.Time `json:"val"`
}

// Parse заполняет значение тэга
func (t *TimeTag) Parse(val []byte) error {
	secs := int64(binary.LittleEndian.Uint32(val))
	t.Val = time.Unix(secs, 0).UTC()

	return nil
}

// TimeTag тип тэга с координатами
type CoordTag struct {
	Nsat      uint8   `json:"nsat"`
	IsValid   uint8   `json:"is_valid"`
	Longitude float64 `json:"longitude"`
	Latitude  float64 `json:"latitude"`
}

// Parse заполняет значение тэга
func (c *CoordTag) Parse(val []byte) error {
	if len(val) != 9 {
		return fmt.Errorf(" Некорректная длин секции координат : %x", val)
	}

	flgByte := val[0]

	c.Latitude = float64(int32(binary.LittleEndian.Uint32(val[1:5]))) / float64(1000000)
	c.Longitude = float64(int32(binary.LittleEndian.Uint32(val[5:]))) / float64(1000000)

	c.Nsat = flgByte & 0xf
	c.IsValid = flgByte >> 4

	return nil
}

// EcoDriveTag (стиль вождения)
type EcoDriveTag struct {
	Acceleration uint8 `json:"acceleration"`
	Braking      uint8 `json:"braking"`
	AngularAcc   uint8 `json:"angular"`
	Hits         uint8 `json:"hits"`
}

func (e *EcoDriveTag) Parse(val []byte) error {
	if len(val) != 4 {
		return fmt.Errorf(" Некорректная длина секции стиля вождения: %x", val)
	}
	e.Acceleration = uint8(val[0])
	e.Braking = uint8(val[1])
	e.AngularAcc = uint8(val[2])
	e.Hits = uint8(val[3])
	return nil
}

// CanA1Tag (Данные CAN-шины (CAN_A1) уровень топлива, температура охлаждающей жидкости и обороты дввигателя )
type CanA1Tag struct {
	FuelLevel uint8  `json:"fulelevel"`
	Cooltemp  int8   `json:"colltemp"`
	Rpm       uint16 `json:"rpm"`
}

func (e *CanA1Tag) Parse(val []byte) error {
	if len(val) != 4 {
		return fmt.Errorf(" Некорректная длина тега CAN_A1: %x", val)
	}
	e.FuelLevel = uint8(float32(val[0]) * 0.4)
	e.Cooltemp = int8(val[1]) - 40
	e.Rpm = uint16(float32(binary.LittleEndian.Uint16(val[2:])) * 0.125)
	return nil
}

// SpeedTag тип тэга со скоростью
type SpeedTag struct {
	Speed  float64 `json:"speed"`
	Course uint16  `json:"course"`
}

// Parse заполняет значение тэга
func (s *SpeedTag) Parse(val []byte) error {
	if len(val) != 4 {
		return fmt.Errorf(" Некорректная длин секции скорости : %x", val)
	}

	s.Speed = float64(binary.LittleEndian.Uint16(val[:2])) / 10
	s.Course = binary.LittleEndian.Uint16(val[2:]) / 10
	return nil
}

// IntTag тип тэга знаковго целого
type IntTag struct {
	Val int `json:"val"`
}

// Parse заполняет значение тэга
func (u *IntTag) Parse(val []byte) error {
	switch size := len(val); {
	case size == 1:
		u.Val = int(val[0])
	case size == 2:
		u.Val = int(binary.LittleEndian.Uint16(val))
	default:
		return fmt.Errorf("входной массив больше 2 байт: %x", val)
	}

	return nil
}

// BitsTag тип тэга с битами
type BitsTag struct {
	Val string `json:"val"`
}

// Parse заполняет значение тэга
func (b *BitsTag) Parse(val []byte) error {

	switch size := len(val); {
	case size == 1:
		b.Val = fmt.Sprintf("%08b", val[0])
	case size == 2:
		b.Val = fmt.Sprintf("%016b", binary.LittleEndian.Uint16(val))
	default:
		return fmt.Errorf("входной массив больше 2 байт: %x", val)
	}

	return nil
}
