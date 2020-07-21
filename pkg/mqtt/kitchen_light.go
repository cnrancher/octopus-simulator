package mqtt

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/256dpi/gomqtt/client"
	"github.com/256dpi/gomqtt/packet"
	"github.com/pkg/errors"

	"github.com/rancher/octopus-simulator/pkg/critical"
	"github.com/rancher/octopus-simulator/pkg/log"
)

func mockKitchenLight(address string, stop <-chan struct{}) (m mocker, err error) {
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

	var in = &kitchenLight{
		on:        "false",
		gear:      "low",
		cli:       cli,
		ctx:       ctx,
		ctxCancel: ctxCancel,
		messages:  messages,
	}
	err = in.init(address)
	return in, err
}

type kitchenLight struct {
	sync.Mutex
	on   string
	gear string

	cli       *client.Client
	ctx       context.Context
	ctxCancel context.CancelFunc
	messages  chan interface{}
}

func (in *kitchenLight) init(address string) error {
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
	var initPublishMessages = map[string]string{
		"cattle.io/octopus/home/status/kitchen/light/switch":              "false",
		"cattle.io/octopus/home/get/kitchen/light/gear":                   "low",
		"cattle.io/octopus/home/status/kitchen/light/parameter_power":     "3.0",
		"cattle.io/octopus/home/status/kitchen/light/parameter_luminance": "245",
		"cattle.io/octopus/home/status/kitchen/light/manufacturer":        "Rancher Octopus Fake Device",
		"cattle.io/octopus/home/status/kitchen/light/production_date":     "2020-07-08T13:24:00.00Z",
		"cattle.io/octopus/home/status/kitchen/light/service_life":        "P10Y0M0D",
	}
	for topic, message := range initPublishMessages {
		var gf, err = cli.Publish(
			topic,
			[]byte(message),
			packet.QOSAtLeastOnce,
			true,
		)
		if err != nil {
			return errors.Wrapf(err, "failed to publish messages for topic %s", topic)
		}
		if err := gf.Wait(10 * time.Second); err != nil {
			return errors.Wrapf(err, "timeout to publish messages for topic %s", topic)
		}
	}

	// subscribes
	go func() {
		for {
			select {
			case <-in.ctx.Done():
				return
			default:
			}
			var msg = <-in.messages

			in.Lock()
			func() {
				defer in.Unlock()

				switch v := msg.(type) {
				case *packet.Message:
					switch v.Topic {
					case "cattle.io/octopus/home/set/kitchen/light/switch":
						var on = string(v.Payload)
						if on != in.on {
							in.on = on
							var switchTopic = "cattle.io/octopus/home/status/kitchen/light/switch"
							var gf, err = in.cli.Publish(
								switchTopic,
								[]byte(on),
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
					case "cattle.io/octopus/home/control/kitchen/light/gear":
						var gear = string(v.Payload)
						if gear != in.gear {
							in.gear = gear
							var gearTopic = "cattle.io/octopus/home/get/kitchen/light/gear"
							var gf, err = in.cli.Publish(
								gearTopic,
								[]byte(gear),
								packet.QOSAtLeastOnce,
								true,
							)
							if err != nil {
								log.Error(err, "failed to publish messages", "topic", gearTopic)
								return
							}
							if err := gf.Wait(10 * time.Second); err != nil {
								log.Error(err, "timeout to publish messages", "topic", gearTopic)
							}
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
		"cattle.io/octopus/home/set/kitchen/light/switch":   packet.QOSAtLeastOnce,
		"cattle.io/octopus/home/control/kitchen/light/gear": packet.QOSAtLeastOnce,
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

func (in *kitchenLight) Close() error {
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

func (in *kitchenLight) Mock(interval time.Duration) error {
	var ticker = time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-in.ctx.Done():
			return nil
		case <-ticker.C:
		}

		in.Lock()
		func() {
			defer in.Unlock()
			if in.on != "true" {
				return
			}

			var luminance string
			switch in.gear {
			case "mid":
				luminance = fmt.Sprintf("%d", rand.Intn(100)+345)
			case "high":
				luminance = fmt.Sprintf("%d", rand.Intn(150)+445)
			default:
				luminance = fmt.Sprintf("%d", rand.Intn(50)+245)
			}
			var luminanceTopic = "cattle.io/octopus/home/status/kitchen/light/parameter_luminance"
			var gf, err = in.cli.Publish(
				luminanceTopic,
				[]byte(luminance),
				packet.QOSAtLeastOnce,
				true,
			)
			if err != nil {
				log.Error(err, "failed to publish messages", "topic", luminanceTopic)
				return
			}
			if err := gf.Wait(10 * time.Second); err != nil {
				log.Error(err, "timeout to publish messages", "topic", luminanceTopic)
				return
			}
		}()

		select {
		case <-in.ctx.Done():
			return nil
		default:
		}
	}
}
