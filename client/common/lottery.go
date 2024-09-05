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

// SendChunks reads data from a file and sends it to the client in chunks.
// It handles shutdown signals to stop the client gracefully.
func SendChunks(c *Client, sigChan chan os.Signal) error {
	filePath := fmt.Sprintf("agency-%s.csv", c.config.ID)
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	maxBatchSize := c.config.MaxAmount
	conn := c.conn

	scanner := bufio.NewScanner(file)
	var buffer bytes.Buffer
	lineCount := 0

	for scanner.Scan() {
		select {
		// Stop sending data if a shutdown signal is received
		case <-sigChan:
			c.StopClient()
			return nil
		default:
			// Process the line and prepare it for sending
			line := scanner.Text()
			lineCount++
			lineWithId := line + "," + c.config.ID

			// Send the current chunk if the batch size is exceeded
			if lineCount > maxBatchSize && buffer.Len() > 0 {
				if err := sendChunk(c, buffer.String(), conn); err != nil {
					return err
				}
				buffer.Reset()
				lineCount = 1
			}

			if buffer.Len() > 0 {
				buffer.WriteString("\n")
			}
			buffer.WriteString(lineWithId)
		}
	}

	// Send any remaining data in the buffer
	if buffer.Len() > 0 {
		if err := sendChunk(c, buffer.String(), conn); err != nil {
			return err
		}
	}

	// Send data size 0 so that the server knows all chunks were sent.
	if err := sendChunk(c, "", conn); err != nil {
		return err
	}

	return nil
}

// sendChunk sends a chunk of data to the client, including its size.
// It handles both the size and data transmission, and logs any errors encountered.
func sendChunk(c *Client, chunk string, conn net.Conn) error {
	chunkBytes := []byte(chunk)
	dataSize := len(chunkBytes)
	var buffer bytes.Buffer

	if err := binary.Write(&buffer, binary.BigEndian, int32(dataSize)); err != nil {
		return fmt.Errorf("failed to write data size: %w", err)
	}

	// Send the data size to the server
	if err := sendAll(conn, buffer.Bytes()); err != nil {
		return fmt.Errorf("failed to send data size: %w", err)
	}

	if dataSize > 0 {
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
	return nil
}
