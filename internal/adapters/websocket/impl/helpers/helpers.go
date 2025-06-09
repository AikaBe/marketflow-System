package helpers

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"net/url"
)

// ConnectAndHandshake устанавливает WebSocket-соединение вручную.
func ConnectAndHandshake(u url.URL, host string) (*tls.Conn, error) {
	conn, err := tls.Dial("tcp", u.Host, nil)
	if err != nil {
		return nil, fmt.Errorf("TLS dial failed: %w", err)
	}

	req := fmt.Sprintf(
		"GET %s HTTP/1.1\r\n"+
			"Host: %s\r\n"+
			"Upgrade: websocket\r\n"+
			"Connection: Upgrade\r\n"+
			"Sec-WebSocket-Version: 13\r\n"+
			"Sec-WebSocket-Key: 123456==\r\n\r\n",
		u.Path, host,
	)

	_, err = conn.Write([]byte(req))
	if err != nil {
		return nil, fmt.Errorf("handshake write failed: %w", err)
	}

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("handshake read failed: %w", err)
	}
	if !bytes.Contains(buf[:n], []byte("101 Switching Protocols")) {
		return nil, fmt.Errorf("unexpected handshake response: %s", buf[:n])
	}

	return conn, nil
}

func ReadFrame(r io.Reader) (opcode byte, payload []byte, err error) {
	header := make([]byte, 2)
	if _, err = io.ReadFull(r, header); err != nil {
		return 0, nil, fmt.Errorf("read header: %w", err)
	}

	fin := header[0]&0x80 != 0
	opcode = header[0] & 0x0F
	masked := header[1]&0x80 != 0
	payloadLen := int(header[1] & 0x7F)

	if !fin {
		return 0, nil, fmt.Errorf("fragmented frames not supported")
	}

	switch payloadLen {
	case 126:
		ext := make([]byte, 2)
		if _, err = io.ReadFull(r, ext); err != nil {
			return 0, nil, fmt.Errorf("read extended payload (126): %w", err)
		}
		payloadLen = int(binary.BigEndian.Uint16(ext))
	case 127:
		ext := make([]byte, 8)
		if _, err = io.ReadFull(r, ext); err != nil {
			return 0, nil, fmt.Errorf("read extended payload (127): %w", err)
		}
		payloadLen = int(binary.BigEndian.Uint64(ext))
	}

	var maskKey []byte
	if masked {
		maskKey = make([]byte, 4)
		if _, err = io.ReadFull(r, maskKey); err != nil {
			return 0, nil, fmt.Errorf("read mask key: %w", err)
		}
	}

	payload = make([]byte, payloadLen)
	if _, err = io.ReadFull(r, payload); err != nil {
		return 0, nil, fmt.Errorf("read payload: %w", err)
	}

	if masked {
		for i := 0; i < payloadLen; i++ {
			payload[i] ^= maskKey[i%4]
		}
	}

	switch opcode {
	case 0x1: // text
		return opcode, payload, nil
	case 0x8: // close
		return opcode, nil, io.EOF
	case 0x9, 0xA: // ping/pong
		return opcode, nil, nil
	default:
		return opcode, nil, fmt.Errorf("unsupported opcode: %d", opcode)
	}
}

// WriteFrame отправляет текстовый WebSocket-фрейм.
func WriteFrame(w io.Writer, data []byte) error {
	header := []byte{0x81} // FIN + text frame
	payloadLen := len(data)

	switch {
	case payloadLen <= 125:
		header = append(header, byte(payloadLen))
	case payloadLen <= 65535:
		header = append(header, 126, byte(payloadLen>>8), byte(payloadLen))
	default:
		header = append(header, 127)
		extended := make([]byte, 8)
		binary.BigEndian.PutUint64(extended, uint64(payloadLen))
		header = append(header, extended...)
	}

	_, err := w.Write(append(header, data...))
	return err
}

// WritePingFrame отправляет WebSocket ping фрейм (opcode 0x9)
// func WritePingFrame(conn net.Conn) error {
// 	// Frame формат: fin=1, opcode=0x9 (ping), no mask, no payload
// 	frame := []byte{0x89, 0x00}
// 	_, err := conn.Write(frame)
// 	if err != nil {
// 		return fmt.Errorf("WritePingFrame error: %w", err)
// 	}
// 	return nil
// }
