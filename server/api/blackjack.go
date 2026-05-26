package api

import (
	"encoding/json"
	"encoding/xml"
	"log"

	"github.com/gin-gonic/gin"

	cfg "github.com/slotopol/server/config"
	"github.com/slotopol/server/game/card"
	"github.com/slotopol/server/game/card/blackjack"
	"github.com/slotopol/server/util"
)

// Returns bet value.
func ApiBlackjackBetGet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}
	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_bj_betget_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_bj_betget_noscene, err)
		return
	}
	var game card.CardGame
	if game, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_bj_betget_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_bj_betget_noaccess, ErrNoAccess)
		return
	}

	ret.Bet = game.GetBet()
	RetOk(c, ret)
}

// Set bet value.
func ApiBlackjackBetSet(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" binding:"required"`
		Bet     float64  `json:"bet" yaml:"bet" xml:"bet" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_bj_betset_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_bj_betset_noscene, err)
		return
	}
	var game card.CardGame
	if game, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_bj_betset_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_bj_betset_noaccess, ErrNoAccess)
		return
	}

	if err = game.SetBet(arg.Bet); err != nil {
		Ret403(c, AEC_bj_betset_badbet, err)
		return
	}

	Ret204(c)
}

// Deal initial hands.
func ApiBlackjackDeal(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
		Bet     float64  `json:"bet,omitempty" yaml:"bet,omitempty" xml:"bet,omitempty"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_bj_deal_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_bj_deal_noscene, err)
		return
	}
	var g card.CardGame
	if g, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_bj_deal_notcard, ErrNotCard)
		return
	}

	var club *Club
	if club, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, AEC_bj_deal_noclub, ErrNoClub)
		return
	}

	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, AEC_bj_deal_nouser, ErrNoUser)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_bj_deal_noaccess, ErrNoAccess)
		return
	}

	if arg.Bet != 0 {
		if err = g.SetBet(arg.Bet); err != nil {
			Ret403(c, AEC_bj_deal_badbet, err)
			return
		}
	}

	var cost = g.GetBet()

	var props *Props
	if props, ok = user.props.Get(scene.CID); !ok {
		Ret500(c, AEC_bj_deal_noprops, ErrNoProps)
		return
	}
	if props.Wallet < cost {
		Ret403(c, AEC_bj_deal_nomoney, ErrNoMoney)
		return
	}

	var bank = club.Bank()
	var mrtp = GetRTP(user, club)

	var debit float64
	var n = 0
	for {
		n++
		if n > cfg.Cfg.MaxSpinAttempts {
			Ret500(c, AEC_bj_deal_badbank, ErrBadBank)
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
		Ret500(c, AEC_bj_deal_sqlbank, err)
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
		XMLName      xml.Name      `json:"-" yaml:"-" xml:"ret"`
		SID          uint64        `json:"sid" yaml:"sid" xml:"sid,attr"`
		Game         card.CardGame `json:"game" yaml:"game" xml:"game"`
		Hand         card.Hand     `json:"hand" yaml:"hand" xml:"hand"`
		DealerHand   card.Hand     `json:"dealerHand" yaml:"dealerHand" xml:"dealerHand"`
		DealerUpcard card.Card     `json:"dealerUpcard" yaml:"dealerUpcard" xml:"dealerUpcard"`
		Wallet       float64       `json:"wallet" yaml:"wallet" xml:"wallet"`
		State        string        `json:"state" yaml:"state" xml:"state"`
		Payout       float64       `json:"payout" yaml:"payout" xml:"payout"`
		Result       string        `json:"result" yaml:"result" xml:"result"`
	}
	ret.SID = sid
	ret.Game = g
	ret.Hand = g.GetHand()
	ret.DealerHand = g.GetDealerHand()
	if bj, ok2 := g.(*blackjack.Game); ok2 {
		ret.DealerUpcard = bj.GetDealerUpcard()
	}
	ret.Wallet = props.Wallet
	ret.State = g.GetGameState()
	ret.Payout = g.GetPayout()
	ret.Result = g.GetResult()

	RetOk(c, ret)
}

// Hit - take another card.
func ApiBlackjackHit(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_bj_hit_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_bj_hit_noscene, err)
		return
	}
	var g card.CardGame
	if g, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_bj_hit_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_bj_hit_noaccess, ErrNoAccess)
		return
	}

	if err = g.(*blackjack.Game).Hit(); err != nil {
		Ret403(c, AEC_bj_hit_nogame, err)
		return
	}

	var ret struct {
		XMLName    xml.Name      `json:"-" yaml:"-" xml:"ret"`
		Game       card.CardGame `json:"game" yaml:"game" xml:"game"`
		Hand       card.Hand     `json:"hand" yaml:"hand" xml:"hand"`
		State      string        `json:"state" yaml:"state" xml:"state"`
		Payout     float64       `json:"payout" yaml:"payout" xml:"payout"`
		Result     string        `json:"result" yaml:"result" xml:"result"`
		DealerHand card.Hand     `json:"dealerHand,omitempty" yaml:"dealerHand,omitempty" xml:"dealerHand,omitempty"`
	}
	ret.Game = g
	ret.Hand = g.GetHand()
	ret.State = g.GetGameState()
	ret.Payout = g.GetPayout()
	ret.Result = g.GetResult()
	if g.GetGameState() == "done" {
		ret.DealerHand = g.GetDealerHand()
	}

	RetOk(c, ret)
}

