package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

type Openai struct {
	client      *openai.Client
	AssistantId string
	ThreadID    string
	FileIds     []string
}

func NewClient() *Openai {
	token := "sk-xxx"
	proxyUrl := "http://127.0.0.1:7890"

	clientConfig := openai.DefaultConfig(token)

	var tr *http.Transport
	var newClient *openai.Client

	if proxyUrl != "" {
		proxy, _ := url.Parse(proxyUrl)
		tr = &http.Transport{
			Proxy:           http.ProxyURL(proxy),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		clientConfig.HTTPClient = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 20,
		}
		newClient = openai.NewClientWithConfig(clientConfig)

	} else {
		newClient = openai.NewClient(token)

	}

	return &Openai{
		client: newClient,
	}
}

func (o *Openai) CreateAssistant(file string) (*openai.Assistant, error) {
	fileRequest := openai.FileRequest{
		FilePath: file,
		Purpose:  "assistants",
	}

	osFile, err := o.client.CreateFile(context.Background(), fileRequest)
	if err != nil {
		fmt.Printf("openai create file error %v", err)
		return nil, err
	}

	assistantName := "文学小助理"
	instructions := "你是一个文学小助理，可以帮我总结一些小说，"

	o.FileIds = []string{osFile.ID}

	request := openai.AssistantRequest{
		Model:        "gpt-3.5-turbo-1106",
		Name:         &assistantName,
		Instructions: &instructions,
		Tools: []openai.AssistantTool{
			{
				Type: "retrieval",
			},
		},
		FileIDs: []string{osFile.ID},
	}

	assistant, err := o.client.CreateAssistant(context.Background(), request)
	if err != nil {
		return nil, err
	}
	o.AssistantId = assistant.ID

	return &assistant, nil
}

func (o *Openai) Run(prompt string) (string, error) {

	if o.ThreadID == "" {
		request := openai.ThreadRequest{}
		thread, err := o.client.CreateThread(context.Background(), request)
		if err != nil {
			return "", err
		}
		o.ThreadID = thread.ID
	}

	messageReq := openai.MessageRequest{
		Role:    "user",
		Content: prompt,
		FileIds: o.FileIds,
	}

	_, err := o.client.CreateMessage(context.Background(), o.ThreadID, messageReq)
	if err != nil {
		return "", err
	}

	runReq := openai.RunRequest{
		AssistantID: o.AssistantId,
	}

	runResp, err := o.client.CreateRun(context.Background(), o.ThreadID, runReq)
	if err != nil {
		return "", err
	}

	var run openai.Run
	for {
		run, err = o.client.RetrieveRun(context.Background(), o.ThreadID, runResp.ID)
		if err != nil {
			return "", err
		}

		if run.Status != openai.RunStatusQueued && run.Status != openai.RunStatusInProgress {
			break
		}

	}

	if run.Status != openai.RunStatusCompleted {
		return "", fmt.Errorf("run not completed %v", run)
	}

	messages, err := o.client.ListMessage(context.Background(), o.ThreadID, nil, nil, nil, nil)
	if err != nil {
		return "", err
	}
	return messages.Messages[0].Content[0].Text.Value, nil

}
