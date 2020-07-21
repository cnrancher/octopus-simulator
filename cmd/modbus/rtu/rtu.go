package rtu

import (
	"github.com/spf13/cobra"

	"github.com/rancher/octopus-simulator/cmd/modbus/rtu/options"
	"github.com/rancher/octopus-simulator/pkg/log"
	"github.com/rancher/octopus-simulator/pkg/modbus"
	"github.com/rancher/octopus-simulator/pkg/util/log/logflag"
	"github.com/rancher/octopus-simulator/pkg/util/version/verflag"
)

const (
	name        = "rtu"
	description = `Modbus RTU protocol simulator`
)

func NewCommand() *cobra.Command {
	var opts = options.NewOptions()

	var c = &cobra.Command{
		Use:  name,
		Long: description,
		RunE: func(cmd *cobra.Command, args []string) error {
			verflag.PrintAndExitIfRequested(name)
			logflag.SetLogger(log.SetLogger)

			return modbus.RunAsRTU(opts.Normalize())
		},
	}

	opts.Flags(c.Flags())
	verflag.AddFlags(c.Flags())
	logflag.AddFlags(c.Flags())
	return c
}
