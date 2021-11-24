package main

import (
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"
)

type myHandler struct{}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//fmt.Print("Request For: " + r.URL.String())
	return
}

func HTTPServer(RuntimeConfig *RuntimeConfigStruct) error {

	mux := http.NewServeMux()

	mux.Handle("/", &myHandler{})

	mux.HandleFunc("/Update", func(w http.ResponseWriter, r *http.Request) { HTTPDUpdate(w, r, RuntimeConfig) })
	mux.HandleFunc("/GetIP", func(w http.ResponseWriter, r *http.Request) { HTTPDGetIP(w, r, RuntimeConfig) })

	var err error
	if len(RuntimeConfig.TLSPrivateKey) > 0 {
		err = http.ListenAndServeTLS(RuntimeConfig.ListenAddr, RuntimeConfig.TLSFullchain, RuntimeConfig.TLSPrivateKey, mux)
	} else {
		err = http.ListenAndServe(RuntimeConfig.ListenAddr, mux)
	}
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func HTTPDGetIP(w http.ResponseWriter, r *http.Request, RuntimeConfig *RuntimeConfigStruct) {
	if r.Method != "GET" {
		return
	}

	RuntimeConfig.Mutex.RLock()

	fmt.Fprintln(w, RuntimeConfig.IPv4)
	fmt.Fprintln(w, RuntimeConfig.IPv6)

	RuntimeConfig.Mutex.RUnlock()
}

func HTTPDUpdate(w http.ResponseWriter, r *http.Request, RuntimeConfig *RuntimeConfigStruct) {

	if r.Method != "GET" {
		fmt.Println("\nHTTPDUpdate expected 'GET' method.")
		return
	}

	fmt.Println("\nReceived API call:", r.URL.String())

	IPv4 := r.URL.Query().Get("IPv4")
	IPv6 := r.URL.Query().Get("IPv6")
	UserID := r.URL.Query().Get("UserID")
	Password := r.URL.Query().Get("Password")

	RuntimeConfig.Mutex.RLock()
	if UserID != RuntimeConfig.UserID {
		RuntimeConfig.Mutex.RUnlock()
		fmt.Println("Rejected UserID")

		// Random delay to frustrate timing attacks
		r, _ := rand.Int(rand.Reader, big.NewInt(300))
		time.Sleep(time.Duration(r.Int64())*time.Millisecond + 500)
		fmt.Fprintf(w, "Invalid UserID")

		return
	}

	if Password != RuntimeConfig.Password {
		RuntimeConfig.Mutex.RUnlock()
		fmt.Println("Rejected Password")

		// Random delay to frustrate timing attacks
		r, _ := rand.Int(rand.Reader, big.NewInt(300))
		time.Sleep(time.Duration(r.Int64())*time.Millisecond + 500)
		fmt.Fprintf(w, "Invalid Password")

		return
	}
	RuntimeConfig.Mutex.RUnlock()

	//	fmt.Println("Authentication succeeded.")

	if len(IPv4) > 0 {
		iip := net.ParseIP(IPv4)
		if iip != nil {
			if strings.Count(iip.String(), ".") != 3 {
				iip = nil
			}
		}
		if iip == nil {
			fmt.Println("Error: Malformed IPv4 Address")
			fmt.Fprintf(w, "Error: Malformed IPv4 Address")
			return
		}
		fmt.Println("Accepted IPv4 Address:", iip)
		RuntimeConfig.Mutex.Lock()
		if RuntimeConfig.IPv4.Equal(iip) {
			fmt.Println("IPv4 Address Unchanged")
			fmt.Fprintf(w, "unchg")
			RuntimeConfig.Mutex.Unlock()
			return
		}
		RuntimeConfig.IPv4 = iip
		if len(RuntimeConfig.IPv4Script.Trigger) == 0 {
			RuntimeConfig.IPv4Script.Trigger <- true
		}

		RuntimeConfig.Mutex.Unlock()
	}

	if len(IPv6) > 0 {
		iip := net.ParseIP(IPv6)
		if iip != nil {
			if strings.Count(iip.String(), ":") < 2 {
				iip = nil
			}
		}
		if iip == nil {
			fmt.Println("Error: Malformed IPv6 Address")
			fmt.Fprintf(w, "Error: Malformed IPv6 Address")
			return
		}
		fmt.Println("Accepted IPv6 Address:", iip)
		RuntimeConfig.Mutex.Lock()
		if RuntimeConfig.IPv6.Equal(iip) {
			fmt.Println("IPv6 Address Unchanged")
			fmt.Fprintf(w, "unchg")
			RuntimeConfig.Mutex.Unlock()
			return
		}
		RuntimeConfig.IPv6 = iip
		if len(RuntimeConfig.IPv6Script.Trigger) == 0 {
			RuntimeConfig.IPv6Script.Trigger <- true
		}

		RuntimeConfig.Mutex.Unlock()
	}
	fmt.Fprintf(w, "good")

	return
}
