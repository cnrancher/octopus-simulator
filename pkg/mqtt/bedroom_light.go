package mqtt

import (
	"context"
	"math/rand"
	"reflect"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/pkg/errors"

	"github.com/rancher/octopus-simulator/pkg/converter"
	"github.com/rancher/octopus-simulator/pkg/critical"
	"github.com/rancher/octopus-simulator/pkg/log"
)

func mockBedroomLight(address string, stop <-chan struct{}) (m mocker, err error) {
	var ctx, ctxCancel = context.WithCancel(critical.Context(stop))
	var cli = client.New()
	defer func() {
		if err != nil {
			_ = cli.Close()
		}
	}()

	var messages = make(chan interface{})
	cli.Callback = func(msg *packet.Message, err error) error {
		if err != nil {
			messages <- err
		} else {
			messages <- *msg
		}
		return nil
	}

	var instance = &bedroomLightJSON{
		Switch: false,
		Action: bedroomLightJSONAction{
			Gear: "low",
		},
		Parameter: bedroomLightJSONParameter{
			Power:     24.3,
			Luminance: 1800,
		},
		Production: bedroomLightJSONProduction{
			Manufacturer: "Rancher Octopus Fake Device",
			Date:         "2020-07-20T13:24:00.00Z",
			ServiceLife:  "P10Y0M0D",
		},
	}

	var in = &bedroomLight{
		cli:       cli,
		ctx:       ctx,
		ctxCancel: ctxCancel,
		messages:  messages,
		instance:  instance,
	}
	err = in.init(address)
	return in, err
}

type bedroomLightJSONAction struct {
	Gear string `json:"gear"`
}

type bedroomLightJSONParameter struct {
	Power     float64 `json:"power"`
	Luminance int     `json:"luminance"`
}

type bedroomLightJSONProduction struct {
	Manufacturer string `json:"manufacturer"`
	Date         string `json:"date"`
	ServiceLife  string `json:"serviceLife"`
}

type bedroomLightJSON struct {
	Switch     bool                       `json:"switch"`
	Action     bedroomLightJSONAction     `json:"action"`
	Parameter  bedroomLightJSONParameter  `json:"parameter"`
	Production bedroomLightJSONProduction `json:"production"`
}

type bedroomLight struct {
	sync.Mutex
	instance *bedroomLightJSON

	cli       *client.Client
	ctx       context.Context
	ctxCancel context.CancelFunc
	messages  chan interface{}
}

func (in *bedroomLight) init(address string) error {
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
		"cattle.io/octopus/home/bedroom/light": converter.TryMarshalJSON(in.instance),
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

	// subscribes
	go func() {
		var msg interface{}
		for {
			select {
			case <-in.ctx.Done():
				return
			case msg = <-in.messages:
			}

			in.Lock()
			func() {
				defer in.Unlock()

				switch v := msg.(type) {
				case packet.Message:
					var lightTopic = "cattle.io/octopus/home/bedroom/light/set"
					if v.Topic == lightTopic {
						var bedroomLightInstance = *in.instance
						if err := converter.UnmarshalJSON(v.Payload, &bedroomLightInstance); err != nil {
							log.Error(err, "failed to unmarshal received data")
							return
						}

						if reflect.DeepEqual(&bedroomLightInstance, in.instance) {
							return
						}
						in.instance = &bedroomLightInstance

						var switchTopic = "cattle.io/octopus/home/bedroom/light"
						var gf, err = in.cli.Publish(
							switchTopic,
							converter.TryMarshalJSON(bedroomLightInstance),
							packet.QOSAtLeastOnce,
							true,
						)
						if err != nil {
							log.Error(err, "failed to publish messages", "topic", switchTopic)
							return
						}
						if err := gf.Wait(10 * time.Second); err != nil {
							log.Error(err, "timeout to publish messages", "topic", switchTopic)
						}

					}
				case error:
					log.Error(v, "subscribed an error")
				}
			}()

			select {
			case <-in.ctx.Done():
			default:
			}
		}
	}()
	var initSubscribeTopics = map[string]packet.QOS{
		"cattle.io/octopus/home/bedroom/light/set": packet.QOSAtLeastOnce,
	}
	for topic, qos := range initSubscribeTopics {
		var gf, err = cli.Subscribe(topic, qos)
		if err != nil {
			return errors.Wrapf(err, "failed to subscribe topic %s", topic)
		}
		if err := gf.Wait(10 * time.Second); err != nil {
			return errors.Wrapf(err, "timeout to publish messages topic %s", topic)
		}
	}

	return nil
}

func (in *bedroomLight) Close() error {
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

func (in *bedroomLight) Mock(interval time.Duration) error {
	var ticker = time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-in.ctx.Done():
			return nil
		case <-ticker.C:
		}

		in.Lock()
		var instance = *in.instance
		in.Unlock()

		if instance.Switch {
			switch instance.Action.Gear {
			case "mid":
				instance.Parameter.Luminance = rand.Intn(100) + 1900
			case "high":
				instance.Parameter.Luminance = rand.Intn(150) + 2000
			default:
				instance.Parameter.Luminance = rand.Intn(50) + 1800
			}

			var lightTopic = "cattle.io/octopus/home/bedroom/light"
			var gf, err = in.cli.Publish(
				lightTopic,
				converter.TryMarshalJSON(instance),
				packet.QOSAtLeastOnce,
				true,
			)
			if err != nil {
				log.Error(err, "failed to publish messages", "topic", lightTopic)
				continue
			}
			if err := gf.Wait(10 * time.Second); err != nil {
				log.Error(err, "timeout to publish messages", "topic", lightTopic)
			}
		}

		select {
		case <-in.ctx.Done():
			return nil
		default:
		}
	}
}
