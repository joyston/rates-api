package functions

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"readconfig"
)

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

// Logs the error passed if not nil
func RequestLogs(currencyRequest string, requestTime string, clientIP string, response string) {
	retrievedConfig := readconfig.GetConfigInfo()
	if retrievedConfig.Miscellaneous.DebugRequest == true {
		year := time.Now().Year()
		month := time.Now().Month()
		filePath := "../tmp/RequestLogs-" + strconv.Itoa(year) + "-" + strconv.Itoa(int(month))

		t := time.Now()
		logString := "\n\nLog Time: " + t.String() + "\nRequest Time: " + requestTime + "\nClient IP: " + clientIP + "\nName of currency requesting: " + currencyRequest + "\nResponse: " + response

		ifFileExist, _ := exists(filePath)
		if ifFileExist == false {
			ioutil.WriteFile(filePath, []byte(logString), 0644)
		} else {
			f, _ := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
			f.WriteString(logString)
			defer f.Close()
		}
	}
}

func SourceLogs(requestTime string, sourceName string, responceCode int, response string) {
	retrievedConfig := readconfig.GetConfigInfo()
	if retrievedConfig.Miscellaneous.DebugSource == true {
		year := time.Now().Year()
		month := time.Now().Month()
		filePath := "../tmp/SourceLogs-" + strconv.Itoa(year) + "-" + strconv.Itoa(int(month))

		t := time.Now()
		logString := "\n\nLog Time: " + t.String() + "\nSource Request Time: " + requestTime + "\nSource Name: " + sourceName + "\nHTTP response code: " + strconv.Itoa(responceCode) + "\nResponse: " + response

		ifFileExist, _ := exists(filePath)
		if ifFileExist == false {
			ioutil.WriteFile(filePath, []byte(logString), 0644)
		} else {
			f, _ := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
			f.WriteString(logString)
			defer f.Close()
		}
	}
}

// Logs the error passed if not nil
func CheckError(e error) int {

	if e != nil {
		retrievedConfig := readconfig.GetConfigInfo()
		if retrievedConfig.Miscellaneous.ServerDebug == true || strings.HasPrefix(e.Error(), "An Unauthorised IP") == true || strings.HasPrefix(e.Error(), "Connection Timed Out") == true {
			year := time.Now().Year()
			month := time.Now().Month()
			filePath := "../tmp/ErrorLogs-" + strconv.Itoa(year) + "-" + strconv.Itoa(int(month))

			t := time.Now()
			errorString := "\n\nTime: " + t.String() + "\n" + e.Error()

			ifFileExist, errExist := exists(filePath)
			if ifFileExist == false {

				err := ioutil.WriteFile(filePath, []byte(errorString), 0644)
				if err != nil {
					SimpleMail(err, e)
				}

			} else {
				if errExist == nil {

					f, errOpen := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600)
					if errOpen != nil {
						SimpleMail(errOpen, e)
					}

					if _, err := f.WriteString(errorString); err != nil {
						SimpleMail(err, e)
					}

					defer f.Close()

				} else {
					SimpleMail(errExist, e)
				}
			}

			if strings.HasPrefix(e.Error(), "Connection Timed Out") == true {
				SimpleMail(e, e)
			}
		}

		return 1
	}
	return 0
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func ToFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output
}

// Function to send errors via mail
func SimpleMail(e error, logerror error) {

	from := mail.Address{"*********", "*************"}
	to := mail.Address{"", "***************"}

	var subj, body string
	if strings.HasPrefix(e.Error(), "Connection Timed Out") == true {
		subj = "Bouncer error due to Connection"
		body = e.Error()
	} else {
		subj = "Bouncer error while logging."
		body = "An error occurred while logging:\n" + e.Error() + "\nError that was to be logged:\n" + logerror.Error()
	}

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	servername := "******************"

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", from.Address, "RoaeIYdjpe3L", host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	c, err := smtp.Dial(servername)
	errFlag := CheckError(err)
	if errFlag == 1 {
		return
	}

	c.StartTLS(tlsconfig)

	// Auth
	if err = c.Auth(auth); err != nil {
		errFlag = CheckError(err)
		if errFlag == 1 {
			return
		}
	}

	// Set the sender and recipient first
	if err := c.Mail(from.Address); err != nil {
		errFlag = CheckError(err)
		if errFlag == 1 {
			return
		}
	}
	if err := c.Rcpt(to.Address); err != nil {
		errFlag = CheckError(err)
		if errFlag == 1 {
			return
		}
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		errFlag = CheckError(err)
		if errFlag == 1 {
			return
		}
	}

	_, err = wc.Write([]byte(message))
	if err != nil {
		errFlag = CheckError(err)
		if errFlag == 1 {
			return
		}
	}

	err = wc.Close()
	if err != nil {
		errFlag = CheckError(err)
		if errFlag == 1 {
			return
		}
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		errFlag = CheckError(err)
		if errFlag == 1 {
			return
		}
	}
}
