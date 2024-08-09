package main

import (
	"crypto/rand"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"math"
	"math/big"
	"net/http"
	"os"
	"time"
)

//const otpLen = 6 // length of token to be returned
const home = `
<!doctype html>
	<html lang="en">
	<head>
		<meta http-equiv="refresh" content="10">
		<meta charset="utf-8">
	</head>
	<body>
		<h1>Verification Code</h1>
		<h2>{{.Token}}<h2>
		<h2>Refresh at: {{.Refresh}}<h2>
	</body>
</html>
` // verification page content

var page *template.Template
var DataStore = make([]string, 2)

// generateOtp creates a n-digit crypto-random token
func generateOtp(max uint32) string {
	bi, err := rand.Int(
		rand.Reader,
		big.NewInt(int64(math.Pow(10, float64(max)))),
	)
	if err != nil {
		_ = fmt.Errorf("generate: %w", err)
	}
	token := fmt.Sprintf("%0*d", max, bi)
	return token
}

// generate updates a slice with a crypto-random number of otpLen
func generate(data []string, delay time.Duration, length int) {
	otpLen := uint32(length)
	// set an initial state
	timer := time.NewTicker(delay)
	data[0] = generateOtp(otpLen)

	for range timer.C {
		// add delay to time now for reporting the refresh time to the template
		t := time.Now()
		refresh := t.Add(delay)

		data[0] = generateOtp(otpLen)
		data[1] = refresh.Format("15:04:05")
	}
}

// pageHandler is a dynamic handlerFunc for http.ServeMux
func pageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	var data struct {
		Token string
		Refresh string
	}

	data.Token   = DataStore[0]
	data.Refresh = DataStore[1]

	err := page.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	var err error

	// commandline flags
	host := flag.String("host", ":3000", "give host:port value, default :3000")
	delay := flag.Int("delay", 30, "token generation delay in seconds")
	length := flag.Int("length", 6, "token length")

	// parse commandline flags
	flag.Parse()

	// generate otp tokens asynchronously
	go generate(DataStore, time.Duration(*delay) * time.Second, *length)

	// parse constant into page data
	page, err = template.New("verify.gohtml").Parse(home)
	if err != nil {
		_ = fmt.Errorf("failed to parse template variable: %w", err)
		return
	}

	// route GET to /verify through pageHandler
	http.HandleFunc("/verify", pageHandler)
	fmt.Printf("Token generation delay: %d\n", *delay)
	fmt.Println("Starting the server on " + *host + "...")

	err = http.ListenAndServe(*host, nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}




