package main

import (
	"fmt"
	"strings"
	"github.com/gin-gonic/gin"
	"net/http"
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
	contextFile string
	logSql bool = false
)

func main() {
	databaseCredentials := GetDatabaseCredentials()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.POST("/", func(c *gin.Context) {
		var requestPayload RequestPayload
		c.BindJSON(&requestPayload)

		dbc, err := Connect(databaseCredentials[0], databaseCredentials[1], databaseCredentials[2])
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
		input := requestPayload.Query
		input = strings.TrimSpace(input)
		answer, err := handlePrompt(input, analyst, queryParser, dbc)

		if err != nil {
			fmt.Println("Error handling prompt", err)
			return
		}

		if answer == "" {
			c.AbortWithStatus(400)
			return
		}
		
		c.JSON(200, gin.H{
			"answer": answer,
		})
		return
	})

	err := router.Run(":80")

	if err != nil {
        panic("[Error] failed to start Gin server due to: " + err.Error())
        return
    }
}
