package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

const (
	baseURL        = "https://speed.cloudflare.com"
	downloadURL    = baseURL + "/__down?bytes=%d"
	uploadURL      = baseURL + "/__up"
	payloadSize    = 10_000_000 // 10MB
	nrTests        = 5
	nrLatencyTests = 20
)

var (
	cfL4RTTRe = regexp.MustCompile(`[?&]rtt=(\d+)`)
)

type Cloudflare struct {
	client *http.Client
}

func NewCloudflare() *Cloudflare {
	return &Cloudflare{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *Cloudflare) RunTest(ctx context.Context) (*result, error) {
	latencyAvg, jitter, err := testLatency(s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to test latency: %w", err)
	}

	downloadSpeed, err := testDownload(s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to test download: %w", err)
	}

	uploadSpeed, err := testUpload(s.client)
	if err != nil {
		return nil, fmt.Errorf("failed to test upload: %w", err)
	}

	result := result{
		upload:   uploadSpeed * 1000 * 1000,
		download: downloadSpeed * 1000 * 1000,
		latency:  latencyAvg,
		jitter:   &jitter,
	}
	return &result, nil
}

func testLatency(client *http.Client) (avg, jitter float64, err error) {
	var latencies []float64

	for range nrLatencyTests {
		start := time.Now()
		resp, err := client.Get(fmt.Sprintf(downloadURL, 0))
		if err != nil {
			return 0, 0, fmt.Errorf("request failed: %w", err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		totalMs := float64(time.Since(start).Milliseconds())

		if serverTiming := resp.Header.Get("Server-Timing"); serverTiming != "" {
			latency, err := parseLatencyFromServerTiming(serverTiming, totalMs)
			if err != nil {
				return 0, 0, fmt.Errorf("parsing latency from header failed: %w", err)
			}
			latencies = append(latencies, latency)
		}
		latencies = append(latencies, totalMs)
	}

	if len(latencies) == 0 {
		return 0, 0, fmt.Errorf("no latencies measured")
	}

	// Calculate average.
	var sum float64
	for _, l := range latencies {
		sum += l
	}
	avg = sum / float64(len(latencies))

	// Calculate jitter (standard deviation).
	var sumSquaredDiff float64
	for _, l := range latencies {
		diff := l - avg
		sumSquaredDiff += diff * diff
	}
	jitter = math.Sqrt(sumSquaredDiff / float64(len(latencies)))

	return avg, jitter, nil
}

func testDownload(client *http.Client) (float64, error) {
	var speeds []float64

	for i := 0; i < nrTests; i++ {
		start := time.Now()
		resp, err := client.Get(fmt.Sprintf(downloadURL, payloadSize))
		if err != nil {
			return 0, fmt.Errorf("request failed: %w", err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		elapsed := time.Since(start)

		speed := float64(payloadSize) * 8.0 / 1_000_000.0 / elapsed.Seconds()
		speeds = append(speeds, speed)
	}

	if len(speeds) == 0 {
		return 0, fmt.Errorf("no speeds measured")
	}

	return median(speeds), nil
}

func testUpload(client *http.Client) (float64, error) {
	var speeds []float64
	data := make([]byte, payloadSize)

	for i := 0; i < nrTests; i++ {
		start := time.Now()
		resp, err := client.Post(uploadURL, "application/octet-stream", bytes.NewReader(data))
		if err != nil {
			return 0, fmt.Errorf("request failed: %w", err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		elapsed := time.Since(start)

		speed := float64(payloadSize) * 8.0 / 1_000_000.0 / elapsed.Seconds()
		speeds = append(speeds, speed)
	}

	if len(speeds) == 0 {
		return 0, fmt.Errorf("no speeds measured")
	}

	return median(speeds), nil
}

func parseLatencyFromServerTiming(header string, totalMs float64) (float64, error) {
	if matches := cfL4RTTRe.FindStringSubmatch(header); len(matches) > 1 {
		rttUs, _ := strconv.ParseFloat(matches[1], 64)
		return rttUs / 1000.0, nil
	}

	return 0, fmt.Errorf("latency couldn't be parsed")
}

func median(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	v := make([]float64, len(values))
	copy(v, values)

	// Sort
	for i := 0; i < len(v); i++ {
		for j := i + 1; j < len(v); j++ {
			if v[i] > v[j] {
				v[i], v[j] = v[j], v[i]
			}
		}
	}

	middle := len(v) / 2
	if len(v)%2 == 1 {
		return v[middle]
	}
	return (v[middle-1] + v[middle]) / 2.0
}
