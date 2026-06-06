package main

import (
	"context"
	"fmt"

	"github.com/bruceharrison1984/cloudflare-speed-test/config"
	"github.com/bruceharrison1984/cloudflare-speed-test/engines"
	"github.com/bruceharrison1984/cloudflare-speed-test/types"
)

type Cloudflare struct {
	engine                  engines.ISpeedTestEngine
	speedTestSummaryChannel chan *types.SpeedTestSummary
	exitChannel             chan struct{}
	errorChannel            chan error
}

func NewCloudflare() *Cloudflare {
	speedTestSummaryChannel := make(chan *types.SpeedTestSummary)
	exitChannel := make(chan struct{})
	errorChannel := make(chan error)

	return &Cloudflare{
		engine:                  engines.NewTestEngine(speedTestSummaryChannel, exitChannel, errorChannel),
		speedTestSummaryChannel: speedTestSummaryChannel,
		exitChannel:             exitChannel,
		errorChannel:            errorChannel,
	}
}

func (s *Cloudflare) runTest(ctx context.Context, cases []types.SpeedTestCase) (*result, error) {
	go s.engine.RunSpeedTest(ctx, cases)

	dlSpeed := float64(0)
	ulSpeed := float64(0)
	ping := float64(0)

	for {
		select {
		case summary, ok := <-s.speedTestSummaryChannel:
			{
				if ok {
					dlSpeed = summary.Bandwidth.DownloadSpeedMbps * 1000 * 1000
					ulSpeed = summary.Bandwidth.UploadSpeedMbps * 1000 * 1000
					ping = summary.Bandwidth.Ping * 1000
				}
			}
		case _, ok := <-s.exitChannel:
			{
				if !ok {
					return &result{
						latency:  ping,
						jitter:   nil,
						download: dlSpeed,
						upload:   ulSpeed,
					}, nil
				}
			}
		case err, ok := <-s.errorChannel:
			{
				if ok {
					return nil, err
				}
			}
		}
	}
}

func (s *Cloudflare) RunTest(ctx context.Context) (*result, error) {
	speedTest, err := s.runTest(ctx, config.GetDefaultConfig())
	if err != nil {
		return nil, fmt.Errorf("failed to run speedtest: %w", err)
	}

	ping, err := s.runTest(ctx, []types.SpeedTestCase{{PayloadSize: 1e5, Iterations: 10, TestType: types.Download}})
	if err != nil {
		return nil, fmt.Errorf("failed to run speedtest: %w", err)
	}

	result := result{
		upload:   speedTest.upload,
		download: speedTest.download,
		latency:  ping.latency,
	}
	return &result, nil
}
