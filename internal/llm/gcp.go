package llm

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"strings"
)

type SystemResponsePart struct {
	DetectedObjectsCount int      `json:"detected_objects_count"`
	NeedMoreData         bool     `json:"need_more_data"`
	StopKeywordsDetected bool     `json:"stop_keywords_detected"`
	StopKeywords         []string `json:"stop_keywords"`
}

type ModelResponse struct {
	AppSpecificResponse json.RawMessage     `json:"app_response"`
	System              *SystemResponsePart `json:"system"`
}

type GcpLlmProvider struct {
	projectID          string
	regionLocation     string
	serviceAccountJson []byte
	modelName          string
	client             *genai.Client
}

// NewGcpLlmProvider creates a new GcpLlmProvider with the provided projectID and regionLocation.
func NewGcpLlmProvider(projectID, regionLocation, modelName string) *GcpLlmProvider {
	return &GcpLlmProvider{projectID: projectID, regionLocation: regionLocation, modelName: modelName}
}

func (gcp *GcpLlmProvider) Init() error {
	ctx := context.Background()
	var err error
	gcp.client, err = genai.NewClient(ctx, gcp.projectID, gcp.regionLocation)

	return err
}

func (gcp *GcpLlmProvider) ImageToText(imageData []byte, prompt string) (*ModelResponse, error) {
	ctx := context.Background()
	model := gcp.client.GenerativeModel(gcp.modelName)
	model.SetTemperature(0)
	var maxOutputTokens int32 = 300
	var candidateCount int32 = 1
	model.MaxOutputTokens = &maxOutputTokens
	model.CandidateCount = &candidateCount
	imgBlob := genai.Blob{
		MIMEType: "image/jpeg",
		Data:     imageData,
	}
	promptPart := genai.Text(prompt)

	res, err := model.GenerateContent(ctx, imgBlob, promptPart)
	if err != nil {
		return nil, fmt.Errorf("unable to generate contents: %v", err)
	}

	if len(res.Candidates) == 0 ||
		len(res.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no content generated")
	}
	rawResponse := ""
	for _, candidate := range res.Candidates {
		for _, part := range candidate.Content.Parts {
			responseText := part.(genai.Text)
			rawResponse += string(responseText)
		}
	}

	response := &ModelResponse{}
	rawResponse = strings.ReplaceAll(rawResponse, "```", "")
	rawResponse = strings.ReplaceAll(rawResponse, "json", "")
	err = json.Unmarshal([]byte(rawResponse), response)
	if err != nil {
		log.Errorf("unable to unmarshal response . Error : %s , raw response: %s ", err.Error(), rawResponse)
		response.AppSpecificResponse = json.RawMessage(rawResponse)
	}
	return response, nil
}

func (gcp *GcpLlmProvider) Close() {
	gcp.client.Close()
}
