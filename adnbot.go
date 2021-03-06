package main

import (
	"encoding/json"
	"flag"
	"log"
	"mxml/makini/api"
	"mxml/makini/listener"
	"mxml/makini/stream"
	"net/url"
	"os"
)

type Config struct {
	ADN struct {
		TokenURLBase      string `json:"token_url_base"`
		TokenHostOverride string `json:"token_host_override"`
		APIURLBase        string `json:"api_url_base"`
		APIHostOverride   string `json:"api_host_override"`
		StreamURLOverride string `json:"stream_url_override"`
		ClientID          string `json:"client_id"`
		ClientSecret      string `json:"client_secret"`
		UserID            string `json:"user_id"`
		StreamKey         string `json:"stream_key"`
	} `json:"adn"`
}

var (
	file = flag.String("config", "config.json", "JSON config file")
)

func main() {
	flag.Parse()

	file, err := os.Open(*file)
	if err != nil {
		log.Fatalf("Error loading config (%q): %s", *file, err)
	}

	decoder := json.NewDecoder(file)
	var config Config
	if err = decoder.Decode(&config); err != nil {
		log.Fatalf("Error decoding config: %s", err)
	}

	api.TokenURLBase = config.ADN.TokenURLBase
	api.TokenHostOverride = config.ADN.TokenHostOverride
	api.APIURLBase = config.ADN.APIURLBase
	api.APIHostOverride = config.ADN.APIHostOverride
	api.ClientID = config.ADN.ClientID
	api.ClientSecret = config.ADN.ClientSecret

	user, err := api.GetUserByID(config.ADN.UserID, []string{
		"messages",
	}, nil)

	if err != nil {
		log.Fatal(err)
	}

	listener.UserID = user.UserID()

	appClient, err := api.GetToken(map[string]string{
		"grant_type": "client_credentials",
	})

	if err != nil {
		log.Fatal(err)
	}

	stream_endpoint := appClient.GetStreamEndpoint(config.ADN.StreamKey)

	if config.ADN.StreamURLOverride != "" {
		stream_url, err := url.Parse(stream_endpoint)
		if err != nil {
			log.Fatal(err)
		}

		override_url, err := url.Parse(config.ADN.StreamURLOverride)
		if err != nil {
			log.Fatal(err)
		}

		// current stream URLs only need to match on path
		override_url.Path = stream_url.Path

		stream_endpoint = override_url.String()
	}

	messages := stream.ProcessStream(stream_endpoint)
	listener.ProcessMessages(user, messages)
}
