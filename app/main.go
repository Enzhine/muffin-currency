package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
)

type Config struct {
	Rates map[string]map[string]float64
	Port  string
}

func parseJsonFile(filename string) (parsed map[string]interface{}) {
	file, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	err = json.Unmarshal(file, &parsed)
	if err != nil {
		panic(err)
	}

	return
}

func getConfig() (cfg Config) {
	json := parseJsonFile("application.json")

	// Config low priority

	if rates, has := json["rates"]; has {
		if value, has := rates.(map[string]map[string]float64); has {
			cfg.Rates = value
		} else {
			panic("Config 'rates' key must contain 'map[string]map[string]float64' value")
		}
	} else {
		// default value
		cfg.Rates = map[string]map[string]float64{
			"CARAMEL":   {"CHOKOLATE": 0.85, "PLAIN": 75.50},
			"CHOKOLATE": {"CARAMEL": 1.18, "PLAIN": 89.00},
			"PLAIN":     {"CHOKOLATE": 0.013, "CARAMEL": 0.011},
		}
	}

	if port, has := json["port"]; has {
		if value, has := port.(string); has {
			cfg.Port = value
		} else {
			panic("Config 'port' key must contain 'string' value")
		}
	} else {
		// default value
		cfg.Port = "8080"
	}

	// Envs high priority

	if port_env, has := os.LookupEnv("PORT"); has {
		cfg.Port = port_env
	}

	return
}

type CurrencyRate struct {
	From string  `json:"from"`
	To   string  `json:"to"`
	Rate float64 `json:"rate"`
}

var config Config

func getRateHandler(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	if from == "" || to == "" {
		http.Error(w, "Missing 'from' or 'to' parameter", http.StatusBadRequest)
		return
	}

	rates := config.Rates

	rate, exists := rates[from][to]
	if !exists {
		http.Error(w, "Currency pair not found", http.StatusNotFound)
		return
	}

	response := CurrencyRate{
		From: from,
		To:   to,
		Rate: rate,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Init global config values
	config = getConfig()

	// REST route handlers
	http.HandleFunc("/rate", getRateHandler)

	// Start app
	port := config.Port
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
