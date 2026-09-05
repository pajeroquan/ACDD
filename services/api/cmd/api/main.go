package main

import (
	"flag"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/local-life/partner/services/api/internal/config"
	"github.com/local-life/partner/services/api/internal/db"
	"github.com/local-life/partner/services/api/internal/handler"
	"github.com/local-life/partner/services/api/internal/llm"
	"github.com/local-life/partner/services/api/internal/middleware"
	"github.com/local-life/partner/services/api/internal/pkg/crypto"
	"github.com/local-life/partner/services/api/internal/service"
)

func main() {
	cfgPath := flag.String("config", "configs/config.yaml", "config path")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	gin.SetMode(cfg.Server.Mode)

	gdb, err := db.Open(cfg.MySQL.DSN, cfg.Server.Mode)
	if err != nil {
		log.Fatalf("mysql: %v", err)
	}
	if err := db.AutoMigrate(gdb); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	if err := db.SeedAdmin(gdb, cfg); err != nil {
		log.Fatalf("seed admin: %v", err)
	}

	cipher, err := crypto.NewAESCipher(cfg.Crypto.AESKey)
	if err != nil {
		log.Fatalf("crypto: %v", err)
	}
	if err := db.SeedDemoData(gdb, cipher); err != nil {
		log.Fatalf("seed demo: %v", err)
	}

	llmClient := llm.New(cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model, cfg.LLM.Mock)
	svc := service.New(gdb, cfg, cipher, llmClient)
	h := handler.New(svc)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())
	r.GET("/healthz", h.Health)

	admin := r.Group("/admin")
	{
		admin.POST("/login", h.AdminLogin)
		auth := admin.Group("")
		auth.Use(middleware.AdminAuth(cfg.JWT.AdminSecret))
		auth.GET("/unions", h.ListUnions)
		auth.POST("/unions", h.CreateUnion)
		auth.PATCH("/unions/:id", middleware.RequireRoles("admin", "finance", "ops"), h.UpdateUnion)
		auth.GET("/partners", h.ListPartners)
		auth.GET("/partners/:id", h.GetPartner)
		auth.POST("/partners", middleware.RequireRoles("admin", "ops"), h.CreatePartner)
		auth.PUT("/partners/:id", middleware.RequireRoles("admin", "ops"), h.UpdatePartner)
		auth.POST("/match/chat", h.MatchChat)
		auth.POST("/match/:id/confirm", h.MatchConfirm)
		auth.GET("/orders", h.ListOrders)
		auth.GET("/orders/:id", h.GetOrder)
		auth.POST("/orders/:id/refund", middleware.RequireRoles("admin", "finance"), h.RefundOrder)
		auth.GET("/commission/report", h.CommissionReport)
		auth.GET("/notifications", h.ListNotifications)
	}

	api := r.Group("/api")
	{
		api.POST("/wx/login", h.WxLogin)
		api.POST("/pay/notify", h.PayNotify)
		api.POST("/pay/mock-notify", h.MockPayNotify)
		user := api.Group("")
		user.Use(middleware.UserAuth(cfg.JWT.UserSecret))
		user.GET("/browse/:sid", h.BrowseSession)
		user.POST("/orders/quote", h.Quote)
		user.POST("/orders", h.CreateOrder)
		user.POST("/orders/:id/pay", h.PayOrder)
	}

	addr := cfg.Server.Addr
	if v := os.Getenv("PORT"); v != "" {
		addr = ":" + v
	}
	log.Printf("api listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}
