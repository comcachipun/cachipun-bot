package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Jugada int8

type Outcome int8

var TIMES_TO_WIN = 10

var Jugadas = [][]Jugada{
	{PIEDRA, PAPEL, PAPEL, TIJERAS, TIJERAS, PIEDRA, TIJERAS, PAPEL, TIJERAS, PAPEL},
	{PAPEL, TIJERAS, PAPEL, PAPEL, PIEDRA, PAPEL, PAPEL, TIJERAS, PAPEL, TIJERAS},
	{TIJERAS, PIEDRA, PAPEL, TIJERAS, PAPEL, PIEDRA, PIEDRA, PAPEL, TIJERAS, PAPEL},
}

var NombresJugadas map[Jugada]string = map[Jugada]string{
	PIEDRA:  "Piedra",
	PAPEL:   "Papel",
	TIJERAS: "Tijeras",
}

const (
	PIEDRA  Jugada = 1
	PAPEL   Jugada = 2
	TIJERAS Jugada = 3
)

const (
	LOSE Outcome = -1
	DRAW Outcome = 0
	WIN  Outcome = 1
)

func (j Jugada) Play(j2 Jugada) Outcome {
	switch j - j2 {
	case 0:
		return DRAW
	case 1, -2:
		return WIN
	default:
		return LOSE
	}
}

func Handle(conn net.Conn) {
	defer conn.Close()
	for {
		conn.SetDeadline(time.Now().Add(240 * time.Second))
		msg := make([]byte, 2)
		WelcomeMessage := fmt.Sprintf(
			"**[comcachipun/cachipun-bot]**\n¡Juguemos Cachipún!\nSi ganas %d veces te ganas un premio.\n\n",
			TIMES_TO_WIN,
		)
		conn.Write([]byte(WelcomeMessage))
		wins := 0
		turn := 1
		gameStyle := rand.Int() % 3
		for {
			if wins == TIMES_TO_WIN {
				conn.Write([]byte(fmt.Sprintf("¡Ganaste!\n\nLa Flag es %s\n\n", os.Getenv("FLAG"))))
				return
			}
			conn.Write([]byte(fmt.Sprintf("[Turno %d / Perdí: %d]\nEscribe tu opción:\n1) PIEDRA\n2) PAPEL\n3) TIJERAS\n\n>>> ", turn, wins)))
			n, err := conn.Read(msg)
			if err != nil {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					conn.Write([]byte("Me aburrí de esperarte.\n"))
					conn.Write([]byte("¡Adiós!\n"))
				} else {
					log.Printf("error with connection %s: %s", conn.RemoteAddr(), err)
				}
				return
			}
			resp, err := strconv.Atoi(strings.TrimSpace(string(msg[:n])))
			if err != nil || resp < 1 || resp > 3 {
				if errors.Is(err, os.ErrDeadlineExceeded) {
					conn.Write([]byte("Opción inválida. No quiero jugar más.\n"))
					conn.Write([]byte("¡Adiós!\n"))
				} else {
					log.Printf("error with connection %s: %s", conn.RemoteAddr(), err)
				}
				return
			}
			opt := Jugada(resp)
			jugada := Jugadas[gameStyle][(turn-1)%TIMES_TO_WIN]
			conn.Write([]byte(fmt.Sprintf("Jugaste %s.\n", NombresJugadas[opt])))
			conn.Write([]byte(fmt.Sprintf("Yo jugué %s.\n\n", NombresJugadas[jugada])))
			switch opt.Play(jugada) {
			case WIN:
				conn.Write([]byte("¡Ganaste esta!\n\n\n"))
				wins += 1
			case LOSE:
				conn.Write([]byte("¡Perdiste!\n¡Adiós!\n\n\n"))
				return
			case DRAW:
				conn.Write([]byte("Empatamos :(\n¡Otra!\n\n\n"))
			}
			turn++
		}
	}
}

func main() {
	addr := os.Getenv("ECHO_ADDRESS")
	if len(addr) == 0 {
		addr = "0.0.0.0:33333"
	}
	s, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}
	log.Printf("Listening in %s", addr)
	for {
		conn, err := s.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %s", err)
		}
		go Handle(conn)
	}
}
