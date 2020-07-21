package modbus

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/goburrow/modbus"
	"github.com/pkg/errors"

	"github.com/rancher/octopus-simulator/pkg/critical"
	"github.com/rancher/octopus-simulator/pkg/log"
)

func mockThermometer(handler modbus.ClientHandler, stop <-chan struct{}) *thermometer {
	var ctx, ctxCancel = context.WithCancel(critical.Context(stop))

	return &thermometer{
		handler:   handler,
		ctx:       ctx,
		ctxCancel: ctxCancel,
	}
}

type thermometer struct {
	handler   modbus.ClientHandler
	ctx       context.Context
	ctxCancel context.CancelFunc
}

func (in *thermometer) Close() error {
	if in.handler != nil {
		if closer, ok := in.handler.(io.Closer); ok {
			return closer.Close()
		}
	}
	if in.ctxCancel != nil {
		in.ctxCancel()
	}
	return nil
}

func (in *thermometer) Mock(interval time.Duration) error {
	var cli = modbus.NewClient(in.handler)

	// defaults temperature limitation is 303.15k
	_, _ = cli.WriteMultipleRegisters(5, 2, parseInt64ToBytes(27315+3000, 2))

	var ticker = time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-in.ctx.Done():
			return nil
		default:
		}

		// mocks absolute temperature, base unit is kevin, at lease 278.15K
		var holdingRegister0 = uint64(rand.Intn(100000)) + 27315 + 500
		_, err := cli.WriteMultipleRegisters(0, 2, parseInt64ToBytes(holdingRegister0, 2))
		if err != nil {
			return errors.Wrapf(err, "failed to write holding register 0, %s:%v", "value", holdingRegister0)
		}
		log.Info(fmt.Sprintf("Mocked absolute temperature as %vK", float64(holdingRegister0)/100))

		// mocks relative humidity, unit is percent, at lease 10%
		var holdingRegister1 = uint64(rand.Intn(10000)) + 1000
		_, err = cli.WriteMultipleRegisters(1, 1, parseInt64ToBytes(holdingRegister1, 1))
		if err != nil {
			return errors.Wrapf(err, "failed to write holding register 1, %s:%v", "value", holdingRegister1)
		}
		log.Info(fmt.Sprintf("Mocked relative humidity as %v%%", float64(holdingRegister1)/100))

		// gets temperature limitation
		holdingRegister5Bytes, err := cli.ReadHoldingRegisters(5, 2)
		if err != nil {
			return errors.Wrap(err, "failed to read holding registers 5")
		}
		var holdingRegister5 = parseBytesToInt64(holdingRegister5Bytes)
		log.Info(fmt.Sprintf("Mocked temperature limiation is %vK", float64(holdingRegister5)/100))

		// reports alarm
		var coilsRegister0 = []byte{0}
		if holdingRegister5 < holdingRegister0 {
			log.Info("Reported high temperature alarm")
			coilsRegister0 = []byte{1}
		}
		_, err = cli.WriteMultipleCoils(0, 1, coilsRegister0)
		if err != nil {
			return errors.Wrapf(err, "failed to write coils register 0, %s:%v", "value", coilsRegister0)
		}

		select {
		case <-in.ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func parseInt64ToBytes(i uint64, quantity int) []byte {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], i)
	return buf[8-quantity*2:]
}

func parseBytesToInt64(bs []byte) uint64 {
	var l = len(bs)
	if l > 8 {
		bs = bs[l-8:]
	} else if l < 8 {
		var tmp = make([]byte, 8)
		copy(tmp[8-l:], bs)
		bs = tmp
	}
	return binary.BigEndian.Uint64(bs)
}
