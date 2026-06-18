package main

import (
	"fmt"
	"runtime"
	"runtime/metrics"
	"sort"
	"time"
)

// startGCLoad runs a background goroutine that continuously allocates and
// discards memory, including pointer-bearing objects, to force frequent garbage
// collections. It exists to demonstrate that GC pauses stay tiny and do not
// disturb the render loop. It is not part of the rendering work.
func startGCLoad() {
	go func() {
		var sink [][]byte
		type node struct {
			next *node
			data [16]float64
		}
		for {
			// Pointer-bearing garbage so the GC has to scan and collect.
			var head *node
			for i := 0; i < 2000; i++ {
				head = &node{next: head}
			}
			_ = head
			// Byte garbage.
			sink = append(sink, make([]byte, 64*1024))
			if len(sink) > 64 { // ~4 MB live, rest becomes garbage
				sink = sink[len(sink)-32:]
			}
			time.Sleep(200 * time.Microsecond)
		}
	}()
}

// gcStats is a snapshot of GC counters and the STW pause histogram.
type gcStats struct {
	numGC        uint64
	pauseTotalNs uint64
	buckets      []float64
	counts       []uint64
}

// gcReport summarizes GC activity over an interval.
type gcReport struct {
	numGC      uint64
	pauseCount uint64
	pauseTotal time.Duration
	maxPause   time.Duration
	meanPause  time.Duration
}

func readGCStats() gcStats {
	samples := []metrics.Sample{
		{Name: "/gc/cycles/total:gc-cycles"},
		{Name: "/gc/pauses:seconds"},
	}
	metrics.Read(samples)
	var s gcStats
	s.numGC = samples[0].Value.Uint64()
	h := samples[1].Value.Float64Histogram()
	s.buckets = append(s.buckets, h.Buckets...)
	s.counts = append(s.counts, h.Counts...)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	s.pauseTotalNs = m.PauseTotalNs
	return s
}

func (s gcStats) sub(start gcStats) gcReport {
	var r gcReport
	r.numGC = s.numGC - start.numGC
	r.pauseTotal = time.Duration(s.pauseTotalNs-start.pauseTotalNs) * time.Nanosecond
	var maxUpper float64
	for i := range s.counts {
		var prev uint64
		if i < len(start.counts) {
			prev = start.counts[i]
		}
		delta := s.counts[i] - prev
		if delta == 0 {
			continue
		}
		r.pauseCount += delta
		upper := s.buckets[i+1]
		if upper > maxUpper && upper < 1e9 { // skip +Inf bucket boundary
			maxUpper = upper
		}
	}
	r.maxPause = time.Duration(maxUpper * float64(time.Second))
	if r.pauseCount > 0 {
		r.meanPause = r.pauseTotal / time.Duration(r.pauseCount)
	}
	return r
}

func report(frameMs []float64, total time.Duration, gc gcReport) {
	n := len(frameMs)
	if n == 0 {
		fmt.Println("no frames rendered")
		return
	}
	sorted := append([]float64(nil), frameMs...)
	sort.Float64s(sorted)
	var sum float64
	for _, v := range frameMs {
		sum += v
	}
	avg := sum / float64(n)
	pct := func(p float64) float64 { return sorted[int(float64(n-1)*p)] }
	fps := float64(n) / total.Seconds()

	fmt.Println()
	fmt.Println("== flythrough report ==")
	fmt.Printf("frames        %d in %.1fs\n", n, total.Seconds())
	fmt.Printf("avg FPS       %.0f\n", fps)
	fmt.Printf("frame time    avg %.2f ms  min %.2f  p50 %.2f  p99 %.2f  max %.2f\n",
		avg, sorted[0], pct(0.50), pct(0.99), sorted[n-1])
	fmt.Println("-- garbage collector --")
	fmt.Printf("GC cycles     %d (%.1f/s)\n", gc.numGC, float64(gc.numGC)/total.Seconds())
	fmt.Printf("STW pauses    %d  total %.2f ms\n", gc.pauseCount, float64(gc.pauseTotal.Microseconds())/1000)
	fmt.Printf("pause/frame   mean %.1f us  max %.1f us\n",
		float64(gc.meanPause.Nanoseconds())/1000, float64(gc.maxPause.Nanoseconds())/1000)
}
