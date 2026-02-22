package main

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Print("missing-mode")
		os.Exit(2)
	}
	mode := os.Args[1]
	switch mode {
	case "good-json":
		fmt.Print("{\"ok\":true,\"mode\":\"good-json\"}")
		os.Exit(0)
	case "bad-json-text":
		fmt.Print("not-json")
		os.Exit(0)
	case "bad-json-multi":
		fmt.Print("{\"a\":1}{\"b\":2}")
		os.Exit(0)
	case "bad-json-array":
		fmt.Print("[1,2,3]")
		os.Exit(0)
	case "good-json-stderr":
		fmt.Print("{\"ok\":true,\"mode\":\"good-json-stderr\"}")
		fmt.Fprint(os.Stderr, "warning")
		os.Exit(0)
	case "fail-exit-3":
		fmt.Print("{\"ok\":false,\"mode\":\"fail-exit-3\"}")
		os.Exit(3)
	case "hang":
		sleepMs := 2000
		if len(os.Args) >= 3 {
			if n, err := strconv.Atoi(os.Args[2]); err == nil {
				sleepMs = n
			}
		}
		time.Sleep(time.Duration(sleepMs) * time.Millisecond)
		fmt.Print("{\"ok\":true,\"mode\":\"hang\"}")
		os.Exit(0)
	default:
		fmt.Print("unknown-mode")
		os.Exit(2)
	}
}
