package modbus

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rancher/octopus-simulator/cmd/modbus/rtu"
	"github.com/rancher/octopus-simulator/cmd/modbus/tcp"
	"github.com/rancher/octopus-simulator/pkg/util/version/verflag"
)

const (
	name = "modbus"
)

var allCommands = []*cobra.Command{
	tcp.NewCommand(),
	rtu.NewCommand(),
}

func NewCommand() *cobra.Command {
	var c = &cobra.Command{
		Use: name,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested(name)

			var (
				basename  = filepath.Base(os.Args[1])
				targetCmd *cobra.Command
			)
			for _, cmd := range allCommands {
				if cmd.Name() == basename {
					targetCmd = cmd
					break
				}
				for _, alias := range cmd.Aliases {
					if alias == basename {
						targetCmd = cmd
						break
					}
				}
			}
			if targetCmd != nil {
				return targetCmd.Execute()
			}
			return cmd.Help()
		},
	}
	c.AddCommand(allCommands...)
	verflag.AddFlags(c.Flags())
	return c
}
