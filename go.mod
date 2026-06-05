module github.com/deanrock/speedtest-exporter

go 1.26.1

replace github.com/bruceharrison1984/cloudflare-speed-test v0.0.0-20230807164728-f97537fef9b5 => github.com/deanrock/cloudflare-speed-test v0.0.0-20260605205647-d99fa3994204

require (
	github.com/bruceharrison1984/cloudflare-speed-test v0.0.0-20230807164728-f97537fef9b5
	github.com/prometheus/client_golang v1.23.2
	github.com/showwin/speedtest-go v1.7.10
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.66.1 // indirect
	github.com/prometheus/procfs v0.16.1 // indirect
	go.yaml.in/yaml/v2 v2.4.2 // indirect
	golang.org/x/sys v0.35.0 // indirect
	google.golang.org/protobuf v1.36.8 // indirect
)
