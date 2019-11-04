package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/jacobsa/go-serial/serial"
)

type clientList struct {
	clients map[*client]*client
	m       sync.Mutex
}

func newClientList() *clientList {
	c := clientList{}
	c.clients = make(map[*client]*client)
	return &c
}

func (cList *clientList) Add(c *client) {
	cList.m.Lock()
	defer cList.m.Unlock()
	cList.clients[c] = c
}

func (cList *clientList) Remove(c *client) {
	cList.m.Lock()
	defer cList.m.Unlock()
	delete(cList.clients, c)
}

type client struct {
	in  chan string
	out io.Writer
	//out     chan string
	conn    net.Conn
	quit    chan bool
	clients *clientList
}

func main() {
	opt := serial.OpenOptions{
		PortName:        "/dev/valvecan",
		BaudRate:        9600,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}
	port, err := serial.Open(opt)
	if err != nil {
		log.Fatalf("Failed to open serial port: %v", err)
	}
	defer port.Close()
	valveCh := make(chan string, 1)
	cList := newClientList()
	go portReader(port, cList, valveCh)
	ln, err := net.Listen("tcp", ":41000")
	if err != nil {
		panic("Unable to listen")
	}
	for {
		conn, err := ln.Accept()
		cli := client{
			in:      make(chan string),
			out:     port,
			conn:    conn,
			quit:    make(chan bool),
			clients: cList,
		}
		cList.Add(&cli)
		if err != nil {
			panic("Accept")
		}
		go clientWriter(&cli)
		go clientReader(&cli)
	}
}

func clientReader(cli *client) {
	reader := bufio.NewScanner(cli.conn)
	for reader.Scan() {
		s := reader.Text()
		fmt.Println(s)
		_, err := io.WriteString(cli.out, s+"\n")
		if err != nil {
			break
		}
	}
	fmt.Println("Quitting")
	cli.clients.Remove(cli)
	cli.quit <- true
}

func clientWriter(cli *client) {
	writer := bufio.NewWriter(cli.conn)
	defer close(cli.in)
	for {
		select {
		case str := <-cli.in:
			writer.WriteString(str)
			writer.Flush()
		case <-cli.quit:
			return
		}
	}
}

func portReader(port io.ReadWriteCloser, cList *clientList, valveCh chan string) {
	scanner := bufio.NewScanner(port)
	for scanner.Scan() {
		str := scanner.Text()
		//fmt.Println(str)
		cList.m.Lock()
		//fmt.Printf("Clients: %d\n", len(cList.clients))
		for cli, _ := range cList.clients {
			cli.in <- str + "\n"
		}
		cList.m.Unlock()
	}
}
