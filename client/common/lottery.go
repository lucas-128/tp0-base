package common

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	MessageTypeWinners   = "WINNERS"
	MessageTypeNoWinn    = "NOWINN"
	MessageTypeBetData   = "BETDATA"
	MessageTypeReqWinner = "REQWINN"
	LengthBytes          = 4
)

type Bet struct {
	Agency    string
	FirstName string
	LastName  string
	Document  string
	Birthdate string
	Number    string
}

func (b Bet) ToString() string {
	return strings.Join([]string{
		b.Agency,
		b.FirstName,
		b.LastName,
		b.Document,
		b.Birthdate,
		b.Number,
	}, "|")
}

func BetFromEnv(id string) Bet {
	nombre := os.Getenv("NOMBRE")
	apellido := os.Getenv("APELLIDO")
	documento := os.Getenv("DOCUMENTO")
	nacimiento := os.Getenv("NACIMIENTO")
	numero := os.Getenv("NUMERO")

	bet := Bet{
		Agency:    id,
		FirstName: nombre,
		LastName:  apellido,
		Document:  documento,
		Birthdate: nacimiento,
		Number:    numero,
	}

	return bet
}

func sendAll(conn net.Conn, data []byte) error {
	totalSent := 0
	for totalSent < len(data) {
		n, err := conn.Write(data[totalSent:])
		if err != nil {
			return fmt.Errorf("failed to send data: %w", err)
		}
		totalSent += n
	}
	return nil
}

func Send(c *Client, bet Bet) error {
	conn := c.conn

	betData := bet.ToString()
	betDataBytes := []byte(betData)
	dataSize := len(betDataBytes)
	var buffer bytes.Buffer

	if err := binary.Write(&buffer, binary.BigEndian, int32(dataSize)); err != nil {
		return fmt.Errorf("failed to write data size: %w", err)
	}

	if err := sendAll(conn, buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to send data size: %w", err)
	}

	if err := sendAll(conn, betDataBytes); err != nil {
		return fmt.Errorf("failed to send bet data: %w", err)
	}

	return nil
}

func SendChunks(c *Client, data string) error {

	maxBatchSize := c.config.MaxAmount
	conn := c.conn
	dataChunks := splitIntoChunks(data, maxBatchSize, c.config.ID)

	betDataMsg := MessageTypeBetData
	betDataBytes := []byte(betDataMsg)
	betDataSize := int32(len(betDataBytes))

	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, betDataSize); err != nil {
		return fmt.Errorf("failed to write BETDATA message size: %w", err)
	}
	if err := sendAll(conn, buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to send BETDATA message size: %w", err)
	}

	if err := sendAll(conn, betDataBytes); err != nil {
		return fmt.Errorf("failed to send BETDATA message: %w", err)
	}

	for _, chunk := range dataChunks {

		chunkBytes := []byte(chunk)
		dataSize := len(chunkBytes)
		buffer.Reset()

		if err := binary.Write(&buffer, binary.BigEndian, int32(dataSize)); err != nil {
			return fmt.Errorf("failed to write data size: %w", err)
		}

		if err := sendAll(conn, buffer.Bytes()); err != nil {
			return fmt.Errorf("failed to send data size: %w", err)
		}

		if err := sendAll(conn, chunkBytes); err != nil {
			return fmt.Errorf("failed to send data chunk: %w", err)
		}

		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		if err != nil {
			log.Errorf("%v", err)
		} else {
			log.Infof("%v", msg)
		}
	}

	// Send data size 0 so that the server knows all chunks were sent.
	buffer.Reset()
	if err := binary.Write(&buffer, binary.BigEndian, int32(0)); err != nil {
		return fmt.Errorf("failed to write: %w", err)
	}

	if err := sendAll(conn, buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}

	return nil
}

