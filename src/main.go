package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"functions"
	"openexchange"
	"readconfig"
)

var clients = make(map[net.Conn]bool)  // connected clients
var ratesMap = make(map[string]string) // data to be updated constantly

func main() {
	retrievedConfig := readconfig.GetConfigInfo()
	service := retrievedConfig.Server.IP + ":" + retrievedConfig.Server.Port
	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)
	errResolveFlag := functions.CheckError(err)
	if errResolveFlag == 1 {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}

	listener, errListenTCP := net.ListenTCP("tcp", tcpAddr)
	//errListenFlag := functions.CheckError(errListenTCP)
	if functions.CheckError(errListenTCP) == 1 {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", errListenTCP.Error())
		os.Exit(1)
	}

	go retrieveData()

	for {
		conn, err := listener.Accept()
		listenerFlag := functions.CheckError(err)
		if listenerFlag == 1 {
			continue
		}

		go HandelRequest(conn)
	}

}

func HandelRequest(conn net.Conn) {
	retrievedConfig := readconfig.GetConfigInfo()
	remoteaddr := strings.Split(conn.RemoteAddr().String(), ":")

	IpPresent := false

	for _, value := range retrievedConfig.Client.Ips {
		if strings.TrimSpace(value) == strings.TrimSpace(remoteaddr[0]) {
			IpPresent = true
			break
		}
	}

	if IpPresent == false {
		err := errors.New("An Unauthorised IP " + conn.RemoteAddr().String() + " tried to access! ")
		errBodyFlag := functions.CheckError(err)
		if errBodyFlag == 1 {
			requestTime := time.Now()
			conn.Write([]byte("UNAUTHORISED" + "\n"))
			functions.RequestLogs("", requestTime.String(), conn.RemoteAddr().String(), "UNAUTHORISED")
			conn.Close()
		}
	} else {
		for {
			message, readErr := bufio.NewReader(conn).ReadString('\n')
			readErrFlag := functions.CheckError(readErr)
			if readErrFlag == 1 {
				break
			}

			retrievedConfig := readconfig.GetConfigInfo()

			if strings.Compare(strings.TrimSpace(message), strings.TrimSpace(retrievedConfig.Client.Token)) == 0 {
				clients[conn] = true
				conn.Write([]byte("AUTHORISED" + "\n"))
				ServeRates(conn)

			} else {

				err := errors.New("Authentication failed! Invalid Token.")
				errBodyFlag := functions.CheckError(err)
				if errBodyFlag == 1 {
					conn.Write([]byte("UNAUTHORISED" + "\n"))
				}
			}
		}

	}
}

func ServeRates(conn net.Conn) {
	retrievedConfig := readconfig.GetConfigInfo()
	timeoutDuration := retrievedConfig.Server.ServerTimeout * time.Second
	for {
		// Read operation will fail if no data is received after deadline.
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))

		message, errDeadline := bufio.NewReader(conn).ReadString('\n')
		if errDeadline != nil {
			fmt.Println(errDeadline)
			_, errWrite := conn.Write([]byte("TIMEOUT" + "\n"))
			functions.CheckError(errWrite)
			DeleteClient(conn)
			break
		}
		requestTime := time.Now()
		clientIP := conn.RemoteAddr()

		currency := strings.TrimSpace(message)
		if val, ok := ratesMap[currency]; ok {
			_, errWrite := conn.Write([]byte(val + "\n"))
			functions.RequestLogs(message, requestTime.String(), clientIP.String(), val)
			errWriteFlag := functions.CheckError(errWrite)
			if errWriteFlag == 1 {
				DeleteClient(conn)
				break
			}
		} else {
			_, errWrite := conn.Write([]byte("INVALID" + "\n"))
			functions.RequestLogs(message, requestTime.String(), clientIP.String(), "INVALID")
			errWriteFlag := functions.CheckError(errWrite)
			if errWriteFlag == 1 {
				DeleteClient(conn)
				break
			}
		}
	}
}

func retrieveData() {
	retrievedConfig := readconfig.GetConfigInfo()
	for {
		deadLine := time.Now().Add(retrievedConfig.Server.ServerTimeout * time.Second)
		fmt.Println("***Time: ", time.Now())

		for {
			var ratesRetrived openexchange.Result
			if retrievedConfig.Sources.OpenExchange.Status == true {
				returnedChannel := openexchange.GetOEdata()
				ratesRetrived = <-returnedChannel
			}

			if ratesRetrived.Err == nil {
				ratesMap = ratesRetrived.ResultMap
				time.Sleep(retrievedConfig.Sources.OpenExchange.Frequency * time.Second)
				break
			} else {
				keepTime := time.Now()
				if keepTime.After(deadLine) {
					fmt.Println("*************Timed out: ", ratesRetrived.Err)
					errTimeout := errors.New("Connection Timed Out due to " + ratesRetrived.Err.Error())
					_ = functions.CheckError(errTimeout)
					break
				} else {
					time.Sleep(2 * time.Second)
					continue
				}

			}

		}

	}
}

// Function to close connection and delete from clients map
func DeleteClient(conn net.Conn) {
	conn.Close()
	_, ok := clients[conn]
	if ok {
		delete(clients, conn)
	}
}
