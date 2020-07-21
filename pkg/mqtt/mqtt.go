package mqtt

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/broker"
	"github.com/256dpi/gomqtt/packet"
	"github.com/256dpi/gomqtt/transport"
	"github.com/pkg/errors"

	"github.com/rancher/octopus-simulator/cmd/mqtt/options"
	"github.com/rancher/octopus-simulator/pkg/log"
	"github.com/rancher/octopus-simulator/pkg/util/log/logflag"
	"github.com/rancher/octopus-simulator/pkg/util/signals"
)

func Run(opts *options.Options) error {
	var sAddress = "tcp://0.0.0.0:1883"
	if podIP := os.Getenv("POD_IP"); podIP != "" {
		sAddress = fmt.Sprintf("tcp://%s:1883", podIP)
	}
	var brk, err = newMemoryBroker(sAddress)
	if err != nil {
		return errors.Wrap(err, "failed to start MQTT memory broker")
	}
	brk.Start()
	defer brk.Close()
	log.Info("Listening on " + sAddress)

	var mockers = make(mockers, 0, 4)
	defer mockers.Close()
	var stop = signals.SetupSignalHandler()

	kitchenDoorMocker, err := mockKitchenDoor(sAddress, stop)
	if err != nil {
		return errors.Wrap(err, "failed to mock kitchen door")
	}
	mockers = append(mockers, kitchenDoorMocker)

	bedroomLightMocker, err := mockBedroomLight(sAddress, stop)
	if err != nil {
		return errors.Wrap(err, "failed to mock bedroom light")
	}
	mockers = append(mockers, bedroomLightMocker)

	kitchenLightMocker, err := mockKitchenLight(sAddress, stop)
	if err != nil {
		return errors.Wrap(err, "failed to mock kitchen light")
	}
	mockers = append(mockers, kitchenLightMocker)

	livingRoomLightMocker, err := mockLivingRoomLight(sAddress, stop)
	if err != nil {
		return errors.Wrap(err, "failed to mock living room light")
	}
	mockers = append(mockers, livingRoomLightMocker)

	return mockers.Mock(time.Duration(opts.Interval) * time.Second)
}

type memoryBroker struct {
	server  transport.Server
	backend *broker.MemoryBackend
}

func (b *memoryBroker) Start() {
	var engine = broker.NewEngine(b.backend)
	engine.Accept(b.server)
}

func (b *memoryBroker) Close() {
	if b.backend != nil {
		b.backend.Close(5 * time.Second)
	}
	if b.server != nil {
		_ = b.server.Close()
	}
}

func newMemoryBroker(address string) (*memoryBroker, error) {
	var server, err = transport.Launch(address)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to launch broker")
	}

	var backend = broker.NewMemoryBackend()
	if logflag.GetLogVerbosity() > 4 {
		backend.Logger = func(e broker.LogEvent, c *broker.Client, pkt packet.Generic, msg *packet.Message, err error) {
			if err != nil {
				log.Error(err, fmt.Sprintf("[%s]", e))
			} else if msg != nil {
				log.Info(fmt.Sprintf("[%s] %s", e, msg.String()))
			} else if pkt != nil {
				log.Info(fmt.Sprintf("[%s] %s", e, pkt.String()))
			} else {
				log.Info(fmt.Sprintf("%s", e))
			}
		}
	}

	return &memoryBroker{
		server:  server,
		backend: backend,
	}, nil
}

type mocker interface {
	io.Closer
	Mock(interval time.Duration) error
}

type mockers []mocker

func (in mockers) Close() error {
	for _, mocker := range in {
		if mocker != nil {
			_ = mocker.Close()
		}
	}
	return nil
}

func (in mockers) Mock(interval time.Duration) error {
	var wg sync.WaitGroup
	defer wg.Wait()
	for _, mocker := range in {
		if mocker != nil {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = mocker.Mock(interval)
			}()
		}
	}
	return nil
}
