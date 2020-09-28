package ble

import (
	"github.com/JuulLabs-OSS/ble"
	"github.com/JuulLabs-OSS/ble/darwin"
)

func newPeripheral(name string, options ...ble.Option) (ble.Device, error) {
	return darwin.NewDevice(append(options, ble.OptPeripheralRole())...)
}
