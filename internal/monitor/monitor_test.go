package monitor

import (
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/poly-oracle/internal/models"
	"github.com/poly-oracle/internal/storage"
)

// ─── Existing DetectChanges tests (T016: updated to remove threshold arg) ────

func TestDetectChanges(t *testing.T) {
	s := storage.New(100, 50, "/tmp/test-monitor.json", 0644, 0755)
	m := New(s)

	now := time.Now()
	event := models.Event{
		ID:             "event-1",
		EventID:        "event-1",
		Title:          "Will X happen?",
		Category:       "politics",
		YesProbability: 0.75,
		NoProbability:  0.25,
		Active:         true,
		LastUpdated:    now,
		CreatedAt:      now.Add(-1 * time.Hour),
	}
	if err := s.AddEvent(&event); err != nil {
		t.Fatalf("Failed to add event: %v", err)
	}

	snapshots := []models.Snapshot{
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.60,
			NoProbability:  0.40,
			Timestamp:      now.Add(-1 * time.Hour),
			Source:         "test",
		},
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.75,
			NoProbability:  0.25,
			Timestamp:      now,
			Source:         "test",
		},
	}
	for _, snap := range snapshots {
		if err := s.AddSnapshot(&snap); err != nil {
			t.Fatalf("Failed to add snapshot: %v", err)
		}
	}

	events := []models.Event{event}
	changes, _, err := m.DetectChanges(events, 2*time.Hour)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if len(changes) == 0 {
		t.Error("Expected at least 1 change, got 0")
		return
	}
	if changes[0].Magnitude < 0.149 || changes[0].Magnitude > 0.151 {
		t.Errorf("Expected magnitude 0.15, got %f", changes[0].Magnitude)
	}
	if changes[0].Direction != "increase" {
		t.Errorf("Expected direction 'increase', got '%s'", changes[0].Direction)
	}
}

func TestDetectChanges_BelowThreshold(t *testing.T) {
	s := storage.New(100, 50, "/tmp/test-threshold.json", 0644, 0755)
	m := New(s)

	now := time.Now()
	event := models.Event{
		ID:             "event-1",
		EventID:        "event-1",
		Title:          "Test?",
		Category:       "politics",
		YesProbability: 0.6001,
		NoProbability:  0.3999,
		Active:         true,
		LastUpdated:    now,
		CreatedAt:      now.Add(-1 * time.Hour),
	}
	if err := s.AddEvent(&event); err != nil {
		t.Fatalf("Failed to add event: %v", err)
	}

	// Very tiny change — below 0.001 floor
	snapshots := []models.Snapshot{
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.6000,
			NoProbability:  0.4000,
			Timestamp:      now.Add(-1 * time.Hour),
			Source:         "test",
		},
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.6001,
			NoProbability:  0.3999,
			Timestamp:      now,
			Source:         "test",
		},
	}
	for _, snap := range snapshots {
		if err := s.AddSnapshot(&snap); err != nil {
			t.Fatalf("Failed to add snapshot: %v", err)
		}
	}

	events := []models.Event{event}
	changes, _, err := m.DetectChanges(events, 2*time.Hour)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("Expected 0 changes (below 0.001 floor), got %d", len(changes))
	}
}

func TestDetectChanges_OutOfOrderSnapshots(t *testing.T) {
	s := storage.New(100, 50, "/tmp/test-out-of-order.json", 0644, 0755)
	m := New(s)

	now := time.Now()
	event := models.Event{
		ID:             "event-1",
		EventID:        "event-1",
		Title:          "Test?",
		Category:       "politics",
		YesProbability: 0.85,
		NoProbability:  0.15,
		Active:         true,
		LastUpdated:    now,
		CreatedAt:      now.Add(-2 * time.Hour),
	}
	if err := s.AddEvent(&event); err != nil {
		t.Fatalf("Failed to add event: %v", err)
	}

	// Add snapshots OUT OF ORDER to test sorting
	snapshots := []models.Snapshot{
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.70,
			NoProbability:  0.30,
			Timestamp:      now.Add(-1 * time.Hour),
			Source:         "test",
		},
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.50,
			NoProbability:  0.50,
			Timestamp:      now.Add(-2 * time.Hour),
			Source:         "test",
		},
		{
			ID:             uuid.New().String(),
			EventID:        "event-1",
			YesProbability: 0.85,
			NoProbability:  0.15,
			Timestamp:      now,
			Source:         "test",
		},
	}
	for _, snap := range snapshots {
		if err := s.AddSnapshot(&snap); err != nil {
			t.Fatalf("Failed to add snapshot: %v", err)
		}
	}

	events := []models.Event{event}
	changes, _, err := m.DetectChanges(events, 3*time.Hour)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}
	if len(changes) == 0 {
		t.Fatal("Expected at least 1 change, got 0")
	}

	expectedMagnitude := 0.35
	if changes[0].Magnitude < expectedMagnitude-0.01 || changes[0].Magnitude > expectedMagnitude+0.01 {
		t.Errorf("Expected magnitude %.2f (0.85 - 0.50), got %.2f", expectedMagnitude, changes[0].Magnitude)
	}
	if changes[0].Direction != "increase" {
		t.Errorf("Expected direction 'increase', got '%s'", changes[0].Direction)
	}
	if changes[0].OldProbability != 0.50 {
		t.Errorf("Expected old probability 0.50, got %.2f", changes[0].OldProbability)
	}
	if changes[0].NewProbability != 0.85 {
		t.Errorf("Expected new probability 0.85, got %.2f", changes[0].NewProbability)
	}
}

