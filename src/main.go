package main

import (
	"strings"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"log"
	"path/filepath"
	"github.com/gin-gonic/contrib/sessions"
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
	cwd, _ := os.Executable()
	cwd = filepath.Dir(cwd)
	
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	
	logFile, logFileError := os.OpenFile(cwd + "/requests_error.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 644)
	sessionStore := sessions.NewCookieStore([]byte(GetSessionKey()))
	router.Use(sessions.Sessions("user-session", sessionStore))

	databaseCredentials := GetDatabaseCredentials()

	if logFileError != nil {
		log.Panic("[Error] failed to open error log file, error: " + logFileError.Error());
	}
	defer logFile.Close()
	log.SetFlags(log.LstdFlags|log.LUTC|log.Lshortfile)
	log.SetOutput(logFile)

	router.GET("/", func(c *gin.Context) {
		router.LoadHTMLFiles(cwd + "/templates/index.html")
		session := sessions.Default(c)
		user := session.Get("user")
		isUserAuthenticated := false

		if user == "Authenticated" {
			isUserAuthenticated = true
		} else {
			session.Set("user", "Authenticated")
			session.Save()
		}

		c.HTML(http.StatusOK, "index.html", gin.H{
			"isUserAuthenticated": isUserAuthenticated,
		})
	})

	router.POST("/", func(c *gin.Context) {
		var requestPayload RequestPayload
		c.BindJSON(&requestPayload)

		dbc, err := Connect(databaseCredentials[0], databaseCredentials[1], databaseCredentials[2])
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

		systemMessages := renderSystemMessages(GetAnalystSystemMessages(), ddl)

		if contextFile != "" {
			contextText, err := readFileContents(contextFile)
			
			if err != nil {
				log.Println("[Error] unable to read context file, error: " + err.Error())
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Unable to read context file!",
					"error": err,
				})
			}
			context := renderTemplate(GetAnalystContextMessages(), &MessageData{Context: contextText})
			systemMessages = append(systemMessages, context)
		}

		analyst := NewOpenAISession(systemMessages, GetAnalystTemperature())
		queryParser := NewOpenAISession(GetQueryParserSystemMessages(), GetQueryParserTemperature())
		input := requestPayload.Query
		input = strings.TrimSpace(input)
		answer, err := handlePrompt(input, analyst, queryParser, dbc)

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

	err := router.Run(":80")

	if err != nil {
    log.Panic("[Error] failed to start Gin server due to: " + err.Error())
    return
  }
}
