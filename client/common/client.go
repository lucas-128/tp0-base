package common

import (
	"bufio"
	"net"
	"os"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
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
func (c *Client) StartClientLoop(sigChan chan os.Signal) {

	select {
	// If a shutdown signal is received, stop the client and exit the loop
	case <-sigChan:
		c.StopClient()
		return

	default:

		err := c.createClientSocket()
		if err != nil {
			return
		}

		// Retrieve the bet from the environment configuration
		bet := BetFromEnv(c.config.ID)

		// Send the bet to the client and check for errors
		if Send(c, bet) != nil {
			return
		}

		// Read a response message
		msg, err := bufio.NewReader(c.conn).ReadString('\n')
		c.conn.Close()

		if err != nil {
			log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("%v",
			msg,
		)
	}

}

// Gracefully shut down the client
func (c *Client) StopClient() {

	if c.conn != nil {
		c.conn.Close()
		//log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}
	log.Infof("action: shutdown | result: success | client_id: %v", c.config.ID)
}