// ─── T011: TestKLDivergence ───────────────────────────────────────────────────

func TestKLDivergence(t *testing.T) {
	tests := []struct {
		name       string
		pOld, pNew float64
		wantMin    float64
		wantMax    float64
		wantPos    bool // result must be > 0
	}{
		{
			name: "5% move at p=0.5 — small positive",
			pOld: 0.50, pNew: 0.55,
			wantMin: 0.001, wantMax: 0.01,
		},
		{
			name: "10% move at p=0.5 — medium positive",
			pOld: 0.50, pNew: 0.60,
			wantMin: 0.005, wantMax: 0.025,
		},
		{
			name: "KL(0.6||0.5) must be positive",
			pOld: 0.50, pNew: 0.60,
			wantPos: true,
		},
		{
			name: "no change — near zero",
			pOld: 0.70, pNew: 0.70,
			wantMin: 0.0, wantMax: 1e-10,
		},
		{
			name: "boundary p=0.0 does not panic or NaN",
			pOld: 0.0, pNew: 0.05,
		},
		{
			name: "boundary p=1.0 does not panic or NaN",
			pOld: 1.0, pNew: 0.95,
		},
		{
			name: "always non-negative",
			pOld: 0.3, pNew: 0.7,
			wantMin: 0.0, wantMax: math.MaxFloat64,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KLDivergence(tt.pOld, tt.pNew)
			if math.IsNaN(got) {
				t.Errorf("KLDivergence(%v, %v) = NaN", tt.pOld, tt.pNew)
				return
			}
			if math.IsInf(got, 0) {
				t.Errorf("KLDivergence(%v, %v) = Inf", tt.pOld, tt.pNew)
				return
			}
			if got < 0 {
				t.Errorf("KLDivergence(%v, %v) = %v, want >= 0", tt.pOld, tt.pNew, got)
			}
			if tt.wantPos && got <= 0 {
				t.Errorf("KLDivergence(%v, %v) = %v, want > 0", tt.pOld, tt.pNew, got)
			}
			if tt.wantMax > 0 && (got < tt.wantMin || got > tt.wantMax) {
				t.Errorf("KLDivergence(%v, %v) = %v, want [%v, %v]", tt.pOld, tt.pNew, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// ─── T012: TestLogVolumeWeight ────────────────────────────────────────────────

func TestLogVolumeWeight(t *testing.T) {
	const vRef = 25000.0
	tests := []struct {
		name             string
		volume24h, vRef  float64
		wantMin, wantMax float64
	}{
		{
			name:      "volume == vRef → 1.0",
			volume24h: vRef, vRef: vRef,
			wantMin: 0.99, wantMax: 1.01,
		},
		{
			name:      "volume = 0 → floor 0.1",
			volume24h: 0, vRef: vRef,
			wantMin: 0.1, wantMax: 0.1,
		},
		{
			name:      "volume = 4×vRef → ~2.32",
			volume24h: 4 * vRef, vRef: vRef,
			wantMin: 2.20, wantMax: 2.40,
		},
		{
			name:      "vRef = 0 treated as 1.0",
			volume24h: 100, vRef: 0,
			wantMin: 0.1, wantMax: 10.0,
		},
		{
			name:      "very large volume → capped by log growth",
			volume24h: 1_000_000, vRef: vRef,
			wantMin: 5.0, wantMax: 6.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LogVolumeWeight(tt.volume24h, tt.vRef)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Errorf("LogVolumeWeight(%v, %v) = %v (invalid)", tt.volume24h, tt.vRef, got)
				return
			}
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("LogVolumeWeight(%v, %v) = %v, want [%v, %v]",
					tt.volume24h, tt.vRef, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// ─── T013: TestHistoricalSNR ──────────────────────────────────────────────────

func makeSnaps(probs []float64) []models.Snapshot {
	snaps := make([]models.Snapshot, len(probs))
	for i, p := range probs {
		snaps[i] = models.Snapshot{
			ID:             uuid.New().String(),
			YesProbability: p,
			Timestamp:      time.Now().Add(time.Duration(i) * time.Hour),
		}
	}
	return snaps
}

func TestHistoricalSNR(t *testing.T) {
	tests := []struct {
		name      string
		probs     []float64
		netChange float64
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "0 snapshots → 1.0",
			probs:     nil,
			netChange: 0.10,
			wantMin:   1.0, wantMax: 1.0,
		},
		{
			name:      "1 snapshot → 1.0",
			probs:     []float64{0.5},
			netChange: 0.10,
			wantMin:   1.0, wantMax: 1.0,
		},
		{
			name:      "stable snapshots (σ ≈ 0) → 1.0",
			probs:     []float64{0.5, 0.50001, 0.50002, 0.50001, 0.5},
			netChange: 0.10,
			wantMin:   1.0, wantMax: 1.0,
		},
		{
			name:      "large move on quiet market → clamp 5.0",
			probs:     []float64{0.50, 0.501, 0.502, 0.501, 0.500},
			netChange: 0.30, // huge net change relative to tiny historical σ
			wantMin:   5.0, wantMax: 5.0,
		},
		{
			name:      "tiny move on volatile market → clamp 0.5",
			probs:     []float64{0.50, 0.60, 0.40, 0.65, 0.35},
			netChange: 0.005, // tiny net change relative to large historical σ
			wantMin:   0.5, wantMax: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HistoricalSNR(makeSnaps(tt.probs), tt.netChange)
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Errorf("HistoricalSNR = %v (invalid)", got)
				return
			}
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("HistoricalSNR = %v, want [%v, %v]", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// ─── T014: TestTrajectoryConsistency ─────────────────────────────────────────

func TestTrajectoryConsistency(t *testing.T) {
	tests := []struct {
		name    string
		probs   []float64
		wantMin float64
		wantMax float64
	}{
		{
			name:    "0 snapshots → 1.0",
			probs:   nil,
			wantMin: 1.0, wantMax: 1.0,
		},
		{
			name:    "1 snapshot → 1.0",
			probs:   []float64{0.5},
			wantMin: 1.0, wantMax: 1.0,
		},
		{
			name:    "2 snapshots (1 pair) → 1.0",
			probs:   []float64{0.5, 0.6},
			wantMin: 1.0, wantMax: 1.0,
		},
		{
			name:    "monotonic rise → 1.0",
			probs:   []float64{0.50, 0.55, 0.60, 0.65},
			wantMin: 1.0, wantMax: 1.0,
		},
		{
			name:    "perfect oscillation (net ≈ 0) → ~0",
			probs:   []float64{0.50, 0.60, 0.50, 0.60, 0.50},
			wantMin: 0.0, wantMax: 0.05,
		},
		{
			name:    "mostly up, one reversal → between 0.5 and 1.0",
			probs:   []float64{0.50, 0.55, 0.60, 0.55, 0.65},
			wantMin: 0.5, wantMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TrajectoryConsistency(makeSnaps(tt.probs))
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Errorf("TrajectoryConsistency = %v (invalid)", got)
				return
			}
			if got < 0 || got > 1.0001 {
				t.Errorf("TrajectoryConsistency = %v, out of [0, 1]", got)
			}
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("TrajectoryConsistency = %v, want [%v, %v]", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// ─── T015: TestScoring — 8 comprehensive cases ───────────────────────────────

func TestScoring(t *testing.T) {
	// buildChange creates a Change with given fields
	buildChange := func(eventID string, pOld, pNew float64, window time.Duration) models.Change {
		mag := math.Abs(pNew - pOld)
		dir := "increase"
		if pNew < pOld {
			dir = "decrease"
		}
		return models.Change{
			ID:             uuid.New().String(),
			EventID:        eventID,
			OldProbability: pOld,
			NewProbability: pNew,
			Magnitude:      mag,
			Direction:      dir,
			TimeWindow:     window,
			DetectedAt:     time.Now(),
		}
	}

	// buildEvent creates an Event with given volume
	buildEvent := func(id string, vol float64) *models.Event {
		return &models.Event{
			ID:         id,
			EventID:    id,
			Volume24hr: vol,
		}
	}

	t.Run("VolumeWins — large market 5% beats small market 9%", func(t *testing.T) {
		klHigh := KLDivergence(0.50, 0.55) // 5% move
		klLow := KLDivergence(0.50, 0.59)  // 9% move
		scoreA := CompositeScore(klHigh, LogVolumeWeight(1_000_000, 25000), 1.0, 1.0)
		scoreB := CompositeScore(klLow, LogVolumeWeight(30_000, 25000), 1.0, 1.0)
		if scoreA <= scoreB {
			t.Errorf("VolumeWins: A(1M vol, 5%%) = %.6f should beat B(30K vol, 9%%) = %.6f", scoreA, scoreB)
		}
	})

	t.Run("SNRWins — quiet market 3% beats volatile market 3%", func(t *testing.T) {
		kl := KLDivergence(0.50, 0.53)
		snrQuiet := HistoricalSNR(makeSnaps([]float64{0.50, 0.501, 0.499, 0.500, 0.501}), 0.03)
		snrNoisy := HistoricalSNR(makeSnaps([]float64{0.50, 0.60, 0.40, 0.65, 0.35}), 0.03)
		scoreQ := CompositeScore(kl, 1.0, snrQuiet, 1.0)
		scoreN := CompositeScore(kl, 1.0, snrNoisy, 1.0)
		if scoreQ <= scoreN {
			t.Errorf("SNRWins: quiet(SNR=%.2f)=%.6f should beat noisy(SNR=%.2f)=%.6f", snrQuiet, scoreQ, snrNoisy, scoreN)
		}
	})

	t.Run("KLRegimeDiff — same magnitude different regime gives different scores", func(t *testing.T) {
		klMid := KLDivergence(0.50, 0.55)  // 5% at mid
		klTail := KLDivergence(0.95, 1.00) // ~5% at tail (near certainty)
		if math.Abs(klMid-klTail) < 1e-6 {
			t.Errorf("KLRegimeDiff: expected different KL values, got %.6f vs %.6f", klMid, klTail)
		}
	})

	t.Run("MonotonicBeatsNoisy — clean path scores higher than oscillating path", func(t *testing.T) {
		kl := KLDivergence(0.50, 0.58)
		tcMono := TrajectoryConsistency(makeSnaps([]float64{0.50, 0.52, 0.55, 0.58}))
		tcNoisy := TrajectoryConsistency(makeSnaps([]float64{0.50, 0.68, 0.42, 0.58}))
		scoreM := CompositeScore(kl, 1.0, 1.0, tcMono)
		scoreN := CompositeScore(kl, 1.0, 1.0, tcNoisy)
		if scoreM <= scoreN {
			t.Errorf("MonotonicBeatsNoisy: mono(TC=%.2f)=%.6f should beat noisy(TC=%.2f)=%.6f",
				tcMono, scoreM, tcNoisy, scoreN)
		}
	})

	t.Run("DegenProbabilities — p=0.0 and p=1.0 do not panic or NaN", func(t *testing.T) {
		kl1 := KLDivergence(0.0, 0.05)
		kl2 := KLDivergence(1.0, 0.95)
		kl3 := KLDivergence(0.0, 1.0)
		for _, kl := range []float64{kl1, kl2, kl3} {
			if math.IsNaN(kl) || math.IsInf(kl, 0) || kl < 0 {
				t.Errorf("DegenProbabilities: KL = %v (invalid)", kl)
			}
		}
	})

	t.Run("ZeroVolumeFloor — volume=0 gets non-zero score", func(t *testing.T) {
		kl := KLDivergence(0.50, 0.55)
		vw := LogVolumeWeight(0, 25000)
		score := CompositeScore(kl, vw, 1.0, 1.0)
		if score <= 0 {
			t.Errorf("ZeroVolumeFloor: expected positive score, got %v", score)
		}
		if vw < 0.1 {
			t.Errorf("ZeroVolumeFloor: volume weight should be at least 0.1, got %v", vw)
		}
	})

	t.Run("SNRFallback — single snapshot gives SNR=1.0 and valid score", func(t *testing.T) {
		snr := HistoricalSNR(makeSnaps([]float64{0.5}), 0.05)
		if snr != 1.0 {
			t.Errorf("SNRFallback: expected SNR=1.0 for single snapshot, got %v", snr)
		}
		score := CompositeScore(KLDivergence(0.5, 0.55), 1.0, snr, 1.0)
		if math.IsNaN(score) || math.IsInf(score, 0) || score <= 0 {
			t.Errorf("SNRFallback: invalid score %v", score)
		}
	})

	t.Run("Determinism — identical inputs produce identical ranked output", func(t *testing.T) {
		store := storage.New(100, 50, "/tmp/test-determinism.json", 0644, 0755)
		mon := New(store)

		events := map[string]*models.Event{
			"evt-a": buildEvent("evt-a", 100_000),
			"evt-b": buildEvent("evt-b", 200_000),
			"evt-c": buildEvent("evt-c", 50_000),
		}
		changes := []models.Change{
			buildChange("evt-a", 0.50, 0.60, time.Hour),
			buildChange("evt-b", 0.40, 0.55, time.Hour),
			buildChange("evt-c", 0.60, 0.75, time.Hour),
		}

		result1 := mon.ScoreAndRank(changes, events, 0.0, 10)
		result2 := mon.ScoreAndRank(changes, events, 0.0, 10)

		if len(result1) != len(result2) {
			t.Fatalf("Determinism: different lengths %d vs %d", len(result1), len(result2))
		}
		for i := range result1 {
			if result1[i].EventID != result2[i].EventID || result1[i].SignalScore != result2[i].SignalScore {
				t.Errorf("Determinism: position %d differs: %s/%.6f vs %s/%.6f",
					i, result1[i].EventID, result1[i].SignalScore,
					result2[i].EventID, result2[i].SignalScore)
			}
		}
	})
}

// ─── ScoreAndRank integration tests ──────────────────────────────────────────

func TestScoreAndRank_TopKLimit(t *testing.T) {
	store := storage.New(100, 50, "/tmp/test-rank-topk.json", 0644, 0755)
	mon := New(store)

	events := map[string]*models.Event{
		"e1": {ID: "e1", Volume24hr: 100_000},
		"e2": {ID: "e2", Volume24hr: 100_000},
		"e3": {ID: "e3", Volume24hr: 100_000},
	}
	changes := []models.Change{
		{ID: "c1", EventID: "e1", OldProbability: 0.50, NewProbability: 0.65, Magnitude: 0.15, Direction: "increase", TimeWindow: time.Hour, DetectedAt: time.Now()},
		{ID: "c2", EventID: "e2", OldProbability: 0.50, NewProbability: 0.70, Magnitude: 0.20, Direction: "increase", TimeWindow: time.Hour, DetectedAt: time.Now()},
		{ID: "c3", EventID: "e3", OldProbability: 0.50, NewProbability: 0.60, Magnitude: 0.10, Direction: "increase", TimeWindow: time.Hour, DetectedAt: time.Now()},
	}

	top := mon.ScoreAndRank(changes, events, 0.0, 2)
	if len(top) != 2 {
		t.Errorf("Expected 2 results (k=2), got %d", len(top))
	}
}

func TestScoreAndRank_NeverNil(t *testing.T) {
	store := storage.New(100, 50, "/tmp/test-rank-nil.json", 0644, 0755)
	mon := New(store)

	result := mon.ScoreAndRank(nil, map[string]*models.Event{}, 0.0, 5)
	if result == nil {
		t.Error("ScoreAndRank should never return nil, got nil")
	}
}

func TestScoreAndRank_MinScoreFilters(t *testing.T) {
	store := storage.New(100, 50, "/tmp/test-rank-minscore.json", 0644, 0755)
	mon := New(store)

	events := map[string]*models.Event{
		"e1": {ID: "e1", Volume24hr: 100_000},
	}
	changes := []models.Change{
		{ID: "c1", EventID: "e1", OldProbability: 0.50, NewProbability: 0.51, Magnitude: 0.01, Direction: "increase", TimeWindow: time.Hour, DetectedAt: time.Now()},
	}

	// With very high minScore, nothing should pass
	result := mon.ScoreAndRank(changes, events, 999.0, 5)
	if len(result) != 0 {
		t.Errorf("Expected 0 results with minScore=999, got %d", len(result))
	}
}

func TestScoreAndRank_TopKZero(t *testing.T) {
	store := storage.New(100, 50, "/tmp/test-rank-k0.json", 0644, 0755)
	mon := New(store)

	events := map[string]*models.Event{
		"e1": {ID: "e1", Volume24hr: 100_000},
	}
	changes := []models.Change{
		{ID: "c1", EventID: "e1", OldProbability: 0.50, NewProbability: 0.70, Magnitude: 0.20, Direction: "increase", TimeWindow: time.Hour, DetectedAt: time.Now()},
	}

	result := mon.ScoreAndRank(changes, events, 0.0, 0)
	if len(result) != 0 {
		t.Errorf("Expected 0 results when k=0, got %d", len(result))
	}
}
