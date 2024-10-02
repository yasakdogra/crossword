package main

import (
	"context"
	"crossword"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

type NewCrosswordResponse struct {
	Rows    int
	Columns int

	Grid []byte

	Across []interface{}
	Down   []interface{}
}

type Target struct {
	Row    int
	Column int
	Across bool
}

var wordList crossword.WordList

func initWordList() {
	wordList = make([]string, 5+1)
	for i := 2; i < 5+1; i++ {
		tmp, err := os.ReadFile("../generator/words-" + strconv.Itoa(i) + ".txt")
		if err != nil {
			log.Fatal(err)
		}
		wordList[i] = string(tmp)
	}
}

func main() {
	initWordList()

	http.HandleFunc("/new", func(w http.ResponseWriter, r *http.Request) {
		c := crossword.New(5, 5)

		numBlanks := rand.Intn(2)
		fmt.Println(numBlanks)
		for i := 0; i < numBlanks+1; i++ {
			x := rand.Intn(5)
			y := rand.Intn(5)
			c.SetCell(x, y, 45)
			c.SetCell(4-x, 4-y, 45)
		}
		fmt.Println(c.ToString())

		if !c.Fill(&wordList) {
			fmt.Println("Failed to fill crossword")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		across := make([]interface{}, 0)
		down := make([]interface{}, 0)

		entries := c.GetEntries()
		answers := []string{}

		for _, entry := range entries {
			answers = append(answers, entry.Word)
		}

		client := openai.NewClient(
			option.WithAPIKey(
				os.Getenv("OPENAI_API_KEY"),
			),
		)

		response, err := client.Chat.Completions.New(
			context.Background(),
			openai.ChatCompletionNewParams{
				Model: openai.F(openai.ChatModelGPT4o),
				Messages: openai.F([]openai.ChatCompletionMessageParamUnion{
					openai.SystemMessage("You are a crossword clue generator. You will read a comma separated list of words and generate clues for them. You will respond in json format with the each word as name and its clue as value. Some words are a combination of multiple words. Pay attention to prefixes, suffixes and plurals. Try to be accurate and concise. Do not use the word or the form of the abbreviations in the clue."),
					openai.UserMessage(`DOG,AG,ETC`),
					openai.AssistantMessage(`{"DOG":"Man's best friend","AG":"Symbol for silver","ETC":"Other similar things (abbr)"}`),
					openai.UserMessage(strings.Join(answers, ",")),
				}),
			},
		)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(response.Choices[0].Message.Content)

		hints := map[string]string{}

		json.Unmarshal([]byte(response.Choices[0].Message.Content), &hints)

		for _, entry := range c.GetEntries() {
			if entry.Horizontal {
				across = append(across, []interface{}{entry.Row, entry.Column, len(entry.Word), hints[entry.Word]})
			} else {
				down = append(down, []interface{}{entry.Row, entry.Column, len(entry.Word), hints[entry.Word]})
			}
		}

		data := NewCrosswordResponse{Rows: c.Height, Columns: c.Width, Grid: c.ToByteArray(), Across: across, Down: down}

		j, err := json.Marshal(data)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")
		w.Write(j)

	})

	http.ListenAndServe(":8080", nil)
}
