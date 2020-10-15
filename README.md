# Octopus Simulator

[![Build Status](http://drone-pandaria.cnrancher.com/api/badges/cnrancher/octopus-simulator/status.svg)](http://drone-pandaria.cnrancher.com/cnrancher/octopus-simulator)

Octopus simulator consists of the simulators for testing [Octopus adaptor](https://github.com/cnrancher/octopus).

## Quick-start

To bring up all Octopus simulator, please apply the installer YAML file under [deploy/e2e](./deploy/e2e).
```shell script
$ kubectl apply -f https://raw.githubusercontent.com/cnrancher/octopus-simulator/master/deploy/e2e/all_in_one.yaml
```

## Simulators

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

- Kitchen door Pub/Sub information

    ```yaml
    # -- sub
    cattle.io/octopus/home/status/kitchen/door/state -> open
    cattle.io/octopus/home/status/kitchen/door/width -> 1.2
    cattle.io/octopus/home/status/kitchen/door/height -> 1.8
    cattle.io/octopus/home/status/kitchen/door/production_material -> wood
    ```

- Kitchen light Pub/Sub information

    ```yaml
    # -- sub
    cattle.io/octopus/home/status/kitchen/light/switch -> false
    cattle.io/octopus/home/get/kitchen/light/gear -> low
    cattle.io/octopus/home/status/kitchen/light/parameter_power -> 3.0
    cattle.io/octopus/home/status/kitchen/light/parameter_luminance -> 245
    cattle.io/octopus/home/status/kitchen/light/manufacturer -> Rancher Octopus Fake Device
    cattle.io/octopus/home/status/kitchen/light/production_date -> 2020-07-08T13:24:00.00Z
    cattle.io/octopus/home/status/kitchen/light/service_life -> P10Y0M0D
    
    # -- pub
    # select from `true, false`
    cattle.io/octopus/home/set/kitchen/light/switch <- true
    # select from `[low, mid, high]`, change to `parameter_luminance`
    cattle.io/octopus/home/control/kitchen/light/gear <- low
    ```

- Living room light Pub/Sub information

    ```yaml
    # -- sub
    cattle.io/octopus/home/livingroom/light/switch -> false
    cattle.io/octopus/home/livingroom/light/gear -> low
    cattle.io/octopus/home/livingroom/light/parameter -> [{"name":"power","value":"70.0w"},{"name":"luminance","value":"4900lm"}]
    cattle.io/octopus/home/livingroom/light/production -> {"manufacturer":"Rancher Octopus Fake Device","date":"2020-07-09T13:00:00.00Z","serviceLife":"P10Y0M0D"}
    
    # -- pub
    # select from `true, false`
    cattle.io/octopus/home/livingroom/light/switch/set <- true
    # select from `[low, mid, high]`, change to `parameter[1].value`
    cattle.io/octopus/home/livingroom/light/gear/set <- low
    ```

- Bedroom light Pub/Sub information

    ```yaml
    # -- sub
    cattle.io/octopus/home/bedroom/light -> {"switch":false,"action":{"gear":"low"},"parameter":{"power":24.3,"luminance":1800},"production":{"manufacturer":"Rancher Octopus Fake Device","date":"2020-07-20T13:24:00.00Z","serviceLife":"P10Y0M0D"}}
    
    # -- pub
    # select from `true, false`
    cattle.io/octopus/home/bedroom/light/set <- {"switch":true}
    # select from `[low, mid, high]`, change to `parameter.luminance`
    cattle.io/octopus/home/bedroom/light/set <- {"action":{"gear":"low"}}
    ```

### OPC-UA Simulator

OPC-UA simulator is [open62541/open62541](https://hub.docker.com/r/open62541/open62541).

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
