package common

import (
	"bufio"
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
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	c.wg.Add(1)
	defer func() {
		c.wg.Done()
		c.StopClient()
	}()

	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		select {
		case <-c.stop:
			log.Infof("action: loop_terminated | result: success | client_id: %v", c.config.ID)
			return
		case sig := <-sigs:
			log.Infof("action: signal_received | signal: %v | client_id: %v", sig, c.config.ID)
			c.StopClient()
			return
		default:
			// Create the connection to the server in every loop iteration
			err := c.createClientSocket()
			if err != nil {
				return
			}

			// Send the message
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

			// Read the response
			msg, err := bufio.NewReader(c.conn).ReadString('\n')
			if err != nil {
				if err, ok := err.(net.Error); ok && err.Timeout() {
					log.Errorf("action: read_message | result: timeout | client_id: %v", c.config.ID)
				} else {
					log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
						c.config.ID,
						err,
					)
				}
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
	close(c.stop)

	if c.conn != nil {
		c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
	}
	log.Infof("action: client_shutdown | result: success | client_id: %v", c.config.ID)
}
