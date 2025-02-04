package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

var (
	clients []net.Conn
	mu      sync.Mutex
	users   []User
	currId  = 0
)

type User struct {
	ID         int
	Username   string
	Messages   []string
	Joined     bool
	PublicKey  *rsa.PublicKey
	PrivateKey *rsa.PrivateKey
}

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide a port number!")
		return
	}

	PORT := ":" + arguments[1]
	l, err := net.Listen("tcp4", PORT)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c)
	}
}

func handleConnection(c net.Conn) {
	mu.Lock()
	privKey, pubKey, err := generateRSAKeys(2048)
	if err != nil {
		fmt.Println("Error generating keys:", err)
		return
	}

	username := fmt.Sprintf("User%d", currId)
	newUser := NewUser(currId, username, make([]string, 0), true, pubKey, privKey)
	users = append(users, *newUser)
	clients = append(clients, c)
	currId += 1
	mu.Unlock()

	fmt.Println("CONNECTED")
	defer c.Close()

	c.Write([]byte("Please enter your username: "))
	usernameInput, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		fmt.Println(err)
		return
	}

	usernameInput = strings.TrimSpace(usernameInput)
	if usernameInput != "" {
		mu.Lock()
		users[currId-1].Username = usernameInput
		mu.Unlock()
	}

	for _, u := range users {
		if u.ID != currId-1 {
			encodedPubKey := base64.StdEncoding.EncodeToString(u.PublicKey.N.Bytes())
			fmt.Fprintf(c, "Public Key of %s: %s\n", u.Username, encodedPubKey)
		}
	}

	for _, client := range clients {
		if client != c {
			fmt.Fprintf(client, "\x1b[1;31m%s has joined the chat!\x1b[0m\n", usernameInput)
		}
	}

	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			break
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			for _, client := range clients {
				if client != c {
					userCount := len(clients)
					msg := fmt.Sprintf("\x1b[1;31mUser %s disconnected, users in the channel: %d\x1b[0m\n", usernameInput, userCount-1)
					client.Write([]byte(msg))
				}
			}
			break
		}

		for _, u := range users {
			if u.Username != usernameInput {
				encryptedMsg, err := encryptMessage(u.PublicKey, temp)
				if err != nil {
					fmt.Println("Error encrypting message:", err)
					continue
				}

				for _, client := range clients {
					if client != c {
						decryptedMsg, err := decryptMessage(u.PrivateKey, encryptedMsg)
						if err != nil {
							fmt.Println("Error decrypting message:", err)
							continue
						}
						fmt.Fprintf(client, "\x1b[34m%s\x1b[0m: %s\n", usernameInput, decryptedMsg)
					}
				}
			}
		}
	}

	mu.Lock()
	for i, client := range clients {
		if client == c {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	mu.Unlock()
}

func generateRSAKeys(bits int) (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}
	return privKey, &privKey.PublicKey, nil
}

func encryptMessage(publicKey *rsa.PublicKey, message string) ([]byte, error) {
	encryptedMessage, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, publicKey, []byte(message), nil)
	if err != nil {
		return nil, err
	}
	return encryptedMessage, nil
}

func decryptMessage(privateKey *rsa.PrivateKey, encryptedMessage []byte) (string, error) {
	decryptedMessage, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, privateKey, encryptedMessage, nil)
	if err != nil {
		return "", err
	}
	return string(decryptedMessage), nil
}

func NewUser(id int, username string, messages []string, joined bool, pubKey *rsa.PublicKey, privKey *rsa.PrivateKey) *User {
	return &User{
		ID:         id,
		Username:   username,
		Messages:   messages,
		Joined:     joined,
		PublicKey:  pubKey,
		PrivateKey: privKey,
	}
}

