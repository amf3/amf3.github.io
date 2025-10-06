package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func hello(name string) {
	fmt.Println("Hello", name)
}

func goodbye(count int) {
	for range count {
		fmt.Println("Goodbye")
	}
}

func runIt(ctx context.Context, cmd *cli.Command) error {
	switch cmd.Name {
	case "hello":
		hello(cmd.String("name"))
	case "goodbye":
		goodbye(cmd.Int("count"))
	default:
		fmt.Println("Unknown command")
	}
	return nil
}

func main() {
	cmd := &cli.Command{
		Commands: []*cli.Command{
			{
				Name:   "hello",
				Usage:  "Greets a person with 'Hello'",
				Action: runIt,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Tell me your name",
						Required: true, // <- throws an error and prints help statement if name is not defined
					},
				},
			},

			{
				Name:   "goodbye",
				Usage:  "Tells a person 'Goodbye'",
				Action: runIt,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:     "count",
						Usage:    "How many times to invoke function",
						Required: false,
						Value:    1,
					},
				},
			},
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
