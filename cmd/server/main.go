package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/holden/sshmasher/internal/handler"
	"github.com/holden/sshmasher/internal/ssh"
	"github.com/holden/sshmasher/static"
)

func printLogo() {
    logo := `
  ____ ____  _   _                      _               
 / ___/ ___|| | | |_ __ ___   __ _  ___| |__   ___ _ __ 
 \___ \___ \| |_| | '_ ` _ \ / _` |/ __| '_ \ / _ \ '__|
  ___) |__) |  _  | | | | | | (_| | (__| | | |  __/ |   
 |____/____/|_| |_|_| |_| |_|\__,_|\___|_| |_|\___|_|   
    `
    fmt.Println(logo)
}

func main() {
	addr := flag.String("addr", "127.0.0.1:8932", "listen address")
	sshPath := flag.String("ssh-dir", "", "SSH directory (default: ~/.ssh)")
	flag.Parse()

	var dir *ssh.SSHDir
	var err error
	if *sshPath != "" {
		dir = ssh.NewSSHDir(*sshPath)
	} else {
		dir, err = ssh.DefaultSSHDir()
		if err != nil {
			log.Fatalf("Failed to resolve SSH directory: %v", err)
		}
	}

	router := handler.NewRouter(dir, static.FS())
	server := handler.WithMiddleware(router)

	printLogo()
	fmt.Printf("SSHmasher listening on http://%s\n", *addr)
	if err := http.ListenAndServe(*addr, server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
