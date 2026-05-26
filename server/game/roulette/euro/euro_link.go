//go:build !prod || full || card

package euro

import (
	"context"
	"fmt"

	"github.com/slotopol/server/game"
)

func CalcStat(ctx context.Context, sp *game.ScanPar) (float64, float64) {
	fmt.Println()
	fmt.Println("European Roulette — single zero")
	fmt.Println("House edge: 2.7%")
	return 0.973, 0
}

func init() {
	var gi = game.AlgInfo{
		Aliases: []game.GameAlias{
			{Prov: "slotopol", Name: "European Roulette"},
		},
		AlgDescr: game.AlgDescr{
			GT:  game.GTcard,
			SX:  0,
			SN:  0,
			RTP: []float64{97.30},
		},
	}
	gi.SetupFactory(func(sel int) game.Gamble { return NewGame() }, CalcStat)
}
