package client

import (
	"fmt"
	"net"
	"os"
	"tp1/pkg/logs"
	"tp1/pkg/utils/id"
	"tp1/pkg/utils/io"
)

func (c *Client) openResultsFile() error {
	fileName := fmt.Sprintf("/app/data/results_%s.txt", c.clientId)
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		logs.Logger.Errorf("Error opening results file: %v", err)
		return err
	}

	c.resultsFile = file
	return nil
}

func (c *Client) fetchClientID(address string) error {
	idsPort := c.cfg.String("gateway.ids_port", "5053")
	idsFullAddress := address + ":" + idsPort
	idConn, err := net.Dial("tcp", idsFullAddress)
	if err != nil {
		logs.Logger.Errorf("ID connection error: %v", err)
		return err
	}
	defer idConn.Close()

	clientIdBuffer := make([]byte, id.ClientIdLen)
	if err := io.ReadFull(idConn, clientIdBuffer, id.ClientIdLen); err != nil {
		logs.Logger.Errorf("Error reading client ID: %v", err)
		return err
	}
	c.clientId = id.DecodeClientId(clientIdBuffer)
	logs.Logger.Infof("Assigned client ID: %s", c.clientId)
	return nil
}

func (c *Client) handleSigterm() {
	logs.Logger.Info("Sigterm Signal Received... Shutting down")
	c.stoppedMutex.Lock()
	c.stopped = true
	c.stoppedMutex.Unlock()
}

func (c *Client) Close(gamesConn net.Conn, reviewsConn net.Conn) {
	gamesConn.Close()
	reviewsConn.Close()
	c.resultsFile.Close()
	close(c.sigChan)
}

func (c *Client) sendClientID(conn net.Conn) error {
	clientIdBuf := id.EncodeClientId(c.clientId)
	if err := io.SendAll(conn, clientIdBuf); err != nil {
		logs.Logger.Errorf("Error sending client ID: %s", err)
		return err
	}
	return nil
}

func readAck(conn net.Conn) error {
	ackBuf := make([]byte, 1)
	_, err := conn.Read(ackBuf)
	if err != nil {
		return fmt.Errorf("error reading ACK: %w", err)
	}

	return nil
}
