# Octopus Simulator

[![Build Status](http://drone-pandaria.cnrancher.com/api/badges/cnrancher/octopus-simulator/status.svg)](http://drone-pandaria.cnrancher.com/cnrancher/octopus-simulator)

Octopus simulator consists of the simulators for testing [Octopus adaptor](https://github.com/cnrancher/octopus).

## Quick-start

To bring up all Octopus simulator, please apply the installer YAML file under [deploy/e2e](./deploy/e2e).
```shell script
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus-simulator/master/deploy/e2e/all_in_one.yaml
```

## Simulators

### Bluetooth Simulator

Bluetooth simulator is mocking a heart rate sensor, there are three [GATT Services](https://www.bluetooth.com/specifications/gatt/services/) in this simulated peripheral:

> The endianness of all property is LittleEndian.

- Device Information Service, UUID `00010000-0001-1000-8000-00805F9B34FB`.

    Characteristic | Type | Property | UUID | Value
    ---|---|---|---|---
    System ID | string | read | `2A23` | `4EBA552F`
    Model Number | string | read | `2A24` | `Rancher Octopus Bluetooth Device 0.0.1`
    Serial Number| string | read | `2A25` | `CB4040E1234567`
    Firmware Revision | string | read | `2A26` | `0.8.0`
    Hardware Revision | read | string | `2A27` | `0.5.7`
    Software Revision | read | string | `2A28` | `0.1.0`
    Manufacturer Name | read | string | `2A29` | `Rancher Octopus Fake Device`

- Battery Service, UUID `00020000-0001-1000-8000-00805F9B34FB`.

    Characteristic | Type | Property | UUID | Value
    ---|---|---|---|---
    Battery Level | int8 | read | `2A19` | `100` - declines per 2 minutes

- Heart Rate Service, UUID `00030000-0001-1000-8000-00805F9B34FB`.

    Characteristic | Type | Property |  UUID | Value
    ---|---|---|---|---
    Heart Rate Measurement | float64 | notify,indicate,read | `2A37` | `control_point ± 20`
    Body Sensor Location | int8 | read | `2A38` | `1` - chest
    Heart Rate Control Point | float64 | write | `2A39` | `80`
    Heart Rate New Alert | boolean | notify,indicate | `2A46` | `0` - green if rate in `control_point ± 20`, `1` - red

### Modbus Simulator

Modbus simulator is mocking a thermometer, the numerical accuracy is two decimal places, and the measurement is Kelvin absolute temperature and relative humidity.

> The endianness of all property is BigEndian.

Name | RegisterType | Type | Property | Address | Quantity  | Value
---|---|---|---|---|---|---
Switch | CoilRegister | boolean | read/write | 1 | 1 | Triggers to mock, the default is `true`.
Temperature | HoldingRegister | float32 | read | 0 | 2 | Represents the realtime absolute temperature, unit is in `kevin`, its range is between `273.15` and `378.15`.
Humidity | HoldingRegister | float32 | read | 2 | 2 | Represents the humidity, unit is in `%`, its range is between `10` and `100`.
High Temperature Threshold | HoldingRegister | int32 | read/write | 4 | 2 | Represents the threshold of absolute temperature, unit is in `kevin`, the default value is `324`.
High Temperature Alarm | CoilRegister | boolean | read | 0 | 1 | Indicates high temperature alarm. When the temperature exceeds the threshold, the high temperature alarm is `true`.
Battery | HoldingRegister | int8 | read | 6 | 1 | Represents the battery, uint is in `%`, the default value is `100`.
Manufacturer | HoldingRegister | string | read | 7 | 14 | Indicates the manufacturer.

### MQTT Simulator

MQTT simulator is mocking kitchen door, kitchen light, living room light and bedroom light.

> The endianness of all bytes property is BigEndian.

- Kitchen door Pub/Sub information

    Name | Topic | Type | Property | Value
    ---|---|---|---|---
    State | `cattle.io/octopus/home/status/kitchen/door/state`(subscribe)<br/> `cattle.io/octopus/home/status/kitchen/door_bytes/state`(subscribe) | string(text)<br/> string(bytes) | read | `open`
    Width | `cattle.io/octopus/home/status/kitchen/door/width`(subscribe)<br/> `cattle.io/octopus/home/status/kitchen/door_bytes/width`(subscribe) | float(text)<br/> float(bytes) | read | unit is meter, `1.2`
    Height | `cattle.io/octopus/home/status/kitchen/door/height`(subscribe)<br/> `cattle.io/octopus/home/status/kitchen/door_bytes/height`(subscribe) | float(text)<br/> float(bytes) | read | unit is meter, `1.8`
    Material | `cattle.io/octopus/home/status/kitchen/door/production_material`(subscribe)<br/> `cattle.io/octopus/home/status/kitchen/door_bytes/production_material`(subscribe) | string(text)<br/> string(bytes) | read | `wood`

- Kitchen light Pub/Sub information

    Name | Topic | Type | Property | Value
    ---|---|---|---|---
    Switch | `cattle.io/octopus/home/status/kitchen/light/switch`(subscribe)<br/> `cattle.io/octopus/home/set/kitchen/light/switch`(publish) | boolean(text) | read<br/> write | the default value is `false`
    Gear | `cattle.io/octopus/home/get/kitchen/light/gear`(subscribe)<br/> `cattle.io/octopus/home/control/kitchen/light/gear`(publish) | string(text) | read<br/> write | select from (low, mid, high), the default value is `low`
    Power | `cattle.io/octopus/home/status/kitchen/light/parameter_power`(subscribe) | float(text) | read | unit is watter, `3.0`
    Luminance | `cattle.io/octopus/home/status/kitchen/light/parameter_luminance`(subscribe) | int(text) | read | unit is luminance, the default value is `245`, changed by gear
    Manufacturer | `cattle.io/octopus/home/status/kitchen/light/manufacturer`(subscribe) | string(text) | read | `Rancher Octopus Fake Device`
    ProductionDate | `cattle.io/octopus/home/status/kitchen/light/production_date`(subscribe) | string(text) | read | `2020-07-08T13:24:00.00Z`
    ServiceLife | `cattle.io/octopus/home/status/kitchen/light/service_life`(subscribe) | string(text) | read | `P10Y0M0D`

- Living room light Pub/Sub information

    Name | Topic | Type | Property | Value
    ---|---|---|---|---
    Switch | `cattle.io/octopus/home/livingroom/light/switch`(subscribe)<br/> `cattle.io/octopus/home/livingroom/light/switch/set`(publish) | boolean(text) | read<br/> write | the default value is `false`
    Gear | `cattle.io/octopus/home/livingroom/light/gear`(subscribe)<br/> `cattle.io/octopus/home/livingroom/light/gear/set`(publish) | string(text) | read<br/> write | select from (low, mid, high), the default value is `low`
    Parameter | `cattle.io/octopus/home/livingroom/light/parameter`(subscribe) | string(text) | read | `[{"name":"power","value":"70.0w"},{"name":"luminance","value":"4900lm"}]`, the luminance value will change by gear
    Production | `cattle.io/octopus/home/livingroom/light/production`(subscribe) | string(text) | read | `{"manufacturer":"Rancher Octopus Fake Device","date":"2020-07-09T13:00:00.00Z","serviceLife":"P10Y0M0D"}`

- Bedroom light Pub/Sub information

    Total properties compress in a JSON-format message, please subscribe the `cattle.io/octopus/home/bedroom/light` topic to get the device information, or publish to `cattle.io/octopus/home/bedroom/light/set` topic to change the device.
    
    ```json
    {
      "switch": false,
      "action": {
        "gear": "low"
      },
      "parameter": {
        "power": 24.3,
        "luminance": 1800
      }, 
      "production": {
        "manufacturer": "Rancher Octopus Fake Device",
        "date": "2020-07-20T13:24:00.00Z",
        "serviceLife": "P10Y0M0D"
      }
    }
    ```
  
    Name | Path | Type | Property | Value
    ---|---|---|---|---
    Switch | `switch` | boolean(text) | read/write | the default value is `false`
    Gear | `action.gear` | string(text) | read/write | select from (low, mid, high), the default value is `low`
    Power | `parameter.power` | float(text) | read | unit is watter, `24.3`
    Luminance | `parameter.luminance` | int(text) | read | unit is luminance, the default value is `1800`
    Manufacturer | `production.manufacturer` | string(text) | read | `Rancher Octopus Fake Device`
    ProductionDate | `production.date` | string(text) | read | `2020-07-08T13:24:00.00Z`
    ServiceLife | `production.serviceLife` | string(text) | read | `P10Y0M0D`
  
    By default, the light is turn off, please publish `{"switch":true}` to turn on it, and publish `{"action":{"gear":"low"}}` to change the default gear if needed.

### OPC-UA Simulator

OPC-UA simulator is [open62541/open62541:1.0](https://hub.docker.com/r/open62541/open62541/tags).

Name | Type | Property | NodeID | Value
---|---|---|---|---
the answer| Int32 | read/write | `ns=1;s=the.answer` | `42`.
example bytestring | ByteString | read/write | `ns=1;s=myByteString` | `0x74657374313233`(string is `test123\0test123`)
Boolean - 0 | Boolean | read | `ns=1;i=43001` | random change
Int16 - 100 | Int16 | read | `ns=1;i=43101` | random change
Uint16 - 200 | Uint16 | read | `ns=1;i=43201` | random change
Int32 - 300 | Int32 | read | `ns=1;i=43301` | random change
Uint32 - 400 | Uint32 | read | `ns=1;i=43401` | random change
Int64 - 500 | Int64 | read | `ns=1;i=43501` | random change
Uint64 - 600 | Uint64 | read | `ns=1;i=43601` | random change
Float - 700 | Float | read | `ns=1;i=43701` | random change
Double - 800 | Double | read | `ns=1;i=43801` | random change
String - 900 | String | read | `ns=1;i=43901` | random change
ByteString - 1100 | ByteString | read | `ns=1;i=44101` | random change

## License
Copyright (c) 2020 [Rancher Labs, Inc.](http://rancher.com)

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at [LICENSE](./LICENSE) file for details.

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
