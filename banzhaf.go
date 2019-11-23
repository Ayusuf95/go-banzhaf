package banzhaf

import (
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/cheggaaa/pb/v3"
)

// zero is the zero value of big.Int.
var zero = big.NewInt(0)

// Banzhaf returns the Banzhaf power index associated with a weighted voting
// system defined by the `weights` and `quota` provided. If `absolute` is set
// to true, then the absolute Banzhaf power index is returned.
func Banzhaf(weights []uint64, quota uint64, absolute bool) (index []*big.Float, err error) {

	var (
		total      uint64     // total votes
		n          uint64     // number of players
		order      uint64     // maximum order of the polynomial
		polynomial []*big.Int // polynomial generator
		i, j, k    uint64     // indices
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

	start := time.Now()

	// polynomial
	polynomial = zeroSlice(total + 1)
	polynomial[0] = big.NewInt(1)

	// get polynomial weights
	for _, w := range weights {
		order += w
		aux := append([]*big.Int{}, polynomial...)
		for j = w; j <= order; j++ {
			polynomial[j] = new(big.Int).Add(aux[j], aux[j-w])
		}
		//log.Printf("p=%v\n", polynomial)
	}

	end := time.Since(start)
	//log.Printf("poly=%v\n", polynomial)
	log.Printf("time: %v\n", end)

	var (
		// an array counting Banzhaf power (swings)
		power = zeroSlice(n)

		// an array counting all swings
		swings = zeroSlice(quota)

		// denominator for the power index
		denom = big.NewInt(0)
	)

	// count swings and banzhaf power
	bar := pb.StartNew(int(n * total))
	for i = 0; i < n; i++ {
		w := weights[i]
		for j = 0; j < quota; j++ {
			if j < w {
				swings[j] = polynomial[j]
			} else {
				swings[j] = new(big.Int).Sub(polynomial[j], swings[j-w])
			}
		}
		for k = 0; k < w; k++ {
			power[i] = new(big.Int).Add(power[i], swings[quota-1-k])
		}
		bar.Add(int(total))
	}
	bar.Finish()

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
