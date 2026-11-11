package fiberinterstate

import (
	"bufio"
	"fmt"
	//"log"
	"net"
	"os"
	"os/signal"
	//"path/filepath"
	//"strings"
	"sync"
	"syscall"
	"time"
)

// -------------------------------------------------------------------------------------------------------------------------------------------------

const (
	sendQueueSize = 64
	readTimeout   = 10 * time.Minute

	broadcastInterval = 3 * time.Second
)

// -------------------------------------------------------------------------------------------------------------------------------------------------

type client struct {
	id     int
	conn   *net.UnixConn
	outbox chan []byte
	done   chan struct{}
	once   sync.Once
}

// closeOnce close connection one time
func (c *client) close() {
	c.once.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

// -------------------------------------------------------------------------------------------------------------------------------------------------

type hub struct {
	mu      sync.RWMutex
	clients map[int]*client
	nextID  int
}

func newHub() *hub {
	return &hub{clients: make(map[int]*client)}
}

func (h *hub) add(c *client) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.nextID++
	c.id = h.nextID
	h.clients[c.id] = c
	return c.id
}

func (h *hub) remove(c *client) {
	h.mu.Lock()
	if _, ok := h.clients[c.id]; ok {
		delete(h.clients, c.id)
	}
	h.mu.Unlock()
	c.close()
}

func (h *hub) broadcast(msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, c := range h.clients {
		select {
		case c.outbox <- msg:
		case <-c.done:
		default:
			// queue is full - dropping packets, not blocking broadcast

			EventLogAdd("IPCServer", "hub.broadcasting", fmt.Sprintf("line of client %d is full, message avoid\n", c.id))
		}
	}
}

// -------------------------------------------------------------------------------------------------------------------------------------------------

// -------------------------------------------------------------------------------------------------------------------------------------------------

type IPCServer struct {
	onmessage_funcs []func(h *hub, c *client, id int, msg string)
}

func NewIPCServer() *IPCServer {

	return &IPCServer{}
}

func (ipcs *IPCServer) OnMessage(fu func(h *hub, c *client, id int, msg string)) {

	ipcs.onmessage_funcs = append(ipcs.onmessage_funcs, fu)
}

func (ipcs *IPCServer) Start() {

	EventLogAdd("IPCServer", "Start", "starting")

	path := socketPath()

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		EventLogAdd("IPCServer", "Start", "Remove old socket file error:", path, " ", err.Error())
	}

	addr := &net.UnixAddr{Name: path, Net: "unix"}
	listener, err := net.ListenUnix("unix", addr)
	if err != nil {

		EventLogAdd("IPCServer", "Start", fmt.Sprintf("net.ListenUnix error %s: %v", path, err))
		os.Exit(1)
	}

	EventLogAdd("IPCServer", "Start", fmt.Sprintf(" server listen non-block unix-socket: %s", path))

	h := newHub()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-shutdown

		EventLogAdd("IPCServer", "Start", fmt.Sprintf(" got a signal %v, finishing...", sig))
		listener.Close()
	}()

	for {
		conn, err := listener.AcceptUnix()
		if err != nil {
			// listener is closed - go to exit
			if opErr, ok := err.(*net.OpError); ok && opErr.Err.Error() == "use of closed network connection" {
				break
			}

			// nonblocking accept can return temporary errors - wait and repeat
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				time.Sleep(10 * time.Millisecond)
				continue
			}

			EventLogAdd("IPCServer", "Start", fmt.Sprintf(" listener accept error: %v", err))
			break
		}

		go ipcs.handleClient(h, conn)
	}

	h.mu.RLock()
	for _, c := range h.clients {
		c.close()
	}
	h.mu.RUnlock()

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {

		EventLogAdd("IPCServer", "Start", fmt.Sprintf("can not remove socket %s: %v", path, err))
	}

	EventLogAdd("IPCServer", "Start", " server is stopped")

}

func (ipcs *IPCServer) handleClient(h *hub, conn *net.UnixConn) {
	c := &client{
		conn:   conn,
		outbox: make(chan []byte, sendQueueSize),
		done:   make(chan struct{}),
	}
	id := h.add(c)

	EventLogAdd("IPCServer", "handleClient", fmt.Sprintf(" client %d connected (count of clients: %d)", id, ipcs.clientCount(h)))

	done := c.done

	go func() {
		for {
			select {
			case msg := <-c.outbox:
				if err := ipcs.writeLineServer(conn, msg); err != nil {

					EventLogAdd("IPCServer", "handleClient", fmt.Sprintf(" client %d write error: %v", id, err))
					h.remove(c)
					return
				}
			case <-done:
				return
			}
		}
	}()

	ipcs.readLoopServer(h, c, id)

	h.remove(c)

	EventLogAdd("IPCServer", "handleClient", fmt.Sprintf("[server] client %d disconnected (count of clients: %d)", id, ipcs.clientCount(h)))

}

func (ipcs *IPCServer) readLoopServer(h *hub, c *client, id int) {

	reader := bufio.NewReader(c.conn)

	for {
		// Read deadline: even a non-blocking socket is polled under the hood;
		// the deadline ensures that a "dead" client won't keep a goroutine occupied indefinitely.
		if err := c.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			return
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			// EOF - client disconnected. Another error - connection problem.
			if err.Error() != "EOF" {
				select {
				case <-c.done: // connection already closed by writer
				default:

					EventLogAdd("IPCServer", "readLoopServer", fmt.Sprintf(" client %d read error: %v", id, err))
				}
			}
			return
		}

		text := ipcs.trimNewline(line)
		if text == "" {
			continue
		}

		// ---------------------------------------------------------------------------------------------------------------------

		for _, fu := range ipcs.onmessage_funcs {

			fu(h, c, id, text)
		}

		/*
			text_command := strings.Split(text, ":")
			if len(text_command) == 2 {

				fmt.Println("text_command=", text_command)

				if text_command[0] == "GET" {
					ret, err := BadgerGetOne(text_command[1])

					if err == nil {
						c.outbox <- ret
					} else {

						fmt.Println("return err=", err.Error(), " for val=", text_command[1])

						c.outbox <- []byte("(empty)")
					}
				}

				if text_command[0] == "DEL" {

					BadgerDel(text_command[1])

					c.outbox <- []byte("(del done)")
				}

			}

			if len(text_command) == 3 {

				fmt.Println("text_command=", text_command)

				if text_command[0] == "SET" {

					BadgerSet(text_command[1], text_command[2])

					c.outbox <- []byte("(set done)")
				}

			}
		*/

		EventLogAdd("IPCServer", "readLoopServer", fmt.Sprintf(" client %d: %q", id, text))

		// ---------------------------------------------------------------------------------------------------------------------

		// Echo to client (nonblock: to queue).

		select {
		case c.outbox <- []byte(fmt.Sprintf("echo[%d]> %s", id, text)):
		case <-c.done:
			return
		}
	}
}

func (ipcs *IPCServer) writeLineServer(conn *net.UnixConn, msg []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeTimeout)); err != nil {
		return err
	}
	data := append(msg, '\n')
	_, err := conn.Write(data)
	return err
}

func (ipcs *IPCServer) clientCount(h *hub) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (ipcs *IPCServer) trimNewline(s string) string {
	for len(s) > 0 && (s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}
