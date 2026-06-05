package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var addr = flag.String("listen-address", ":9100", "The address to listen on for HTTP requests.")
var interval = flag.String("interval", "1h", "Interval at which to run tests.")

type metrics struct {
	up                      *prometheus.GaugeVec
	uploadBitsPerSecond     *prometheus.GaugeVec
	downloadBitsPerSecond   *prometheus.GaugeVec
	pingLatencyMilliseconds *prometheus.GaugeVec
	jitterMilliseconds      *prometheus.GaugeVec
}

func main() {
	flag.Parse()

	reg := prometheus.NewRegistry()

	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	parsedInterval, err := time.ParseDuration(*interval)
	if err != nil {
		panic(fmt.Errorf("invalid interval '%s': %w", *interval, err))
	}

	labels := []string{"name"}
	m := metrics{
		up: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "speedtest_up",
			Help: "Speedtest status whether the scrape worked",
		}, labels),
		uploadBitsPerSecond: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "speedtest_upload_bits_per_second",
			Help: "Speedtest current upload speed in bits/s",
		}, labels),
		downloadBitsPerSecond: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "speedtest_download_bits_per_second",
			Help: "Speedtest current download speed in bit/s",
		}, labels),
		pingLatencyMilliseconds: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "speedtest_ping_latency_milliseconds",
			Help: "Speedtest current ping in ms",
		}, labels),
		jitterMilliseconds: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "speedtest_jitter_milliseconds",
			Help: "Speedtest current jitter in ms",
		}, labels),
	}
	reg.MustRegister(m.up, m.uploadBitsPerSecond, m.downloadBitsPerSecond, m.pingLatencyMilliseconds, m.jitterMilliseconds)

	providers := map[string]provider{
		//	"speedtest_net": NewSpeedtestNet(),
		"cloudflare": NewCloudflare(),
	}

	go func() {
		for {
			for name, provider := range providers {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

				result, err := provider.RunTest(ctx)
				if err != nil {
					slog.Error("running test failed", slog.String("provider", name), slog.String("err", err.Error()))
					m.up.WithLabelValues(name).Set(0)
				} else {
					slog.Info("running test succeeded", slog.String("provider", name))
					m.up.WithLabelValues(name).Set(1)
					m.downloadBitsPerSecond.WithLabelValues(name).Set(result.download)
					m.uploadBitsPerSecond.WithLabelValues(name).Set(result.upload)
					m.pingLatencyMilliseconds.WithLabelValues(name).Set(result.latency)
					m.jitterMilliseconds.WithLabelValues(name).Set(result.jitter)
				}

				cancel()
			}

			time.Sleep(parsedInterval)
		}
	}()

	slog.Info("Speedtest Exporter started", slog.String("address", *addr), slog.String("interval", parsedInterval.String()))

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))
	log.Fatal(http.ListenAndServe(*addr, nil))
}
