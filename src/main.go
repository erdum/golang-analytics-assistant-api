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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Unable to connect to database!",
				"error": err,
			})
		}
		defer dbc.Close()

		ddl, err := dbc.GetDDL()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Unable to get database schema!",
				"error": err,
			})
		}

		systemMessages := renderSystemMessages(GetAnalystSystemMessages(), ddl)

		if contextFile != "" {
			c, err := readFileContents(contextFile)
			
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Unable to read context file!",
					"error": err,
				})
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to handle prompt!",
				"error": err,
			})
		}

		if answer == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "OpenAI returned empty response!",
				"error": nil,
			})
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