func splitIntoChunks(data string, maxBatchSize int, id string) []string {
	var chunks []string
	scanner := bufio.NewScanner(strings.NewReader(data))
	var buffer bytes.Buffer
	lineCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++
		lineWithId := line + "," + id

		if lineCount > maxBatchSize && buffer.Len() > 0 {
			chunks = append(chunks, buffer.String())
			buffer.Reset()
			lineCount = 1
		}

		if buffer.Len() > 0 {
			buffer.WriteString("\n")
		}
		buffer.WriteString(lineWithId)
	}

	if buffer.Len() > 0 {
		chunks = append(chunks, buffer.String())
	}
	return chunks
}
func readAgencyBets(id string) (string, error) {

	filePath := fmt.Sprintf("agency-%s.csv", id)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func recvAll(conn net.Conn, length int) ([]byte, error) {
	data := make([]byte, 0, length)
	buf := make([]byte, length)

	for len(data) < length {
		n, err := conn.Read(buf[len(data):])
		if err != nil {
			return nil, fmt.Errorf("failed to receive data: %w", err)
		}
		data = append(data, buf[:n]...)
	}
	return data, nil
}

func requestWinner(c *Client) (bool, error) {

	if err := c.createClientSocket(); err != nil {
		return false, fmt.Errorf("failed to create client socket: %w", err)
	}
	defer c.conn.Close()

	reqWinMsg := MessageTypeReqWinner
	reqWinBytes := []byte(reqWinMsg)
	reqWinSize := int32(len(reqWinBytes))

	var buffer bytes.Buffer
	if err := binary.Write(&buffer, binary.BigEndian, reqWinSize); err != nil {
		return false, fmt.Errorf("failed to write REQWINN message size: %w", err)
	}
	if err := sendAll(c.conn, buffer.Bytes()); err != nil {
		return false, fmt.Errorf("failed to send REQWINN message size: %w", err)
	}

	if err := sendAll(c.conn, reqWinBytes); err != nil {
		return false, fmt.Errorf("failed to send REQWINN message: %w", err)
	}

	id := c.config.ID
	idBytes := []byte(id)
	idSize := int32(len(idBytes))

	buffer.Reset()
	if err := binary.Write(&buffer, binary.BigEndian, idSize); err != nil {
		return false, fmt.Errorf("failed to write ID size: %w", err)
	}
	if err := sendAll(c.conn, buffer.Bytes()); err != nil {
		return false, fmt.Errorf("failed to send ID size: %w", err)
	}

	if err := sendAll(c.conn, idBytes); err != nil {
		return false, fmt.Errorf("failed to send ID: %w", err)
	}

	lengthBytes, err := recvAll(c.conn, LengthBytes)
	if err != nil {
		return false, fmt.Errorf("failed to read response length: %w", err)
	}

	var responseSize int32
	if err := binary.Read(bytes.NewReader(lengthBytes), binary.BigEndian, &responseSize); err != nil {
		return false, fmt.Errorf("failed to parse response size: %w", err)
	}

	responseBytes, err := recvAll(c.conn, int(responseSize))
	if err != nil {
		return false, fmt.Errorf("failed to read response data: %w", err)
	}

	responseMessage := string(responseBytes)
	if responseMessage == MessageTypeWinners {
		if err := handleWinnerData(c.conn); err != nil {
			return false, fmt.Errorf("failed to handle winner data: %w", err)
		}
		return true, nil
	} else if responseMessage == MessageTypeNoWinn {
		return false, nil
	} else {
		return false, fmt.Errorf("unexpected response: %s", responseMessage)
	}
}

func handleWinnerData(conn net.Conn) error {

	lengthBytes, err := recvAll(conn, LengthBytes)
	if err != nil {
		return fmt.Errorf("failed to read winner data length: %w", err)
	}

	var dataSize int32
	if err := binary.Read(bytes.NewReader(lengthBytes), binary.BigEndian, &dataSize); err != nil {
		return fmt.Errorf("failed to parse winner data size: %w", err)
	}

	dataBytes, err := recvAll(conn, int(dataSize))
	if err != nil {
		return fmt.Errorf("failed to read winner data: %w", err)
	}

	documents := strings.Split(string(dataBytes), ",")
	documentCount := len(documents)

	logMessage := fmt.Sprintf("action: consulta_ganadores | result: success | cant_ganadores: %d", documentCount)
	log.Infof("%v", logMessage)
	return nil
}
