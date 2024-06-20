package main

import (
    "fmt"
    "os"

    "github.com/urfave/cli/v2"
)

func main() {
    app := &cli.App{
        Name: "MultLdr CLI client",

        Commands: []*cli.Command{
            {
                Name:  "plugs",
                Usage: "Retrieve list of plugins present",

                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:        "server",
                        Aliases:     []string{"s"},
                        Value:       "127.0.0.1",
                        Usage:       "Use provided `IP` for the MultLdr server",
                        DefaultText: "127.0.0.1",
                    },
                    &cli.IntFlag{
                        Name:        "port",
                        Aliases:     []string{"p"},
                        Value:       5000,
                        Usage:       "Use provided `PORT` for the MultLdr server",
                        DefaultText: "5000",
                    },
                },
                Action: func(ctx *cli.Context) error {
                    plugins, err  := getPlugins(ctx.String("server"), ctx.Int("port"))

                    if err != nil {
                        printLog(logError, fmt.Sprintf("%v", err))
                        return nil
                    }

                    displayPlugins(plugins)

                    return nil
                },
            },
            {
                Name:  "gen",
                Usage: "Generates loader with the options provided",

                Flags: []cli.Flag{
                    &cli.StringFlag{
                        Name:        "server",
                        Aliases:     []string{"s"},
                        Value:       "127.0.0.1",
                        Usage:       "Use provided `IP` for the MultLdr server",
                        DefaultText: "127.0.0.1",
                    },
                    &cli.IntFlag{
                        Name:        "port",
                        Aliases:     []string{"p"},
                        Value:       5000,
                        Usage:       "Use provided `PORT` for the MultLdr server",
                        DefaultText: "5000",
                    },
                    &cli.StringFlag{
                        Name:        "config",
                        Aliases:     []string{"c"},
                        Usage:       "Load configuration from `FILE` (optional)",
                    },
                    &cli.StringFlag{
                        Name:        "bin",
                        Aliases:     []string{"b"},
                        Usage:       "Use payload binary from `FILE`",
                        Required:    true,
                    },
                },
                Action: func(ctx *cli.Context) error {
					var config map[string][]string
                    var err error
					if ctx.String("config") != "" {
						config, err = readConfig(ctx.String("config"))
                        if err != nil {
                            printLog(logError, fmt.Sprintf("%v", err))
                            return nil
                        }

					} else {
						config, err = getConfig(ctx.String("server"), ctx.Int("port"))
                        if err != nil {
                            printLog(logError, fmt.Sprintf("%v", err))
                            return nil
                        }

                        saveConfigFile(config)
					}

                    printPlugins("Using the following settings", config)
                    
					payloadFile := ctx.String("bin")

                    id, err := generateID()
                    if err != nil {
                        printLog(logError, fmt.Sprintf("%v", err))
                        return nil
                    }

                    sendPayload(ctx.String("server"), ctx.Int("port"), id, payloadFile)

                    generateLoader(ctx.String("server"), ctx.Int("port"), id, config)
                    
                    requestLoader(ctx.String("server"), ctx.Int("port"), id)
                    
                    return nil
                },
            },
        },
    }

    if err := app.Run(os.Args); err != nil {
        printLog(logError, fmt.Sprintf("%v", err))
    }
}
