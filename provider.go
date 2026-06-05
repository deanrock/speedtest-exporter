package main

import "context"

type result struct {
	latency  float64
	jitter   float64
	download float64
	upload   float64
}

type provider interface {
	RunTest(ctx context.Context) (*result, error)
}
