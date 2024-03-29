package main

import (
	"log"
	"math/rand"

	"github.com/347255699/comfyapi/pkg/comfyctl"
)

func main() {
	queuePrompt("workflows/prompt2image.json")
}

func queuePrompt(filePath string) {
	cli := comfyctl.NewWithPlainText("localhost:8188", "")
	log.Printf("Client Id: %s", cli.Id())
	values := map[string]interface{}{
		"positivePrompt": "beautiful scenery nature glass bottle landscape, , purple galaxy bottle,",
		"negativePrompt": "text, watermark",
		"seed":           rand.Intn(999999999999999),
	}

	// block way
	if ret, err := cli.QueuePrompt(filePath, "output", "", values, true); err != nil {
		log.Fatal(err)
		return
	} else {
		log.Printf("Ret: %v", ret)
	}
}
