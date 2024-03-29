package cmd

import (
	"log"

	"github.com/347255699/comfyapi/pkg/comfyctl"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

func Watch() *cobra.Command {

	watchCmd := &cobra.Command{
		Use:   "watch",
		Short: "Track the execution status of the workflow in comfyui",
	}

	flags := watchCmd.Flags()
	host := flags.StringP("connect", "c", "127.0.0.1:8188", "connect to compyui server")
	id := flags.StringP("id", "", "", "client id")
	plaintext := flags.BoolP("plaintext", "", false, "use plaintext connection")

	watchCmd.RunE = func(cmd *cobra.Command, args []string) error {
		var cli *comfyctl.ComfyCtl
		if *plaintext {
			cli = comfyctl.NewWithPlainText(*host, *id)
		} else {
			cli = comfyctl.New(*host, *id)
		}
		log.Printf("Client Id: %s", cli.Id())
		c, _, err := websocket.DefaultDialer.Dial(cli.MakeWsUrl(), nil)
		if err != nil {
			return err
		}

		return cli.Watch(c, "", func(cr *comfyctl.ComfyResult) {})
	}

	return watchCmd
}
