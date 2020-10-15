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

	// defaults allowing mocking is true
	_, err := cli.WriteSingleCoil(1, 0xFF00)
	if err != nil {
		log.Error(err, "Failed to configure switch to default value")
	}

	// defaults temperature limitation is 324K
	_, err = cli.WriteMultipleRegisters(4, 2, convertInt32ToBytes(324))
	if err != nil {
		log.Error(err, "Failed to configure temperature limitation to default value")
	}

	// defaults battery is 100
	_, err = cli.WriteMultipleRegisters(6, 1, convertInt8ToBytes(100))
	if err != nil {
		log.Error(err, "Failed to configure battery to default value")
	}
	var start = time.Now()

	// defaults manufacturer is Rancher Octopus Fake Factory
	var manufacturer = "Rancher Octopus Fake Factory"
	_, err = cli.WriteMultipleRegisters(7, (uint16(len(manufacturer))+1)/2, []byte(manufacturer))
	if err != nil {
		log.Error(err, "Failed to configure manufacturer to default value")
	}

	var ticker = time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-in.ctx.Done():
			return nil
		default:
		}

		// mocks or not
		var coilsRegister1, err = cli.ReadCoils(1, 1)
		if err != nil {
			return errors.Wrap(err, "failed to read coils registers 1")
		}
		if coilsRegister1[0]&0x01 != 0x01 {
			log.Info("Mocking is stopped")
			select {
			case <-in.ctx.Done():
				return nil
			case <-ticker.C:
			}
			continue
		} else {
			log.Info("Mocking is starting")
		}

		// mocks absolute temperature, base unit is kevin, range is (273.15K, 378.15K]
		var holdingRegister0 = rand.Float32()*100 + 274.15
		_, err = cli.WriteMultipleRegisters(0, 2, convertFloat32ToBytes(holdingRegister0))
		if err != nil {
			return errors.Wrapf(err, "failed to write holding register 0, %s:%v", "value", holdingRegister0)
		}
		log.Info(fmt.Sprintf("Mocked absolute temperature as %vK", holdingRegister0))

		// mocks relative humidity, unit is percent, range is [10%, 100%)
		var holdingRegister2 = rand.Float32()*90 + 10
		_, err = cli.WriteMultipleRegisters(2, 2, convertFloat32ToBytes(holdingRegister2))
		if err != nil {
			return errors.Wrapf(err, "failed to write holding register 2, %s:%v", "value", holdingRegister2)
		}
		log.Info(fmt.Sprintf("Mocked relative humidity as %v%%", holdingRegister2))

		// gets temperature limitation
		holdingRegister4Bytes, err := cli.ReadHoldingRegisters(4, 2)
		if err != nil {
			return errors.Wrap(err, "failed to read holding registers 4")
		}
		var holdingRegister4 = parseBytesToInt32(holdingRegister4Bytes)
		log.Info(fmt.Sprintf("Mocked temperature limiation is %vK", holdingRegister4))

		// mocks battery, unit is percent, range is [20%, 100%]
		var holdingRegister6 = 100 - int8(time.Since(start)/time.Hour)
		if holdingRegister6 < 20 {
			holdingRegister6 = 20
		}
		_, err = cli.WriteMultipleRegisters(6, 1, convertInt8ToBytes(holdingRegister6))
		if err != nil {
			return errors.Wrapf(err, "failed to write holding register 6, %s:%v", "value", holdingRegister6)
		}
		log.Info(fmt.Sprintf("Mocked battery as %v%%", holdingRegister6))

		// gets manufacturer
		holdingRegister7, err := cli.ReadHoldingRegisters(7, 14)
		if err != nil {
			return errors.Wrap(err, "failed to read holding registers 7")
		}
		log.Info(fmt.Sprintf("Mocked manufacturer is %v", string(holdingRegister7)))

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

func convertInt8ToBytes(i int8) []byte {
	var ret = make([]byte, 2)
	binary.BigEndian.PutUint16(ret, uint16(i))
	return ret
}

func convertInt32ToBytes(i int32) []byte {
	var ret = make([]byte, 4)
	binary.BigEndian.PutUint32(ret, uint32(i))
	return ret
}

func convertFloat32ToBytes(f float32) []byte {
	var ret = make([]byte, 4)
	binary.BigEndian.PutUint32(ret, math.Float32bits(f))
	return ret
}

func convertFloat64ToBytes(f float64) []byte {
	var ret = make([]byte, 8)
	binary.BigEndian.PutUint64(ret, math.Float64bits(f))
	return ret
}

func parseBytesToInt32(bs []byte) int32 {
	return int32(binary.BigEndian.Uint32(bs))
}

func parseBytesToFloat32(bs []byte) float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(bs))
}
