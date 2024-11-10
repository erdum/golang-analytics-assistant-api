package main

import (
	"analytics/config"
	"analytics/openai"
	"analytics/db"

	"strings"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"log"
	// "fmt"
)

type RequestPayload struct {
	Query string `json:"query"`
}

var (
	contextFile string
)

func main() {
	databaseCredentials := config.GetDatabaseCredentials()
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	cwd, _ := os.Getwd()

	router.LoadHTMLGlob(cwd + "/templates/*.html")
	// logFile, logFileError := os.OpenFile(cwd + "/requests_error.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 644)

	// if logFileError != nil {
	// 	log.Panic("[Error] failed to open error log file, error: " + logFileError.Error());
	// }
	// defer logFile.Close()
	// log.SetFlags(log.LstdFlags|log.LUTC|log.Lshortfile)
	// log.SetOutput(logFile)

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.POST("/", func(c *gin.Context) {
		var requestPayload RequestPayload
		c.BindJSON(&requestPayload)

		dbc, err := db.Connect(databaseCredentials[0], databaseCredentials[1], databaseCredentials[2])
		if err != nil {
			log.Println("[Error] unable to connect to database, error: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Unable to connect to database!",
				"error": err,
			})
		}
		defer dbc.Close()

		ddl, err := dbc.GetDDL()
		if err != nil {
			log.Println("[Error] unable to get database schema, error: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Unable to get database schema!",
				"error": err,
			})
		}

		systemMessages := openai.RenderSystemMessages(config.GetAnalystSystemMessages(), ddl)

		if contextFile != "" {
			contextText, err := openai.ReadFileContents(contextFile)
			
			if err != nil {
				log.Println("[Error] unable to read context file, error: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Unable to read context file!",
					"error": err,
				})
			}
			context := openai.RenderTemplate(config.GetAnalystContextMessages(), &openai.MessageData{Context: contextText})
			systemMessages = append(systemMessages, context)
		}

		analyst := openai.NewOpenAISession(systemMessages, config.GetAnalystTemperature())
		queryParser := openai.NewOpenAISession(config.GetQueryParserSystemMessages(), config.GetQueryParserTemperature())
		input := requestPayload.Query
		input = strings.TrimSpace(input)
		answer, err := openai.HandlePrompt(input, analyst, queryParser, dbc)

		if err != nil {
			log.Println("[Error] failed to handle prompt, error: " + err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "Failed to handle prompt!",
				"error": err,
			})
		}

		if answer == "" {
			log.Println("[Error] OpenAI returned empty response")
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

	err := router.Run(":8000")

	if err != nil {
    log.Panic("[Error] failed to start Gin server due to: " + err.Error())
    return
  }
}
