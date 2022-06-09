package main

import (
	"github.com/hwiewie/APIServer/cmd/k8sailor/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	cmd.Execute()
}

func init() {
	logrus.SetLevel(logrus.InfoLevel)
}
