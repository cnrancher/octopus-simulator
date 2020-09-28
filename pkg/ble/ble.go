package ble

import (
	"os"

	"github.com/JuulLabs-OSS/ble"
	"github.com/pkg/errors"

	"github.com/rancher/octopus-simulator/cmd/ble/options"
	"github.com/rancher/octopus-simulator/pkg/log"
	"github.com/rancher/octopus-simulator/pkg/util/signals"
)

func Run(opts *options.Options) error {
	var peripheralName = opts.Name
	if name := os.Getenv("NAME"); name != "" {
		peripheralName = name
	}

	var s, err = mockHeartRateSensor(peripheralName, signals.SetupSignalHandler())
	if err != nil {
		return errors.Wrap(err, "failed to mock heart rate sensor")
	}
	defer s.Close()
	log.Info("Listening on " + peripheralName)

	return s.Mock()
}

type ChainCharacteristic ble.Characteristic

func (c *ChainCharacteristic) AddDescriptor(uuid ble.UUID, configFn func(descriptor *ble.Descriptor)) *ChainCharacteristic {
	var descriptor = (*ble.Characteristic)(c).NewDescriptor(uuid)
	if configFn != nil {
		configFn(descriptor)
	}
	return c
}

func (c *ChainCharacteristic) SetValue(b []byte) *ChainCharacteristic {
	(*ble.Characteristic)(c).SetValue(b)
	return c
}

func (c *ChainCharacteristic) HandleRead(h ble.ReadHandler) *ChainCharacteristic {
	(*ble.Characteristic)(c).HandleRead(h)
	return c
}

func (c *ChainCharacteristic) HandleWrite(h ble.WriteHandler) *ChainCharacteristic {
	(*ble.Characteristic)(c).HandleWrite(h)
	return c
}

func (c *ChainCharacteristic) HandleNotify(h ble.NotifyHandler) *ChainCharacteristic {
	(*ble.Characteristic)(c).HandleNotify(h)
	return c
}

func (c *ChainCharacteristic) HandleIndicate(h ble.NotifyHandler) *ChainCharacteristic {
	(*ble.Characteristic)(c).HandleIndicate(h)
	return c
}

func (c *ChainCharacteristic) HandleNotifyAndIndicate(h ble.NotifyHandler) *ChainCharacteristic {
	(*ble.Characteristic)(c).HandleNotify(h)
	(*ble.Characteristic)(c).HandleIndicate(h)
	return c
}

func (c *ChainCharacteristic) Origin() *ble.Characteristic {
	return (*ble.Characteristic)(c)
}
