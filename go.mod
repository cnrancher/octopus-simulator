module github.com/rancher/octopus-simulator

go 1.13

require (
	github.com/256dpi/gomqtt v0.14.2
	github.com/JuulLabs-OSS/ble v0.0.0-20200716215611-d4fcc9d598bb
	github.com/go-logr/logr v0.1.0
	github.com/goburrow/modbus v0.1.0
	github.com/goburrow/serial v0.1.0
	github.com/json-iterator/go v1.1.8
	github.com/pkg/errors v0.8.1
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/tbrandon/mbserver v0.0.0-20170611213546-993e1772cc62
	go.uber.org/atomic v1.4.0 // indirect
	go.uber.org/multierr v1.1.0 // indirect
	go.uber.org/zap v1.10.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
	k8s.io/apimachinery v0.18.5
	k8s.io/client-go v0.18.5
)
