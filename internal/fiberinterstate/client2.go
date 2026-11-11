package fiberinterstate

import (
	"bufio"
	"context"
	"fmt"
	//"log"
	"net"
	"os"
	//"path/filepath"
	"errors"
	"strconv"
	"strings"
	"time"
)

func ClientImmediateRequest(ctx context.Context, msg_ []byte) (string, error) {

	clientID := strconv.Itoa(os.Getpid())

	path := socketPath()
	addr := &net.UnixAddr{Name: path, Net: "unix"}

	EventLogAdd("ClientImmediateRequest", fmt.Sprintf("[client %s] connecting to %s ...", clientID, path))

	ctxDial, cancel := context.WithTimeout(ctx, dialTimeout)
	defer cancel()

	d := net.Dialer{Timeout: dialTimeout}
	conn, err := d.DialUnix(ctxDial, "unix", nil, addr)
	if err != nil {

		EventLogAdd("ClientImmediateRequest", fmt.Sprintf("[client %s] connect error: %v", clientID, err))
		os.Exit(1)
	}
	defer conn.Close()

	EventLogAdd("ClientImmediateRequest", fmt.Sprintf("[client %s] connected", clientID))

	// --------------------------------------------------------------------------------------

	msg := string(msg_)

	if err := writeLineClient(conn, msg); err != nil {

		EventLogAdd("ClientImmediateRequest", fmt.Sprintf("[client %s] write error: %v", clientID, err))
		return "", errors.New(fmt.Sprintf("[client %s] write error: %v", clientID, err))
	}

	reader := bufio.NewReader(conn)

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Minute)); err != nil {

		EventLogAdd("ClientImmediateRequest", fmt.Sprintf("[client %s] SetReadDeadline: %v", clientID, err))
		return "", errors.New(fmt.Sprintf("[client %s] SetReadDeadline: %v", clientID, err))
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		if os.IsTimeout(err) {
			panic(err)

		}

		EventLogAdd("ClientImmediateRequest", fmt.Sprintf("[client %s] server disconnected (%v)", clientID, err))
		os.Exit(0)
	}

	line = strings.TrimRight(line, "\r\n")

	EventLogAdd("ClientImmediateRequest", fmt.Sprintf("[client %s] >>> %s\n", clientID, line))

	return line, nil

}

func writeLineClient(conn *net.UnixConn, msg string) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return err
	}
	_, err := conn.Write([]byte(msg + "\n"))
	return err
}
