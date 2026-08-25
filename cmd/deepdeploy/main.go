package main

import (
	"flag"
	"fmt"
	"os"

	"deepdeploy/internal/application"
	"deepdeploy/internal/persistence"
	"deepdeploy/internal/transport"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:19081", "监听地址")
	selfcheck := flag.Bool("selfcheck", false, "运行有界冒烟流程")
	dataDir := flag.String("data-dir", ".deepdeploy-data", "事件账本与快照目录")
	flag.Parse()
	if p := os.Getenv("PORT"); p != "" && flag.Lookup("addr").Value.String() == "127.0.0.1:19081" {
		*addr = "127.0.0.1:" + p
	}
	if *selfcheck {
		app := application.NewService(persistence.NewMemoryStore())
		server := transport.NewServer(app, *addr)
		if err := server.SelfCheck(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	store, err := persistence.NewFileStore(*dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	app := application.NewService(store)
	server := transport.NewServer(app, *addr)
	if err := server.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
