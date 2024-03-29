package comfyctl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/347255699/comfyapi/pkg/types"

	"log"

	"github.com/gofrs/uuid"
	"github.com/gorilla/websocket"
)

type ComfyCtl struct {
	ClientId    string
	Host        string
	HttpCli     *http.Client
	isPlaintext bool
}

func NewWithPlainText(host, id string) *ComfyCtl {
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
	}
	return &ComfyCtl{
		ClientId:    id,
		Host:        host,
		HttpCli:     http.DefaultClient,
		isPlaintext: true,
	}
}

func New(host, id string) *ComfyCtl {
	if id == "" {
		id = uuid.Must(uuid.NewV4()).String()
	}
	return &ComfyCtl{
		ClientId:    id,
		Host:        host,
		HttpCli:     http.DefaultClient,
		isPlaintext: false,
	}
}

func (ctl *ComfyCtl) Id() string {
	return ctl.ClientId
}

type queuePromptParams struct {
	Prompt   map[string]interface{} `json:"prompt"`
	ClientId string                 `json:"client_id"`
}

func (ctl *ComfyCtl) render(tmplPath string, values map[string]interface{}) (map[string]interface{}, error) {
	tmpl := types.TemplateFile(tmplPath)

	var prompt map[string]interface{}
	if err := tmpl.ParseObject(values, func(data []byte) error {
		return json.Unmarshal(data, &prompt)
	}); err != nil {
		return nil, err
	}
	return prompt, nil
}

func (ctl *ComfyCtl) QueuePrompt(workflowPath, outputPath, imagePath string, values map[string]interface{}, watch bool) ([]string, error) {
	if imagePath != "" {
		if err := ctl.uploadImage(imagePath); err != nil {
			return nil, err
		}
	}

	prompt, err := ctl.render(workflowPath, values)
	if err != nil {
		return nil, err
	}

	var c *websocket.Conn
	if watch {
		addr := ctl.MakeWsUrl()
		if c, _, err = websocket.DefaultDialer.Dial(addr, nil); err != nil {
			return nil, err
		}
	}

	resp, err := ctl.queuePrompt(prompt)
	if err != nil {
		return nil, err
	}

	if !watch {
		return []string{resp.PromptID}, nil
	}

	var ret []string
	if err != ctl.Watch(c, resp.PromptID, func(cr *ComfyResult) {
		images := cr.Data.Output.Images
		for _, image := range images {
			ret = append(ret, image.Filename)
		}
		err = ctl.saveImages(images, outputPath)
	}) {
		return nil, err
	}
	log.Printf("Output: %v", ret)
	return ret, nil
}

func (ctl *ComfyCtl) saveImages(images []*ComfyImage, outputPath string) error {
	if err := makeSureOutputDir(outputPath); err != nil {
		return err
	}
	for _, image := range images {
		if err := ctl.saveImage(image, outputPath); err != nil {
			return err
		}
	}
	return nil
}

