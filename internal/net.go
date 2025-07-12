// Package internal
package internal

import (
	"log"
	"net"
	"sync"
	"time"
)

const (
	multicastAddr = "239.0.0.0:9999" // Адрес для multicast рассылки
	responseAddr  = "239.0.0.1:9998" // Адрес для ответов
	networkWait   = 2 * time.Second  // Время ожидания ответов
)

func GetLocalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}

// Главная функция, запускающая sender и listener в отдельных горутинах
func DiscoverPeers() ([]string, error) {
	peersFound := make(chan string)
	var peers []string
	var wg sync.WaitGroup

	// Запускаем горутины
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := sendDiscoveryMessage(); err != nil {
			log.Printf("Sender error: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := listenForPeers(peersFound); err != nil {
			log.Printf("Listener error: %v", err)
		}
	}()

	// Собираем результаты
	go func() {
		wg.Wait()
		close(peersFound)
	}()

	// Собираем уникальные адреса
	uniquePeers := make(map[string]struct{})
	timer := time.NewTimer(networkWait)
	defer timer.Stop()

	for {
		select {
		case peer, ok := <-peersFound:
			if !ok {
				return peers, nil
			}
			if _, exists := uniquePeers[peer]; !exists {
				uniquePeers[peer] = struct{}{}
				peers = append(peers, peer)
			}
		case <-timer.C:
			return peers, nil
		}
	}
}

// Функция для отправки multicast сообщения в сеть
func sendDiscoveryMessage() error {
	addr, err := net.ResolveUDPAddr("udp", multicastAddr)
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.Write([]byte("DISCOVER"))
	return err
}

// Функция для прослушивания ответов от других пиров
func listenForPeers(peers chan<- string) error {
	addr, err := net.ResolveUDPAddr("udp", responseAddr)
	if err != nil {
		return err
	}

	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(networkWait))

	buffer := make([]byte, 1024)
	for {
		n, src, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				return nil // Таймаут - не ошибка в нашем случае
			}
			return err
		}

		if string(buffer[:n]) == "DISCOVER" {
			// Отправляем ответ
			respConn, err := net.DialUDP("udp", nil, src)
			if err != nil {
				continue
			}
			respConn.Write([]byte("HERE"))
			respConn.Close()
		} else if string(buffer[:n]) == "HERE" {
			peers <- src.IP.String()
		}
	}
}
