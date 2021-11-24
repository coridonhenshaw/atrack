package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

func ExecTimeout(Cmd string, Timeout time.Duration) error {

	var err error = nil

	if Timeout == 0 {
		Timeout = time.Minute * 15
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	r := csv.NewReader(strings.NewReader(Cmd))
	r.Comma = ' '
	record, err := r.Read()
	if err != nil {
		log.Fatal(err)
	}

	Bin := record[0]
	Args := record[1:]

	c := exec.CommandContext(ctx, Bin, Args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if ctx.Err() == context.DeadlineExceeded {
		log.Fatal("Command \"", Cmd, "\" timed out.")
	}
	if err != nil {
		log.Fatal("Command \"", Cmd, "\" failed. Err: ", err)
	}

	return err
}

func ExecuteQueue(RuntimeConfig *RuntimeConfigStruct, Script *TriggeredScript) {
	for {
		_ = <-Script.Trigger

		var ExecList []ScriptEntry

		RuntimeConfig.Mutex.RLock()

		for _, i := range Script.Script {

			re := strings.NewReplacer("$IPv4", RuntimeConfig.IPv4.String(), "$IPv6", RuntimeConfig.IPv6.String())

			var S ScriptEntry
			S.Exec = re.Replace(i.Exec)
			S.Timeout = i.Timeout

			ExecList = append(ExecList, S)
		}

		RuntimeConfig.Mutex.RUnlock()

		for _, Cmd := range ExecList {
			fmt.Println("Exec:", Cmd.Exec)
			fmt.Println()

			err := ExecTimeout(Cmd.Exec, Cmd.Timeout)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
