package options

import (
	"strings"

	flag "github.com/spf13/pflag"
)

type Options struct {
	ID       uint8
	Parity   string
	BaudRate int
	DataBits int
	StopBits int
	Interval int
}

func (in *Options) Flags(fs *flag.FlagSet) {
	fs.Uint8VarP(&in.ID, "id", "", in.ID, "ID of the Modbus worker")
	fs.StringVarP(&in.Parity, "parity", "p", in.Parity, "Parity: N - None, E - Even, O - Odd (default E, the use of None parity requires 2 stop bits)")
	fs.IntVarP(&in.BaudRate, "baud-rate", "b", in.BaudRate, "RTU baudRate of serial port")
	fs.IntVarP(&in.DataBits, "data-bits", "d", in.DataBits, "Data bits: 5, 6, 7 or 8 (default 8)")
	fs.IntVarP(&in.StopBits, "stop-bits", "s", in.StopBits, "Stop bits: 1 or 2 (default 1)")
	fs.IntVarP(&in.Interval, "interval", "i", in.Interval, "Change cycle in seconds")
	return
}

func (in *Options) Normalize() *Options {
	var parity = strings.ToUpper(in.Parity)
	if parity == "N" || parity == "NONE" {
		in.StopBits = 2
	}
	return in
}

func NewOptions() *Options {
	return &Options{
		ID:       1,
		Parity:   "E",
		BaudRate: 19200,
		DataBits: 8,
		StopBits: 1,
		Interval: 10,
	}
}
