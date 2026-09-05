package db

import (
	"fmt"
	"time"

	"github.com/local-life/partner/services/api/internal/config"
	"github.com/local-life/partner/services/api/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Open(dsn string, mode string) (*gorm.DB, error) {
	logLevel := logger.Warn
	if mode == "debug" {
		logLevel = logger.Info
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logLevel)})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.AdminUser{},
		&models.Union{},
		&models.Partner{},
		&models.PartnerTag{},
		&models.PartnerAvailability{},
		&models.WxUser{},
		&models.MatchRequest{},
		&models.MatchMessage{},
		&models.MatchCandidate{},
		&models.BrowseSession{},
		&models.BrowseSessionPartner{},
		&models.Order{},
		&models.Payment{},
		&models.CommissionLedger{},
		&models.AuditLog{},
		&models.PartnerNotification{},
	)
}

func SeedAdmin(db *gorm.DB, cfg *config.Config) error {
	var count int64
	if err := db.Model(&models.AdminUser{}).Where("username = ?", cfg.Seed.AdminUsername).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(cfg.Seed.AdminPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return db.Create(&models.AdminUser{
		Username:     cfg.Seed.AdminUsername,
		PasswordHash: string(hash),
		DisplayName:  "管理员",
		Role:         "admin",
		Status:       1,
	}).Error
}

func SeedDemoData(db *gorm.DB, cipher interface{ Encrypt(string) (string, error) }) error {
	var n int64
	if err := db.Model(&models.Partner{}).Count(&n).Error; err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	phone, _ := cipher.Encrypt("13800001111")
	u := models.Union{Name: "浦东兴趣工会", CommissionRate: 1500, ContactName: "王经理", ContactPhoneEnc: phone, SettleAccount: "union-pd-001", Status: 1}
	if err := db.Create(&u).Error; err != nil {
		return err
	}
	uid := u.ID
	partners := []struct {
		p    models.Partner
		tags []string
	}{
		{models.Partner{UnionID: &uid, Nickname: "阿禾", Gender: "female", City: "上海", Bio: "咖啡与胶片摄影爱好者，擅长带你发现城市角落。", Highlight: "同城咖啡漫步搭子", AvatarURL: "https://picsum.photos/seed/ahe/800/1000", GalleryJSON: models.JSONStringSlice{"https://picsum.photos/seed/ahe1/800/1000"}, PhoneEnc: phone, Status: "online", HourlyPriceFen: 12800, MinHours: 2, WeekendSurchargeRate: 1000, NightSurchargeRate: 1500, ProfileText: "上海 女 咖啡 摄影 漫步"}, []string{"咖啡", "摄影", "漫步"}},
		{models.Partner{UnionID: &uid, Nickname: "小林", Gender: "male", City: "上海", Bio: "周末徒步向导，熟悉余山与崇明路线。", Highlight: "徒步爬山好搭子", AvatarURL: "https://picsum.photos/seed/xiaolin/800/1000", GalleryJSON: models.JSONStringSlice{"https://picsum.photos/seed/xl1/800/1000"}, PhoneEnc: phone, Status: "online", HourlyPriceFen: 9800, MinHours: 3, WeekendSurchargeRate: 800, NightSurchargeRate: 0, ProfileText: "上海 男 徒步 爬山 户外"}, []string{"徒步", "爬山", "户外"}},
		{models.Partner{Nickname: "Momo", Gender: "female", City: "上海", Bio: "桌游社交达人，德式美式都玩。", Highlight: "桌游局开黑搭子", AvatarURL: "https://picsum.photos/seed/momo/800/1000", GalleryJSON: models.JSONStringSlice{"https://picsum.photos/seed/momo1/800/1000"}, PhoneEnc: phone, Status: "online", HourlyPriceFen: 8800, MinHours: 2, WeekendSurchargeRate: 500, NightSurchargeRate: 1000, ProfileText: "上海 女 桌游 剧本杀"}, []string{"桌游", "剧本杀"}},
		{models.Partner{UnionID: &uid, Nickname: "Ace", Gender: "male", City: "上海", Bio: "羽毛球陪练，业余赛事经验丰富。", Highlight: "羽毛球陪练搭子", AvatarURL: "https://picsum.photos/seed/ace/800/1000", GalleryJSON: models.JSONStringSlice{"https://picsum.photos/seed/ace1/800/1000"}, PhoneEnc: phone, Status: "online", HourlyPriceFen: 15000, MinHours: 1, WeekendSurchargeRate: 1000, NightSurchargeRate: 500, ProfileText: "上海 男 羽毛球 运动"}, []string{"羽毛球", "运动"}},
		{models.Partner{Nickname: "小满", Gender: "female", City: "上海", Bio: "美食探店，擅长安利小馆子。", Highlight: "美食探店搭子", AvatarURL: "https://picsum.photos/seed/xiaoman/800/1000", GalleryJSON: models.JSONStringSlice{"https://picsum.photos/seed/xm1/800/1000"}, PhoneEnc: phone, Status: "online", HourlyPriceFen: 10800, MinHours: 2, WeekendSurchargeRate: 1000, NightSurchargeRate: 1200, ProfileText: "上海 女 美食 探店"}, []string{"美食", "探店"}},
		{models.Partner{UnionID: &uid, Nickname: "老K", Gender: "male", City: "北京", Bio: "北京城区骑行向导。", Highlight: "骑行搭子", AvatarURL: "https://picsum.photos/seed/laok/800/1000", GalleryJSON: models.JSONStringSlice{"https://picsum.photos/seed/lk1/800/1000"}, PhoneEnc: phone, Status: "online", HourlyPriceFen: 9000, MinHours: 2, WeekendSurchargeRate: 800, NightSurchargeRate: 0, ProfileText: "北京 男 骑行 户外"}, []string{"骑行", "户外"}},
	}
	for _, item := range partners {
		p := item.p
		if err := db.Create(&p).Error; err != nil {
			return err
		}
		for _, t := range item.tags {
			if err := db.Create(&models.PartnerTag{PartnerID: p.ID, Tag: t}).Error; err != nil {
				return err
			}
		}
		d := time.Now().AddDate(0, 0, 2)
		_ = db.Create(&models.PartnerAvailability{PartnerID: p.ID, AvailDate: d, StartTime: "10:00:00", EndTime: "22:00:00"}).Error
	}
	return nil
}
