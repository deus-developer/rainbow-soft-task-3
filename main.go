package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/gorilla/websocket"
)

type generatorOptions struct {
	CountNumbers int `schema:"countNumbers,required"`
	CountThreads int `schema:"countThreads,required"`
}

const (
	maxCountNumbers = 2147483647
	maxCountThreads = 8
)

// Decoder http url query params
var decoder = schema.NewDecoder()

// Upgrader http to websocket
var upgrader = websocket.Upgrader{
	ReadBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "http://127.0.0.1:8080" // Allow only site origin
	},
}

// Filter unique random numbers and send to output channel
func uniqueNumbersGen(numbersGen chan int, outputNumbersGen chan int, count int, quit chan struct{}) {
	numbers := make(map[int]bool)
	defer close(quit)
	defer close(outputNumbersGen)

	for number := range numbersGen {
		if numbers[number] {
			continue
		}
		numbers[number] = true

		outputNumbersGen <- number
		count--

		if count < 1 {
			return
		}
	}
}

// Worker for generation random numbers to channel
func randomGenWorker(numbersGen chan int, quit chan struct{}, maxNumberLimit int) {
	for {
		select {
		case <-quit:
			return
		default:
			numbersGen <- rand.Intn(maxNumberLimit)
		}
	}
}

// Create random generators and return filtered(uniqued) numbers channel
func generateRandomNumbers(countNumbers int, threadsCount int) (<-chan int, chan struct{}) {
	numbersGen := make(chan int, threadsCount)
	quit := make(chan struct{})

	for i := 1; i <= threadsCount; i++ {
		go randomGenWorker(numbersGen, quit, countNumbers)
	}

	outputNumbersGen := make(chan int)
	go uniqueNumbersGen(numbersGen, outputNumbersGen, countNumbers, quit)
	return outputNumbersGen, quit
}

// Create socket random generator numbers handler
func getRandomGenerator(w http.ResponseWriter, r *http.Request) {
	var options generatorOptions

	// Parsing generator options from GET query params
	err := decoder.Decode(&options, r.URL.Query())
	if err != nil {
		http.Error(w, "Invalid generator options", 400)
		return
	}

	if options.CountNumbers < 1 || options.CountNumbers > maxCountNumbers {
		http.Error(w, "Invalid countNumbers", 400)
		return
	}

	if options.CountThreads < 1 || options.CountThreads > maxCountThreads {
		http.Error(w, "Invalid countThreads", 400)
		return
	}

	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade connection from %s error", r.RemoteAddr)
		return
	}

	// Start generation numbers
	go generatorWriter(connection, options)
}

// generate random numbers by options and write to websocket
func generatorWriter(connection *websocket.Conn, options generatorOptions) {
	defer connection.Close()

	numbersGen, quit := generateRandomNumbers(options.CountNumbers, options.CountThreads)

	for number := range numbersGen {
		err := connection.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(number)))
		if err != nil {
			// Close connection if error on write to socket
			close(quit)
			return
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/static/", http.StripPrefix("/static", neuter(fs)))
	http.HandleFunc("/generator", getRandomGenerator)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/index.html")
	})

	log.Println("Starting server at port 8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

// Middleware for hide listing files in directory
func neuter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}
