package preview

import "testing"

func TestAngleFrameOrdinal(t *testing.T) {
	cases := []struct {
		rotation, angles, frames, want int
	}{
		{4, 42, 1, 4},     // static art: rotation is the frame (oak)
		{6, 3, 1, -1},     // out-of-range static facing is rejected
		{0, 16, 60, 0},    // animated: first frame of the first facing
		{7, 16, 60, 420},  // first frame of facing 7
		{15, 16, 60, 900}, // last facing
		{16, 16, 60, -1},  // past the last facing: reported, not wrapped
		{3, 1, 10, -1},    // animation frames cannot create facing slots
		{4, 1, 90, -1},    // fish trap: stored facing 4 is not animation frame 4
		{3, 1, 90, -1},    // flag: stored facing 3 is unavailable
		{2, 1, 180, -1},   // turtles: stored facing 2 is unavailable
		{0, 1, 180, 0},    // single-facing animations retain their first frame
		{0, 0, 0, 0},      // absent metadata permits the default facing only
		{1, 0, 0, -1},
		{-1, 16, 60, -1},
		{0, -1, 1, -1},
		{0, 1, -1, -1},
	}
	for _, c := range cases {
		if got := angleFrameOrdinal(c.rotation, c.angles, c.frames); got != c.want {
			t.Errorf("angleFrameOrdinal(%d, %d, %d) = %d, want %d", c.rotation, c.angles, c.frames, got, c.want)
		}
	}
}
