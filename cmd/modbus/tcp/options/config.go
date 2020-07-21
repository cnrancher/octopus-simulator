package options

import (
	flag "github.com/spf13/pflag"
)

type Options struct {
	ID       uint8
	Interval int
}

func (in *Options) Flags(fs *flag.FlagSet) {
	fs.Uint8VarP(&in.ID, "id", "", in.ID, "ID of the Modbus worker")
	fs.IntVarP(&in.Interval, "interval", "i", in.Interval, "Change cycle in seconds")
	return
}

func NewOptions() *Options {
	return &Options{
		ID:       1,
		Interval: 10,
	}
}
