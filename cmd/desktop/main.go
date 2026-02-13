package main

import (
	"log"

	"github.com/holden/sshmasher/internal/handler"
	"github.com/holden/sshmasher/internal/ssh"
	"github.com/holden/sshmasher/static"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

func main() {
	dir, err := ssh.DefaultSSHDir()
	if err != nil {
		log.Fatalf("Failed to resolve SSH directory: %v", err)
	}

	router := handler.NewRouter(dir, static.FS())

	err = wails.Run(&options.App{
		Title:     "SSHmasher",
		Width:     1200,
		Height:    800,
		MinWidth:  800,
		MinHeight: 600,
		AssetServer: &assetserver.Options{
			Handler: handler.WithMiddleware(router),
		},
		OnStartup: nil,
		Bind:       []interface{}{},
	})
	if err != nil {
		log.Fatalf("Wails failed: %v", err)
	}
}
