// Copyright 2021-2022 The Memphis Authors
// Licensed under the Apache License, Version 2.0 (the “License”);
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an “AS IS” BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"memphis-broker/analytics"
	"memphis-broker/background_tasks"
	"memphis-broker/broker"
	"memphis-broker/db"
	"memphis-broker/handlers"
	"memphis-broker/http_server"
	"memphis-broker/logger"
	"memphis-broker/tcp_server"
	"net/http"
	"os"
	"sync"
)

func main() {
	err := logger.InitializeLogger()
	handleError("Failed initializing logger: ", err)

	err = analytics.InitializeAnalytics()
	handleError("Failed initializing analytics: ", err)

	err = handlers.CreateRootUserOnFirstSystemLoad()
	handleError("Failed to create root user: ", err)

	defer db.Close()
	defer broker.Close()
	defer analytics.Close()

	wg := new(sync.WaitGroup)
	wg.Add(4)

	go background_tasks.ConsumeSysLogs(wg)
	go tcp_server.InitializeTcpServer(wg)
	go http_server.InitializeHttpServer(wg)
	go background_tasks.KillZombieResources(wg)
	go background_tasks.ListenForPoisonMessages()

	var env string
	if os.Getenv("DOCKER_ENV") != "" {
		env = "Docker"
		logger.Info("\n**********\n\nDashboard: http://localhost:9000\nMemphis broker: localhost:5555 (Management Port) / 7766 (Data Port) / 6666 (TCP Port)\nUI/CLI root username - root\nUI/CLI root password - memphis\nSDK root connection token - memphis  \n\n**********")
	} else {
		env = "K8S"
	}

	// Set the router as the default one shipped with Gin
	// uiRouter := gin.Default()
	// uiRouter := mux.NewRouter()

	// uiRouter.HandleFunc("/", index).Methods("GET")
	// uiBuildHandler := http.FileServer(http.Dir("./memphis-ui/build"))
	// uiRouter.PathPrefix("/").Handler(uiBuildHandler)

	// srv := &http.Server{
	// 	Handler:      uiRouter,
	// 	Addr:         "127.0.0.1:9000",
	// 	WriteTimeout: 15 * time.Second,
	// 	ReadTimeout:  15 * time.Second,
	// }

	// logger.Info("Starting UI server on port 9000")
	// log.Fatal(srv.ListenAndServe())

	// Serve frontend static files
	// uiRouter.Use(static.Serve("/", static.LocalFile("./memphis-ui/build", true)))

	// Setup route group for the API
	// api := uiRouter.Group("/")
	// {
	// 	api.GET("/", func(c *gin.Context) {
	// 		c.JSON(200, gin.H{
	// 			"message": "pong",
	// 		})
	// 	})
	// }

	// Start and run the server
	// uiRouter.Run(":9000")

	logger.Info("Memphis broker is up and running, ENV: " + env)
	wg.Wait()
}

func handleError(message string, err error) {
	if err != nil {
		logger.Error(message + " " + err.Error())
		panic(message + " " + err.Error())
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./memphis-ui/build/index.html")
}
