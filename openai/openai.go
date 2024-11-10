package openai

import (
	"analytics/config"
	"analytics/db"

	"context"
	"os"
	"text/template"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"strings"
)

type Session struct {
	temperature float32
	messages    []openai.ChatCompletionMessage
}

type MessageData struct {
	Ddl string
	Query string
	QueryResults string
	Prompt string
	Context string
}

var (
	openAiClient = openai.NewClient(config.GetAPIKey())
	logSql bool = false
)

func NewOpenAISession(systemMessages []string, temperature float32) *Session {
	messages := make([]openai.ChatCompletionMessage, len(systemMessages))

	for i := range systemMessages {
		messages[i] = openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemMessages[i],
		}
	}

	return &Session{
		temperature: temperature,
		messages:    messages,
	}
}

func (s *Session) SystemPrompt(prompt string) string {
	return s.prompt(prompt, openai.ChatMessageRoleSystem)
}

func (s *Session) UserPrompt(prompt string) string {
	return s.prompt(prompt, openai.ChatMessageRoleUser)
}

func (s *Session) prompt(prompt string, role string) string {
	s.messages = append(s.messages, openai.ChatCompletionMessage{
		Role:    role,
		Content: prompt,
	})

	resp, err := openAiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT3Dot5Turbo,
			Messages:    s.messages,
			Temperature: s.temperature,
		},
	)

	if err != nil {
		fmt.Println(err)
		panic("Error prompting open ai")
	}

	if len(resp.Choices) == 0 {
		panic("no choices returned from the OpenAI API")
	}

	s.messages = append(s.messages, resp.Choices[0].Message)

	return strings.TrimSpace(resp.Choices[0].Message.Content)
}

func HandlePrompt(input string, analyst *Session, queryParser *Session, dbc *db.DBConnection) (string, error) {
	response := analyst.UserPrompt(input)

	m := RenderTemplate(config.GetQueryParserMessage(), &MessageData{
		Query: response,
	})

	query := queryParser.UserPrompt(m)
	if query == "No query was found." {
		return response, nil
	}

	queryResult, err := dbc.ExecuteQuery(query, logSql)
	if err != nil {
		//TODO: implement retry
		return "", err
	}

	m = RenderTemplate(config.GetAnalystQueryResultsMessage(), &MessageData{
		QueryResults: queryResult,
		Query:        query,
		Prompt:       input,
	})

	return analyst.SystemPrompt(m), nil
}

func RenderSystemMessages(messageTemplates []string, ddl string) []string {
	m := make([]string, len(messageTemplates))

	data := &MessageData{
		Ddl: ddl,
	}

	for i := range messageTemplates {
		renderedTemplate := RenderTemplate(messageTemplates[i], data)

		m[i] = renderedTemplate
	}

	return m
}

func RenderTemplate(tmpl string, data *MessageData) string {
	t, err := template.New("message").Parse(tmpl)
	if err != nil {
		panic("Error parsing template")
	}

	var buf strings.Builder
	err = t.Execute(&buf, data)
	if err != nil {
		panic("Error rendering the template")
	}

	return buf.String()
}

func ReadFileContents(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
