package main

import (
	"fmt"
	"github.com/Toorop/tmail/api"
	"github.com/Toorop/tmail/scope"
	"github.com/codegangsta/cli"
	"os"
	"strconv"
)

var cliCommands = []cli.Command{
	{
		// SMTPD
		Name:  "smtpd",
		Usage: "commands to interact with smtpd process",
		Subcommands: []cli.Command{
			// users
			{
				Name:        "addUser",
				Usage:       "Add a smtpd user",
				Description: "tmail smtpd addUser USER CLEAR_PASSWD [RELAY_ALLOWED]",
				Action: func(c *cli.Context) {
					var err error
					if len(c.Args()) < 2 {
						cliDieBadArgs(c)
					}
					relayAllowed := false
					if len(c.Args()) > 2 {
						relayAllowed, err = strconv.ParseBool(c.Args()[2])
						cliHandleErr(err)
					}

					err = api.SmtpdAddUser(c.Args()[0], c.Args()[1], relayAllowed)
					cliHandleErr(err)
					cliDieOk()
				},
			},
			{
				Name:        "delUser",
				Usage:       "Delete a smtpd user",
				Description: "tmail smtpd delUser USER",
				Action: func(c *cli.Context) {
					var err error
					if len(c.Args()) != 1 {
						cliDieBadArgs(c)
					}
					err = api.SmtpdDelUser(c.Args()[0])
					cliHandleErr(err)
					cliDieOk()
				},
			},
			{
				Name:        "listAutorizedUsers",
				Usage:       "Return a list of authorized users (users who can send mail after authentification)",
				Description: "",
				Action: func(c *cli.Context) {
					users, err := api.SmtpdGetAllowedUsers()
					cliHandleErr(err)
					if len(users) == 0 {
						println("There is no smtpd users yet.")
						return
					}
					println("Relay access granted to: ", c.Args().First())
					for _, user := range users {
						println(user.Login)
					}
				},
			},
			{
				Name:        "addRcpthost",
				Usage:       "Add a 'rcpthost' which is a hostname that tmail have to handle mails for",
				Description: "tmail smtpd addRcpthost HOSTNAME",
				Action: func(c *cli.Context) {
					var err error
					if len(c.Args()) != 1 {
						cliDieBadArgs(c)
					}
					err = api.SmtpdAddRcptHost(c.Args()[0])
					cliHandleErr(err)
					cliDieOk()
				},
			},
			{
				Name:        "delRcpthost",
				Usage:       "Delete a rcpthost",
				Description: "tmail smtpd delRcpthost HOSTNAME",
				Action: func(c *cli.Context) {
					var err error
					if len(c.Args()) != 1 {
						cliDieBadArgs(c)
					}
					err = api.SmtpdDelRcptHost(c.Args()[0])
					cliHandleErr(err)
					cliDieOk()
				},
			},
			{
				Name:        "getRcpthosts",
				Usage:       "Returns all the rcpthosts ",
				Description: "tmail smtpd getRcpthost",
				Action: func(c *cli.Context) {
					var err error
					if len(c.Args()) != 0 {
						cliDieBadArgs(c)
					}
					rcptHosts, err := api.SmtpdGetRcptHosts()
					cliHandleErr(err)
					for _, h := range rcptHosts {
						println(h.Hostname)
					}
				},
			},
		},
	}, {
		// QUEUE
		Name:  "queue",
		Usage: "commands to interact with tmail queue",
		Subcommands: []cli.Command{
			// list queue
			{
				Name:        "list",
				Usage:       "List messages in queue",
				Description: "tmail queue list",
				Action: func(c *cli.Context) {
					var status string
					messages, err := api.QueueGetMessages()
					cliHandleErr(err)
					if len(messages) == 0 {
						println("There is no message in queue.")
					} else {
						fmt.Printf("%d messages in queue.\r\n", len(messages))
						for _, m := range messages {
							switch m.Status {
							case 0:
								status = "Delivery in progress"
							case 1:
								status = "Will be discarded"
							case 2:
								status = "Scheduled"
							case 3:
								status = "Will be bounced"
							}

							msg := fmt.Sprintf("%d - From: %s - To: %s - Status: %s - Added: %v ", m.Id, m.MailFrom, m.RcptTo, status, m.AddedAt)
							if m.Status != 0 {
								msg += fmt.Sprintf("- Next delivery process scheduled at: %v", m.NextDeliveryScheduledAt)
							}
							println(msg)
						}
					}
					os.Exit(0)
				},
			},
			{
				Name:        "discard",
				Usage:       "Discard (delete without bouncing) a message in queue",
				Description: "tmail queue discard MESSAGE_ID",
				Action: func(c *cli.Context) {
					if len(c.Args()) != 1 {
						cliDieBadArgs(c)
					}
					id, err := strconv.ParseInt(c.Args()[0], 10, 64)
					cliHandleErr(err)
					cliHandleErr(api.QueueDiscardMsg(id))
					cliDieOk()
				},
			},
			{
				Name:        "bounce",
				Usage:       "Bounce a message in queue",
				Description: "tmail queue bounce MESSAGE_ID",
				Action: func(c *cli.Context) {
					if len(c.Args()) != 1 {
						cliDieBadArgs(c)
					}
					id, err := strconv.ParseInt(c.Args()[0], 10, 64)
					cliHandleErr(err)
					cliHandleErr(api.QueueBounceMsg(id))
					cliDieOk()
				},
			},
		},
	}, {
		// ROUTES
		Name:  "routes",
		Usage: "commands to manage outgoing SMTP routes",
		Subcommands: []cli.Command{
			{
				Name:        "list",
				Usage:       "List routes",
				Description: "tmail routes list",
				Action: func(c *cli.Context) {
					routes, err := api.RoutesGet()
					cliHandleErr(err)
					scope.Log.Debug(routes)
					if len(routes) == 0 {
						println("There is no routes configurated, all mails are routed following MX records")
					} else {
						for _, route := range routes {
							scope.Log.Debug(route)
							line := "Destination host: " + route.Host

							// Priority
							line += " - Prority: "
							if route.Priority.Valid && route.Priority.Int64 != 0 {
								line += fmt.Sprintf("%d", route.Priority.Int64)
							} else {
								line += "1"
							}

							// Local IPs
							line += " - Local IPs: "
							if route.LocalIp.Valid && route.LocalIp.String != "" {
								line += route.LocalIp.String
							} else {
								line += "default"
							}
							// Remote Host
							line += " - Remote host: " + route.RemoteHost
							if route.RemotePort.Valid && route.RemotePort.Int64 != 0 {
								line += fmt.Sprintf(":%d", route.RemotePort.Int64)
							} else {
								line += ":25"
							}

							println(line)
						}
					}
					os.Exit(0)
				},
			},
		},
	},
}

// gotError handle error from cli
func cliHandleErr(err error) {
	if err != nil {
		println("Error: ", err.Error())
		os.Exit(1)
	}
}

// cliDieBadArgs die on bad arg
func cliDieBadArgs(c *cli.Context) {
	println("Error: bad args")
	cli.ShowAppHelp(c)
	os.Exit(1)
}

func cliDieOk() {
	//println("Success")
	os.Exit(0)
}
