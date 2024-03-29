package cmd

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"

	"github.com/347255699/comfyapi/pkg/comfyctl"

	"github.com/spf13/cobra"
)

func QueuePrompt() *cobra.Command {
	qpCmd := &cobra.Command{
		Use:   "qp",
		Short: "execute a workflow with prompt and output image",
	}
	flags := qpCmd.Flags()
	host := flags.StringP("connect", "c", "127.0.0.1:8188", "connect to compyui server")
	data := flags.StringP("data", "d", "", "data of workflow parameters")
	file := flags.StringP("file", "f", "", "workflow file path")
	seed := flags.IntP("seed", "s", 0, "seed for sampler")
	output := flags.StringP("output", "o", "output", "image output directory")
	input := flags.StringP("input", "i", "", "input image file")
	watch := flags.BoolP("watch", "w", false, "track the execution status of the workflow in comfyui")
	id := flags.StringP("id", "", "", "client id")
	plaintext := flags.BoolP("plaintext", "", false, "use plaintext connection")

	qpCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !verifyJson(*data) {
			return errors.New("data is not a valid json")
		}

		var values map[string]interface{}
		if err := json.Unmarshal([]byte(*data), &values); err != nil {
			return err
		}

		if *seed == 0 {
			*seed = rand.Intn(999999999999999)
			log.Printf("Seed: %d", *seed)
		}
		values["seed"] = *seed

		var cli *comfyctl.ComfyCtl
		if *plaintext {
			cli = comfyctl.NewWithPlainText(*host, *id)
		} else {
			cli = comfyctl.New(*host, *id)
		}
		log.Printf("Client Id: %s", cli.Id())
		if ret, err := cli.QueuePrompt(*file, *output, *input, values, *watch); err != nil {
			return err
		} else {
			if !*watch {
				log.Printf("PromptId: %s", ret[0])
			}
		}
		return nil
	}

	return qpCmd
}

func verifyJson(s string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(s), &js) == nil
}
