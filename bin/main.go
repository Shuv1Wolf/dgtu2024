package main

import (
	"context"
	"hack/containers"
	"os"
)

func main() {
	proc := containers.NewFitProcess()
	proc.SetConfigPath("./config/config.yml")
	proc.Run(context.Background(), os.Args)
}
