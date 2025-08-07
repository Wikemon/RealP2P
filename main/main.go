package main

import (
	"P2P/internal"
	"fmt"
)

func main() {
	// Получаем локальный IP
	fmt.Printf("Starting P2P chat. Your IP: %s\n", internal.GetLocalIP())

	// Запускаем UDP-broadcast сервер для обнаружения пиров
	go internal.StartUDPServer()

	// Запускаем TCP сервер для обнаружения пиров
	go internal.StartTCPServer()

	// Запускаем broadcast клиент для оповещения о себе
	go internal.StartBroadcastClient()

	internal.StartUserInterface()
}
