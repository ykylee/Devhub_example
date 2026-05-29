package view

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

// fakeCounter — OnboardingPendingReviewCounter test double. Count 값 또는 error
// 둘 중 하나만 반환. atomic counter 로 호출 횟수 추적.
type fakeCounter struct {
	count    int
	err      error
	callsCnt int64
}

func (f *fakeCounter) CountPendingReview(ctx context.Context) (int, error) {
	atomic.AddInt64(&f.callsCnt, 1)
	if f.err != nil {
		return 0, f.err
	}
	return f.count, nil
}

func (f *fakeCounter) calls() int64 { return atomic.LoadInt64(&f.callsCnt) }

// TestRunOnboardingPendingReviewGauge_InitialTick — startup 직후 1차 tick 이
// 즉시 실행되는지 + Gauge update 검증.
func TestRunOnboardingPendingReviewGauge_InitialTick(t *testing.T) {
	initOnboardingMetrics()
	fake := &fakeCounter{count: 7}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = RunOnboardingPendingReviewGauge(ctx, fake, OnboardingPendingGaugeOptions{Interval: time.Hour})
		close(done)
	}()

	// 초기 1차 tick 이 즉시 실행돼야 함 — 짧은 대기 후 호출 횟수 확인.
	time.Sleep(50 * time.Millisecond)
	if fake.calls() < 1 {
		t.Fatalf("initial tick not invoked; calls=%d", fake.calls())
	}

	// Gauge 가 fake.count 로 업데이트 되었는지 확인.
	got := gaugeValue(t, onboardingPendingReviewCount)
	if got != 7 {
		t.Fatalf("gauge value = %v; want 7", got)
	}

	cancel()
	<-done
}

// TestRunOnboardingPendingReviewGauge_CtxCancel — ctx cancel 시 즉시 종료 검증.
func TestRunOnboardingPendingReviewGauge_CtxCancel(t *testing.T) {
	initOnboardingMetrics()
	fake := &fakeCounter{count: 0}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		err := RunOnboardingPendingReviewGauge(ctx, fake, OnboardingPendingGaugeOptions{Interval: time.Hour})
		if !errors.Is(err, context.Canceled) {
			t.Errorf("err = %v; want context.Canceled", err)
		}
		close(done)
	}()

	// 초기 1차 tick 후 cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// OK
	case <-time.After(time.Second):
		t.Fatal("worker did not exit within 1s after ctx cancel")
	}
}

// TestRunOnboardingPendingReviewGauge_ErrorRecovery — Counter 가 error 반환해도
// Gauge stale 보존 + worker 가 다음 tick 으로 진행.
func TestRunOnboardingPendingReviewGauge_ErrorRecovery(t *testing.T) {
	initOnboardingMetrics()
	fake := &fakeCounter{err: errors.New("simulated DB error")}

	// 기존 Gauge 값 record (이전 test 의 영향)
	before := gaugeValue(t, onboardingPendingReviewCount)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = RunOnboardingPendingReviewGauge(ctx, fake, OnboardingPendingGaugeOptions{Interval: time.Hour})
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)

	// error 발생 → Gauge 변경 안 됨 (직전 값 보존)
	after := gaugeValue(t, onboardingPendingReviewCount)
	if after != before {
		t.Fatalf("gauge changed on error: before=%v after=%v (want stale)", before, after)
	}
	if fake.calls() < 1 {
		t.Fatalf("counter not called: calls=%d", fake.calls())
	}

	cancel()
	<-done
}

// TestRunOnboardingPendingReviewGauge_DefaultInterval — Interval=0 시 default 60s
// 가 적용되는지 (interval 자체 검증은 불가하나, worker 가 panic 없이 시작 + tick
// 됐는지 확인).
func TestRunOnboardingPendingReviewGauge_DefaultInterval(t *testing.T) {
	initOnboardingMetrics()
	fake := &fakeCounter{count: 3}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = RunOnboardingPendingReviewGauge(ctx, fake, OnboardingPendingGaugeOptions{}) // Interval 0
		close(done)
	}()

	time.Sleep(50 * time.Millisecond)
	if fake.calls() < 1 {
		t.Fatalf("initial tick not invoked with default interval; calls=%d", fake.calls())
	}
	cancel()
	<-done
}

// gaugeValue — prometheus Gauge 의 현재 값 read.
func gaugeValue(t *testing.T, g prometheus.Gauge) float64 {
	t.Helper()
	pb := &dto.Metric{}
	if err := g.Write(pb); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if pb.Gauge == nil {
		return 0
	}
	return pb.Gauge.GetValue()
}
