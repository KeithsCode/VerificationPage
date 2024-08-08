package main

import (
	"crypto/rand"
	"errors"
	"fmt"
	"html/template"
	"math"
	"math/big"
	"net/http"
	"os"
	"time"
)

const otpLen = 6
const delay = 30 * time.Second
const home = `
<!doctype html>
	<html lang="en">
	<head>
		<meta http-equiv="refresh" content="10">
		<meta charset="utf-8">
	</head>
	<body>
		<h1>Verification Code</h1>
		<h2>{{.}}<h2>
	</body>
</html>
`

var page *template.Template
var DataStore = make([]string, 1)

// generateOtp creates a 6-digit crypto-random token
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
func generate(data []string) {
	timer := time.NewTicker(delay)
	data[0] = generateOtp(otpLen)

	for range timer.C {
		data[0] = generateOtp(otpLen)
	}
}

// pageHandler is a dynamic handlerFunc for http.ServeMux
func pageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	data := "caching number"

	data = string(DataStore[0])
	err := page.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	var err error

	// generate otp tokens asynchronously
	go generate(DataStore)

	// parse constant into page data
	page, err = template.New("verify.gohtml").Parse(home)
	if err != nil {
		_ = fmt.Errorf("failed to parse template variable: %w", err)
		return
	}

	// route GET to /verify through pageHandler
	http.HandleFunc("/verify", pageHandler)
	fmt.Println("Starting the server on :3333...")

	err = http.ListenAndServe(":3333", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}