func (ctl *ComfyCtl) saveImage(image *ComfyImage, outputPath string) error {
	query := url.Values{}
	query.Add("filename", image.Filename)
	query.Add("subfolder", image.Subfolder)
	query.Add("type", image.Type)

	if resp, err := ctl.HttpCli.Get(ctl.makeHttpUrl(fmt.Sprintf("view?%s", query.Encode()))); err != nil {
		return err
	} else {
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return errors.New("failed to get image, code: " + resp.Status)
		}

		if b, err := io.ReadAll(resp.Body); err != nil {
			return err
		} else {
			outputImage := filepath.Join(outputPath, image.Filename)
			if err := os.WriteFile(outputImage, b, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}

func makeSureOutputDir(outputPath string) error {
	if _, err := os.Stat(outputPath); err != nil && os.IsNotExist(err) {
		if errDir := os.MkdirAll(outputPath, 0755); errDir != nil {
			return errDir
		}
	} else {
		if err != nil {
			return err
		}
	}
	return nil
}

func (ctl *ComfyCtl) queuePrompt(prompt map[string]interface{}) (*QueuePromptResp, error) {
	body := &queuePromptParams{
		Prompt:   prompt,
		ClientId: ctl.ClientId,
	}

	b, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	url := ctl.makeHttpUrl("prompt")
	resp, err := ctl.HttpCli.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to queue prompt, code: " + resp.Status)
	}

	if b, err = io.ReadAll(resp.Body); err != nil {
		return nil, err
	}

	var ret QueuePromptResp
	return &ret, json.Unmarshal(b, &ret)
}

type QueuePromptResp struct {
	PromptID   string                 `json:"prompt_id"`
	Number     int                    `json:"number"`
	NodeErrors map[string]interface{} `json:"node_errors"`
}

type ComfyMessage struct {
	Type string                 `json:"type"`
	Data map[string]interface{} `json:"data"`
}

const (
	TYPE_PROGRESS         = "progress"
	TYPE_EXECUTION_CACHED = "execution_cached"
	TYPE_EXECUTION_START  = "execution_start"
	TYPE_EXECUTING        = "executing"
	TYPE_STATUS           = "status"
	TYPE_EXECUTED         = "executed"
)

type ComfyResult struct {
	Type string              `json:"type"`
	Data *ComfySaveImageData `json:"data"`
}

type ComfySaveImageData struct {
	Node   string                    `json:"node"`
	Output *ComfySaveImageNodeOutput `json:"output"`
}

type ComfySaveImageNodeOutput struct {
	Images   []*ComfyImage `json:"images"`
	PromptId string        `json:"prompt_id"`
}

type ComfyImage struct {
	Filename  string `json:"filename"`
	Subfolder string `json:"subfolder"`
	Type      string `json:"type"`
}

// type ComfyExcutionCallback interface {
// 	Callback(*ComfyResult)
// }

type ComfyExcutionCallback func(*ComfyResult)

func (ctl *ComfyCtl) Watch(c *websocket.Conn, promptId string, cb ComfyExcutionCallback) error {
	defer c.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Fatalln("read:", err)
				return
			}
			if mt != websocket.TextMessage {
				continue
			}
			var cm ComfyMessage
			if err = json.Unmarshal(message, &cm); err != nil {
				log.Fatalln("invalid json:", err)
				return
			}

			switch cm.Type {
			case TYPE_EXECUTION_START:
				log.Printf("Execution Start: from prompt %s", cm.Data["prompt_id"].(string))
			case TYPE_EXECUTION_CACHED:
				log.Printf("Cached: %v from prompt %s", cm.Data["nodes"].([]interface{}), cm.Data["prompt_id"].(string))
			case TYPE_EXECUTING:
				if cm.Data["node"] != nil {
					log.Printf("Executing: Node %s from prompt %s", cm.Data["node"].(string), cm.Data["prompt_id"].(string))
				} else {
					if cm.Data["prompt_id"].(string) == promptId {
						return
					}
				}
			case TYPE_PROGRESS:
				log.Printf("Progress: Step %d of %d in node %s from prompt %s", int(cm.Data["value"].(float64)), int(cm.Data["max"].(float64)), cm.Data["node"].(string), cm.Data["prompt_id"].(string))
			case TYPE_EXECUTED:
				var cr ComfyResult
				if err = json.Unmarshal(message, &cr); err != nil {
					log.Fatalln("invalid json:", err)
					return
				}
				var images []string
				for _, i := range cr.Data.Output.Images {
					images = append(images, i.Filename)
				}
				log.Printf("Executed: Output %v from prompt %s", images, cm.Data["prompt_id"].(string))
				cb(&cr)
			default:
				continue
			}
		}
	}()

	for {
		select {
		case <-done:
			return closeWs(c)
		case <-interrupt:
			log.Println("Interrupt")
			return closeWs(c)
		}
	}
}

func (ctl *ComfyCtl) makeHttpUrl(paths ...string) string {
	strings.Join(paths, "/")
	scheme := "https"
	if ctl.isPlaintext {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s/%s", scheme, ctl.Host, strings.Join(paths, "/"))
}

func (ctl *ComfyCtl) MakeWsUrl() string {
	scheme := "wss"
	if ctl.isPlaintext {
		scheme = "ws"
	}
	return fmt.Sprintf("%s://%s/ws?clientId=%s", scheme, ctl.Host, ctl.ClientId)
}

func (ctl *ComfyCtl) uploadImage(filePath string) error {
	name := filepath.Base(filePath)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("type", "input")
	writer.WriteField("overwrite", "true")
	part, err := writer.CreateFormFile("image", name)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	writer.Close()

	req, err := http.NewRequest("POST", ctl.makeHttpUrl("upload", "image"), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := ctl.HttpCli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to upload image, code: " + resp.Status)
	}
	return nil
}

func closeWs(c *websocket.Conn) error {
	err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Fatal("write close:", err)
		return err
	}
	return nil
}
