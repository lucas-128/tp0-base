package common

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

const retryInterval = 2 * time.Second

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ServerAddress string
	ID            string
	LoopAmount    int
	LoopPeriod    time.Duration
	MaxAmount     int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	stop   chan struct{}
	wg     sync.WaitGroup
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
		stop:   make(chan struct{}),
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {

	err := c.createClientSocket()
	if err != nil {
		return
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	c.wg.Add(1)
	defer c.wg.Done()

	data, err := readAgencyBets(c.config.ID)
	if err != nil {
		fmt.Println("Error reading data:", err)
		return
	}

	SendChunks(c, data)
	for {
		success := requestWinner(c)
		if !success {
			time.Sleep(retryInterval)
			continue
		}
		break
	}
}

// Gracefully shut down the client
func (c *Client) StopClient() {
	close(c.stop)
	c.wg.Wait()

	if c.conn != nil {
		c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}
	log.Infof("action: client_shutdown | result: success | client_id: %v", c.config.ID)
}
