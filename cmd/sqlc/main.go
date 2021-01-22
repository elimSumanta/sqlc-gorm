package main

import (
	"github.com/ujunglangit-id/sqlc/internal/cmd"
	"os"
)

func main() {
	os.Exit(cmd.Do(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}
