// Copyright 2013-2016 Adam Presley. All rights reserved
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

//go:generate esc -o ./www/www.go -pkg www -ignore DS_Store|README\.md|LICENSE|www\.go -prefix /www/ ./www

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mailslurper/mailslurper/pkg/mailslurper"
	"github.com/mailslurper/mailslurper/pkg/ui"
	"github.com/sirupsen/logrus"
)

const (
	// Version of the MailSlurper Server application
	SERVER_VERSION string = "1.12.0"

	// Set to true while developing
	DEBUG_ASSETS bool = false

	CONFIGURATION_FILE_NAME string = "config.json"
)

var config *mailslurper.Configuration
var database mailslurper.IStorage
var logger *logrus.Entry
var serviceTierConfig *mailslurper.ServiceTierConfiguration
var renderer *ui.TemplateRenderer
var mailItemChannel chan *mailslurper.MailItem
var smtpListener *mailslurper.SMTPListener
var connectionManager *mailslurper.ConnectionManager

var logFormat = flag.String("logformat", "simple", "Format for logging. 'simple' or 'json'. Default is 'simple'")
var logLevel = flag.String("loglevel", "info", "Level of logs to write. Valid values are 'debug', 'info', or 'error'. Default is 'info'")

func main() {
	var err error
	flag.Parse()

	logger = mailslurper.GetLogger(*logLevel, *logFormat, "MailSlurper")
	logger.Infof("Starting MailSlurper Server v%s", SERVER_VERSION)

	renderer = ui.NewTemplateRenderer(DEBUG_ASSETS)

	setupConfig()
	setupDatabase()
	setupSMTP()
	setupAdminListener()
	setupServicesListener()

	if config.AutoStartBrowser {
		ui.StartBrowser(config)
	}

	/*
	 * Block this thread until we get an interrupt signal. Once we have that
	 * start shutting everything down
	 */
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT)

	<-quit

	ctx, cancel := context.WithTimeout(smtpListenerContext, 10*time.Second)
	defer cancel()

	smtpListenerCancel()

	if err = admin.Shutdown(ctx); err != nil {
		logger.Fatalf("Error shutting down admin listener: %s", err.Error())
	}

	if err = service.Shutdown(ctx); err != nil {
		logger.Fatalf("Error shutting down service listener: %s", err.Error())
	}
}
