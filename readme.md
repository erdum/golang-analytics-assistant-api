# Analytics Assistant CLI
The Data Analytics Assistant is a command-line interface (CLI) tool designed to assist users in obtaining insights from their MySQL databases using natural language questions. By leveraging the power of OpenAI's API, users can easily ask questions about their data, and the assistant will provide clear, concise answers based on the information stored in the database. This tool eliminates the need for users to write complex SQL queries, empowering users with little or no SQL knowledge to make data-driven decisions.

## Requirements
- Go 1.18 or later
- A MySQL database
- An OpenAI API key
## Installation
- Clone the repository:
```bash
git clone https://github.com/yourusername/openai-analytics-assistant.git
```
Navigate to the project directory:
```bash
cd openai-analytics-assistant
```
Build the CLI tool:
```bash
go build -o analytics-assistant
```
This will create an executable named analytics-assistant in the current directory.

## Usage
```
Usage:
analytics-assistant session [flags]

Flags:
-c, --context-file string   Path to a file containing business context
-p, --db-password string    MySQL database password
-u, --db-url string         MySQL database URL
-n, --db-username string    MySQL database username
-h, --help                  help for session
-s, --log-sql               Log SQL
```

To use the CLI tool, you will first need to set and environment variable with the openai API key:
```bash
export OPENAI_API_KEY="{you_openai_api_key}"
```

And then run:
```bash
./analytics-assistant session --db-url "tcp(localhost:3306)/your_database_name" --db-username "your_username" --db-password "your_password"
```
Replace `your_database_name`, `your_username`, and `your_password` with the appropriate values for your MySQL database. This will initiate an interactive session with the data analyst assistant.

You can also provide a file containing business rules or database field descriptions to give the assistant more context when answering questions. Use the --context-file flag to achieve this.

### Demo
![DEMO_1](./assets/demo_1.gif)

## License
This project is licensed under the [MIT License](./LICENSE).