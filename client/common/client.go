package common

import (
	"bufio"
	"fmt"
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
	timeout := 2 * time.Second
	conn, err := net.DialTimeout("tcp", c.config.ServerAddress, timeout)
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

func (c *Client) StartClientLoop(sigChan chan os.Signal) {

	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {

		select {
		case <-sigChan:
			c.StopClient()
			return
		default:

			err := c.createClientSocket()
			if err != nil {
				c.StopClient()
				return
			}
			_, err = fmt.Fprintf(
				c.conn,
				"[CLIENT %v] Message N°%v\n",
				c.config.ID,
				msgID,
			)
			if err != nil {
				log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				c.StopClient()
				return
			}

			msg, err := bufio.NewReader(c.conn).ReadString('\n')
			c.conn.Close()

			if err != nil {
				log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				c.StopClient()
				return
			}
			log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
				c.config.ID,
				msg,
			)
			// Wait a time between sending one message and the next one
			time.Sleep(c.config.LoopPeriod)
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

// Gracefully shut down the client
func (c *Client) StopClient() {
	if c.conn != nil {
		c.conn.Close()
		//log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}
	log.Infof("action: exit | result: success | client_id: %v", c.config.ID)
}
