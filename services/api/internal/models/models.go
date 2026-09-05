package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type JSONMap map[string]any

func (JSONMap) GormDataType() string { return "json" }

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONMap) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported type %T for JSONMap", value)
	}
	if len(b) == 0 {
		*j = JSONMap{}
		return nil
	}
	return json.Unmarshal(b, j)
}

type JSONStringSlice []string

func (JSONStringSlice) GormDataType() string { return "json" }

func (j JSONStringSlice) Value() (driver.Value, error) {
	if j == nil {
		return "[]", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (j *JSONStringSlice) Scan(value any) error {
	if value == nil {
		*j = []string{}
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return fmt.Errorf("unsupported type %T for JSONStringSlice", value)
	}
	if len(b) == 0 {
		*j = []string{}
		return nil
	}
	return json.Unmarshal(b, j)
}

type AdminUser struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	Role         string    `json:"role"`
	Status       int       `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (AdminUser) TableName() string { return "admin_users" }

type Union struct {
	ID               uint64    `gorm:"primaryKey" json:"id"`
	Name             string    `json:"name"`
	CommissionRate   int       `json:"commission_rate"`
	ContactName      string    `json:"contact_name"`
	ContactPhoneEnc  string    `json:"-"`
	SettleAccount    string    `json:"settle_account"`
	Status           int       `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (Union) TableName() string { return "unions" }

type Partner struct {
	ID                   uint64          `gorm:"primaryKey" json:"id"`
	UnionID              *uint64         `json:"union_id"`
	Nickname             string          `json:"nickname"`
	Gender               string          `json:"gender"`
	City                 string          `json:"city"`
	Bio                  string          `json:"bio"`
	Highlight            string          `json:"highlight"`
	AvatarURL            string          `json:"avatar_url"`
	GalleryJSON          JSONStringSlice `gorm:"column:gallery_json" json:"gallery"`
	PhoneEnc             string          `json:"-"`
	Status               string          `json:"status"`
	HourlyPriceFen       int64           `json:"hourly_price_fen"`
	MinHours             float64         `json:"min_hours"`
	WeekendSurchargeRate int             `json:"weekend_surcharge_rate"`
	NightSurchargeRate   int             `json:"night_surcharge_rate"`
	ProfileText          string          `json:"profile_text"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	Tags                 []PartnerTag    `gorm:"foreignKey:PartnerID" json:"tags,omitempty"`
	Union                *Union          `gorm:"foreignKey:UnionID" json:"union,omitempty"`
}

func (Partner) TableName() string { return "partners" }

type PartnerTag struct {
	ID        uint64 `gorm:"primaryKey" json:"id"`
	PartnerID uint64 `json:"partner_id"`
	Tag       string `json:"tag"`
}

func (PartnerTag) TableName() string { return "partner_tags" }

type PartnerAvailability struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	PartnerID uint64    `json:"partner_id"`
	AvailDate time.Time `gorm:"type:date" json:"avail_date"`
	StartTime string    `gorm:"type:varchar(16)" json:"start_time"`
	EndTime   string    `gorm:"type:varchar(16)" json:"end_time"`
}

func (PartnerAvailability) TableName() string { return "partner_availabilities" }

type WxUser struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	OpenID    string    `gorm:"column:openid" json:"openid"`
	UnionID   string    `gorm:"column:unionid" json:"unionid"`
	Nickname  string    `json:"nickname"`
	AvatarURL string    `json:"avatar_url"`
	PhoneEnc  string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (WxUser) TableName() string { return "wx_users" }

type MatchRequest struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	AdminUserID uint64    `json:"admin_user_id"`
	RawText     string    `json:"raw_text"`
	ParsedJSON  JSONMap   `gorm:"column:parsed_json" json:"parsed"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (MatchRequest) TableName() string { return "match_requests" }

type MatchMessage struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	MatchRequestID uint64    `json:"match_request_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

func (MatchMessage) TableName() string { return "match_messages" }

type MatchCandidate struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	MatchRequestID uint64    `json:"match_request_id"`
	PartnerID      uint64    `json:"partner_id"`
	RankNo         int       `json:"rank_no"`
	Score          float64   `json:"score"`
	Reason         string    `json:"reason"`
	Selected       int       `json:"selected"`
	CreatedAt      time.Time `json:"created_at"`
	Partner        *Partner  `gorm:"foreignKey:PartnerID" json:"partner,omitempty"`
}

func (MatchCandidate) TableName() string { return "match_candidates" }

type BrowseSession struct {
	ID             uint64    `gorm:"primaryKey" json:"id"`
	Code           string    `json:"code"`
	MatchRequestID *uint64   `json:"match_request_id"`
	WxUserID       *uint64   `json:"wx_user_id"`
	ExpiresAt      time.Time `json:"expires_at"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (BrowseSession) TableName() string { return "browse_sessions" }

type BrowseSessionPartner struct {
	ID              uint64   `gorm:"primaryKey" json:"id"`
	BrowseSessionID uint64   `json:"browse_session_id"`
	PartnerID       uint64   `json:"partner_id"`
	Reason          string   `json:"reason"`
	SortNo          int      `json:"sort_no"`
	Partner         *Partner `gorm:"foreignKey:PartnerID" json:"partner,omitempty"`
}

func (BrowseSessionPartner) TableName() string { return "browse_session_partners" }

type Order struct {
	ID                  uint64     `gorm:"primaryKey" json:"id"`
	OrderNo             string     `json:"order_no"`
	BrowseSessionID     *uint64    `json:"browse_session_id"`
	WxUserID            uint64     `json:"wx_user_id"`
	PartnerID           uint64     `json:"partner_id"`
	UnionID             *uint64    `json:"union_id"`
	ScheduleDate        time.Time  `gorm:"type:date" json:"schedule_date"`
	StartTime           string     `gorm:"type:varchar(16)" json:"start_time"`
	DurationHours       float64    `json:"duration_hours"`
	BaseAmountFen       int64      `json:"base_amount_fen"`
	SurchargeFen        int64      `json:"surcharge_fen"`
	TotalAmountFen      int64      `json:"total_amount_fen"`
	CommissionRate      int        `json:"commission_rate"`
	CommissionAmountFen int64      `json:"commission_amount_fen"`
	ContactPhoneEnc     string     `json:"-"`
	Remark              string     `json:"remark"`
	Status              string     `json:"status"`
	NotifiedAt          *time.Time `json:"notified_at"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	Partner             *Partner   `gorm:"foreignKey:PartnerID" json:"partner,omitempty"`
}

func (Order) TableName() string { return "orders" }

type Payment struct {
	ID            uint64     `gorm:"primaryKey" json:"id"`
	OrderID       uint64     `json:"order_id"`
	OutTradeNo    string     `json:"out_trade_no"`
	TransactionID string     `json:"transaction_id"`
	AmountFen     int64      `json:"amount_fen"`
	Status        string     `json:"status"`
	PrepayID      string     `json:"prepay_id"`
	NotifyRaw     JSONMap    `gorm:"column:notify_raw" json:"notify_raw,omitempty"`
	PaidAt        *time.Time `json:"paid_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (Payment) TableName() string { return "payments" }

type CommissionLedger struct {
	ID                  uint64    `gorm:"primaryKey" json:"id"`
	OrderID             uint64    `json:"order_id"`
	UnionID             uint64    `json:"union_id"`
	PartnerID           uint64    `json:"partner_id"`
	OrderAmountFen      int64     `json:"order_amount_fen"`
	CommissionRate      int       `json:"commission_rate"`
	CommissionAmountFen int64     `json:"commission_amount_fen"`
	Status              string    `json:"status"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

func (CommissionLedger) TableName() string { return "commission_ledgers" }

type AuditLog struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	AdminUserID *uint64   `json:"admin_user_id"`
	Action      string    `json:"action"`
	TargetType  string    `json:"target_type"`
	TargetID    string    `json:"target_id"`
	DetailJSON  JSONMap   `gorm:"column:detail_json" json:"detail"`
	CreatedAt   time.Time `json:"created_at"`
}

func (AuditLog) TableName() string { return "audit_logs" }

type PartnerNotification struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	OrderID   uint64    `json:"order_id"`
	PartnerID uint64    `json:"partner_id"`
	Channel   string    `json:"channel"`
	Content   string    `json:"content"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func (PartnerNotification) TableName() string { return "partner_notifications" }
