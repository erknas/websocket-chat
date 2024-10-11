package main

import (
	"io"
	"log"
	"log/slog"
	"net/http"

	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]bool
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWS(ws *websocket.Conn) {
	log.Println(ws.RemoteAddr())

	s.conns[ws] = true

	s.read(ws)
}

func (s *Server) read(ws *websocket.Conn) {
	buf := make([]byte, 1024)

	for {
		n, err := ws.Read(buf)
		if err != nil {
			slog.Error("reading error", "err", err)
		}
		if err == io.EOF {
			break
		}
		msg := buf[:n]
		s.broadcast(msg)
	}
}

func (s *Server) broadcast(b []byte) {
	for ws := range s.conns {
		go func(ws *websocket.Conn) {
			if _, err := ws.Write(b); err != nil {
				slog.Error("writing error", "err", err)
			}
		}(ws)
	}
}

func main() {
	server := NewServer()

	http.Handle("/ws", websocket.Handler(server.handleWS))
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}
}
