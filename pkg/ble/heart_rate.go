package ble

import (
	"context"
	"encoding/binary"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/JuulLabs-OSS/ble"
	"github.com/pkg/errors"

	"github.com/rancher/octopus-simulator/pkg/critical"
	"github.com/rancher/octopus-simulator/pkg/log"
)

func mockHeartRateSensor(name string, stop <-chan struct{}) (*heartRateSensor, error) {
	var protocol, err = newPeripheral(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create heart rate sensor device")
	}

	var ctx, ctxCancel = context.WithCancel(critical.Context(stop))

	return &heartRateSensor{
		ctx:       ctx,
		ctxCancel: ctxCancel,
		name:      name,
		protocol:  protocol,
		device:    &heartRateDevice{},
	}, nil
}

type heartRateSensor struct {
	ctx       context.Context
	ctxCancel context.CancelFunc
	name      string
	protocol  ble.Device
	device    *heartRateDevice
}

func (in *heartRateSensor) Close() error {
	if in.protocol != nil {
		_ = in.protocol.Stop()
	}
	if in.ctxCancel != nil {
		in.ctxCancel()
	}
	return nil
}

func (in *heartRateSensor) Mock() error {
	var err = in.registerServices()
	if err != nil {
		return errors.Wrap(err, "failed to register services into heart rate sensor device")
	}

	in.device.start(in.ctx)

	err = in.protocol.AdvertiseNameAndServices(in.ctx, in.name)
	if err != nil && err != context.Canceled {
		return errors.Wrap(err, "failed to advertise services of heart rate sensor device")
	}
	return nil
}

func (in *heartRateSensor) registerServices() error {
	var logger = log.WithName("heart_rate_sensor")
	var services = make(map[string]*ble.Service)

	// device information service
	var deviceInformationService = ble.NewService(ble.MustParse("00010000-0001-1000-8000-00805F9B34FB"))
	services["device information"] = deviceInformationService
	// system id
	deviceInformationService.NewCharacteristic(ble.UUID16(0x2A23)).
		SetValue([]byte("4EBA552F"))
	// model number
	deviceInformationService.NewCharacteristic(ble.UUID16(0x2A24)).
		SetValue([]byte("Rancher Octopus Bluetooth Device 0.0.1"))
	// serial number
	deviceInformationService.NewCharacteristic(ble.UUID16(0x2A25)).
		SetValue([]byte("CB4040E1234567"))
	// firmware revision
	deviceInformationService.NewCharacteristic(ble.UUID16(0x2A26)).
		SetValue([]byte("0.8.0"))
	// hardware revision
	deviceInformationService.NewCharacteristic(ble.UUID16(0x2A27)).
		SetValue([]byte("0.5.7"))
	// software revision
	deviceInformationService.NewCharacteristic(ble.UUID16(0x2A28)).
		SetValue([]byte("0.1.0"))
	// manufacturer name
	deviceInformationService.NewCharacteristic(ble.UUID16(0x2A29)).
		SetValue([]byte("Rancher Octopus Fake Device"))

	// battery service
	var batteryServiceLogger = logger.WithValues("service", "battery")
	var batteryService = ble.NewService(ble.MustParse("00020000-0001-1000-8000-00805F9B34FB"))
	services["battery"] = batteryService
	// battery level
	batteryService.NewCharacteristic(ble.UUID16(0x2A19)).
		HandleRead(ble.ReadHandlerFunc(func(req ble.Request, resp ble.ResponseWriter) {
			var id = req.Conn().RemoteAddr().String()
			var logger = batteryServiceLogger.WithValues("id", id, "char", "battery level", "handle", "read")
			logger.V(1).Info("Start")
			defer func() {
				logger.V(1).Info("End")
			}()

			var _, err = resp.Write(in.device.getPower())
			if err != nil {
				logger.Error(err, "Failed to response the read handle")
				resp.SetStatus(ble.ErrInvalidHandle)
				return
			}
			resp.SetStatus(ble.ErrSuccess)
		}))

	// heart rate service
	var heartRateServiceLogger = logger.WithValues("service", "heart rate")
	var heartRateService = ble.NewService(ble.MustParse("00030000-0001-1000-8000-00805F9B34FB"))
	services["heart rate"] = heartRateService
	// heart rate measurement
	(*ChainCharacteristic)(heartRateService.NewCharacteristic(ble.UUID16(0x2A37))).
		HandleNotifyAndIndicate(ble.NotifyHandlerFunc(func(req ble.Request, n ble.Notifier) {
			var id = req.Conn().RemoteAddr().String()
			var logger = heartRateServiceLogger.WithValues("id", id, "char", "heart rate measurement", "handle", "notify")
			logger.V(1).Info("Start")
			defer func() {
				logger.V(1).Info("End")
			}()

			for {
				select {
				case <-n.Context().Done():
					return
				case <-time.After(time.Second):
					var _, err = n.Write(in.device.getHeartRate())
					if err != nil {
						log.Error(err, "Failed to response the notify handle")
						return
					}
				}
			}
		})).
		HandleRead(ble.ReadHandlerFunc(func(req ble.Request, resp ble.ResponseWriter) {
			var id = req.Conn().RemoteAddr().String()
			var logger = heartRateServiceLogger.WithValues("id", id, "char", "heart rate measurement", "handle", "read")
			logger.V(1).Info("Start")
			defer func() {
				logger.V(1).Info("End")
			}()

			var _, err = resp.Write(in.device.getHeartRate())
			if err != nil {
				logger.Error(err, "Failed to response the head handle")
				resp.SetStatus(ble.ErrInvalidHandle)
				return
			}
			resp.SetStatus(ble.ErrSuccess)
		}))
	// body sensor location
	heartRateService.NewCharacteristic(ble.UUID16(0x2A38)).
		HandleRead(ble.ReadHandlerFunc(func(req ble.Request, resp ble.ResponseWriter) {
			var id = req.Conn().RemoteAddr().String()
			var logger = heartRateServiceLogger.WithValues("id", id, "char", "body sensor location", "handle", "read")
			logger.V(1).Info("Start")
			defer func() {
				logger.V(1).Info("End")
			}()

			var _, err = resp.Write(in.device.getSensorLocation())
			if err != nil {
				logger.Error(err, "Failed to response the head handle")
				resp.SetStatus(ble.ErrInvalidHandle)
				return
			}
			resp.SetStatus(ble.ErrSuccess)
		}))
	// heart rate control point
	heartRateService.NewCharacteristic(ble.UUID16(0x2A39)).
		HandleWrite(ble.WriteHandlerFunc(func(req ble.Request, resp ble.ResponseWriter) {
			var id = req.Conn().RemoteAddr().String()
			var logger = heartRateServiceLogger.WithValues("id", id, "char", "heart rate control point", "handle", "write")
			logger.V(1).Info("Start")
			defer func() {
				logger.V(1).Info("End")
			}()

			var valBits = binary.LittleEndian.Uint64(req.Data())
			var valFloat64 = math.Float64frombits(valBits)
			in.device.setControlPoint(valFloat64)
			resp.SetStatus(ble.ErrSuccess)
		}))
	// heart rate new alert
	(*ChainCharacteristic)(heartRateService.NewCharacteristic(ble.UUID16(0x2A46))).
		HandleNotifyAndIndicate(ble.NotifyHandlerFunc(func(req ble.Request, n ble.Notifier) {
			var id = req.Conn().RemoteAddr().String()
			var logger = heartRateServiceLogger.WithValues("id", id, "char", "heart rate new alert", "handle", "notify")
			logger.V(1).Info("Start")
			defer func() {
				logger.V(1).Info("End")
			}()

			for {
				select {
				case <-n.Context().Done():
					return
				case <-time.After(time.Second):
					var _, err = n.Write(in.device.getAlert())
					if err != nil {
						log.Error(err, "Failed to response the notify handle")
						return
					}
				}
			}
		}))

	// registers services
	for name, service := range services {
		if err := in.protocol.AddService(service); err != nil {
			return errors.Wrapf(err, "failed to register %s service", name)
		}
		logger.Info("Registered", "service", name)
	}
	return nil
}

