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

func defaultConfig() Config {
	return Config{
		Rates: map[string]map[string]float64{
			"CARAMEL":   {"CHOKOLATE": 0.85, "PLAIN": 75.50},
			"CHOKOLATE": {"CARAMEL": 1.18, "PLAIN": 89.00},
			"PLAIN":     {"CHOKOLATE": 0.013, "CARAMEL": 0.011},
		},
		Port: "8080",
	}
}

func populateConfig(cfg *Config, path string) error {
	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var parsedConfig Config
	err = json.Unmarshal(file, &parsedConfig)
	if err != nil {
		return err
	}

	if parsedConfig.Port != "" {
		cfg.Port = parsedConfig.Port
	}
	if len(parsedConfig.Rates) > 0 {
		cfg.Rates = parsedConfig.Rates
	}

	return nil
}

func populateConfigByEnvs(cfg *Config) {
	if port_env, has := os.LookupEnv("PORT"); has {
		cfg.Port = port_env
	}
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
	config = defaultConfig()
	for _, path := range []string{"application.json", "config/application.json"} {
		err := populateConfig(&config, path)
		if err == nil {
			log.Printf("Applied '%s' config", path)
		}
	}
	populateConfigByEnvs(&config)

	// REST route handlers
	http.HandleFunc("/rate", getRateHandler)

	// Start app
	port := config.Port
	log.Printf("Server running on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
