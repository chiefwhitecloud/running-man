package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chiefwhitecloud/running-man/service"
	"log"
	"os"
)

func main() {
	// Get Arguments
	var cfgPath string

	flag.StringVar(&cfgPath, "config", "./config.yaml", "Path to Config File")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [arguments] <command> \n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	// Load Config
	type Configuration struct {
		Bind     string
		Database string
	}

	file, _ := os.Open(cfgPath)
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		fmt.Println("error:", err)
	}

	// pull desired command/operation from args
	if flag.NArg() == 0 {
		flag.Usage()
		log.Fatal("Command argument required")
	}
	cmd := flag.Arg(0)

	// Configure Server
	s, err := service.NewRunningManService(configuration.Bind, configuration.Database)
	if err != nil {
		log.Fatal(err)
	}
	// Run Main App
	switch cmd {
	case "serve":

		// Start Server
		if err := s.Run(); err != nil {
			log.Fatal(err)
		}
	case "migrate-db":

		// Start Server
		if err := s.MigrateDb(); err != nil {
			log.Fatal(err)
		}
	default:
		flag.Usage()
		log.Fatalf("Unknown Command: %s", cmd)
	}

}
