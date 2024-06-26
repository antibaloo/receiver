package galileo

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
)

// Packet структура пакета протокола GalileoSky
type Packet struct {
	Header byte   `json:"header"`
	Length uint16 `json:"length"`
	Tags   tags   `json:"tags"`
	Crc16  uint16 `json:"crc"`
}

// Decode декодирует пакет
func (g *Packet) Decode(pkg []byte) error {
	var (
		err error
	)

	paketBodyLen := len(pkg) - 2

	g.Crc16 = binary.LittleEndian.Uint16(pkg[paketBodyLen:])

	if crc16(pkg[:paketBodyLen]) != g.Crc16 {
		return fmt.Errorf("crc пакета не совпадает")
	}

	buf := bytes.NewReader(pkg[:paketBodyLen])

	if g.Header, err = buf.ReadByte(); err != nil {
		return fmt.Errorf("ошибка чтения залоговка пакета: %v", err)
	}

	lenBytes := make([]byte, 2)
	if _, err = buf.Read(lenBytes); err != nil {
		return fmt.Errorf("ошибка чтения длины пакета: %v", err)
	}

	g.Length = binary.LittleEndian.Uint16(lenBytes)

	lenBits := strconv.FormatUint(uint64(g.Length), 2)
	if len(lenBits) < 1 {
		return fmt.Errorf("некорректная длина пакета: %v", err)
	}

	if lenBits[:1] == "1" {
		// если есть не отправленные данные, зануляем старший бит
		g.Length = g.Length << 1 >> 1
	}

	for buf.Len() > 0 {
		t := tag{}
		if t.Tag, err = buf.ReadByte(); err != nil {
			return fmt.Errorf("ошибка чтения тэга: %v", err)
		}

		if tagInfo, ok := tagsTable[t.Tag]; ok {
			tagVal := make([]byte, tagInfo.Len)
			if _, err := buf.Read(tagVal); err != nil {
				return fmt.Errorf("не удалось считать значение тега %x: %v", t.Tag, err)
			}
			if err := t.SetValue(tagInfo.Type, tagVal); err != nil {
				return err
			}
			g.Tags = append(g.Tags, t)
		} else {
			return fmt.Errorf("не найдена информаци по тегу %x", t.Tag)
		}

	}

	return err
}
