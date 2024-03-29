# CFY

cfy is a command-line tool that lets you interact with comfyui servers. It's basically curl for comfyui servers.

Its main function is to template the workflow JSON files of comfyui, separating the parameters for easy modification.


## Installation

macOS:

```shell
make install
```

## Quick Start

Create your workflow JSON template. 
For more details, please refer to [workflows/](./workflows/) folder.

### prompt2image

```shell
cfy qp -c 127.0.0.1:8188 -f workflows/prompt2image.json -d '{"positivePrompt":"beautiful scenery nature glass bottle landscape, , purple galaxy bottle,", "negativePrompt":"text, watermark"}' -o output -w
```

### image2image

```shell
cfy qp -c 127.0.0.1:8188 -f workflows/image2image.json -d '{"negativePrompt":"nsfw","image":"1.png"}' -o output -i input/1.png -w
```

### hold watch

Open the terminal.

```shell
cfy watch -c 127.0.0.1:8188
```
```txt
2024/03/29 16:46:47 Client Id: 45a9251e-e265-4d3a-92b3-dbd702bd0290
```

Open anothers terminal.

```shell
cfy qp -f workflows/prompt2image.json -d '{"positivePrompt":"beautiful scenery nature glass bottle landscape, , purple galaxy bottle,", "negativePrompt":"text, watermark"}' -o output --id [same with watch client id]
```

### code

```golang
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
```