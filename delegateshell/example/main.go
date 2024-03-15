package main

import (
	"context"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/drone/go-task/delegateshell/client"
	"github.com/drone/go-task/delegateshell/heartbeat"
	"github.com/drone/go-task/delegateshell/poller"
)

func main() {
	// Create a delegate client
	managerClient := client.New("https://localhost:9090", "kmpySmUISimoRrJL6NL73w", "2f6b0988b6fb3370073c3d0505baee59", true, "")

	// The poller needs a client that interacts with the task management system and a router to route the tasks
	keepAlive := heartbeat.New("kmpySmUISimoRrJL6NL73w", "2f6b0988b6fb3370073c3d0505baee59", "dlite-xingchi", []string{"k8s-runner"}, managerClient)

	// // Register the poller
	ctx := context.Background()
	info, _ := keepAlive.Register(ctx)

	logrus.Info("Runner registered")

	requestsChan := make(chan *client.RunnerRequest, 3)

	// Start polling for bijou events
	eventsServer := poller.New(managerClient, requestsChan)
	// TODO: we don't need hb if we poll for task. Isn't it ? : )
	eventsServer.PollRunnerEvents(ctx, 3, info.ID, time.Second*10)

	// Just to keep it running
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		logrus.Info(w, *info)
	})

	logrus.Info("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logrus.Fatal(err)
	}
}
