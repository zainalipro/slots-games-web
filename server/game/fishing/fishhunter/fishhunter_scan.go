package fishhunter

// ReelsMap defines the available RTP levels. Each cannon level has its own
// RTP computed analytically from the fish distribution, catch probabilities,
// and cost multipliers.
var ReelsMap = map[float64]struct{}{
	94.6: {}, // Cannon 1
	87.8: {}, // Cannon 2
	91.1: {}, // Cannon 3
	92.0: {}, // Cannon 4
	94.1: {}, // Cannon 5
}
