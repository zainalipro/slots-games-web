package api

import (
	"encoding/json"
	"encoding/xml"
	"log"

	"github.com/gin-gonic/gin"

	cfg "github.com/slotopol/server/config"
	"github.com/slotopol/server/game/card"
	"github.com/slotopol/server/game/card/videopoker"
	"github.com/slotopol/server/util"
)

// Returns bet value.
func ApiPokerBetGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_pok_betget_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_pok_betget_noscene, err)
		return
	}
	var game card.CardGame
	if game, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_pok_betget_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_pok_betget_noaccess, ErrNoAccess)
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
func ApiPokerBetSet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_pok_betset_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_pok_betset_noscene, err)
		return
	}
	var game card.CardGame
	if game, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_pok_betset_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_pok_betset_noaccess, ErrNoAccess)
		return
	}

	if err = game.SetBet(arg.Bet); err != nil {
		Ret403(c, AEC_pok_betset_badbet, err)
		return
	}

	Ret204(c)
}

// Deal initial 5 cards.
func ApiPokerDeal(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
		Bet     float64  `json:"bet,omitempty" yaml:"bet,omitempty" xml:"bet,omitempty"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_pok_deal_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_pok_deal_noscene, err)
		return
	}
	var g card.CardGame
	if g, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_pok_deal_notcard, ErrNotCard)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, AEC_pok_deal_noclub, ErrNoClub)
		return
	}

	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, AEC_pok_deal_nouser, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_pok_deal_noaccess, ErrNoAccess)
		return
	}

	if arg.Bet != 0 {
		if err = g.SetBet(arg.Bet); err != nil {
			Ret403(c, AEC_pok_deal_badbet, err)
			return
		}
	}

	var cost = g.GetBet()

	var props *Props
	if props, ok = user.props.Get(scene.CID); !ok {
		Ret500(c, AEC_pok_deal_noprops, ErrNoProps)
		return
	}
	if props.Wallet < cost {
		Ret403(c, AEC_pok_deal_nomoney, ErrNoMoney)
		return
	}

	var bank = club.Bank()
	_ = bank
	var mrtp = GetRTP(user, club)
	_ = mrtp

	// Deal (no bank validation needed for initial deal, only wallet)
	if err = g.Deal(); err != nil {
		Ret500(c, AEC_pok_deal_badbank, err)
		return
	}

	// Debit the bet immediately
	var debit = cost

	if Cfg.ClubUpdateBuffer > 1 {
		go BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit)
	} else if err = BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, debit); err != nil {
		Ret500(c, AEC_pok_deal_sqlbank, err)
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
		Gain:   0,
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
		XMLName xml.Name      `json:"-" yaml:"-" xml:"ret"`
		SID     uint64        `json:"sid" yaml:"sid" xml:"sid,attr"`
		Game    card.CardGame `json:"game" yaml:"game" xml:"game"`
		Hand    card.Hand     `json:"hand" yaml:"hand" xml:"hand"`
		Wallet  float64       `json:"wallet" yaml:"wallet" xml:"wallet"`
		State   string        `json:"state" yaml:"state" xml:"state"`
	}
	ret.SID = sid
	ret.Game = g
	ret.Hand = g.GetHand()
	ret.Wallet = props.Wallet
	ret.State = g.GetGameState()

	RetOk(c, ret)
}

// Draw - replace unheld cards and evaluate.
func ApiPokerDraw(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName  xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID      uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
		HoldMask [5]bool  `json:"hold" yaml:"hold" xml:"hold" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_pok_draw_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_pok_draw_noscene, err)
		return
	}
	var g card.CardGame
	if g, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_pok_draw_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_pok_draw_noaccess, ErrNoAccess)
		return
	}

	var pokGame = g.(*videopoker.Game)
	if err = pokGame.SetHold(arg.HoldMask); err != nil {
		Ret403(c, AEC_pok_draw_nogame, err)
		return
	}
	if err = pokGame.Draw(); err != nil {
		Ret403(c, AEC_pok_draw_nogame, err)
		return
	}

	// Credit winnings to wallet
	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, AEC_pok_draw_nogame, ErrNoUser)
		return
	}
	var props *Props
	if props, ok = user.props.Get(scene.CID); !ok {
		Ret500(c, AEC_pok_draw_nogame, ErrNoProps)
		return
	}
	var club *Club
	if club, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, AEC_pok_draw_nogame, ErrNoClub)
		return
	}

	var gain = pokGame.GetPayout()
	var credit = gain // positive credit = back to wallet

	if credit > 0 {
		if Cfg.ClubUpdateBuffer > 1 {
			go BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, -credit)
		} else if err = BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, -credit); err != nil {
			Ret500(c, AEC_pok_draw_nogame, err)
			return
		}
		club.AddBank(-credit)
		props.Wallet += credit
	}

	// Update spin log with draw result
	var sid = SpinCounter.Inc()
	scene.SID = sid
	var rec = Spinlog{
		SID:    sid,
		GID:    arg.GID,
		MRTP:   0,
		Gain:   gain,
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
		XMLName    xml.Name            `json:"-" yaml:"-" xml:"ret"`
		Game       card.CardGame       `json:"game" yaml:"game" xml:"game"`
		Hand       card.Hand           `json:"hand" yaml:"hand" xml:"hand"`
		HandResult card.PokerHandResult `json:"handResult" yaml:"handResult" xml:"handResult"`
		Payout     float64             `json:"payout" yaml:"payout" xml:"payout"`
		Result     string              `json:"result" yaml:"result" xml:"result"`
		Wallet     float64             `json:"wallet" yaml:"wallet" xml:"wallet"`
	}
	ret.Game = g
	ret.Hand = pokGame.GetHand()
	ret.HandResult = pokGame.HandResult
	ret.Payout = pokGame.GetPayout()
	ret.Result = pokGame.GetResult()
	ret.Wallet = props.Wallet

	RetOk(c, ret)
}
