package banzhaf

import (
	"fmt"
	"math/big"

	"github.com/cheggaaa/pb/v3"
)

// zero is the zero value of big.Int.
var zero = big.NewInt(0)

// ProgressBar determines whether a progress is output
// to the standard error to show progress of the calculation.
var ProgressBar = false

// Banzhaf returns the Banzhaf power index associated with a weighted voting
// system defined by the `weights` and `quota` provided. If `absolute` is set
// to true, then the absolute Banzhaf power index is returned. Otherwise, the
// relative Banzhaf power index is returned.
//
// The quota should be an integer greater than half the total number of votes
// and less than the total number of votes.
//
// This implementation of the Banzhaf calculation uses a generator function
// approach and should run in around O(nt) where n is the number
// of players and t is the total voting weight of the system.
func Banzhaf(weights []uint64, quota uint64, absolute bool) (index []*big.Float, err error) {

	var (
		total   uint64          // total votes
		n       uint64          // number of players
		order   uint64          // maximum order of the polynomial
		P       []*big.Int      // polynomial generator function
		i, j, k uint64          // indices
		bar     *pb.ProgressBar // progress bar
	)

	// calculate the total votes
	for _, w := range weights {
		total += w
	}

	// check quota
	if quota > total || quota <= total/2 {
		return nil, fmt.Errorf("the quota is out of bounds: [%d,%d]", total/2+1, total)
	}

	// n
	n = uint64(len(weights))

	// polynomial
	P = zeroSlice(total + 1)
	P[0] = big.NewInt(1)

	// progress bar
	if ProgressBar {
		bar = pb.StartNew(int(n * total))
	}

	// Get polynomial weights. This function multiplies out
	//
	//   (1+x^w(0))(1+x^w(1))...(1+x^w(n-1))
	//
	// where w(k) are the weights (k=0,...,n-1).
	for _, w := range weights {
		order += w
		for j = order; j >= w; j-- {
			P[j] = new(big.Int).Add(P[j], P[j-w])
		}

		// progress bar
		if ProgressBar {
			bar.Add(int(total))
		}
	}

	// finish progress bars
	if ProgressBar {
		bar.Finish()
	}

	var (
		// an array counting Banzhaf power (swings)
		power = zeroSlice(n)

		// an array counting all swings
		swings = zeroSlice(quota)

		// denominator for the power index
		denom = big.NewInt(0)
	)

	// count swings and banzhaf power
	if ProgressBar {
		bar = pb.StartNew(int(n * total))
	}
	for i = 0; i < n; i++ {
		w := weights[i]
		for j = 0; j < quota; j++ {
			if j < w {
				swings[j] = P[j]
			} else {
				swings[j] = new(big.Int).Sub(P[j], swings[j-w])
			}
		}
		for k = 0; k < w; k++ {
			power[i] = new(big.Int).Add(power[i], swings[quota-1-k])
		}

		// progress bar
		if ProgressBar {
			bar.Add(int(total))
		}
	}

	// progress bar
	if ProgressBar {
		bar.Finish()
	}

	if absolute {
		// absolute Banzhaf power index takes the
		// denominator as all possible votes where
		// everyone else other than this player participates
		// which is 2^(n-1)
		denom.Exp(big.NewInt(2), new(big.Int).SetUint64(n-1), nil)
	} else {
		// normalized Banzhaf power index takes the
		// denominator as all possible swings
		for _, p := range power {
			denom.Add(denom, p)
		}
	}

	index = make([]*big.Float, n)
	d := new(big.Float).SetInt(denom)
	for i := range index {
		p := new(big.Float).SetInt(power[i])
		index[i] = new(big.Float).Quo(p, d)
	}

	return index, nil
}

// zeroSlice creates a new []*big.Int slice of size n and sets
// each item to big.NewInt(0)
func zeroSlice(n uint64) []*big.Int {
	v := make([]*big.Int, n)
	for i := range v {
		v[i] = zero
	}
	return v
}
