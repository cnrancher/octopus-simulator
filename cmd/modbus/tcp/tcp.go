package tcp

import (
	"github.com/spf13/cobra"

	"github.com/rancher/octopus-simulator/cmd/modbus/tcp/options"
	"github.com/rancher/octopus-simulator/pkg/log"
	"github.com/rancher/octopus-simulator/pkg/modbus"
	"github.com/rancher/octopus-simulator/pkg/util/log/logflag"
	"github.com/rancher/octopus-simulator/pkg/util/version/verflag"
)

const (
	name        = "tcp"
	description = `Modbus TCP protocol simulator`
)

func NewCommand() *cobra.Command {
	var opts = options.NewOptions()

	var c = &cobra.Command{
		Use:  name,
		Long: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested(name)
			logflag.SetLogger(log.SetLogger)

			return modbus.RunAsTCP(opts)
		},
	}

	opts.Flags(c.Flags())
	verflag.AddFlags(c.Flags())
	logflag.AddFlags(c.Flags())
	return c
}
