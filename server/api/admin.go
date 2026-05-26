package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"

	cfg "github.com/slotopol/server/config"
	"github.com/slotopol/server/game"
)

// AdminSettings stores all admin-configurable values in memory.
type AdminSettings struct {
	PaymentMethods   []PaymentMethod `json:"paymentMethods" yaml:"paymentMethods" xml:"paymentMethods"`
	MinDeposit       float64         `json:"minDeposit" yaml:"minDeposit" xml:"minDeposit"`
	MaxDeposit       float64         `json:"maxDeposit" yaml:"maxDeposit" xml:"maxDeposit"`
	MinWithdrawal    float64         `json:"minWithdrawal" yaml:"minWithdrawal" xml:"minWithdrawal"`
	MaxWithdrawal    float64         `json:"maxWithdrawal" yaml:"maxWithdrawal" xml:"maxWithdrawal"`
	DefaultCommission float64        `json:"defaultCommission" yaml:"defaultCommission" xml:"defaultCommission"`
	WinSchedule      string         `json:"winSchedule" yaml:"winSchedule" xml:"winSchedule"`
	RegistrationBonus float64        `json:"registrationBonus" yaml:"registrationBonus" xml:"registrationBonus"`
	DepositBonus     float64         `json:"depositBonus" yaml:"depositBonus" xml:"depositBonus"`
	WelcomeMessage   string         `json:"welcomeMessage" yaml:"welcomeMessage" xml:"welcomeMessage"`
	SiteName         string         `json:"siteName" yaml:"siteName" xml:"siteName"`
	SiteLogo         string         `json:"siteLogo" yaml:"siteLogo" xml:"siteLogo"`
}

type PaymentMethod struct {
	Name   string `json:"name" yaml:"name" xml:"name"`
	Logo   string `json:"logo" yaml:"logo" xml:"logo"`
	Active bool   `json:"active" yaml:"active" xml:"active"`
}

var AdminConf = AdminSettings{
	PaymentMethods: []PaymentMethod{
		{Name: "Easypaisa", Logo: "/uploads/easypaisa.png", Active: true},
		{Name: "JazzCash", Logo: "/uploads/jazzcash.png", Active: true},
	},
	MinDeposit:        100,
	MaxDeposit:        100000,
	MinWithdrawal:     500,
	MaxWithdrawal:     50000,
	DefaultCommission: 5.0,
	WinSchedule:       "08:00-23:00",
	RegistrationBonus: 50,
	DepositBonus:      10,
	WelcomeMessage:    "Welcome to the casino!",
	SiteName:          "Slotopol Casino",
	SiteLogo:          "/uploads/logo.png",
}

// AdminConfPath is the path to the admin config YAML file.
var AdminConfPath string

// SaveAdminConf persists the admin configuration to a YAML file.
func SaveAdminConf() error {
	if AdminConfPath == "" {
		return nil
	}
	var data, err = yaml.Marshal(AdminConf)
	if err != nil {
		return err
	}
	return os.WriteFile(AdminConfPath, data, 0644)
}

// LoadAdminConf loads the admin configuration from a YAML file.
// If the file does not exist, the defaults are kept.
func LoadAdminConf(path string) error {
	AdminConfPath = path
	var data, err = os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // keep defaults
		}
		return err
	}
	return yaml.Unmarshal(data, &AdminConf)
}

// GET /admin/settings - get admin settings
func ApiAdminSettingsGet(c *gin.Context) {
	RetOk(c, AdminConf)
}

