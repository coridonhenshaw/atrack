package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ConfigurationFile struct {
	XMLName        xml.Name       `xml:"Atrack"`
	ListenAddress  ListenAddress  `xml:"ListenAddress"`
	TLSCredentials TLSCredentials `xml:"TLSCredentials"`
	PIDFile        PIDFile        `xml:"PIDFile"`
	Credentials    Credentials    `xml:"Credentials"`
	IPv4Commands   []IPv4Commands `xml:"IPv4Commands"`
	IPv6Commands   []IPv6Commands `xml:"IPv6Commands"`
}

type ListenAddress struct {
	Address string `xml:"Value,attr"`
}

type Credentials struct {
	UserID   string `xml:"UserID,attr"`
	Password string `xml:"Password,attr"`
}

type TLSCredentials struct {
	Fullchain  string `xml:"Fullchain,attr"`
	PrivateKey string `xml:"PrivateKey,attr"`
}

type PIDFile struct {
	PIDFile string `xml:"Value,attr"`
}

type IPv4Commands struct {
	Command string `xml:"Exec,attr"`
	Timeout int64  `xml:"Timeout,attr"`
}

type IPv6Commands struct {
	Command string `xml:"Exec,attr"`
	Timeout int64  `xml:"Timeout,attr"`
}

func LoadConfig(Filename string) (*RuntimeConfigStruct, error) {
	var RuntimeConfig RuntimeConfigStruct

	RuntimeConfig.IPv4Script.Trigger = make(chan bool, 16)
	RuntimeConfig.IPv6Script.Trigger = make(chan bool, 16)
	RuntimeConfig.ConfigurationFilename = Filename

	err := ReloadConfig(&RuntimeConfig)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)

	go func() {
		for {
			_ = <-c
			fmt.Println("Configuration hot reload.")
			ReloadConfig(&RuntimeConfig)
		}
	}()

	return &RuntimeConfig, err
}

func ReloadConfig(RuntimeConfig *RuntimeConfigStruct) error {

	RuntimeConfig.Mutex.Lock()

	xmlFile, err := os.Open(RuntimeConfig.ConfigurationFilename)
	if err != nil {
		log.Fatal(err)
	}
	defer xmlFile.Close()

	CFBytes, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		log.Fatal(err)
	}

	var CF ConfigurationFile
	err = xml.Unmarshal(CFBytes, &CF)
	if err != nil {
		log.Fatal(err)
	}

	RuntimeConfig.ListenAddr = CF.ListenAddress.Address
	RuntimeConfig.UserID = CF.Credentials.UserID
	RuntimeConfig.Password = CF.Credentials.Password
	RuntimeConfig.TLSFullchain = CF.TLSCredentials.Fullchain
	RuntimeConfig.TLSPrivateKey = CF.TLSCredentials.PrivateKey
	RuntimeConfig.PIDFilename = CF.PIDFile.PIDFile

	RuntimeConfig.IPv4Script.Script = nil
	for _, i := range CF.IPv4Commands {
		var SE ScriptEntry
		SE.Exec = i.Command
		SE.Timeout = time.Second * time.Duration(i.Timeout)
		RuntimeConfig.IPv4Script.Script = append(RuntimeConfig.IPv4Script.Script, SE)
	}

	RuntimeConfig.IPv6Script.Script = nil
	for _, i := range CF.IPv6Commands {
		var SE ScriptEntry
		SE.Exec = i.Command
		SE.Timeout = time.Second * time.Duration(i.Timeout)
		RuntimeConfig.IPv6Script.Script = append(RuntimeConfig.IPv6Script.Script, SE)
	}

	if len(RuntimeConfig.IPv4Script.Script)+len(RuntimeConfig.IPv6Script.Script) == 0 {
		log.Fatal("No commands specified.")
	}

	if len(RuntimeConfig.PIDFilename) == 0 {
		RuntimeConfig.PIDFilename = "atrack.pid"
	}

	if len(RuntimeConfig.ListenAddr) == 0 {
		log.Fatal("No listen address provided.")
	}

	if (len(RuntimeConfig.TLSPrivateKey) > 0) != (len(RuntimeConfig.TLSFullchain) > 0) {
		log.Fatal("Inconsistent TLS configuration. Both a certificate chain and a private key must be provided to enable TLS. Alternatively, do not provide either a certificate or private key to disable TLS.")
	}

	if len(RuntimeConfig.TLSFullchain) > 0 {
		_, err := os.Stat(RuntimeConfig.TLSFullchain)
		if err != nil {
			log.Fatal("The specified TLS certificate fullchain file is not accessible.")
		}
	}

	if len(RuntimeConfig.TLSPrivateKey) > 0 {
		_, err := os.Stat(RuntimeConfig.TLSPrivateKey)
		if err != nil {
			log.Fatal("The specified TLS private key is not accessible.")
		}
	}

	RuntimeConfig.Mutex.Unlock()

	return nil
}
