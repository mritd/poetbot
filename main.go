package main

import (
	"bufio"
	"fmt"
	"github.com/Arman92/go-tdlib/client"
	"github.com/Arman92/go-tdlib/tdlib"
	"github.com/sirupsen/logrus"
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
			&cli.StringFlag{
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

	client.SetLogVerbosityLevel(1)
	client.SetFilePath(os.Stdout.Name())
	// Create new instance of client
	tdCli := client.NewClient(client.Config{
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
		currentState, _ := tdCli.Authorize()
		switch currentState.GetAuthorizationStateEnum() {
		case tdlib.AuthorizationStateWaitPhoneNumberType:
			fmt.Print("Enter phone: ")
			var number string
			_, _ = fmt.Scanln(&number)
			_, err := tdCli.SendPhoneNumber(number)
			if err != nil {
				logrus.Errorf("Error sending phone number: %v", err)
			}
		case tdlib.AuthorizationStateWaitCodeType:
			fmt.Print("Enter code: ")
			var code string
			_, _ = fmt.Scanln(&code)
			_, err := tdCli.SendAuthCode(code)
			if err != nil {
				logrus.Errorf("Error sending auth code : %v", err)
			}
		case tdlib.AuthorizationStateWaitPasswordType:
			fmt.Print("Enter Password: ")
			var password string
			_, _ = fmt.Scanln(&password)
			_, err := tdCli.SendAuthPassword(password)
			if err != nil {
				logrus.Errorf("Error sending auth password: %v", err)
			}
		case tdlib.AuthorizationStateReadyType:
			logrus.Info("Authorization Ready! Let's rock")
			goto AuthSuccess
		}
	}

AuthSuccess:

	cn := cron.New()
	_, err = cn.AddFunc(c.String("cron"), func() {
		rand.Seed(time.Now().Unix())
		name := data[rand.Intn(len(data)-1)]
		logrus.Infof("update name to [%s]...", name)
		_, err := tdCli.SetName(name, "")
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
