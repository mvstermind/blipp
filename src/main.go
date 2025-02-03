package main

import (
	"fmt"

	"github.com/mvstermind/mini-cli"
)

type Action int

const (
	Join Action = iota + 1
	Create
)

func main() {

	argCreate := mini.Arg{
		ShortCmd: "c",
		LongCmd:  "create",
		Usage:    "create a chat room",
		Required: false,
	}

	argJoin := mini.Arg{
		ShortCmd: "j",
		LongCmd:  "join",
		Usage:    "join a room",
		Required: false,
	}

	cmds := mini.AddArguments(&argCreate, &argJoin)
	argValues := cmds.Execute()

	action, err := mustCreateOrJoin(argValues)
	if err != nil {
		fmt.Println("err: ", err)
	}
	println(action)

}

func mustCreateOrJoin(args map[string]string) (Action, error) {

	for k, v := range args {
		if k == "-c" && v != "" {
			return Create, nil
		}
		if k == "-j" && v != "" {
			return Join, nil
		}

	}
	panic("action is needed")
}
