//go:build !prod || full || crash

package aviator

import (
	"context"
	"fmt"

	"github.com/slotopol/server/game"
)

func CalcStat(ctx context.Context, sp *game.ScanPar) (float64, float64) {
	fmt.Println()
	fmt.Println("Aviator — crash game")
	fmt.Println("Analytic RTP calculation not available via slot scanner.")
	fmt.Printf("House edge: %.1f%%, Expected RTP: %.1f%%\n", HouseEdge*100, (1-HouseEdge)*100)
	return 1 - HouseEdge, 0
}

func init() {
	var gi = game.AlgInfo{
		Aliases: []game.GameAlias{
			{Prov: "slotopol", Name: "Aviator"},
		},
		AlgDescr: game.AlgDescr{
			GT:  game.GTcrash,
			SX:  0,
			SN:  0,
			RTP: []float64{97.0}, // 3% house edge
		},
	}
	gi.SetupFactory(func(sel int) game.Gamble { return NewGame() }, CalcStat)
}
