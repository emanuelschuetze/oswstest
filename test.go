package main

import (
	"fmt"
	"log"
	"time"
)

// Test is a function, that expect a slice of clients and returns a slice of
// test results.
type Test func(clients []Client) (r []TestResult)

// RunTests runs some tests for a slice of clients. It returns the TestResults
// for each test.
func RunTests(clients []Client, tests []Test) (r []TestResult) {
	start := time.Now()
	defer func() { fmt.Printf("\nAll tests took %dms\n\n", time.Since(start)/time.Millisecond) }()
	for _, test := range tests {
		r = append(r, test(clients)...)
	}
	return
}

// ConnectTest opens connections for any given client. It returns two TestResults
// The first measures the time until the connection was open, the second measures the
// time until the fire data was received.
// Expects, that the wsconnection of the clients are closed.
func ConnectTest(clients []Client) (r []TestResult) {
	log.Println("Start ConnectTest")
	startTest := time.Now()
	defer func() { log.Printf("ConnectionTest took %dms", time.Since(startTest)/time.Millisecond) }()

	// Connect all Clients
	connected := make(chan time.Duration)
	connectedError := make(chan error)
	connectFinished := connectClients(clients, connectedError, connected)

	// Listen to all clients to receive the response.
	dataReceived := make(chan time.Duration)
	errorReceived := make(chan error)
	receivedFinished := listenToClients(clients, dataReceived, errorReceived, 1, nil, nil)

	connectedResult := TestResult{description: "Time to established connection"}
	dataReceivedResult := TestResult{description: "Time until data has been reveiced since the connection"}
	tick := time.Tick(time.Second)

	for {
		select {
		case value := <-connected:
			connectedResult.Add(value)

		case value := <-connectedError:
			connectedResult.AddError(value)

		case value := <-dataReceived:
			dataReceivedResult.Add(value)

		case value := <-errorReceived:
			dataReceivedResult.AddError(value)

		case <-tick:
			if LogStatus {
				log.Println(connectedResult.CountBoth(), dataReceivedResult.CountBoth())
			}
		}

		if *connectFinished && *receivedFinished {
			break
		}
	}
	return []TestResult{connectedResult, dataReceivedResult}
}

// OneWriteTest tests, that all clients get a response when there is one write
// request.
// Expects, that the first client is a logged-in admin client and that all
// clients have open websocket connections.
func OneWriteTest(clients []Client) (r []TestResult) {
	log.Println("Start OneWriteTest")
	startTest := time.Now()
	defer func() { log.Printf("OneWriteTest took %dms\n", time.Since(startTest)/time.Millisecond) }()

	// Find the admin client.
	admin, ok := clients[0].(AdminClient)
	if !ok || !admin.IsAdmin() || !admin.IsConnected() {
		log.Fatalf("Fatal: Expect the first client in OneWriteTest to be a connected AdminClient")
	}

	// Send the request.
	err := admin.Send()
	if err != nil {
		log.Fatalf("Can not send request, %s", err)
	}

	// Listen to all clients to receive the response.
	dataReceived := make(chan time.Duration)
	errorReceived := make(chan error)
	finished := listenToClients(clients, dataReceived, errorReceived, 1, nil, nil)

	dataReceivedResult := TestResult{description: "Time until data is received after one write request"}
	tick := time.Tick(time.Second)

	// Listn to all channels until the listeing is finished
	for {
		select {
		case value := <-dataReceived:
			dataReceivedResult.Add(value)

		case value := <-errorReceived:
			dataReceivedResult.AddError(value)

		case <-tick:
			if LogStatus {
				log.Println(dataReceivedResult.Count() + dataReceivedResult.ErrCount())
			}
		}

		if *finished {
			break
		}
	}

	return []TestResult{dataReceivedResult}
}

// ManyWriteTest tests behave like the OneWriteTest but send many write request.
// The first clients have to be admin clients. Sends one write request for each
// admin client.
// Expects, that at least one client is a logged-in admin client and that all
// clients have open websocket connections.
func ManyWriteTest(clients []Client) (r []TestResult) {
	log.Println("Start ManyWriteTest")
	startTest := time.Now()
	defer func() { log.Printf("ManyWriteTest took %dms\n", time.Since(startTest)/time.Millisecond) }()

	// Find all admins in the clients
	var admins []AdminClient
	for _, client := range clients {
		admin, ok := client.(AdminClient)
		if ok && admin.IsAdmin() && admin.IsConnected() {
			admins = append(admins, admin)
		}
	}
	if len(admins) == 0 {
		log.Fatalf("Fatal: Expect one client in ManyWriteTest to be a connected AdminClient")
	}

	// Send requests for all admin clients
	dataSended := make(chan time.Duration)
	errorSended := make(chan error)
	sendFinished := sendClients(admins, errorSended, dataSended)

	// Listen for all clients to receive messages
	dataReceived := make(chan time.Duration)
	errorReceived := make(chan error)
	// TODO: Use the sinceReceived or remove the API from the client.
	// var sinceReceived time.Time
	// sinceSet := make(chan bool)
	receiveFinished := listenToClients(clients, dataReceived, errorReceived, len(admins), nil, nil)

	sendedResult := TestResult{description: "Time until all requests have been sended"}
	receivedResult := TestResult{description: "Time until all responses have been received"}
	tick := time.Tick(time.Second)

	for {
		select {
		case value := <-dataSended:
			sendedResult.Add(value)

		case value := <-errorSended:
			sendedResult.AddError(value)

		case value := <-dataReceived:
			receivedResult.Add(value)

		case value := <-errorReceived:
			receivedResult.AddError(value)

		case <-tick:
			if LogStatus {
				log.Println(sendedResult.CountBoth(), receivedResult.CountBoth())
			}
		}

		// // Set sinceReceived when all requests are send. Close the channel
		// if *sendFinished && sinceReceived.IsZero() {
		// 	sinceReceived = time.Now()
		// 	close(sinceSet)
		// }

		// End the test when all admins have sended there data and each client got
		// as many responces as there are admins.
		if *sendFinished && *receiveFinished {
			break
		}
	}

	return []TestResult{sendedResult, receivedResult}
}

// KeepOpenTest is not a normal test. All id does is keeps the connection
// open for all given clients forever. You have to kill the program to exit.
func KeepOpenTest(clients []Client) (r []TestResult) {
	log.Println("Start KeepOpenTest")
	startTest := time.Now()
	defer func() { log.Printf("KeepOpenTest took %dms\n", time.Since(startTest)/time.Millisecond) }()

	readChan := make(chan []byte)
	errChan := make(chan error)
	for _, c := range clients {
		c.SetChannels(readChan, errChan)
		defer c.ClearChannels()
	}

	tick := time.Tick(time.Second)
	counter := 0
	errCounter := 0

	// Listn to all channels
	for {
		select {
		case <-readChan:
			counter++

		case <-errChan:
			errCounter++

		case <-tick:
			if LogStatus {
				log.Println(counter, errCounter)
			}
		}
	}

	return []TestResult{}
}
