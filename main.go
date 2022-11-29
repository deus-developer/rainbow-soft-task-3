package main

import (
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
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
func generateRandomNumbers(countNumbers int, threadsCount int) <-chan int {
	numbersGen := make(chan int, threadsCount)
	quit := make(chan struct{})

	for i := 1; i <= threadsCount; i++ {
		go randomGenWorker(numbersGen, quit, countNumbers)
	}

	outputNumbersGen := make(chan int)
	go uniqueNumbersGen(numbersGen, outputNumbersGen, countNumbers, quit)
	return outputNumbersGen
}

// Create socket random generator numbers handler
func getRandomGenerator(w http.ResponseWriter, r *http.Request) {
	connection, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade connection from %s error", r.RemoteAddr)
		return
	}
	go generatorListener(connection)
}

// listen options and generate random numbers
func generatorListener(connection *websocket.Conn) {
	defer connection.Close()

	type generatorOptions struct {
		CountNumbers int `json:"countNumbers"`
		CountThreads int `json:"countThreads"`
	}
	type generatorResult struct {
		Ok     bool   `json:"ok"`
		Err    string `json:"err"`
		Result []int  `json:"result"`
	}
	for {
		var message generatorOptions
		err := connection.ReadJSON(&message)

		if err != nil {
			log.Println(err)
			break // Break cycle and close connection if error
		}
		var response generatorResult

		if message.CountNumbers < 1 || message.CountNumbers > 2147483647 {
			response.Ok = false
			response.Err = "Кол-во чисел должно быть больше 0 и не более 2147483647"
		} else if message.CountThreads < 1 || message.CountThreads > 32 {
			response.Ok = false
			response.Err = "Кол-во потоков должно быть больше 0 и не более 32"
		} else {
			response.Ok = true

			// Read random unique numbers from channel
			numbersGen := generateRandomNumbers(message.CountNumbers, message.CountThreads)
			for number := range numbersGen {
				response.Result = append(response.Result, number)
			}
		}

		err = connection.WriteJSON(response)
		if err != nil {
			log.Println(err)
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
