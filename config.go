package main

// NormalClients and AdminClients are all clients, that are logged in. For the
// ConnectionTest there is no difference between the to clients. The AdminClient
// is needed to write data.
const (
	NormalClients = 10
	AdminClients  = 10
)

const (
	// BaseURL is the URL to the server. It is used for websocket and http. The
	// Placeholders are filled in by the code.
	BaseURL = "%s://localhost:8000/%s"

	// SSL is a flag for the BaseURL if http or https should be used.
	SSL = false

	// LoginURLPath is the path to build the url for login. It has no leading slash.
	LoginURLPath = "users/login/"

	// WSURLPath is the path to build the websocket url. It has no leading slash.
	WSURLPath = "ws/site/"

	// LoginPassword is the password to login the normal clients and also the admin clients.
	LoginPassword = "password"

	// MaxLoginAttemts is the number of tries for each client to login. If one
	// client fails more then this number, then the program is quit with a fatal
	// error.
	MaxLoginAttemts = 5

	// MaxConnectionAttemts is th enumber of tries for each client, to connect via
	// websocket. If a client fails, is program is not quit, but the error is shoun
	// in the end.
	MaxConnectionAttemts = 3

	// CSRFCookieName is the name of the CSRF cookie of OpenSlides. Make sure, that
	// this is the same as in the OpenSlides config.
	CSRFCookieName = "OpenSlidesCsrfToken"

	// ParallelConnections defines the number of connections, that are done in
	// parallel. The number should be similar as the number of openslides workers.
	ParallelConnections = 2

	// Same for logins
	ParallelLogins = 10

	// Same for sends in the ManySendTest
	ParallelSends = 10
)

const (
	// If ShowAllErros is true, then all errors that happen are shoun after a result
	// Else, only the first error is shown.
	ShowAllErros = true

	// If LogStatus is true, then the program shows some output while the tests are
	// running
	LogStatus = true
)

// List of all tests to performe
var Tests = []Test{
	// ConnectTest connects all clients. Measures the time until all clients are
	// connected and until they all got there first data.
	ConnectTest,

	// OneWriteTest expects the first client to be an admin client and all clients
	// to be connected. Therefore the test requires, that the ConnectTest is run
	// before. This test sends one write request with the first client and measures
	// the time until all clients get the changed data.
	OneWriteTest,

	// ManyWriteTests expects at least one client to be an admin client and all clients
	// to be connected. Therefore the test requires, that the ConnectTest is run
	// before. This test sends one write request for each admin client and measures
	// the time until all write requests are send and until all data is received.
	ManyWriteTest,

	// Keeps the connections open. This is not usual for a testrun of this program
	// but can help to open a lot of connections with this tool to test manuely
	// how OpenSlides reacts with a lot of open connections.
	//KeepOpenTest,
}
