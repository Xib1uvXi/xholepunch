package main

import (
	"flag"
	"github.com/Xib1uvXi/xholepunch/pkg/traversalv2"
	"github.com/go-kratos/kratos/v2/log"
)

func main() {
	logger := log.With(log.DefaultLogger,
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
	)

	log.SetLogger(logger)
	log.DefaultLogger = logger

	var token string
	var server string
	var nat int

	flag.StringVar(&token, "token", "", "token")
	flag.StringVar(&server, "server", "", "server address")
	flag.IntVar(&nat, "nat", 0, "nat type")
	flag.Parse()

	if token == "" {
		panic("token is required")
	}

	if server == "" {
		panic("server address is required")
	}

	if nat == 0 {
		panic("nat type is required")
	}

	nattype := int8(nat)

	cLocal, err := traversalv2.BuilderDemo(server, nattype)
	if err != nil {
		panic(err)
	}

	defer cLocal.Close()

	if err = cLocal.HolePunching(token); err != nil {
		panic(err)
	}

}
