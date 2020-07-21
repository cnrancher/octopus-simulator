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

func mockLivingRoomLight(address string, stop <-chan struct{}) (m mocker, err error) {
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

	var in = &livingRoomLight{
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

type livingRoomLight struct {
	sync.Mutex
	on   string
	gear string

	cli       *client.Client
	ctx       context.Context
	ctxCancel context.CancelFunc
	messages  chan interface{}
}

func (in *livingRoomLight) init(address string) error {
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
		"cattle.io/octopus/home/livingroom/light/switch":     "false",
		"cattle.io/octopus/home/livingroom/light/gear":       "low",
		"cattle.io/octopus/home/livingroom/light/parameter":  `[{"name":"power","value":"70.0w"},{"name":"luminance","value":"4900lm"}]`,
		"cattle.io/octopus/home/livingroom/light/production": `{"manufacturer":"Rancher Octopus Fake Device","date":"2020-07-09T13:00:00.00Z","serviceLife":"P10Y0M0D"}`,
	}
	for topic, message := range initPublishMessages {
		var gf, err = cli.Publish(
			topic,
			[]byte(message),
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
				case packet.Message:
					switch v.Topic {
					case "cattle.io/octopus/home/livingroom/light/switch/set":
						var on = string(v.Payload)
						if on != in.on {
							in.on = on
							var switchTopic = "cattle.io/octopus/home/livingroom/light/switch"
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
					case "cattle.io/octopus/home/livingroom/light/gear/set":
						var gear = string(v.Payload)
						if gear != in.gear {
							in.gear = gear
							var gearTopic = "cattle.io/octopus/home/livingroom/light/gear"
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
		"cattle.io/octopus/home/livingroom/light/switch/set": packet.QOSAtLeastOnce,
		"cattle.io/octopus/home/livingroom/light/gear/set":   packet.QOSAtLeastOnce,
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

func (in *livingRoomLight) Close() error {
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

func (in *livingRoomLight) Mock(interval time.Duration) error {
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
				luminance = fmt.Sprintf("%d", rand.Intn(100)+5000)
			case "high":
				luminance = fmt.Sprintf("%d", rand.Intn(150)+5100)
			default:
				luminance = fmt.Sprintf("%d", rand.Intn(50)+4900)
			}
			var parameterTopic = "cattle.io/octopus/home/livingroom/light/parameter"
			var gf, err = in.cli.Publish(
				parameterTopic,
				[]byte(fmt.Sprintf(`[{"name":"power","value":"70.0w"},{"name":"luminance","value":"%slm"}]`, luminance)),
				packet.QOSAtLeastOnce,
				true,
			)
			if err != nil {
				log.Error(err, "failed to publish messages", "topic", parameterTopic)
				return
			}
			if err := gf.Wait(10 * time.Second); err != nil {
				log.Error(err, "timeout to publish messages", "topic", parameterTopic)
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
