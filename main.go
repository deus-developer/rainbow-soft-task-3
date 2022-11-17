package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

// Render JSON response
func renderJson(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

// Generate random numbers handler
func getRandom(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(1 << 10) // 1kb form limit
	if err != nil {
		log.Println(err)
		renderJson(w, make([]int, 0))
		return
	}

	countNumbers, err := strconv.Atoi(r.PostFormValue("countNumbers"))
	if err != nil || countNumbers < 1 || countNumbers > 2147483647 {
		log.Println(err)
		renderJson(w, make([]int, 0))
		return
	}

	countThreads, err := strconv.Atoi(r.PostFormValue("countThreads"))
	if err != nil || countThreads < 1 || countThreads > 32 {
		log.Println(err)
		renderJson(w, make([]int, 0))
		return
	}

	// Read random unique numbers from channel
	var numbers []int
	numbersGen := generateRandomNumbers(countNumbers, countThreads)
	for number := range numbersGen {
		numbers = append(numbers, number)
	}
	log.Println(numbers)
	renderJson(w, numbers)
}

func main() {
	rand.Seed(time.Now().UnixNano())

	fs := http.FileServer(http.Dir("./public"))
	http.Handle("/static/", http.StripPrefix("/static", neuter(fs)))
	http.HandleFunc("/random", getRandom)
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