// Stand - end turn and let dealer play.
func ApiBlackjackStand(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_bj_stand_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_bj_stand_noscene, err)
		return
	}
	var g card.CardGame
	if g, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_bj_stand_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_bj_stand_noaccess, ErrNoAccess)
		return
	}

	if err = g.(*blackjack.Game).Stand(); err != nil {
		Ret403(c, AEC_bj_stand_nogame, err)
		return
	}

	var ret struct {
		XMLName    xml.Name      `json:"-" yaml:"-" xml:"ret"`
		Game       card.CardGame `json:"game" yaml:"game" xml:"game"`
		Hand       card.Hand     `json:"hand" yaml:"hand" xml:"hand"`
		DealerHand card.Hand     `json:"dealerHand" yaml:"dealerHand" xml:"dealerHand"`
		State      string        `json:"state" yaml:"state" xml:"state"`
		Payout     float64       `json:"payout" yaml:"payout" xml:"payout"`
		Result     string        `json:"result" yaml:"result" xml:"result"`
	}
	ret.Game = g
	ret.Hand = g.GetHand()
	ret.DealerHand = g.GetDealerHand()
	ret.State = g.GetGameState()
	ret.Payout = g.GetPayout()
	ret.Result = g.GetResult()

	RetOk(c, ret)
}

// Double - double down. Debits the extra bet from wallet.
func ApiBlackjackDouble(c *gin.Context) {
	var err error
	var ok bool
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}

	if err = c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_bj_double_nobind, err)
		return
	}

	var scene *Scene
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_bj_double_noscene, err)
		return
	}
	var g card.CardGame
	if g, ok = scene.Game.(card.CardGame); !ok {
		Ret403(c, AEC_bj_double_notcard, ErrNotCard)
		return
	}

	var admin, al = MustAdmin(c, scene.CID)
	if admin.UID != scene.UID && al&ALdealer == 0 {
		Ret403(c, AEC_bj_double_noaccess, ErrNoAccess)
		return
	}

	// Check wallet for extra double bet
	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, AEC_bj_double_nomoney, ErrNoUser)
		return
	}
	var props *Props
	if props, ok = user.props.Get(scene.CID); !ok {
		Ret500(c, AEC_bj_double_nomoney, ErrNoProps)
		return
	}
	var club *Club
	if club, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, AEC_bj_double_nomoney, ErrNoClub)
		return
	}

	if props.Wallet < g.GetBet() {
		Ret403(c, AEC_bj_double_nomoney, ErrNoMoney)
		return
	}

	if err = g.(*blackjack.Game).Double(); err != nil {
		Ret403(c, AEC_bj_double_nogame, err)
		return
	}

	// Debit the extra bet from wallet
	var extraDebit = g.GetBet() // additional bet = original bet
	if Cfg.ClubUpdateBuffer > 1 {
		go BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, extraDebit)
	} else if err = BankBat[scene.CID].Put(cfg.XormStorage, scene.UID, extraDebit); err != nil {
		Ret500(c, AEC_bj_double_nomoney, err)
		return
	}
	club.AddBank(extraDebit)
	props.Wallet -= extraDebit

	var ret struct {
		XMLName    xml.Name      `json:"-" yaml:"-" xml:"ret"`
		Game       card.CardGame `json:"game" yaml:"game" xml:"game"`
		Hand       card.Hand     `json:"hand" yaml:"hand" xml:"hand"`
		DealerHand card.Hand     `json:"dealerHand" yaml:"dealerHand" xml:"dealerHand"`
		State      string        `json:"state" yaml:"state" xml:"state"`
		Payout     float64       `json:"payout" yaml:"payout" xml:"payout"`
		Result     string        `json:"result" yaml:"result" xml:"result"`
		Wallet     float64       `json:"wallet" yaml:"wallet" xml:"wallet"`
	}
	ret.Game = g
	ret.Hand = g.GetHand()
	ret.DealerHand = g.GetDealerHand()
	ret.State = g.GetGameState()
	ret.Payout = g.GetPayout()
	ret.Result = g.GetResult()
	ret.Wallet = props.Wallet

	RetOk(c, ret)
}
