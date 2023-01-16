package main

import (
	"github.com/sirupsen/logrus"
	"xxtuitui.com/filesvr/config"
	"xxtuitui.com/filesvr/source"
	"xxtuitui.com/filesvr/websvr"
)

func main() {
	var (
		configFilename = "config.json"
	)

	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: true,
	})

	if err := config.LoadContextFromConfigFile(configFilename); err != nil {
		logrus.WithFields(logrus.Fields{
			"filename": configFilename,
			"err":      err,
		}).Fatal("LoadContextFailed")
		return
	}
	source.RestoreFromContext(&config.App)
	websvr.Run()
}
