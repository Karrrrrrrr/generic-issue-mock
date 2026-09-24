package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const maxBodyBytes = 4 << 20

type acknowledgement struct {
	Roger bool `json:"roger"`
}

func main() {
	address := flag.String("addr", "127.0.0.1:9000", "HTTP listen address")
	flag.Parse()
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lmicroseconds)

	handler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost {
			writer.Header().Set("Allow", http.MethodPost)
			http.Error(writer, "use POST to deliver a webhook", http.StatusMethodNotAllowed)
			return
		}

		request.Body = http.MaxBytesReader(writer, request.Body, maxBodyBytes)
		defer request.Body.Close()
		dump, err := httputil.DumpRequest(request, true)
		if err != nil {
			var tooLarge *http.MaxBytesError
			if errors.As(err, &tooLarge) {
				http.Error(writer, "webhook body exceeds 4 MiB", http.StatusRequestEntityTooLarge)
				return
			}
			logger.Printf("read webhook from %s: %v", request.RemoteAddr, err)
			http.Error(writer, "cannot read webhook request", http.StatusBadRequest)
			return
		}

		logger.Printf(
			"\n========== WEBHOOK ==========\nRemote: %s\n%s\n========== END ==========",
			request.RemoteAddr,
			dump,
		)
		writer.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := json.NewEncoder(writer).Encode(acknowledgement{Roger: true}); err != nil {
			logger.Printf("write webhook acknowledgement: %v", err)
		}
	})

	server := &http.Server{
		Addr:              *address,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          logger,
	}
	logger.Printf("webhook receiver listening on %s; POST to any path", *address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Fatal(err)
	}
}
