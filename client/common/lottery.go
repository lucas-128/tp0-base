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

// SendChunks reads data from a file and sends it to the client in chunks.
// It handles shutdown signals and sends an initial BETDATA message.
func SendChunks(c *Client, sigChan chan os.Signal) error {
	filePath := fmt.Sprintf("agency-%s.csv", c.config.ID)
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	maxBatchSize := c.config.MaxAmount
	conn := c.conn
	betDataMsg := MessageTypeBetData
	betDataBytes := []byte(betDataMsg)
	betDataSize := int32(len(betDataBytes))

	// Send BETDATA message type
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

	scanner := bufio.NewScanner(file)
	var chunkBuffer bytes.Buffer
	lineCount := 0

	for scanner.Scan() {
		select {
		case <-sigChan:
			// Stop sending data if a shutdown signal is received
			c.StopClient()
			return fmt.Errorf("SIGTERM Received")
		default:
			line := scanner.Text()
			lineCount++
			lineWithId := line + "," + c.config.ID

			if lineCount > maxBatchSize && chunkBuffer.Len() > 0 {
				// Send the current chunk
				if err := sendChunk(c, chunkBuffer.String(), conn); err != nil {
					return err
				}
				chunkBuffer.Reset()
				lineCount = 1
			}

			if chunkBuffer.Len() > 0 {
				chunkBuffer.WriteString("\n")
			}
			chunkBuffer.WriteString(lineWithId)
		}
	}

	// Send any remaining data in the buffer
	if chunkBuffer.Len() > 0 {
		if err := sendChunk(c, chunkBuffer.String(), conn); err != nil {
			return err
		}
	}

	// Send data size 0 to indicate completion
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

// Sends a request to the server to get the winners and processes the response.
func requestWinner(c *Client) bool {

	// Create a request message for winners
	reqWinMsg := MessageTypeReqWinner
	reqWinBytes := []byte(reqWinMsg)
	reqWinSize := int32(len(reqWinBytes))

	var buffer bytes.Buffer
	binary.Write(&buffer, binary.BigEndian, reqWinSize)
	sendAll(c.conn, buffer.Bytes())
	sendAll(c.conn, reqWinBytes)

	id := c.config.ID
	idBytes := []byte(id)
	idSize := int32(len(idBytes))

	buffer.Reset()
	binary.Write(&buffer, binary.BigEndian, idSize)
	sendAll(c.conn, buffer.Bytes())
	sendAll(c.conn, idBytes)

	// Receive the length of the server's response
	lengthBytes, _ := recvAll(c.conn, LengthBytes)

	var responseSize int32
	binary.Read(bytes.NewReader(lengthBytes), binary.BigEndian, &responseSize)

	responseBytes, _ := recvAll(c.conn, int(responseSize))
	responseMessage := string(responseBytes)

	// Check the type of response message
	if responseMessage == MessageTypeWinners {
		handleWinnerData(c.conn)
		return true
	} else if responseMessage == MessageTypeNoWinn {
		return false
	}
	return false
}

// Processes the winner data received from the server and logs the result.
func handleWinnerData(conn net.Conn) {

	lengthBytes, _ := recvAll(conn, LengthBytes)

	var dataSize int32
	// Read the size of the winner data from the buffer
	binary.Read(bytes.NewReader(lengthBytes), binary.BigEndian, &dataSize)

	// Receive the actual winner data
	dataBytes, _ := recvAll(conn, int(dataSize))

	// Split the data into documents and count
	documents := strings.Split(string(dataBytes), ",")
	documentCount := len(documents)

	logMessage := fmt.Sprintf("action: consulta_ganadores | result: success | cant_ganadores: %d", documentCount)
	log.Infof("%v", logMessage)
}
