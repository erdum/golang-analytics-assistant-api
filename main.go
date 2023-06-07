package main

import (
	"fmt"
	"strings"
	"github.com/gin-gonic/gin"
)

type RequestPayload struct {
	Query string `json:"query"`
}

type MessageData struct {
	Ddl string
	Query string
	QueryResults string
	Prompt string
	Context string
}

var (
	dbURL string
	dbUsername string
	dbPassword string
	contextFile string
	logSql bool = false
)

func main() {
	router := gin.Default()

	router.POST("/", func(c *gin.Context) {
		var requestPayload RequestPayload
		c.BindJSON(&requestPayload)

		dbc, err := Connect(dbURL, dbUsername, dbPassword)
		if err != nil {
			fmt.Println("Error connecting to the MySQL database:", err)
			return
		}

		defer dbc.Close()

		ddl, err := dbc.GetDDL()
		if err != nil {
			fmt.Println("Error getting the database schema:", err)
			return
		}

		systemMessages := renderSystemMessages(GetAnalystSystemMessages(), ddl)

		if contextFile != "" {
			c, err := readFileContents(contextFile)
			if err != nil {
				fmt.Println("Error reading file", err)
				return
			}
			context := renderTemplate(GetAnalystContextMessages(), &MessageData{Context: c})

			systemMessages = append(systemMessages, context)
		}

		analyst := NewOpenAISession(systemMessages, GetAnalystTemperature())
		queryParser := NewOpenAISession(GetQueryParserSystemMessages(), GetQueryParserTemperature())

		for {
			input := requestPayload.Query
			input = strings.TrimSpace(input)

			answer, err := handlePrompt(input, analyst, queryParser, dbc)
			if err != nil {
				fmt.Println("Error handling prompt", err)
				return
			}
			c.Data(200, "application/json", []byte(answer))
			break
		}

	})

	router.Run()
}
