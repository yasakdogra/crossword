module server

go 1.22.6

replace crossword => ../crossword

require (
	crossword v0.0.0-00010101000000-000000000000
	github.com/openai/openai-go v0.1.0-alpha.19
)

require (
	github.com/tidwall/gjson v1.17.3 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.1 // indirect
	github.com/tidwall/sjson v1.2.5 // indirect
)