type heartRateDevice struct {
	sync.RWMutex

	power            int
	rate             float64
	rateControlPoint float64
	alert            byte
}

func (hrd *heartRateDevice) start(ctx context.Context) {
	var stopC = ctx.Done()

	// declines battery
	go func() {
		hrd.power = 100
		for {
			select {
			case <-stopC:
				return
			case <-time.After(2 * time.Minute):
				hrd.Lock()
				hrd.power--
				hrd.Unlock()
			}
		}
	}()

	// changes heart rate
	go func() {
		hrd.rate = 60
		hrd.rateControlPoint = 80
		for {
			select {
			case <-stopC:
				return
			case <-time.After(2 * time.Second):
				hrd.Lock()
				// NB(thxCode) boxing alert
				//    top +20: ------------
				//             >>>>>>>>>>>> -> alert
				//        +15: ============
				//             =-=-=-=-=-=-
				//  control 0: ============
				//             -=-=-=-=-=-=
				//        -15: ============
				//             >>>>>>>>>>>> -> alert
				// bottom -20: ------------
				hrd.rate = hrd.rateControlPoint + math.Cos(math.Pi*rand.Float64())*20
				if hrd.rateControlPoint-hrd.rate-15 > 1e-9 || hrd.rate-hrd.rateControlPoint-15 > 1e-9 {
					hrd.alert = 0x01
				} else {
					hrd.alert = 0x00
				}
				hrd.Unlock()
			}
		}
	}()

}

func (hrd *heartRateDevice) getPower() []byte {
	hrd.RLock()
	defer hrd.RUnlock()
	return []byte{byte(hrd.power)}
}

func (hrd *heartRateDevice) getHeartRate() []byte {
	hrd.RLock()
	defer hrd.RUnlock()
	var data = make([]byte, 8)
	binary.LittleEndian.PutUint64(data, math.Float64bits(hrd.rate))
	return data
}

func (hrd *heartRateDevice) getAlert() []byte {
	hrd.RLock()
	defer hrd.RUnlock()
	return []byte{hrd.alert}
}

func (hrd *heartRateDevice) setControlPoint(cp float64) {
	hrd.Lock()
	defer hrd.Unlock()
	hrd.rateControlPoint = cp
}

func (hrd *heartRateDevice) getSensorLocation() []byte {
	// 0 Other
	// 1 Chest
	// 2 Wrist
	// 3 Finger
	// 4 Hand
	// 5 Ear Lobe
	// 6 Foot
	// 7 ~ 255 Reserved for future use
	return []byte{0x01}
}