// POST /admin/settings - update admin settings
func ApiAdminSettingsSet(c *gin.Context) {
	var arg struct {
		XMLName          xml.Name `json:"-" yaml:"-" xml:"arg"`
		MinDeposit       *float64 `json:"minDeposit,omitempty" yaml:"minDeposit,omitempty" xml:"minDeposit,omitempty"`
		MaxDeposit       *float64 `json:"maxDeposit,omitempty" yaml:"maxDeposit,omitempty" xml:"maxDeposit,omitempty"`
		MinWithdrawal    *float64 `json:"minWithdrawal,omitempty" yaml:"minWithdrawal,omitempty" xml:"minWithdrawal,omitempty"`
		MaxWithdrawal    *float64 `json:"maxWithdrawal,omitempty" yaml:"maxWithdrawal,omitempty" xml:"maxWithdrawal,omitempty"`
		DefaultCommission *float64 `json:"defaultCommission,omitempty" yaml:"defaultCommission,omitempty" xml:"defaultCommission,omitempty"`
		WinSchedule      *string  `json:"winSchedule,omitempty" yaml:"winSchedule,omitempty" xml:"winSchedule,omitempty"`
		RegistrationBonus *float64 `json:"registrationBonus,omitempty" yaml:"registrationBonus,omitempty" xml:"registrationBonus,omitempty"`
		DepositBonus     *float64 `json:"depositBonus,omitempty" yaml:"depositBonus,omitempty" xml:"depositBonus,omitempty"`
		WelcomeMessage   *string  `json:"welcomeMessage,omitempty" yaml:"welcomeMessage,omitempty" xml:"welcomeMessage,omitempty"`
		SiteName         *string  `json:"siteName,omitempty" yaml:"siteName,omitempty" xml:"siteName,omitempty"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_admin_settings_nobind, err)
		return
	}

	if arg.MinDeposit != nil {
		AdminConf.MinDeposit = *arg.MinDeposit
	}
	if arg.MaxDeposit != nil {
		AdminConf.MaxDeposit = *arg.MaxDeposit
	}
	if arg.MinWithdrawal != nil {
		AdminConf.MinWithdrawal = *arg.MinWithdrawal
	}
	if arg.MaxWithdrawal != nil {
		AdminConf.MaxWithdrawal = *arg.MaxWithdrawal
	}
	if arg.DefaultCommission != nil {
		AdminConf.DefaultCommission = *arg.DefaultCommission
	}
	if arg.WinSchedule != nil {
		AdminConf.WinSchedule = *arg.WinSchedule
	}
	if arg.RegistrationBonus != nil {
		AdminConf.RegistrationBonus = *arg.RegistrationBonus
	}
	if arg.DepositBonus != nil {
		AdminConf.DepositBonus = *arg.DepositBonus
	}
	if arg.WelcomeMessage != nil {
		AdminConf.WelcomeMessage = *arg.WelcomeMessage
	}
	if arg.SiteName != nil {
		AdminConf.SiteName = *arg.SiteName
	}

	// Persist settings to disk
	if err := SaveAdminConf(); err != nil {
		log.Printf("failed to persist admin config: %s", err.Error())
	}

	RetOk(c, AdminConf)
}

// GET /admin/payments - list payment methods
func ApiAdminPaymentsGet(c *gin.Context) {
	RetOk(c, AdminConf.PaymentMethods)
}

// POST /admin/payments/add - add payment method
func ApiAdminPaymentAdd(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		Name    string   `json:"name" yaml:"name" xml:"name" binding:"required"`
		Logo    string   `json:"logo,omitempty" yaml:"logo,omitempty" xml:"logo,omitempty"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_admin_payment_nobind, err)
		return
	}

	AdminConf.PaymentMethods = append(AdminConf.PaymentMethods, PaymentMethod{
		Name:   arg.Name,
		Logo:   arg.Logo,
		Active: true,
	})

	if err := SaveAdminConf(); err != nil {
		log.Printf("failed to persist admin config: %s", err.Error())
	}

	RetOk(c, AdminConf.PaymentMethods)
}

// POST /admin/payments/toggle - toggle payment method active state
func ApiAdminPaymentToggle(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		Name    string   `json:"name" yaml:"name" xml:"name" binding:"required"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_admin_payment_nobind, err)
		return
	}

	for i, pm := range AdminConf.PaymentMethods {
		if strings.EqualFold(pm.Name, arg.Name) {
			AdminConf.PaymentMethods[i].Active = !AdminConf.PaymentMethods[i].Active
			break
		}
	}

	if err := SaveAdminConf(); err != nil {
		log.Printf("failed to persist admin config: %s", err.Error())
	}

	RetOk(c, AdminConf.PaymentMethods)
}

// POST /admin/payments/remove - remove payment method
func ApiAdminPaymentRemove(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		Name    string   `json:"name" yaml:"name" xml:"name" binding:"required"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_admin_payment_nobind, err)
		return
	}

	var filtered []PaymentMethod
	for _, pm := range AdminConf.PaymentMethods {
		if !strings.EqualFold(pm.Name, arg.Name) {
			filtered = append(filtered, pm)
		}
	}
	AdminConf.PaymentMethods = filtered

	if err := SaveAdminConf(); err != nil {
		log.Printf("failed to persist admin config: %s", err.Error())
	}

	RetOk(c, AdminConf.PaymentMethods)
}

