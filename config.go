package main

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/viper"
)

//go:embed config.yaml
var configFile embed.FS

func GetAPIKey() string {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// Modify the init function to use the embedded config.yaml
	f, err := fs.ReadFile(configFile, "config.yaml")
	if err != nil {
		panic(err)
	}

	err = viper.ReadConfig(bytes.NewReader(f))
	if err != nil {
		panic(err)
	}

	apiKey := viper.GetString("openai.api_key")
	
	if apiKey == "" {
		fmt.Println("OPENAI_API_KEY env var not set")
		os.Exit(-1)
	}
	return apiKey
}

func GetDatabaseCredentials() []string {
	return viper.GetStringSlice("openai.database")
}

func GetAnalystSystemMessages() []string {
	return viper.GetStringSlice("openai.analyst.messages.system_messages")
}

func GetAnalystContextMessages() string {
	return viper.GetString("openai.analyst.messages.context_message")
}

func GetAnalystQueryResultsMessage() string {
	return viper.GetString("openai.analyst.messages.query_results_message")
}

func GetQueryParserSystemMessages() []string {
	return viper.GetStringSlice("openai.query_parser.messages.system_messages")
}

func GetQueryParserMessage() string {
	return viper.GetString("openai.query_parser.messages.parse_query_message")
}

func GetAnalystTemperature() float32 {
	return float32(viper.GetFloat64("openai.analyst.temperature"))
}

func GetQueryParserTemperature() float32 {
	return float32(viper.GetFloat64("openai.query_parser.temperature"))
}
