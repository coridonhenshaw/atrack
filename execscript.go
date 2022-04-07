package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/shlex"
)

func ExecTimeout(Cmd string, Timeout time.Duration) error {

	var err error = nil

	if Timeout == 0 {
		Timeout = time.Minute * 15
	}

	ctx, cancel := context.WithTimeout(context.Background(), Timeout)
	defer cancel()

	record, err := shlex.Split(Cmd)
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
	RuntimeConfig.Mutex.RLock()
	var Last4 = RuntimeConfig.IPv4.String()
	var Last6 = RuntimeConfig.IPv6.String()
	RuntimeConfig.Mutex.RUnlock()

	for {
		for {
			_ = <-Script.Trigger

			// Ignore link blips of less than 3 seconds
			time.Sleep(time.Duration(3 * time.Second))

			RuntimeConfig.Mutex.RLock()
			var Flag4 = (Last4 != RuntimeConfig.IPv4.String())
			var Flag6 = (Last6 != RuntimeConfig.IPv6.String())
			RuntimeConfig.Mutex.RUnlock()

			if Flag4 || Flag6 {
				break
			}
		}
		var ExecList []ScriptEntry

		RuntimeConfig.Mutex.RLock()

		for _, i := range Script.Script {

			re := strings.NewReplacer("$IPv4", RuntimeConfig.IPv4.String(), "$IPv6", RuntimeConfig.IPv6.String())

			var S ScriptEntry
			S.Exec = re.Replace(i.Exec)
			S.Timeout = i.Timeout

			ExecList = append(ExecList, S)
		}

		Last4 = RuntimeConfig.IPv4.String()
		Last6 = RuntimeConfig.IPv6.String()

		RuntimeConfig.Mutex.RUnlock()

		for _, Cmd := range ExecList {
			fmt.Println(Last4, Last6, "Exec:", Cmd.Exec)
			fmt.Println()

			err := ExecTimeout(Cmd.Exec, Cmd.Timeout)
			if err != nil {
				log.Fatal(err)
			}
		}
		time.Sleep(time.Duration(7 * time.Second))
	}
}
