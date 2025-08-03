package main

import (
	"P2P/internal"
	"fmt"
)

func main() {
	// Получаем локальный IP
	fmt.Printf("Starting P2P chat. Your IP: %s\n", internal.GetLocalIP())
	fmt.Printf("BroadcastAddr: %s\n", internal.GetBroadcastAddr())

	// Запускаем сервер для приема чат-сообщений
	//go internal.StartChatServer()

	// Запускаем broadcast сервер для обнаружения пиров
	go internal.StartBroadcastServer()

	// Запускаем broadcast клиент для оповещения о себе
	go internal.StartBroadcastClient()

	for {
		fmt.Scan()
	}
	//internal.StartUserInterface()
}
