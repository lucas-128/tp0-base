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

func SendChunks(c *Client, data string, sigChan chan os.Signal) error {

	maxBatchSize := c.config.MaxAmount
	conn := c.conn
	dataChunks := splitIntoChunks(data, maxBatchSize, c.config.ID)

	for _, chunk := range dataChunks {

		select {
		case <-sigChan:
			c.StopClient()
			return nil

		default:

			chunkBytes := []byte(chunk)
			dataSize := len(chunkBytes)
			var buffer bytes.Buffer

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
				log.Errorf("%v",
					err,
				)
			} else {
				log.Infof("%v",
					msg,
				)
			}
		}
	}

	// Send data size 0 so that the server knows all chunks were sent.
	var buffer bytes.Buffer
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
