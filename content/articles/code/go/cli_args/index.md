---
title: "Go CLIs: Creating Subcommands and Flags"
description: "Using urfave/cli to create polished CLI applications"
date: 2025-09-29T17:17:10-04:00
draft: false
searchHidden: false
showtoc: true
categories: [Go, Code]
---

## Adding Subcommands to Go CLIs

Command Line Interfaces (CLIs) use subcommands and flags to enable different program features.  A
subcommand is a grouping of related features, and flags are options for controlling those features.  The openssl
command provides a great example of subcommands and flags. `openssl rand -base64 8` will generate 8 random bytes of
data with hexadecimal output.  The subcommand is "rand" and "-base64" is the flag.  Other openssl subcommands
like "s_client" or "x509", provide different features and each has their own options.

When running the openssl command, the shell passes each space separated value into an argument list.  It's possible to
parse the list by checking if the second value is a subcommand, then looping over the rest to figure out
which are flags.  Some programs like `git` take this further by treating subcommands as seperate executables.
`git status` actually runs the `git-status` command in a subshell.  The approach works but keeping help output and
flag parsing consistent between executables gets messy.

Fortunately several libraries are available to make subcommands and flags easier to manage.  The one that I finally chose is [urfave/cli](https://cli.urfave.org) which offers a builder style API.  The examples below use 
version 3 of the library which can be imported as:

```go
import "github.com/urfave/cli/v3"
```

## urfave/cli with flags

This is how simple the urfave/cli library is to use. I configure a [cli.Command](https://pkg.go.dev/github.com/urfave/cli#Command)
struct, then call [Run()](https://pkg.go.dev/github.com/urfave/cli#App.Run) to execute it.

```go
func main() {
    cmd := &cli.Command{
        ... struct goes here ...
    }
    if err := cmd.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}
```

So what does the struct inside cli.Command look like? It's a slice of [Flag](https://pkg.go.dev/github.com/urfave/cli#Flag) interfaces.
allowing us to freely mix different flag types like
[StringFlag](https://pkg.go.dev/github.com/urfave/cli#StringFlag), [BoolFlag](https://pkg.go.dev/github.com/urfave/cli#BoolFlag),
or [IntFlag](https://pkg.go.dev/github.com/urfave/cli#IntFlag).

```go
    Flags: []cli.Flag{
        &cli.IntFlag{
            ...
        },
        &cli.BoolFlag{
            ...
        },
    },
```

The other struct field is [Action](https://pkg.go.dev/github.com/urfave/cli/v3#ActionFunc) which runs a function after parsing the flags.
The function can either be inline or passed as a reference. Here is an inline example.  

```go
    Action: func(ctx context.Context, cmd *cli.Command) error {
        if cmd.Bool("hello") {
            hello(cmd.String("name"), cmd.Int("count"))
        }
        return nil
    },
```

This is what my populated main() looks like.  

```go
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
        Action: func(ctx context.Context, cmd *cli.Command) error {
            if cmd.Bool("hello") {
                hello(cmd.String("name"), cmd.Int("count"))
            }
            return nil
        },
    }

    if err := cmd.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}
```

A complete working example using flags can be [downloaded](./assets/flag_example.go).

### Flags Demo

When invoked without the --name flag which is marked required, urfave/cli will throw an error and print the help statement.

```shell
$ go run flag_example.go --count 2 --hello 
NAME:
   flag_example - A new cli application

USAGE:
   flag_example

OPTIONS:
   --hello        Will greet a person with 'Hello' (default: false)
   --goodbye      Will tell a person 'Goodbye' (default: false)
   --count int    How many times to invoke function (default: 1)
   --name string  Tell me your name
   --help, -h     show help
2025/10/02 08:36:19 Required flag "name" not set
exit status 1
```

Adding the --name flag results in it working as expected.

```shell
$ go run flag_example.go --count 2 --hello --name Adam
Hello Adam
Hello Adam
```

## urfave/cli with subcommands

Subcommands are added in a similar way, except instead of using a slice of Flag, we are using a slice of
Command which I described above.  Instead of using --hello or --goodbye as a flag, this example creates the
hello and goodbye subcommands.  

```go
func hello(ctx context.Context, cmd *cli.Command) error {
    fmt.Println("Hello")
    return nil
}

func goodbye(ctx context.Context, cmd *cli.Command) error {
    fmt.Println("Goodbye")
    return nil
}

func main() {
    cmd := &cli.Command{
        Commands: []*cli.Command{
            {
                Name:   "hello",
                Usage:  "Greets a person with 'Hello'",
                Action: hello,
            },
            {
                Name:   "goodbye",
                Usage:  "Tells a person 'Goodbye'",
                Action: goodbye,
            },
        },
    }

    if err := cmd.Run(context.Background(), os.Args); err != nil {
        log.Fatal(err)
    }
}
```

### Commands Demo

The autogenerated help statement when a subcommand isn't used.

```shell
$ go run command_example.go
NAME:
   command_example - A new cli application

USAGE:
   command_example [global options] [command [command options]]

COMMANDS:
   hello    Greets a person with 'Hello'
   goodbye  Tells a person 'Goodbye'
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

Invoking the goodbye subcommand

```shell
$ go run command_example.go goodbye
Goodbye
```

## Combining flags and subcommands

Now that we know how flags and subcommands work, lets combine them to get that openssl like experience.
But this is the great part, because subcommands are a slice of Command, and we know that Command accepts
a slice of Flag, we already know how this works.  Simply define a Flag for each subcommand.

Here's an example where the hello subcommand supports a name flag and the goodbye subcommand supports a count flag.

```go
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
```

A complete example that combines flags and subcommands is available for [download](./assets/combo_example.go).

### Combination Demo

Running the demo without any options

```shell
$ go run combo_example.go 
NAME:
   combo_example - A new cli application

USAGE:
   combo_example [global options] [command [command options]]

COMMANDS:
   hello    Greets a person with 'Hello'
   goodbye  Tells a person 'Goodbye'
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

Running the demo with a subcommand missing a required flag

```shell
$ go run combo_example.go hello                                                 
NAME:
   combo_example hello - Greets a person with 'Hello'

USAGE:
   combo_example hello

OPTIONS:
   --name string  Tell me your name
   --help, -h     show help
2025/10/05 09:33:19 Required flag "name" not set
exit status 1
```

Running the demo successfully

```shell
% go run combo_example.go hello --name Adam
Hello Adam
```

## urfave is now myfave

I started this post as a few notes for myself while learning **urfave/cli**, but it turned into a great reminder
of how flexible the library is.  It keeps command structure clear, couples help statements and actions to flags and subcommands,
resulting in polished CLIs.

If you're building a Go CLI and want a simple API for managing subcommands and options, give **urfave/cli** a try.
You'll spend more time thinking about command logic and less time manaaging subcommands or flag options.  

I can be found on [Bluesky](https://bsky.app/profile/af9.us) if you want to trade notes on building CLIs.
