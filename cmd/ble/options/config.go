package options

import (
	flag "github.com/spf13/pflag"
)

type Options struct {
	Name string
}

func (in *Options) Flags(fs *flag.FlagSet) {
	fs.StringVarP(&in.Name, "name", "n", in.Name, "Specify the name of the Bluetooth Heart Rate Sensor")
	return
}

func NewOptions() *Options {
	return &Options{
		Name: "Polar_H7",
	}
}
