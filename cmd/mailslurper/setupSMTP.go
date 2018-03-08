package main

import (
	"context"

	"github.com/mailslurper/mailslurper/pkg/mailslurper"
)

func setupSMTP() {
	/*
	 * Setup a channel to receive parsed mail items
	 */
	mailItemChannel = make(chan *mailslurper.MailItem, 1000)

	/*
	 * Setup the server pool
	 */
	pool := mailslurper.NewServerPool(mailslurper.GetLogger(*logLevel, *logFormat, "SMTP Server Pool"), config.MaxWorkers)

	/*
	 * Setup receivers (subscribers) to handle new mail items.
	 */
	receivers := []mailslurper.IMailItemReceiver{
		mailslurper.NewDatabaseReceiver(database, mailslurper.GetLogger(*logLevel, *logFormat, "Database Receiver")),
	}

	/*
	 * Setup a context for controlling shutdown of SMTP services
	 */
	smtpListenerContext, smtpListenerCancel := context.WithCancel(context.Background())

	/*
	 * Setup the connection manager
	 */
	if connectionManager, err = mailslurper.NewConnectionManager(mailslurper.GetLogger(*logLevel, *logFormat, "Connection Manager"), config, smtpListenerContext, mailItemChannel, pool); err != nil {
		logger.WithError(err).Fatalf("Error creating connection manager. Exiting...")
	}

	/*
	 * Setup the SMTP listener
	 */

	if smtpListener, err = mailslurper.NewSMTPListener(
		mailslurper.GetLogger(*logLevel, *logFormat, "SMTP Listener"),
		config,
		mailItemChannel,
		pool,
		receivers,
		connectionManager,
	); err != nil {
		logger.WithError(err).Fatalf("There was a problem starting the SMTP listener. Exiting...")
	}

	/*
	 * Start the SMTP listener
	 */
	if err = smtpListener.Start(); err != nil {
		logger.WithError(err).Fatalf("Error starting SMTP listener. Exiting...")
	}

	smtpListener.Dispatch(smtpListenerContext)
}
