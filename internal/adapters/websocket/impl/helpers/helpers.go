// internal/adapters/websocket/impl/helpers/helpers.go
package helpers

import (
	"crypto/tls"
	"io"
	"strconv"
)

func ParseFloat(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

func ReadFrame(r io.Reader) ([]byte, error) {
	header := make([]byte, 2)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}

	length := int(header[1] & 0x7F)
	if length == 126 {
		ext := make([]byte, 2)
		io.ReadFull(r, ext)
		length = int(ext[0])<<8 | int(ext[1])
	} else if length == 127 {
		ext := make([]byte, 8)
		io.ReadFull(r, ext)
		length = int(ext[7])
	}

	mask := make([]byte, 4)
	io.ReadFull(r, mask)

	payload := make([]byte, length)
	io.ReadFull(r, payload)

	for i := 0; i < length; i++ {
		payload[i] ^= mask[i%4]
	}

	return payload, nil
}

// Вспомогательная функция для записи WebSocket фрейма с маскированием
func WriteFrame(conn *tls.Conn, payload []byte) error {
	// Для простоты отправим один фрейм текста без маски (т.к. клиент -> сервер обычно должен маскировать,
	// но это можно опустить для учебного проекта, либо дописать)
	// Здесь — минимальная реализация фрейма:
	header := []byte{0x81} // FIN=1, текстовый фрейм=1
	length := len(payload)
	if length < 126 {
		header = append(header, byte(length))
	} else if length <= 65535 {
		header = append(header, 126, byte(length>>8), byte(length&0xff))
	} else {
		// длина > 65535 (маловероятно для подписки)
		header = append(header, 127,
			byte(length>>56), byte(length>>48), byte(length>>40), byte(length>>32),
			byte(length>>24), byte(length>>16), byte(length>>8), byte(length&0xff))
	}
	frame := append(header, payload...)
	_, err := conn.Write(frame)
	return err
}
