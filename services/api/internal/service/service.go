package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/local-life/partner/services/api/internal/config"
	"github.com/local-life/partner/services/api/internal/llm"
	"github.com/local-life/partner/services/api/internal/middleware"
	"github.com/local-life/partner/services/api/internal/models"
	"github.com/local-life/partner/services/api/internal/pkg/crypto"
	"github.com/local-life/partner/services/api/internal/pkg/sid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type Services struct {
	DB     *gorm.DB
	Cfg    *config.Config
	Cipher *crypto.AESCipher
	LLM    *llm.Client
}

func New(db *gorm.DB, cfg *config.Config, cipher *crypto.AESCipher, llmClient *llm.Client) *Services {
	return &Services{DB: db, Cfg: cfg, Cipher: cipher, LLM: llmClient}
}

func (s *Services) Audit(adminID *uint64, action, targetType, targetID string, detail models.JSONMap) {
	_ = s.DB.Create(&models.AuditLog{
		AdminUserID: adminID,
		Action:      action,
		TargetType:  targetType,
		TargetID:    targetID,
		DetailJSON:  detail,
	}).Error
}

func (s *Services) AdminLogin(username, password string) (*models.AdminUser, string, error) {
	var user models.AdminUser
	if err := s.DB.Where("username = ? AND status = 1", username).First(&user).Error; err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, "", errors.New("用户名或密码错误")
	}
	token, err := middleware.SignAdmin(s.Cfg.JWT.AdminSecret, s.Cfg.JWT.AdminTTLHours, user.ID, user.Username, user.Role)
	return &user, token, err
}

type UnionInput struct {
	Name           string `json:"name" binding:"required"`
	CommissionRate int    `json:"commission_rate"`
	ContactName    string `json:"contact_name"`
	ContactPhone   string `json:"contact_phone"`
	SettleAccount  string `json:"settle_account"`
	Status         *int   `json:"status"`
}

func (s *Services) ListUnions() ([]models.Union, error) {
	var list []models.Union
	return list, s.DB.Order("id desc").Find(&list).Error
}

