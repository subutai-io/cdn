package main

import (
	"github.com/subutai-io/cdn"
	"github.com/urfave/cli"
	"os"
	"github.com/subutai-io/agent/log"
)

var (
	AppVersion = "7.0.1"
)

// Command-line flags
var (
	Log string
)

/*
var (
	stopServer *http.Server
	stopped    chan bool
)
*/

func Init() {
	log.Info("Initialization started")
	app.SetLogLevel(Log)
	app.InitConfig()
	app.InitDB()
	app.InitFilters()
	log.Info("Initialization ended")
}

// main starts/stops CDN server
func main() {
	log.Info("Starting main")
	application := cli.NewApp()
	application.Name = "CDN"
	application.Version = AppVersion
	application.Authors = []cli.Author{
		{
			Name: "Subutai.io",
		},
	}
	application.Copyright = "Copyright 2018 Subutai.io"
	application.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "log",
			Usage:       "Specify log level. Possible options: panic, fatal, error, warn, info, debug",
			Value:       "",
			Destination: &Log,
		},

	}
	application.Action = func(c *cli.Context) error {
		Init()
		app.RunServer()
		<-app.Stop
		close(app.Stop)
		return nil
	}
	/*
	application.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Run CDN in daemon mode",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:        "log",
					Usage:       "Specify log level",
					Value:       "",
					Destination: &Log,
				},
			},
			Action: func(c *cli.Context) error {
				Init()
				app.RunServer()
				stopHandler := http.NewServeMux()
				stopHandler.HandleFunc("/kurjun/rest/stop", StopServer)
				stopServer = &http.Server{
					Addr:    "localhost:32525",
					Handler: stopHandler,
				}
				stopped = make(chan bool)
				go func() {
					stopServer.ListenAndServe()
					log.Info("ListenAndServe finished")
				}()
				<-stopped
				close(stopped)
				stopServer.Shutdown(context.Background())
				return nil
			},
		},
		{
			Name:  "stop",
			Usage: "Stop CDN",
			Action: func(c *cli.Context) error {
				request, _ := http.NewRequest("", "http://localhost:32525/kurjun/rest/stop", nil)
				client := http.Client{}
				response, err := client.Do(request)
				if err != nil || response.StatusCode != 200 {
					if err == nil {
						err = fmt.Errorf("response didn't return status code 200 (OK)")
					}
					log.Warn(fmt.Sprintf("failed to stop CDN: %v", err))
					return err
				}
				response.Body.Close()
				return nil
			},
		},
	}
	*/
	application.Run(os.Args)
}

/*
func StopServer(w http.ResponseWriter, r *http.Request) {
	if app.Stop != nil {
		log.Info("Handling stop server request")
		app.Stop <- true
		<-app.Stop
		close(app.Stop)
		log.Info("Stop channel closed")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server stopped successfully"))
		stopped <- true
	} else {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Server is not running"))
	}
}
*/