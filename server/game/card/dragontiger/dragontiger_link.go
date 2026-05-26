//go:build !prod || full || card

package dragontiger

import (
	"context"
	"fmt"

	"github.com/slotopol/server/game"
)

func CalcStat(ctx context.Context, sp *game.ScanPar) (float64, float64) {
	fmt.Println()
	fmt.Println("Dragon Tiger — card game")
	fmt.Println("Analytic RTP calculation not available via slot scanner.")
	fmt.Println("Standard Dragon Tiger RTP: Dragon/Tiger ~96.27%, Tie ~3.73%")
	return 0.9627, 0
}

func init() {
	var gi = game.AlgInfo{
		Aliases: []game.GameAlias{
			{Prov: "slotopol", Name: "Dragon Tiger"},
		},
		AlgDescr: game.AlgDescr{
			GT:  game.GTcard,
			SX:  0,
			SN:  0,
			RTP: []float64{96.27},
		},
	}
	gi.SetupFactory(func(sel int) game.Gamble { return NewGame() }, CalcStat)
}
