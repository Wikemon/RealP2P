// Package internal
package internal

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"
)

const (
	BroadcastPort = 9999
	ChatPort      = 9998
	PingInterval  = 5 * time.Second
)

type Peer struct {
	IP   string
	Port int
}

type Message struct {
	Type    string // "ping", "pong", "chat"
	From    Peer
	Content string
}

func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println("Error getting local IP:", err)
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func GetBroadcastAddr() string {
	tmp := strings.Split(localIP, ".")
	tmp[3] = "255"
	return strings.Join(tmp, ".")
}

var (
	peers         = make(map[string]Peer)
	peersLock     sync.Mutex
	localIP       = GetLocalIP()
	BroadcastAddr = GetBroadcastAddr()
)

func StartBroadcastServer() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", BroadcastPort))
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error listening UDP:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Broadcast server started on port", BroadcastPort)

	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if n == 0 || err != nil {
			fmt.Println("Read error:", err)
			return
		}
		var msg Message
		err = json.Unmarshal(buffer[:n], &msg)
		if err != nil {
			fmt.Println("Error decoding message:", err)
			return
		}

		fmt.Println(msg)

		if msg.From.IP == localIP {
			continue
		}

		switch msg.Type {
		case "ping":
			// Отвечаем на ping сообщением pong
			response := Message{
				Type: "pong",
				From: Peer{IP: localIP, Port: ChatPort},
			}
			data, _ := json.Marshal(response)
			_, err = conn.WriteToUDP(data, clientAddr)
			if err != nil {
				fmt.Println("Error sending pong:", err)
			}
			fmt.Println("pong sended")

		case "pong":
			// Добавляем пир в список
			addPeer(msg.From)
		}
	}

}

func StartBroadcastClient() {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", BroadcastAddr, BroadcastPort))
	if err != nil {
		fmt.Println("Error resolving broadcast address:", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		fmt.Println("Error dialing UDP:", err)
		return
	}
	defer conn.Close()

	// Периодически отправляем ping сообщения
	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for range ticker.C {
		msg := Message{
			Type: "ping",
			From: Peer{IP: localIP, Port: ChatPort},
		}
		data, _ := json.Marshal(msg)
		_, err := conn.Write(data)
		if err != nil {
			fmt.Println("Error sending ping:", err)
		}
		fmt.Println("ping sended")
	}
}

// func StartChatServer() {
// 	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", ChatPort))
// 	if err != nil {
// 		fmt.Println("Error starting chat server:", err)
// 		return
// 	}
// 	defer listener.Close()

// 	fmt.Println("Chat server started on port", ChatPort)

// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			fmt.Println("Error accepting connection:", err)
// 			continue
// 		}

// 		go handleChatConnection(conn)
// 	}
// }

// func handleChatConnection(conn net.Conn) {
// 	defer conn.Close()

// 	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
// 	peer := Peer{IP: remoteAddr.IP.String(), Port: remoteAddr.Port}

// 	var msg Message
// 	decoder := json.NewDecoder(conn)
// 	err := decoder.Decode(&msg)
// 	if err != nil {
// 		fmt.Println("Error decoding chat message:", err)
// 		return
// 	}

// 	if msg.Type == "chat" {
// 		fmt.Printf("\n[%s] %s\n> ", peer.IP, msg.Content)
// 	}
// }

func addPeer(peer Peer) {
	peersLock.Lock()
	defer peersLock.Unlock()

	key := fmt.Sprintf("%s:%d", peer.IP, peer.Port)
	if _, exists := peers[key]; !exists && peer.IP != localIP {
		peers[key] = peer
		fmt.Printf("Discovered new peer: %s\n", key)
	}
}

// func sendChatMessage(peer Peer, text string) {
// 	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", peer.IP, peer.Port))
// 	if err != nil {
// 		fmt.Println("Error connecting to peer:", err)
// 		return
// 	}
// 	defer conn.Close()

// 	msg := Message{
// 		Type:    "chat",
// 		From:    Peer{IP: localIP, Port: ChatPort},
// 		Content: text,
// 	}

// 	encoder := json.NewEncoder(conn)
// 	err = encoder.Encode(msg)
// 	if err != nil {
// 		fmt.Println("Error sending chat message:", err)
// 	}
// }

// func StartUserInterface() {
// 	reader := bufio.NewReader(os.Stdin)

// 	for {
// 		fmt.Print("> ")
// 		input, err := reader.ReadString('\n')
// 		if err != nil {
// 			fmt.Println("Error reading input:", err)
// 			continue
// 		}

// 		input = strings.TrimSpace(input)

// 		// Специальные команды
// 		if input == "/list" {
// 			peersLock.Lock()
// 			fmt.Println("\nConnected peers:")
// 			for _, peer := range peers {
// 				fmt.Printf("- %s:%d\n", peer.IP, peer.Port)
// 			}
// 			peersLock.Unlock()
// 			continue
// 		}

// 		if strings.HasPrefix(input, "/send ") {
// 			parts := strings.SplitN(input, " ", 3)
// 			if len(parts) != 3 {
// 				fmt.Println("Usage: /send <IP> <message>")
// 				continue
// 			}

// 			ip := parts[1]
// 			message := parts[2]

// 			peersLock.Lock()
// 			var foundPeer *Peer
// 			for _, peer := range peers {
// 				if peer.IP == ip {
// 					foundPeer = &peer
// 					break
// 				}
// 			}
// 			peersLock.Unlock()

// 			if foundPeer != nil {
// 				sendChatMessage(*foundPeer, message)
// 				fmt.Printf("Message sent to %s\n", ip)
// 			} else {
// 				fmt.Printf("Peer %s not found\n", ip)
// 			}
// 			continue
// 		}

// 		// Отправка сообщения всем пирам
// 		if input != "" {
// 			peersLock.Lock()
// 			for _, peer := range peers {
// 				go sendChatMessage(peer, input)
// 			}
// 			peersLock.Unlock()
// 		}
// 	}
// }
