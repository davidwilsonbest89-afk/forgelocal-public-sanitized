package launch

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"sync/atomic"
	"time"
)

// newID generates a collision-resistant session identifier without
// exposing any host secret (no hostname, PID or user info). T08 IDs are
// purely opaque tokens; they are not cryptographic nonces.
var seq uint64

// idRand is lazily populated so tests without entropy (e.g. sealed envs)
// still get a deterministic fallback.
var (
	randOnce sync.Once
	randBits atomic.Value // uint64
)

func readRand() uint64 {
	var buf [8]byte
	_, err := rand.Read(buf[:])
	if err != nil {
		// Fallback: monotonic counter + time, deterministic in tests.
		return 0
	}
	var v uint64
	for i := 0; i < 8; i++ {
		v = v<<8 | uint64(buf[i])
	}
	return v
}

func newID() string {
	seq := atomic.AddUint64(&seq, 1)
	now := uint64(time.Now().UnixMilli())
	var v uint64
	randOnce.Do(func() { randBits.Store(readRand()) })
	if r, ok := randBits.Load().(uint64); ok {
		v = r
	}
	if v == 0 {
		v = seq ^ (now * 0x9e3779b97f4a7c15)
	}
	buf := make([]byte, 16)
	for i := 0; i < 8; i++ {
		buf[i] = byte(v >> (56 - 8*i))
	}
	for i := 8; i < 16; i++ {
		buf[i] = byte(now >> (120 - 8*i))
	}
	buf[0] ^= byte(seq)
	buf[1] ^= byte(seq >> 8)
	return "sess-" + hex.EncodeToString(buf)[:32]
}
