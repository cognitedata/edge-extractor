package llm

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type Pipe struct {
	Text string `json:"text"`
	Rust bool   `json:"rust"`
}

type PipesData struct {
	Pipes      []Pipe `json:"pipes"`
	TotalPipes int    `json:"total_pipes"`
}

func TestImageToTextNoText(t *testing.T) {
	provider := NewGcpLlmProvider("cognite-sa-sandbox", "europe-west4", "gemini-1.5-flash")
	err := provider.Init()
	if err != nil {
		t.Errorf("Can't init llm provider.unexpected error: %v", err)
		t.Fail()
	}
	imageData, err := os.ReadFile("testdata/pipe-no-text.jpeg")
	if err != nil {
		t.Errorf("Can't read image file. Unexpected error: %v", err)
		t.Fail()
	}

	prompt := `The picture contains horizontal metal pipe with printed text on each pipe.
	 Extract text from each pipe and total number of pipes on picture. Set stop_keyword to true if the keyword is found in the text otherwise .
	 Ignore text:Not enough power, product requires POE class 4 or higher. 
	 Stop keyword: "HEAT NO"
	 Response in JSON format:
	  { "pipes": [ { "text": "text1","rust":true }, { "text": "text2","rust":false } ], "total_pipes": 0 }`

	response, err := provider.ImageToText(imageData, prompt)
	if err != nil {
		t.Errorf("Can't generate content. Unexpected error: %v", err)
		t.Fail()
	}
	if response == "" {
		t.Errorf("No content generated")
		t.Fail()
	}
	t.Logf("Response: %s", response)
	pipeResponse := PipesData{}
	err = json.Unmarshal([]byte(response), &pipeResponse)
	if err != nil {
		t.Errorf("Can't unmarshal response. Unexpected error: %v", err)
		t.Fail()
	}
	if pipeResponse.TotalPipes != 1 {
		t.Errorf("Expected total pipes: 1, got: %d", pipeResponse.TotalPipes)
		t.Fail()
	}
	if len(pipeResponse.Pipes) != 0 {
		t.Errorf("Response should not contain any pipes. Got: %v", pipeResponse.Pipes)
		t.Fail()
	}
}

func TestImageToText(t *testing.T) {
	provider := NewGcpLlmProvider("cognite-sa-sandbox", "europe-west4", "gemini-1.5-flash")
	err := provider.Init()
	if err != nil {
		t.Errorf("Can't init llm provider.unexpected error: %v", err)
		t.Fail()
	}
	imageData, err := os.ReadFile("testdata/pipe-with-text.jpg")
	if err != nil {
		t.Errorf("Can't read image file. Unexpected error: %v", err)
		t.Fail()
	}

	prompt := `The picture contains several horizontal metal pipes with printed text on each pipe.
	 Extract text from each pipe and count total number of pipes on the picture. Report text for each pipe separately.
	 Response in JSON format:
	  { "pipes": [ { "text": "text1","rust":true }, { "text": "text2","rust":false } ], "total_pipes": 0 }`

	response, err := provider.ImageToText(imageData, prompt)
	if err != nil {
		t.Errorf("Can't generate content. Unexpected error: %v", err)
		t.Fail()
	}
	if response == "" {
		t.Errorf("No content generated")
		t.Fail()
	}
	t.Logf("Response: %s", response)
	pipeResponse := PipesData{}
	err = json.Unmarshal([]byte(response), &pipeResponse)
	if err != nil {
		t.Errorf("Can't unmarshal response. Unexpected error: %v", err)
		t.Fail()
	}
	if pipeResponse.TotalPipes != 1 {
		t.Errorf("Expected total pipes: 1, got: %d", pipeResponse.TotalPipes)
		t.Fail()
	}
	if len(pipeResponse.Pipes) != 1 {
		t.Errorf("Response should not contain any pipes. Got: %v", pipeResponse.Pipes)
		t.Fail()
	}
	responseText := pipeResponse.Pipes[0].Text
	responseText = strings.ReplaceAll(responseText, "\n", "")
	if responseText != "TC 153-21-2007426 10 0972C4677 A" {
		t.Errorf("Wrong response text , got: %s", pipeResponse.Pipes[0].Text)
		t.Fail()
	}
}
