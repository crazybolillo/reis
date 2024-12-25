package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/crazybolillo/reis/event/twilio"
	"github.com/crazybolillo/reis/storage"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"os"
)

type config struct {
	Port       int    `env:"PORT" envDefault:"8080"`
	StorageURI string `env:"STORAGE_URI"`
}

func main() {
	os.Exit(run(context.Background()))
}

func run(_ context.Context) int {
	cfg := config{}
	if err := env.Parse(&cfg); err != nil {
		slog.Error("Failed to parse config", "reason", err)
		return 1
	}

	backend, err := storage.Parse(cfg.StorageURI)
	if err != nil {
		slog.Error("Failed to parse storage URI", "reason", err)
		return 1
	}

	upgrader := websocket.Upgrader{}
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			slog.Error("Failed to upgrade connection", "reason", err)
			return
		}
		defer c.Close()
		handleStream(c, backend)
	})

	slog.Info("Starting http server", "port", cfg.Port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil)
	if err != nil {
		slog.Error("Failed to start http server", "reason", err)
		return 1
	}

	return 0
}

func handleStream(c *websocket.Conn, backend storage.Backend) {
	var callSid string
	var record storage.Record
	defer func() {
		if record != nil {
			err := record.Close()
			if err != nil {
				slog.Error("Failed to close record", "reason", err, "callSid", callSid)
			}
		}
	}()

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			slog.Error("Failed to read message", "reason", err)
			return
		}

		var event twilio.Event
		if err := json.Unmarshal(message, &event); err != nil {
			slog.Error("Failed to obtain event from message", "reason", err)
			return
		}

		switch event.Event {
		case "start":
			var start twilio.Start
			if err := json.Unmarshal(message, &start); err != nil {
				slog.Error("Failed to parse start event", "reason", err)
				return
			}
			callSid = start.Start.CallSid
			record, err = backend.New(callSid)
			if err != nil {
				slog.Error("Failed to create storage record", "reason", err, "callSid", callSid)
				return
			}
			slog.Info("Created storage record", "callSid", callSid)
		case "media":
			var media twilio.Media
			if err := json.Unmarshal(message, &media); err != nil {
				slog.Error("Failed to decode media message", "reason", err, "callSid", callSid)
				continue
			}
			dst := make([]byte, base64.StdEncoding.DecodedLen(len(media.Media.Payload)))
			size, err := base64.StdEncoding.Decode(dst, []byte(media.Media.Payload))
			if err != nil {
				slog.Error("Failed to decode media payload", "reason", err)
				continue
			}
			_, err = record.Write(dst[:size])
			if err != nil {
				slog.Error("Failed to write record", "reason", err, "callSid", callSid)
			}
		case "stop":
			slog.Info("Received stop message", "callSid", callSid)
			return
		}
	}
}
