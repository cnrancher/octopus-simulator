package mqtt

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/pkg/errors"

	"github.com/rancher/octopus-simulator/pkg/critical"
)

func mockKitchenDoor(address string, stop <-chan struct{}) (m mocker, err error) {
	var ctx, ctxCancel = context.WithCancel(critical.Context(stop))
	var cli = client.New()
	defer func() {
		if err != nil {
			_ = cli.Close()
		}
	}()

	var in = &kitchenDoor{
		cli:       cli,
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}
	err = in.init(address)
	return in, err
}

type kitchenDoor struct {
	cli       *client.Client
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (in *kitchenDoor) init(address string) error {
	var cli = in.cli

	// connects
	var cf, err = cli.Connect(client.NewConfig(address))
	if err != nil {
		return errors.Wrap(err, "failed to connect broker")
	}
	if err := cf.Wait(10 * time.Second); err != nil {
		return errors.Wrap(err, "timeout to connect broker")
	}

	// publishes
	var initPublishMessages = map[string][]byte{
		// text
		"cattle.io/octopus/home/status/kitchen/door/state":               []byte("open"),
		"cattle.io/octopus/home/status/kitchen/door/width":               []byte("1.2"),
		"cattle.io/octopus/home/status/kitchen/door/height":              []byte("1.8"),
		"cattle.io/octopus/home/status/kitchen/door/production_material": []byte("wood"),

		// bytes
		"cattle.io/octopus/home/status/kitchen/door_bytes/state": func() (data []byte) {
			var s = []int32("open")
			var bs bytes.Buffer
			_ = binary.Write(&bs, binary.BigEndian, s)
			return bs.Bytes()
		}(),
		"cattle.io/octopus/home/status/kitchen/door_bytes/width": func() (data []byte) {
			data = make([]byte, 4)
			binary.BigEndian.PutUint32(data, math.Float32bits(float32(1.2)))
			return
		}(),
		"cattle.io/octopus/home/status/kitchen/door_bytes/height": func() (data []byte) {
			data = make([]byte, 4)
			binary.BigEndian.PutUint32(data, math.Float32bits(float32(1.8)))
			return
		}(),
		"cattle.io/octopus/home/status/kitchen/door_bytes/production_material": func() (data []byte) {
			var s = []int32("wood")
			var bs bytes.Buffer
			_ = binary.Write(&bs, binary.BigEndian, s)
			return bs.Bytes()
		}(),
	}
	for topic, message := range initPublishMessages {
		var gf, err = cli.Publish(
			topic,
			message,
			packet.QOSAtLeastOnce,
			true,
		)
		if err != nil {
			return errors.Wrapf(err, "failed to init messages for topic %s", topic)
		}
		if err := gf.Wait(10 * time.Second); err != nil {
			return errors.Wrapf(err, "timeout to init messages for topic %s", topic)
		}
	}
	return nil
}

func (in *kitchenDoor) Close() error {
	if in.cli != nil {
		if err := in.cli.Close(); err != nil {
			return err
		}
	}
	if in.ctxCancel != nil {
		in.ctxCancel()
	}
	return nil
}

func (in *kitchenDoor) Mock(interval time.Duration) error {
	<-in.ctx.Done()
	return nil
}