// POST /admin/users/list - list all users
func ApiAdminUsersList(c *gin.Context) {
	type userItem struct {
		XMLName   xml.Name `json:"-" yaml:"-" xml:"user"`
		UID       uint64   `json:"uid" yaml:"uid" xml:"uid,attr"`
		Email     string   `json:"email" yaml:"email" xml:"email"`
		Name      string   `json:"name,omitempty" yaml:"name,omitempty" xml:"name,omitempty"`
		Status    UF       `json:"status" yaml:"status" xml:"status"`
		GlobalAL  AL       `json:"gal" yaml:"gal" xml:"gal"`
		CreatedAt string   `json:"createdAt" yaml:"createdAt" xml:"createdAt"`
	}

	var ret struct {
		XMLName xml.Name   `json:"-" yaml:"-" xml:"ret"`
		Users   []userItem `json:"users" yaml:"users" xml:"users>user"`
	}

	for _, user := range Users.Items() {
		ret.Users = append(ret.Users, userItem{
			UID:       user.UID,
			Email:     user.Email,
			Name:      user.Name,
			Status:    user.Status,
			GlobalAL:  user.GAL,
			CreatedAt: user.CTime.Format(time.RFC3339),
		})
	}

	RetOk(c, ret)
}

// POST /admin/user/block - block/unblock user
func ApiAdminUserBlock(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" binding:"required"`
		Blocked bool     `json:"blocked" yaml:"blocked" xml:"blocked"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_admin_user_nobind, err)
		return
	}

	var user *User
	var ok bool
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, AEC_admin_user_nouser, ErrNoUser)
		return
	}

	if arg.Blocked {
		user.GAL &^= ALmember
	} else {
		user.GAL |= ALmember
	}

	Ret204(c)
}

// POST /admin/user/al/set - set user access level at club
func ApiAdminUserALSet(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" binding:"required"`
		CID     uint64   `json:"cid" yaml:"cid" xml:"cid,attr" binding:"required"`
		Access  AL       `json:"access" yaml:"access" xml:"access"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_admin_user_nobind, err)
		return
	}

	var user *User
	var ok bool
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, AEC_admin_user_nouser, ErrNoUser)
		return
	}

	var props *Props
	if props, ok = user.props.Get(arg.CID); ok {
		props.Access = arg.Access
	}

	Ret204(c)
}

// POST /admin/user/commission/set - set user commission %
func ApiAdminUserCommissionSet(c *gin.Context) {
	// Commission is stored as a special property - for simplicity, store in a map
	type args struct {
		XMLName    xml.Name `json:"-" yaml:"-" xml:"arg"`
		UID        uint64   `json:"uid" yaml:"uid" xml:"uid,attr" binding:"required"`
		Commission float64  `json:"commission" yaml:"commission" xml:"commission"`
	}

	var arg args
	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_admin_user_nobind, err)
		return
	}

	var user *User
	var ok bool
	if user, ok = Users.Get(arg.UID); !ok {
		Ret404(c, AEC_admin_user_nouser, ErrNoUser)
		return
	}
	_ = user

	// Store commission in UserCommission map (defined below)
	UserCommission[arg.UID] = arg.Commission

	var ret struct {
		XMLName    xml.Name `json:"-" yaml:"-" xml:"ret"`
		Commission float64  `json:"commission" yaml:"commission" xml:"commission"`
	}
	ret.Commission = arg.Commission
	RetOk(c, ret)
}

// GET /admin/user/commission/get - get user commission %
func ApiAdminUserCommissionGet(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		UID     uint64   `json:"uid" yaml:"uid" xml:"uid,attr" binding:"required"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_admin_user_nobind, err)
		return
	}

	var ret struct {
		XMLName    xml.Name `json:"-" yaml:"-" xml:"ret"`
		Commission float64  `json:"commission" yaml:"commission" xml:"commission"`
	}

	if c, ok := UserCommission[arg.UID]; ok {
		ret.Commission = c
	} else {
		ret.Commission = AdminConf.DefaultCommission
	}

	RetOk(c, ret)
}

