package main

import (
	"bufio"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/zelenin/go-tdlib/client"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/urfave/cli/v2"
)

var (
	version   string
	buildDate string
	commitID  string
)

func main() {
	app := &cli.App{
		Name:    "poetbot",
		Usage:   "Telegram auto update name bot",
		Version: fmt.Sprintf("%s %s %s", version, buildDate, commitID),
		Authors: []*cli.Author{
			{
				Name:  "mritd",
				Email: "mritd@linux.com",
			},
		},
		Copyright: "Copyright (c) 2020 mritd, All rights reserved.",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "txtfile",
				Aliases: []string{"t"},
				Value:   "poet.txt",
				Usage:   "Names file",
				EnvVars: []string{"POETBOT_TXTFILE"},
			},
			&cli.StringFlag{
				Name:    "cron",
				Aliases: []string{"c"},
				Value:   "@every 30s",
				Usage:   "Update crontab",
				EnvVars: []string{"POETBOT_CRON"},
			},
			&cli.StringFlag{
				Name:    "data",
				Aliases: []string{"d"},
				Value:   "data",
				Usage:   "Data storage dir",
				EnvVars: []string{"POETBOT_DATA"},
			},
			&cli.IntFlag{
				Name:     "appid",
				Usage:    "Telegram app id",
				Required: true,
				EnvVars:  []string{"POETBOT_APPID"},
			},
			&cli.StringFlag{
				Name:     "apphash",
				Usage:    "Telegram app hash",
				Required: true,
				EnvVars:  []string{"POETBOT_APPHASH"},
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
				Usage: "Debug mode",
			},
		},
		Before: func(c *cli.Context) error {
			if c.Bool("debug") {
				logrus.SetLevel(logrus.DebugLevel)
			}
			logrus.SetFormatter(&logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			})
			return nil
		},
		Action: update,
	}
	err := app.Run(os.Args)
	if err != nil {
		logrus.Error(err)
	}
}

func update(c *cli.Context) error {
	var data []string

	f, err := os.OpenFile(c.String("txtfile"), os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len([]rune(line)) > 64 {
			continue
		}
		data = append(data, line)
	}

	// client authorizer
	authorizer := client.ClientAuthorizer()
	go client.CliInteractor(authorizer)

	authorizer.TdlibParameters <- &client.TdlibParameters{
		UseTestDc:              false,
		DatabaseDirectory:      filepath.Join(c.String("data"), "tdlib-db"),
		FilesDirectory:         filepath.Join(c.String("data"), "tdlib-files"),
		UseFileDatabase:        true,
		UseChatInfoDatabase:    true,
		UseMessageDatabase:     true,
		UseSecretChats:         false,
		ApiId:                  int32(c.Int("appid")),
		ApiHash:                c.String("apphash"),
		SystemLanguageCode:     "en",
		DeviceModel:            "Server",
		SystemVersion:          "1.0.0",
		ApplicationVersion:     "1.0.0",
		EnableStorageOptimizer: true,
		IgnoreFileNames:        false,
	}

	logVerbosity := client.WithLogVerbosity(&client.SetLogVerbosityLevelRequest{
		NewVerbosityLevel: 0,
	})

	tdlibClient, err := client.NewClient(authorizer, logVerbosity)
	if err != nil {
		logrus.Fatalf("NewClient error: %s", err)
	}

	optionValue, err := tdlibClient.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		logrus.Fatalf("GetOption error: %s", err)
	}

	logrus.Infof("TDLib version: %s", optionValue.(*client.OptionValueString).Value)

	me, err := tdlibClient.GetMe()
	if err != nil {
		logrus.Fatalf("GetMe error: %s", err)
	}

	logrus.Infof("Me: %s %s [%s]", me.FirstName, me.LastName, me.Username)

	cn := cron.New()
	_, err = cn.AddFunc(c.String("cron"), func() {
		rand.Seed(time.Now().Unix())
		name := data[rand.Intn(len(data)-1)]
		logrus.Infof("update name to [%s]...", name)
		_, err := tdlibClient.SetName(&client.SetNameRequest{
			FirstName: name,
		})
		if err != nil {
			logrus.Error(err)
		}
	})
	if err != nil {
		return err
	}

	cn.Start()
	logrus.Info("Poet Bot running...")

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cn.Stop()

	return nil
}
