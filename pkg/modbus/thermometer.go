package modbus

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
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

	// defaults temperature limitation is 324K
	_, _ = cli.WriteMultipleRegisters(4, 2, convertInt32ToByteArray(324))

	var ticker = time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-in.ctx.Done():
			return nil
		default:
		}

		// mocks absolute temperature, base unit is kevin, range is (273.15K, 378.15K]
		var holdingRegister0 = rand.Float32()*100 + 274.15
		_, err := cli.WriteMultipleRegisters(0, 2, convertFloat32ToByteArray(holdingRegister0))
		if err != nil {
			return errors.Wrapf(err, "failed to write holding register 0, %s:%v", "value", holdingRegister0)
		}
		log.Info(fmt.Sprintf("Mocked absolute temperature as %vK", holdingRegister0))

		// mocks relative humidity, unit is percent, range is [10%, 100%)
		var holdingRegister2 = rand.Float32()*90 + 10
		_, err = cli.WriteMultipleRegisters(2, 2, convertFloat32ToByteArray(holdingRegister2))
		if err != nil {
			return errors.Wrapf(err, "failed to write holding register 2, %s:%v", "value", holdingRegister2)
		}
		log.Info(fmt.Sprintf("Mocked relative humidity as %v%%", holdingRegister2))

		// gets temperature limitation
		holdingRegister4Bytes, err := cli.ReadHoldingRegisters(4, 2)
		if err != nil {
			return errors.Wrap(err, "failed to read holding registers 4")
		}
		var holdingRegister4 = parseByteArrayToInt32(holdingRegister4Bytes)
		log.Info(fmt.Sprintf("Mocked temperature limiation is %vK", holdingRegister4))

		// reports alarm
		var coilsRegister0 uint16 = 0x0000
		if holdingRegister0-float32(holdingRegister4) > 0.1 {
			log.Info("++ Reported high temperature alarm ++")
			coilsRegister0 = 0xFF00
		} else {
			log.Info("-- Removed high temperature alarm --")
		}
		_, err = cli.WriteSingleCoil(0, coilsRegister0)
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

func convertInt32ToByteArray(i int32) []byte {
	var ret = make([]byte, 4)
	binary.BigEndian.PutUint32(ret, uint32(i))
	return ret
}

func parseByteArrayToInt32(bs []byte) int32 {
	return int32(binary.BigEndian.Uint32(bs))
}

func convertFloat32ToByteArray(f float32) []byte {
	var ret = make([]byte, 4)
	binary.BigEndian.PutUint32(ret, math.Float32bits(f))
	return ret
}

func parseByteArrayToFloat32(bs []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(bs))
}
