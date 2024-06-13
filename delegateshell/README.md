# About

A golang client which can interact and acquire tasks from the Harness manager (or any task management system which implements the same interface)

delegate-shell library provides two utlities.
1. Register Runner with Runner Manager and sending heartbeats
2. Poll Runner task events

The idea is `delegate-shell` library should be a standalone library. Runner can use this library to handle all lifecycle events and interations with Harness.

# Usage
Example codes in `delegateshell/example` folder. You can run it by `go run main.go`

A way to use this client would be:
1. Registeration & heartbeat
```
	// Create a manager http client
	managerClient := client.New(...)

	keepAlive := heartbeat.New(..., managerClient)

	// Register & heartbeat
	ctx := context.Background()
	resp, _ := keepAlive.Register(ctx)
```
2. Poll tasks
```
  requestsChan := make(chan *client.RunnerRequest, 3)

  // Start polling for events
  eventsServer := poller.New(managerClient, requestsChan)
  eventsServer.PollRunnerEvents(ctx, 3, info.ID, time.Second*10)
  // next: process requests from 'requestsChan'
```

## TODOs
1. It uses register call for heartbeat. Maybe we should use polling tasks as heartbeat. This reduces network load on Harness gateway significantly
2. thread pool abstraction should be there for handle thread&resource allocation&isolation
3. Create a top level package for logger. Use that everywhere
4. Shutdownhook is not there
5. CPU memory related rejecting tasks is not there 
6. Add error library for all places 
7. Central Config: implement a global context where it has all the delegate configurations
