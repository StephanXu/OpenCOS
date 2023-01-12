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

	for i := range config.App.Sources {
		context := &config.App.Sources[i]
		if err := source.Manager.Restore(context); err != nil {
			logrus.WithFields(logrus.Fields{
				"sourceName": context.Name,
				"sourceType": context.Type,
				"err":        err,
			}).Error("RestoreSourceFailed")
			continue
		}
		if err := config.SaveContext(); err != nil {
			logrus.WithFields(logrus.Fields{
				"contextFilename": config.App.ContextFile,
				"err":             err,
			}).Error("SaveContextFailed")
			continue
		}
	}

	websvr.Run()
}