// POST /admin/upload - upload an image/file
func ApiAdminUpload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		Ret400(c, AEC_admin_upload_nofile, err)
		return
	}

	// Ensure uploads directory exists
	uploadDir := filepath.Join(".", "uploads")
	os.MkdirAll(uploadDir, 0755)

	// Generate safe filename
	ext := filepath.Ext(file.Filename)
	safeName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savePath := filepath.Join(uploadDir, safeName)

	// Save the file
	src, err := file.Open()
	if err != nil {
		Ret500(c, AEC_admin_upload_fail, err)
		return
	}
	defer src.Close()

	dst, err := os.Create(savePath)
	if err != nil {
		Ret500(c, AEC_admin_upload_fail, err)
		return
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		Ret500(c, AEC_admin_upload_fail, err)
		return
	}

	var ret struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"ret"`
		Path    string   `json:"path" yaml:"path" xml:"path"`
		URL     string   `json:"url" yaml:"url" xml:"url"`
	}
	ret.Path = savePath
	ret.URL = "/uploads/" + safeName

	RetOk(c, ret)
}

// GET /admin/analytics - basic analytics
func ApiAdminAnalytics(c *gin.Context) {
	var ret struct {
		XMLName     xml.Name `json:"-" yaml:"-" xml:"ret"`
		TotalUsers  int      `json:"totalUsers" yaml:"totalUsers" xml:"totalUsers"`
		TotalGames  int      `json:"totalGames" yaml:"totalGames" xml:"totalGames"`
		ActiveGames int      `json:"activeGames" yaml:"activeGames" xml:"activeGames"`
	}

	ret.TotalUsers = Users.Len()
	ret.TotalGames = len(game.GameFactory)
	ret.ActiveGames = Scenes.Len()

	RetOk(c, ret)
}

// POST /admin/bonus/claim - let a user claim registration bonus
func ApiAdminBonusClaim(c *gin.Context) {
	var arg struct {
		XMLName xml.Name `json:"-" yaml:"-" xml:"arg"`
		GID     uint64   `json:"gid" yaml:"gid" xml:"gid,attr" form:"gid" binding:"required"`
	}

	if err := c.ShouldBind(&arg); err != nil {
		Ret400(c, AEC_admin_bonus_nobind, err)
		return
	}

	var scene *Scene
	var err error
	if scene, err = GetScene(arg.GID); err != nil {
		Ret404(c, AEC_admin_bonus_noscene, err)
		return
	}

	var ok bool
	if _, ok = Clubs.Get(scene.CID); !ok {
		Ret500(c, AEC_admin_bonus_noclub, ErrNoClub)
		return
	}

	var user *User
	if user, ok = Users.Get(scene.UID); !ok {
		Ret500(c, AEC_admin_bonus_nouser, ErrNoUser)
		return
	}

	// Add registration bonus to wallet
	if props, ok := user.props.Get(scene.CID); ok {
		var bonus = AdminConf.RegistrationBonus
		if Cfg.ClubInsertBuffer > 1 {
			go BankBat[scene.CID].Add(cfg.XormStorage, scene.UID, scene.UID, props.Wallet+bonus, bonus)
		} else {
			BankBat[scene.CID].Add(cfg.XormStorage, scene.UID, scene.UID, props.Wallet+bonus, bonus)
		}
		props.Wallet += bonus
	}

	var ret struct {
		XMLName  xml.Name `json:"-" yaml:"-" xml:"ret"`
		Bonus    float64  `json:"bonus" yaml:"bonus" xml:"bonus"`
		Wallet   float64  `json:"wallet" yaml:"wallet" xml:"wallet"`
	}
	ret.Bonus = AdminConf.RegistrationBonus
	ret.Wallet = user.GetWallet(scene.CID)

	RetOk(c, ret)
}

// UserCommission stores per-user commission percentages.
var UserCommission = map[uint64]float64{}

const (
	// Admin error codes (starting at 2000+ to avoid conflict)
	AEC_admin_settings_nobind = 2000 + iota
	AEC_admin_payment_nobind
	AEC_admin_user_nobind
	AEC_admin_user_nouser
	AEC_admin_upload_nofile
	AEC_admin_upload_fail
	AEC_admin_bonus_nobind
	AEC_admin_bonus_noscene
	AEC_admin_bonus_noclub
	AEC_admin_bonus_nouser
)
