package main

import (
	"context"
	"fmt"

	"github.com/showwin/speedtest-go/speedtest"
)

type SpeedtestNet struct {
	client *speedtest.Speedtest
}

func NewSpeedtestNet() *SpeedtestNet {
	return &SpeedtestNet{
		client: speedtest.New(),
	}
}

func (s *SpeedtestNet) RunTest(context.Context) (*result, error) {
	serverList, _ := s.client.FetchServers()
	targets, err := serverList.FindServer([]int{})
	if err != nil {
		return nil, fmt.Errorf("failed to find servers: %w", err)
	}

	if len(targets) <= 0 {
		return nil, fmt.Errorf("no target servers returned")
	}

	target := targets[0]

	if err := target.PingTest(nil); err != nil {
		return nil, fmt.Errorf("failed ping test: %w", err)
	}
	if err := target.DownloadTest(); err != nil {
		return nil, fmt.Errorf("failed download test: %w", err)
	}
	if err := target.UploadTest(); err != nil {
		return nil, fmt.Errorf("failed upload test: %w", err)
	}

	target.Context.Reset()

	return &result{
		latency:  float64(target.Latency.Milliseconds()),
		jitter:   new(float64(target.Latency.Milliseconds())),
		download: float64(target.DLSpeed) * 8,
		upload:   float64(target.ULSpeed) * 8,
	}, nil
}
