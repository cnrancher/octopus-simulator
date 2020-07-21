package modbus

import (
	stdlog "log"
	"os"
	"time"

	"github.com/goburrow/modbus"
	"github.com/goburrow/serial"
	"github.com/pkg/errors"
	"github.com/tbrandon/mbserver"

	rtu "github.com/rancher/octopus-simulator/cmd/modbus/rtu/options"
	tcp "github.com/rancher/octopus-simulator/cmd/modbus/tcp/options"
	"github.com/rancher/octopus-simulator/pkg/log"
	"github.com/rancher/octopus-simulator/pkg/util/log/logflag"
	"github.com/rancher/octopus-simulator/pkg/util/signals"
)

func RunAsRTU(opts *rtu.Options) error {
	var s = mbserver.NewServer()

	var (
		cAddress = "/dev/ttyS001"
		sAddress = "/dev/ttyS002"
	)
	if err := s.ListenRTU(&serial.Config{
		Address:  sAddress,
		BaudRate: opts.BaudRate,
		DataBits: opts.DataBits,
		StopBits: opts.StopBits,
		Parity:   opts.Parity,
	}); err != nil {
		return errors.Wrap(err, "failed to start Modbus RTU server")
	}
	defer s.Close()
	log.Info("Listening on " + sAddress)

	var handler = modbus.NewRTUClientHandler(cAddress)
	if logflag.GetLogVerbosity() > 4 {
		handler.Logger = stdlog.New(os.Stdout, "modbus.client", stdlog.LstdFlags)
	}
	handler.BaudRate = opts.BaudRate
	handler.DataBits = opts.DataBits
	handler.StopBits = opts.StopBits
	handler.Parity = opts.Parity
	handler.SlaveId = opts.ID

	var t = mockThermometer(handler, signals.SetupSignalHandler())
	defer t.Close()

	return t.Mock(time.Duration(opts.Interval) * time.Second)
}

func RunAsTCP(opts *tcp.Options) error {
	var s = mbserver.NewServer()

	var sAddress = "0.0.0.0:5020"
	if err := s.ListenTCP(sAddress); err != nil {
		return errors.Wrap(err, "failed to start Modbus TCP server")
	}
	defer s.Close()
	log.Info("Listening on " + sAddress)

	var handler = modbus.NewTCPClientHandler(sAddress)
	if logflag.GetLogVerbosity() > 4 {
		handler.Logger = stdlog.New(os.Stdout, "modbus.client", stdlog.LstdFlags)
	}
	handler.SlaveId = opts.ID
	defer handler.Close()

	var t = mockThermometer(handler, signals.SetupSignalHandler())
	defer t.Close()

	return t.Mock(time.Duration(opts.Interval) * time.Second)
}
