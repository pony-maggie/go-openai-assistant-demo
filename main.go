package main

import (
	"fmt"
)

func main() {

	fmt.Println("assistant api demo")

	openaiClient := NewClient()

	_, err := openaiClient.CreateAssistant("D:\\code\\golang\\go-openai-assistant-demo\\最后一个问题 .txt")
	if err != nil {
		fmt.Printf("openai create assistant error %v", err)
		return
	}

	answer, err := openaiClient.Run("帮我总结这篇小说不超过100字")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(answer)

}
