package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/Arman92/go-tdlib"
	"github.com/mritd/logger"
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
			},
			&cli.StringFlag{
				Name:    "cron",
				Aliases: []string{"c"},
				Value:   "@every 30s",
				Usage:   "Update crontab",
			},
			&cli.StringFlag{
				Name:    "data",
				Aliases: []string{"d"},
				Value:   "data",
				Usage:   "Data storage dir",
			},
			&cli.StringFlag{
				Name:     "appid",
				Usage:    "Telegram app id",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "apphash",
				Usage:    "Telegram app hash",
				Required: true,
			},
			&cli.BoolFlag{
				Name:  "debug",
				Value: false,
				Usage: "Debug mode",
			},
		},
		Before: func(c *cli.Context) error {
			if c.Bool("debug") {
				logger.SetDevelopment()
			}
			return nil
		},
		Action: update,
	}
	err := app.Run(os.Args)
	if err != nil {
		logger.Error(err)
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

	tdlib.SetLogVerbosityLevel(1)
	tdlib.SetFilePath(os.Stdout.Name())
	// Create new instance of client
	client := tdlib.NewClient(tdlib.Config{
		APIID:               c.String("appid"),
		APIHash:             c.String("apphash"),
		SystemLanguageCode:  "en",
		DeviceModel:         "Server",
		SystemVersion:       "1.0.0",
		ApplicationVersion:  "1.0.0",
		UseMessageDatabase:  true,
		UseFileDatabase:     true,
		UseChatInfoDatabase: true,
		DatabaseDirectory:   filepath.Join(c.String("data"), "tdlib-db"),
		FileDirectory:       filepath.Join(c.String("data"), "tdlib-files"),
	})

	for {
		currentState, _ := client.Authorize()
		switch currentState.GetAuthorizationStateEnum() {
		case tdlib.AuthorizationStateWaitPhoneNumberType:
			fmt.Print("Enter phone: ")
			var number string
			_, _ = fmt.Scanln(&number)
			_, err := client.SendPhoneNumber(number)
			if err != nil {
				logger.Errorf("Error sending phone number: %v", err)
			}
		case tdlib.AuthorizationStateWaitCodeType:
			fmt.Print("Enter code: ")
			var code string
			_, _ = fmt.Scanln(&code)
			_, err := client.SendAuthCode(code)
			if err != nil {
				logger.Errorf("Error sending auth code : %v", err)
			}
		case tdlib.AuthorizationStateWaitPasswordType:
			fmt.Print("Enter Password: ")
			var password string
			_, _ = fmt.Scanln(&password)
			_, err := client.SendAuthPassword(password)
			if err != nil {
				logger.Errorf("Error sending auth password: %v", err)
			}
		case tdlib.AuthorizationStateReadyType:
			logger.Info("Authorization Ready! Let's rock")
			goto AuthSuccess
		}
	}

AuthSuccess:

	cn := cron.New()
	_, err = cn.AddFunc(c.String("cron"), func() {
		rand.Seed(time.Now().Unix())
		name := data[rand.Intn(len(data)-1)]
		logger.Infof("update name to [%s]...", name)
		_, err := client.SetName(name, "")
		if err != nil {
			logger.Error(err)
		}
	})
	if err != nil {
		return err
	}

	cn.Start()
	logger.Info("Poet Bot running...")

	sigs := make(chan os.Signal)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cn.Stop()

	return nil
}
