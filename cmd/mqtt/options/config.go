package options

import (
	flag "github.com/spf13/pflag"
)

type Options struct {
	Interval int
}

func (in *Options) Flags(fs *flag.FlagSet) {
	fs.IntVarP(&in.Interval, "interval", "i", in.Interval, "Change cycle in seconds")
	return
}

func NewOptions() *Options {
	return &Options{
		Interval: 10,
	}
}
