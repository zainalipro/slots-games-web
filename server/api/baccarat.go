package api

import (
	"encoding/json"
	"encoding/xml"
	"log"

	"github.com/gin-gonic/gin"

	cfg "github.com/slotopol/server/config"
	"github.com/slotopol/server/game/card"
	"github.com/slotopol/server/game/card/baccarat"
	"github.com/slotopol/server/util"
)

// Returns bet value.
func ApiBaccaratBetGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_bac_betget_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_bac_betget_noscene, err)
		return
	}
	var game card.CardGame
	if game, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_bac_betget_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_bac_betget_noaccess, ErrNoAccess)
		return
	}

	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet"`
	}
	ret.Bet = game.GetBet()
	RetOk(c, ret)
}

// Set bet value.
func ApiBaccaratBetSet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_bac_betset_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_bac_betset_noscene, err)
		return
	}
	var game card.CardGame
	if game, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_bac_betset_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_bac_betset_noaccess, ErrNoAccess)
		return
	}

	if err = game.SetBet(arg.Bet); err != nil {
		Ret403(c, AEC_bac_betset_badbet, err)
		return
	}

	Ret204(c)
}

// Deal baccarat hands.
func ApiBaccaratDeal(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName   xml.Name          `json:"-" yaml:"-" xml:"arg"`
		GID       uint64            `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
		Bet       float64           `json:"bet,omitempty" yaml:"bet,omitempty" xml:"bet,omitempty"`
		BetTarget baccarat.BetTarget `json:"betTarget,omitempty" yaml:"betTarget,omitempty" xml:"betTarget,omitempty"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_bac_deal_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_bac_deal_noscene, err)
		return
	}
	var g card.CardGame
	if g, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_bac_deal_notcard, ErrNotCard)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, AEC_bac_deal_noclub, ErrNoClub)
		return
	}

	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, AEC_bac_deal_nouser, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_bac_deal_noaccess, ErrNoAccess)
		return
	}

	if arg.Bet != 0 {
		if err = g.SetBet(arg.Bet); err != nil {
			Ret403(c, AEC_bac_deal_badbet, err)
			return
		}
	}

	// Set bet target
	if arg.BetTarget != "" {
		if bacGame, ok2 := g.(*baccarat.Game); ok2 {
			bacGame.BetTarget = arg.BetTarget
		}
	}

	var cost = g.GetBet()

	var props *Props
	if props, ok = user.props.Get(scene.CID); !ok {
		Ret500(c, AEC_bac_deal_noprops, ErrNoProps)
		return
	}
	if props.Wallet < cost {
		Ret403(c, AEC_bac_deal_nomoney, ErrNoMoney)
		return
	}

	var bank = club.Bank()
	var mrtp = GetRTP(user, club)

	var debit float64
	var n = 0
	for {
		n++
		if n > cfg.Cfg.MaxSpinAttempts {
			Ret500(c, AEC_bac_deal_badbank, ErrBadBank)
			return
		}
		if err = g.Deal(); err != nil {
			continue
		}
		debit = cost - g.GetPayout()
		if bank+debit >= 0 || (bank < 0 && debit > 0) {
			break
		}
	}

	if Cfg.ClubUpdateBuffer > 1 {
		go BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit)
	} else if err = BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit); err != nil {
		Ret500(c, AEC_bac_deal_sqlbank, err)
		return
	}

	club.AddBank(debit)
	props.Wallet -= debit

	var sid = SpinCounter.Inc()
	scene.SID = sid
	var rec = Spinlog{
		SID:    sid,
		GID:    arg.GID,
		MRTP:   mrtp,
		Gain:   g.GetPayout(),
		Wallet: props.Wallet,
	}
	var b []byte
	if b, err = json.Marshal(scene.Game); err != nil {
		return
	}
	rec.Game = util.B2S(b)

	if Cfg.UseSpinLog {
		go func() {
			if err = SpinBuf.Put(cfg.XormSpinlog, rec); err != nil {
				log.Printf("can not write to spin log: %s", err.Error())
			}
		}()
	}

	var ret struct {
		XMLName     xml.Name      `json:"-" yaml:"-" xml:"ret"`
		SID         uint64        `json:"sid" yaml:"sid" xml:"sid,attr"`
		Game        card.CardGame `json:"game" yaml:"game" xml:"game"`
		PlayerHand  card.Hand     `json:"playerHand" yaml:"playerHand" xml:"playerHand"`
		BankerHand  card.Hand     `json:"bankerHand" yaml:"bankerHand" xml:"bankerHand"`
		PlayerValue int           `json:"playerValue" yaml:"playerValue" xml:"playerValue"`
		BankerValue int           `json:"bankerValue" yaml:"bankerValue" xml:"bankerValue"`
		BetTarget   string        `json:"betTarget" yaml:"betTarget" xml:"betTarget"`
		Wallet      float64       `json:"wallet" yaml:"wallet" xml:"wallet"`
		Payout      float64       `json:"payout" yaml:"payout" xml:"payout"`
		Result      string        `json:"result" yaml:"result" xml:"result"`
	}
	ret.SID = sid
	ret.Game = g
	ret.PlayerHand = g.GetHand()
	ret.BankerHand = g.GetDealerHand()
	ret.PlayerValue = card.BaccaratValue(ret.PlayerHand)
	ret.BankerValue = card.BaccaratValue(ret.BankerHand)
	if bacGame, ok2 := g.(*baccarat.Game); ok2 {
		ret.BetTarget = string(bacGame.BetTarget)
	}
	ret.Wallet = props.Wallet
	ret.Payout = g.GetPayout()
	ret.Result = g.GetResult()

	RetOk(c, ret)
}
