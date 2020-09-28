package ble

import (
	"github.com/JuulLabs-OSS/ble"
	"github.com/JuulLabs-OSS/ble/linux"
)

func newPeripheral(name string, options ...ble.Option) (ble.Device, error) {
	return linux.NewDeviceWithName(name, append(options, ble.OptPeripheralRole())...)
}
