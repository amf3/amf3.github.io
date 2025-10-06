package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func hello(name string, count int) {
	for range count {
		fmt.Println("Hello", name)
	}
}

func goodbye(name string, count int) {
	for range count {
		fmt.Println("Goodbye", name)
	}
}

func runIt(ctx context.Context, cmd *cli.Command) error {
	if cmd.Bool("hello") {
		hello(cmd.String("name"), cmd.Int("count"))
	}
	return nil
}

func main() {
	cmd := &cli.Command{
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:     "hello",
				Required: false,
				Usage:    "Will greet a person with 'Hello'",
			},
			&cli.BoolFlag{
				Name:     "goodbye",
				Required: false,
				Usage:    "Will tell a person 'Goodbye'",
			},
			&cli.IntFlag{
				Name:     "count",
				Usage:    "How many times to invoke function",
				Required: false,
				Value:    1,
			},
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Tell me your name",
				Required: true, // <- throws an error and prints help statement if name is not defined
			},
		},
		Action: runIt, // <-- reference to the named function
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