func (s *Services) CreateUnion(in UnionInput) (*models.Union, error) {
	if in.CommissionRate < 0 || in.CommissionRate > 10000 {
		return nil, errors.New("分成比例需在 0-10000 万分比之间")
	}
	enc, err := s.Cipher.Encrypt(in.ContactPhone)
	if err != nil {
		return nil, err
	}
	status := 1
	if in.Status != nil {
		status = *in.Status
	}
	u := models.Union{
		Name: in.Name, CommissionRate: in.CommissionRate, ContactName: in.ContactName,
		ContactPhoneEnc: enc, SettleAccount: in.SettleAccount, Status: status,
	}
	if err := s.DB.Create(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Services) UpdateUnion(id uint64, in UnionInput) (*models.Union, error) {
	var u models.Union
	if err := s.DB.First(&u, id).Error; err != nil {
		return nil, errors.New("工会不存在")
	}
	if in.CommissionRate < 0 || in.CommissionRate > 10000 {
		return nil, errors.New("分成比例需在 0-10000 万分比之间")
	}
	enc, err := s.Cipher.Encrypt(in.ContactPhone)
	if err != nil {
		return nil, err
	}
	u.Name, u.CommissionRate, u.ContactName = in.Name, in.CommissionRate, in.ContactName
	u.ContactPhoneEnc, u.SettleAccount = enc, in.SettleAccount
	if in.Status != nil {
		u.Status = *in.Status
	}
	if err := s.DB.Save(&u).Error; err != nil {
		return nil, err
	}
	return &u, nil
}

type AvailInput struct {
	AvailDate string `json:"avail_date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type PartnerInput struct {
	UnionID              *uint64      `json:"union_id"`
	Nickname             string       `json:"nickname" binding:"required"`
	Gender               string       `json:"gender"`
	City                 string       `json:"city" binding:"required"`
	Bio                  string       `json:"bio"`
	Highlight            string       `json:"highlight"`
	AvatarURL            string       `json:"avatar_url"`
	Gallery              []string     `json:"gallery"`
	Phone                string       `json:"phone"`
	Status               string       `json:"status"`
	HourlyPriceFen       int64        `json:"hourly_price_fen"`
	MinHours             float64      `json:"min_hours"`
	WeekendSurchargeRate int          `json:"weekend_surcharge_rate"`
	NightSurchargeRate   int          `json:"night_surcharge_rate"`
	ProfileText          string       `json:"profile_text"`
	Tags                 []string     `json:"tags"`
	Availabilities       []AvailInput `json:"availabilities"`
}

func uniqueStrings(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func (s *Services) syncPartnerExtras(tx *gorm.DB, partnerID uint64, tags []string, avails []AvailInput, replaceAvail bool) error {
	if err := tx.Where("partner_id = ?", partnerID).Delete(&models.PartnerTag{}).Error; err != nil {
		return err
	}
	for _, tag := range uniqueStrings(tags) {
		if err := tx.Create(&models.PartnerTag{PartnerID: partnerID, Tag: tag}).Error; err != nil {
			return err
		}
	}
	if !replaceAvail {
		return nil
	}
	if err := tx.Where("partner_id = ?", partnerID).Delete(&models.PartnerAvailability{}).Error; err != nil {
		return err
	}
	for _, a := range avails {
		d, err := time.ParseInLocation("2006-01-02", a.AvailDate, time.Local)
		if err != nil {
			continue
		}
		if err := tx.Create(&models.PartnerAvailability{
			PartnerID: partnerID, AvailDate: d, StartTime: a.StartTime, EndTime: a.EndTime,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Services) CreatePartner(in PartnerInput) (*models.Partner, error) {
	if in.Status == "" {
		in.Status = "draft"
	}
	if in.Gender == "" {
		in.Gender = "unknown"
	}
	if in.MinHours <= 0 {
		in.MinHours = 1
	}
	phoneEnc, err := s.Cipher.Encrypt(in.Phone)
	if err != nil {
		return nil, err
	}
	profile := in.ProfileText
	if profile == "" {
		profile = fmt.Sprintf("%s %s %s %s", in.Nickname, in.City, in.Highlight, in.Bio)
	}
	p := models.Partner{
		UnionID: in.UnionID, Nickname: in.Nickname, Gender: in.Gender, City: in.City,
		Bio: in.Bio, Highlight: in.Highlight, AvatarURL: in.AvatarURL,
		GalleryJSON: models.JSONStringSlice(in.Gallery), PhoneEnc: phoneEnc, Status: in.Status,
		HourlyPriceFen: in.HourlyPriceFen, MinHours: in.MinHours,
		WeekendSurchargeRate: in.WeekendSurchargeRate, NightSurchargeRate: in.NightSurchargeRate,
		ProfileText: profile,
	}
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&p).Error; err != nil {
			return err
		}
		return s.syncPartnerExtras(tx, p.ID, in.Tags, in.Availabilities, true)
	})
	if err != nil {
		return nil, err
	}
	return s.GetPartner(p.ID)
}

func (s *Services) UpdatePartner(id uint64, in PartnerInput) (*models.Partner, error) {
	var p models.Partner
	if err := s.DB.First(&p, id).Error; err != nil {
		return nil, errors.New("搭子不存在")
	}
	phoneEnc, err := s.Cipher.Encrypt(in.Phone)
	if err != nil {
		return nil, err
	}
	if in.Status == "" {
		in.Status = p.Status
	}
	if in.MinHours <= 0 {
		in.MinHours = p.MinHours
	}
	if in.Gender == "" {
		in.Gender = p.Gender
	}
	profile := in.ProfileText
	if profile == "" {
		profile = fmt.Sprintf("%s %s %s %s", in.Nickname, in.City, in.Highlight, in.Bio)
	}
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		p.UnionID, p.Nickname, p.Gender, p.City = in.UnionID, in.Nickname, in.Gender, in.City
		p.Bio, p.Highlight, p.AvatarURL = in.Bio, in.Highlight, in.AvatarURL
		p.GalleryJSON, p.PhoneEnc, p.Status = models.JSONStringSlice(in.Gallery), phoneEnc, in.Status
		p.HourlyPriceFen, p.MinHours = in.HourlyPriceFen, in.MinHours
		p.WeekendSurchargeRate, p.NightSurchargeRate = in.WeekendSurchargeRate, in.NightSurchargeRate
		p.ProfileText = profile
		if err := tx.Save(&p).Error; err != nil {
			return err
		}
		return s.syncPartnerExtras(tx, p.ID, in.Tags, in.Availabilities, len(in.Availabilities) > 0)
	})
	if err != nil {
		return nil, err
	}
	return s.GetPartner(p.ID)
}

func (s *Services) GetPartner(id uint64) (*models.Partner, error) {
	var p models.Partner
	if err := s.DB.Preload("Tags").Preload("Union").First(&p, id).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *Services) ListPartners(city, status string, page, pageSize int) ([]models.Partner, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	q := s.DB.Model(&models.Partner{})
	if city != "" {
		q = q.Where("city = ?", city)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []models.Partner
	err := q.Preload("Tags").Preload("Union").Order("id desc").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

type CandidateView struct {
	PartnerID uint64          `json:"partner_id"`
	RankNo    int             `json:"rank_no"`
	Score     float64         `json:"score"`
	Reason    string          `json:"reason"`
	Selected  bool            `json:"selected"`
	Partner   *models.Partner `json:"partner"`
}

type MatchChatResult struct {
	MatchRequestID uint64              `json:"match_request_id"`
	Parsed         *llm.ParsedIntent   `json:"parsed"`
	Candidates     []CandidateView     `json:"candidates"`
	AssistantMsg   string              `json:"assistant_message"`
}

func (s *Services) MatchChat(ctx context.Context, adminID uint64, matchID *uint64, text string) (*MatchChatResult, error) {
	var mr models.MatchRequest
	if matchID != nil && *matchID > 0 {
		if err := s.DB.First(&mr, *matchID).Error; err != nil {
			return nil, errors.New("匹配会话不存在")
		}
		mr.RawText = mr.RawText + "\n" + text
	} else {
		mr = models.MatchRequest{AdminUserID: adminID, RawText: text, Status: "draft"}
		if err := s.DB.Create(&mr).Error; err != nil {
			return nil, err
		}
	}
	_ = s.DB.Create(&models.MatchMessage{MatchRequestID: mr.ID, Role: "user", Content: text}).Error

	intent, err := s.LLM.ParseIntent(ctx, mr.RawText)
	if err != nil {
		return nil, err
	}
	b, _ := json.Marshal(intent)
	var jm models.JSONMap
	_ = json.Unmarshal(b, &jm)
	mr.ParsedJSON, mr.Status = jm, "matched"
	_ = s.DB.Save(&mr).Error

	candidates, err := s.recallPartners(intent, 50)
	if err != nil {
		return nil, err
	}
	briefs := make([]llm.PartnerBrief, 0, len(candidates))
	byID := map[uint64]models.Partner{}
	for _, p := range candidates {
		byID[p.ID] = p
		tags := make([]string, 0, len(p.Tags))
		for _, t := range p.Tags {
			tags = append(tags, t.Tag)
		}
		briefs = append(briefs, llm.PartnerBrief{
			ID: p.ID, Nickname: p.Nickname, City: p.City, Gender: p.Gender, Tags: tags,
			Highlight: p.Highlight, Bio: p.Bio, HourlyPriceFen: p.HourlyPriceFen, ProfileText: p.ProfileText,
		})
	}
	ranked, err := s.LLM.Rank(ctx, intent, briefs)
	if err != nil {
		return nil, err
	}
	_ = s.DB.Where("match_request_id = ?", mr.ID).Delete(&models.MatchCandidate{}).Error
	views := make([]CandidateView, 0, len(ranked))
	for i, r := range ranked {
		p := byID[r.PartnerID]
		pp := p
		mc := models.MatchCandidate{
			MatchRequestID: mr.ID, PartnerID: r.PartnerID, RankNo: i + 1,
			Score: r.Score, Reason: r.Reason, Selected: 1,
		}
		if err := s.DB.Create(&mc).Error; err != nil {
			return nil, err
		}
		views = append(views, CandidateView{
			PartnerID: r.PartnerID, RankNo: i + 1, Score: r.Score, Reason: r.Reason, Selected: true, Partner: &pp,
		})
	}
	assistant := fmt.Sprintf("已解析城市「%s」、兴趣 %v，推荐 %d 位搭子，请确认后生成小程序链接。", intent.City, intent.Interests, len(views))
	_ = s.DB.Create(&models.MatchMessage{MatchRequestID: mr.ID, Role: "assistant", Content: assistant}).Error
	return &MatchChatResult{MatchRequestID: mr.ID, Parsed: intent, Candidates: views, AssistantMsg: assistant}, nil
}

func (s *Services) recallPartners(intent *llm.ParsedIntent, limit int) ([]models.Partner, error) {
	q := s.DB.Model(&models.Partner{}).Where("status = ?", "online")
	if intent.City != "" {
		q = q.Where("city = ?", intent.City)
	}
	if intent.GenderPref == "male" || intent.GenderPref == "female" {
		q = q.Where("gender = ?", intent.GenderPref)
	}
	if intent.BudgetFenMax > 0 {
		q = q.Where("hourly_price_fen <= ?", intent.BudgetFenMax)
	}
	var list []models.Partner
	if err := q.Preload("Tags").Limit(limit).Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 && intent.City != "" {
		if err := s.DB.Where("status = ?", "online").Preload("Tags").Limit(limit).Find(&list).Error; err != nil {
			return nil, err
		}
	}
	if len(intent.Interests) == 0 {
		return list, nil
	}
	interest := map[string]struct{}{}
	for _, i := range intent.Interests {
		interest[i] = struct{}{}
	}
	type scored struct {
		p models.Partner
		n int
	}
	arr := make([]scored, 0, len(list))
	for _, p := range list {
		n := 0
		for _, t := range p.Tags {
			if _, ok := interest[t.Tag]; ok {
				n++
			}
		}
		arr = append(arr, scored{p: p, n: n})
	}
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[j].n > arr[i].n {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}
	out := make([]models.Partner, 0, len(arr))
	for _, a := range arr {
		out = append(out, a.p)
	}
	return out, nil
}

type ConfirmMatchInput struct {
	PartnerIDs []uint64 `json:"partner_ids"`
}

type ConfirmMatchResult struct {
	SID             string    `json:"sid"`
	MiniProgramURL  string    `json:"mini_program_url"`
	MiniProgramPath string    `json:"mini_program_path"`
	ExpiresAt       time.Time `json:"expires_at"`
}

func (s *Services) ConfirmMatch(adminID, matchID uint64, in ConfirmMatchInput) (*ConfirmMatchResult, error) {
	var mr models.MatchRequest
	if err := s.DB.First(&mr, matchID).Error; err != nil {
		return nil, errors.New("匹配会话不存在")
	}
	var cands []models.MatchCandidate
	if err := s.DB.Where("match_request_id = ?", matchID).Order("rank_no asc").Find(&cands).Error; err != nil {
		return nil, err
	}
	selected := map[uint64]models.MatchCandidate{}
	for _, c := range cands {
		selected[c.PartnerID] = c
	}
	ids := in.PartnerIDs
	if len(ids) == 0 {
		for _, c := range cands {
			if c.Selected == 1 {
				ids = append(ids, c.PartnerID)
			}
		}
	}
	if len(ids) == 0 {
		return nil, errors.New("请至少选择一位搭子")
	}
	if len(ids) > 5 {
		ids = ids[:5]
	}
	code, err := sid.New(10)
	if err != nil {
		return nil, err
	}
	expires := time.Now().Add(time.Duration(s.Cfg.MiniProgram.SessionTTLDays) * 24 * time.Hour)
	var session models.BrowseSession
	err = s.DB.Transaction(func(tx *gorm.DB) error {
		session = models.BrowseSession{Code: code, MatchRequestID: &matchID, ExpiresAt: expires, Status: "active"}
		if err := tx.Create(&session).Error; err != nil {
			return err
		}
		for i, pid := range ids {
			reason := ""
			if c, ok := selected[pid]; ok {
				reason = c.Reason
				_ = tx.Model(&models.MatchCandidate{}).Where("id = ?", c.ID).Update("selected", 1).Error
			}
			if err := tx.Create(&models.BrowseSessionPartner{
				BrowseSessionID: session.ID, PartnerID: pid, Reason: reason, SortNo: i + 1,
			}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&mr).Update("status", "confirmed").Error
	})
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("%s?sid=%s", s.Cfg.MiniProgram.PathPrefix, code)
	url := strings.TrimRight(s.Cfg.MiniProgram.LinkBase, "/") + path
	aid := adminID
	s.Audit(&aid, "match.confirm", "match_request", fmt.Sprintf("%d", matchID), models.JSONMap{"sid": code, "partner_ids": ids})
	return &ConfirmMatchResult{SID: code, MiniProgramURL: url, MiniProgramPath: path, ExpiresAt: expires}, nil
}

type BrowseCard struct {
	PartnerID      uint64   `json:"partner_id"`
	Nickname       string   `json:"nickname"`
	Gender         string   `json:"gender"`
	City           string   `json:"city"`
	Bio            string   `json:"bio"`
	Highlight      string   `json:"highlight"`
	AvatarURL      string   `json:"avatar_url"`
	Gallery        []string `json:"gallery"`
	Tags           []string `json:"tags"`
	HourlyPriceFen int64    `json:"hourly_price_fen"`
	MinHours       float64  `json:"min_hours"`
	Reason         string   `json:"reason"`
	SortNo         int      `json:"sort_no"`
}

func (s *Services) GetBrowseSession(code string, wxUserID uint64) ([]BrowseCard, *models.BrowseSession, error) {
	var session models.BrowseSession
	if err := s.DB.Where("code = ?", code).First(&session).Error; err != nil {
		return nil, nil, errors.New("推荐链接无效")
	}
	if session.Status != "active" || time.Now().After(session.ExpiresAt) {
		return nil, nil, errors.New("推荐链接已过期")
	}
	if session.WxUserID == nil && wxUserID > 0 {
		_ = s.DB.Model(&session).Update("wx_user_id", wxUserID).Error
		session.WxUserID = &wxUserID
	}
	var links []models.BrowseSessionPartner
	if err := s.DB.Preload("Partner.Tags").Where("browse_session_id = ?", session.ID).Order("sort_no asc").Find(&links).Error; err != nil {
		return nil, nil, err
	}
	cards := make([]BrowseCard, 0, len(links))
	for _, l := range links {
		if l.Partner == nil || l.Partner.Status != "online" {
			continue
		}
		tags := make([]string, 0, len(l.Partner.Tags))
		for _, t := range l.Partner.Tags {
			tags = append(tags, t.Tag)
		}
		gallery := []string(l.Partner.GalleryJSON)
		if gallery == nil {
			gallery = []string{}
		}
		cards = append(cards, BrowseCard{
			PartnerID: l.Partner.ID, Nickname: l.Partner.Nickname, Gender: l.Partner.Gender,
			City: l.Partner.City, Bio: l.Partner.Bio, Highlight: l.Partner.Highlight,
			AvatarURL: l.Partner.AvatarURL, Gallery: gallery, Tags: tags,
			HourlyPriceFen: l.Partner.HourlyPriceFen, MinHours: l.Partner.MinHours,
			Reason: l.Reason, SortNo: l.SortNo,
		})
	}
	return cards, &session, nil
}

type QuoteInput struct {
	PartnerID     uint64  `json:"partner_id" binding:"required"`
	ScheduleDate  string  `json:"schedule_date" binding:"required"`
	StartTime     string  `json:"start_time" binding:"required"`
	DurationHours float64 `json:"duration_hours" binding:"required"`
}

type QuoteResult struct {
	BaseAmountFen  int64   `json:"base_amount_fen"`
	SurchargeFen   int64   `json:"surcharge_fen"`
	TotalAmountFen int64   `json:"total_amount_fen"`
	DurationHours  float64 `json:"duration_hours"`
	HourlyPriceFen int64   `json:"hourly_price_fen"`
	Breakdown      string  `json:"breakdown"`
}

func (s *Services) Quote(in QuoteInput) (*QuoteResult, error) {
	var p models.Partner
	if err := s.DB.First(&p, in.PartnerID).Error; err != nil {
		return nil, errors.New("搭子不存在")
	}
	if p.Status != "online" {
		return nil, errors.New("搭子未上架")
	}
	if in.DurationHours < p.MinHours {
		return nil, fmt.Errorf("时长不能少于 %.1f 小时", p.MinHours)
	}
	date, err := time.ParseInLocation("2006-01-02", in.ScheduleDate, time.Local)
	if err != nil {
		return nil, errors.New("日期格式错误")
	}
	base := int64(math.Round(float64(p.HourlyPriceFen) * in.DurationHours))
	surcharge := int64(0)
	parts := []string{fmt.Sprintf("基础 %d分/时 × %.1f时 = %d分", p.HourlyPriceFen, in.DurationHours, base)}
	wd := date.Weekday()
	if (wd == time.Saturday || wd == time.Sunday) && p.WeekendSurchargeRate > 0 {
		extra := base * int64(p.WeekendSurchargeRate) / 10000
		surcharge += extra
		parts = append(parts, fmt.Sprintf("周末加价 %d", extra))
	}
	hour := 0
	fmt.Sscanf(in.StartTime, "%d", &hour)
	if (hour >= 21 || hour < 6) && p.NightSurchargeRate > 0 {
		extra := base * int64(p.NightSurchargeRate) / 10000
		surcharge += extra
		parts = append(parts, fmt.Sprintf("夜间加价 %d", extra))
	}
	return &QuoteResult{
		BaseAmountFen: base, SurchargeFen: surcharge, TotalAmountFen: base + surcharge,
		DurationHours: in.DurationHours, HourlyPriceFen: p.HourlyPriceFen, Breakdown: strings.Join(parts, "；"),
	}, nil
}

type CreateOrderInput struct {
	SID           string  `json:"sid" binding:"required"`
	PartnerID     uint64  `json:"partner_id" binding:"required"`
	ScheduleDate  string  `json:"schedule_date" binding:"required"`
	StartTime     string  `json:"start_time" binding:"required"`
	DurationHours float64 `json:"duration_hours" binding:"required"`
	ContactPhone  string  `json:"contact_phone" binding:"required"`
	Remark        string  `json:"remark"`
}

func (s *Services) CreateOrder(wxUserID uint64, in CreateOrderInput) (*models.Order, error) {
	var session models.BrowseSession
	if err := s.DB.Where("code = ?", in.SID).First(&session).Error; err != nil {
		return nil, errors.New("推荐会话无效")
	}
	if session.Status != "active" || time.Now().After(session.ExpiresAt) {
		return nil, errors.New("推荐会话已过期")
	}
	var link models.BrowseSessionPartner
	if err := s.DB.Where("browse_session_id = ? AND partner_id = ?", session.ID, in.PartnerID).First(&link).Error; err != nil {
		return nil, errors.New("该搭子不在本次推荐中")
	}
	startTime := normalizeTime(in.StartTime)
	quote, err := s.Quote(QuoteInput{
		PartnerID: in.PartnerID, ScheduleDate: in.ScheduleDate, StartTime: startTime, DurationHours: in.DurationHours,
	})
	if err != nil {
		return nil, err
	}
	var partner models.Partner
	if err := s.DB.First(&partner, in.PartnerID).Error; err != nil {
		return nil, err
	}
	date, _ := time.ParseInLocation("2006-01-02", in.ScheduleDate, time.Local)
	phoneEnc, err := s.Cipher.Encrypt(in.ContactPhone)
	if err != nil {
		return nil, err
	}
	var commissionRate int
	var commissionFen int64
	var unionID *uint64
	if partner.UnionID != nil {
		var u models.Union
		if err := s.DB.First(&u, *partner.UnionID).Error; err == nil && u.Status == 1 {
			unionID = partner.UnionID
			commissionRate = u.CommissionRate
			commissionFen = quote.TotalAmountFen * int64(commissionRate) / 10000
		}
	}
	order := models.Order{
		OrderNo: fmt.Sprintf("P%s%04d", time.Now().Format("20060102150405"), wxUserID%10000),
		BrowseSessionID: &session.ID, WxUserID: wxUserID, PartnerID: in.PartnerID, UnionID: unionID,
		ScheduleDate: date, StartTime: startTime, DurationHours: in.DurationHours,
		BaseAmountFen: quote.BaseAmountFen, SurchargeFen: quote.SurchargeFen, TotalAmountFen: quote.TotalAmountFen,
		CommissionRate: commissionRate, CommissionAmountFen: commissionFen,
		ContactPhoneEnc: phoneEnc, Remark: in.Remark, Status: "pending_pay",
	}
	if err := s.DB.Create(&order).Error; err != nil {
		return nil, err
	}
	return &order, nil
}

type PayParams struct {
	OrderID    uint64 `json:"order_id"`
	OrderNo    string `json:"order_no"`
	TimeStamp  string `json:"timeStamp"`
	NonceStr   string `json:"nonceStr"`
	Package    string `json:"package"`
	SignType   string `json:"signType"`
	PaySign    string `json:"paySign"`
	Mock       bool   `json:"mock"`
	OutTradeNo string `json:"out_trade_no"`
}

func (s *Services) CreatePayment(wxUserID, orderID uint64) (*PayParams, error) {
	var order models.Order
	if err := s.DB.First(&order, orderID).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	if order.WxUserID != wxUserID {
		return nil, errors.New("无权支付该订单")
	}
	if order.Status != "pending_pay" && order.Status != "created" {
		return nil, errors.New("订单状态不可支付")
	}
	outTradeNo := fmt.Sprintf("T%s%d", time.Now().Format("20060102150405"), order.ID)
	prepayID := "mock_prepay_" + outTradeNo
	_ = s.DB.Where("order_id = ?", order.ID).Delete(&models.Payment{}).Error
	pay := models.Payment{
		OrderID: order.ID, OutTradeNo: outTradeNo, AmountFen: order.TotalAmountFen,
		Status: "pending", PrepayID: prepayID,
	}
	if err := s.DB.Create(&pay).Error; err != nil {
		return nil, err
	}
	_ = s.DB.Model(&order).Update("status", "pending_pay").Error
	nonce := outTradeNo
	if len(nonce) > 8 {
		nonce = nonce[len(nonce)-8:]
	}
	return &PayParams{
		OrderID: order.ID, OrderNo: order.OrderNo, TimeStamp: fmt.Sprintf("%d", time.Now().Unix()),
		NonceStr: nonce, Package: "prepay_id=" + prepayID, SignType: "RSA", PaySign: "mock_sign",
		Mock: s.Cfg.WeChat.MockPay, OutTradeNo: outTradeNo,
	}, nil
}

func (s *Services) HandlePaySuccess(outTradeNo, transactionID string, notifyRaw models.JSONMap) error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var pay models.Payment
		if err := tx.Where("out_trade_no = ?", outTradeNo).First(&pay).Error; err != nil {
			return errors.New("支付单不存在")
		}
		if pay.Status == "paid" {
			return nil
		}
		var order models.Order
		if err := tx.First(&order, pay.OrderID).Error; err != nil {
			return err
		}
		if pay.AmountFen != order.TotalAmountFen {
			return errors.New("支付金额不匹配")
		}
		now := time.Now()
		pay.Status, pay.TransactionID, pay.NotifyRaw, pay.PaidAt = "paid", transactionID, notifyRaw, &now
		if err := tx.Save(&pay).Error; err != nil {
			return err
		}
		order.Status = "paid"
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		if order.UnionID != nil && order.CommissionAmountFen > 0 {
			var cnt int64
			_ = tx.Model(&models.CommissionLedger{}).Where("order_id = ?", order.ID).Count(&cnt)
			if cnt == 0 {
				if err := tx.Create(&models.CommissionLedger{
					OrderID: order.ID, UnionID: *order.UnionID, PartnerID: order.PartnerID,
					OrderAmountFen: order.TotalAmountFen, CommissionRate: order.CommissionRate,
					CommissionAmountFen: order.CommissionAmountFen, Status: "pending",
				}).Error; err != nil {
					return err
				}
			}
		}
		phone, _ := s.Cipher.Decrypt(order.ContactPhoneEnc)
		content := fmt.Sprintf("新订单 %s：预约 %s %s，时长 %.1f 小时，用户电话 %s，请尽快联系。",
			order.OrderNo, order.ScheduleDate.Format("2006-01-02"), order.StartTime, order.DurationHours, maskPhone(phone))
		if err := tx.Create(&models.PartnerNotification{
			OrderID: order.ID, PartnerID: order.PartnerID, Channel: "inbox", Content: content, Status: "sent",
		}).Error; err != nil {
			return err
		}
		order.Status, order.NotifiedAt = "notified", &now
		return tx.Save(&order).Error
	})
}

func (s *Services) MockPayNotify(outTradeNo string) error {
	return s.HandlePaySuccess(outTradeNo, "mock_txn_"+outTradeNo, models.JSONMap{"mock": true})
}

func (s *Services) WxLogin(code, nickname string) (*models.WxUser, string, error) {
	openid := code
	if s.Cfg.WeChat.MockLogin || code == "" {
		if code == "" {
			code = "mock_code"
		}
		openid = "mock_openid_" + code
	}
	var user models.WxUser
	err := s.DB.Where("openid = ?", openid).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		user = models.WxUser{OpenID: openid, Nickname: nickname}
		if user.Nickname == "" {
			user.Nickname = "微信用户"
		}
		if err := s.DB.Create(&user).Error; err != nil {
			return nil, "", err
		}
	} else if err != nil {
		return nil, "", err
	}
	token, err := middleware.SignUser(s.Cfg.JWT.UserSecret, s.Cfg.JWT.UserTTLHours, user.ID, user.OpenID)
	return &user, token, err
}

func (s *Services) ListOrders(status string, page, pageSize int) ([]models.Order, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	q := s.DB.Model(&models.Order{})
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	_ = q.Count(&total)
	var list []models.Order
	err := q.Preload("Partner").Order("id desc").Offset((page-1)*pageSize).Limit(pageSize).Find(&list).Error
	return list, total, err
}

func (s *Services) GetOrder(id uint64) (*models.Order, error) {
	var o models.Order
	if err := s.DB.Preload("Partner").First(&o, id).Error; err != nil {
		return nil, err
	}
	return &o, nil
}

// RefundOrder marks a paid/notified order as refunded, cancels its commission ledger, and audits.
func (s *Services) RefundOrder(orderID uint64) (*models.Order, error) {
	var order models.Order
	if err := s.DB.First(&order, orderID).Error; err != nil {
		return nil, errors.New("订单不存在")
	}
	switch order.Status {
	case "paid", "notified", "in_service", "pending_pay":
		// allowed
	case "refunded", "cancelled":
		return &order, nil
	default:
		return nil, errors.New("当前状态不可退款")
	}
	err := s.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		order.Status = "refunded"
		order.UpdatedAt = now
		if err := tx.Save(&order).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.CommissionLedger{}).
			Where("order_id = ? AND status = ?", order.ID, "pending").
			Update("status", "cancelled").Error; err != nil {
			return err
		}
		_ = tx.Model(&models.Payment{}).
			Where("order_id = ? AND status = ?", order.ID, "paid").
			Update("status", "refunded").Error
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.GetOrder(orderID)
}

func (s *Services) CommissionReport() ([]map[string]any, error) {
	type row struct {
		UnionID    uint64
		UnionName  string
		OrderCount int64
		AmountFen  int64
		PendingFen int64
	}
	var rows []row
	err := s.DB.Raw(`
		SELECT cl.union_id as union_id, u.name as union_name,
		       COUNT(*) as order_count,
		       COALESCE(SUM(cl.commission_amount_fen),0) as amount_fen,
		       COALESCE(SUM(CASE WHEN cl.status='pending' THEN cl.commission_amount_fen ELSE 0 END),0) as pending_fen
		FROM commission_ledgers cl
		JOIN unions u ON u.id = cl.union_id
		GROUP BY cl.union_id, u.name
		ORDER BY amount_fen DESC`).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]any{
			"union_id": r.UnionID, "union_name": r.UnionName,
			"order_count": r.OrderCount, "amount_fen": r.AmountFen, "pending_fen": r.PendingFen,
		})
	}
	return out, nil
}

func (s *Services) ListNotifications(partnerID uint64) ([]models.PartnerNotification, error) {
	var list []models.PartnerNotification
	q := s.DB.Order("id desc").Limit(100)
	if partnerID > 0 {
		q = q.Where("partner_id = ?", partnerID)
	}
	return list, q.Find(&list).Error
}

func maskPhone(phone string) string {
	r := []rune(phone)
	if len(r) < 7 {
		return phone
	}
	return string(r[:3]) + "****" + string(r[len(r)-4:])
}

// normalizeTime ensures MySQL TIME columns accept HH:MM or HH:MM:SS.
func normalizeTime(t string) string {
	t = strings.TrimSpace(t)
	switch len(t) {
	case 5: // HH:MM
		return t + ":00"
	case 8: // HH:MM:SS
		return t
	default:
		if hour, min, ok := parseHM(t); ok {
			return fmt.Sprintf("%02d:%02d:00", hour, min)
		}
		return t
	}
}

func parseHM(t string) (int, int, bool) {
	var h, m int
	if _, err := fmt.Sscanf(t, "%d:%d", &h, &m); err != nil {
		return 0, 0, false
	}
	return h, m, true
}
