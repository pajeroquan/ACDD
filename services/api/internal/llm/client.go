package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	APIKey  string
	Model   string
	Mock    bool
	HTTP    *http.Client
}

type ParsedIntent struct {
	City          string   `json:"city"`
	Interests     []string `json:"interests"`
	GenderPref    string   `json:"gender_pref"`
	BudgetFenMax  int64    `json:"budget_fen_max"`
	DurationHours float64  `json:"duration_hours"`
	Personality   string   `json:"personality"`
	DateHint      string   `json:"date_hint"`
	RawSummary    string   `json:"raw_summary"`
}

type RankItem struct {
	PartnerID uint64  `json:"partner_id"`
	Score     float64 `json:"score"`
	Reason    string  `json:"reason"`
}

type PartnerBrief struct {
	ID            uint64   `json:"id"`
	Nickname      string   `json:"nickname"`
	City          string   `json:"city"`
	Gender        string   `json:"gender"`
	Tags          []string `json:"tags"`
	Highlight     string   `json:"highlight"`
	Bio           string   `json:"bio"`
	HourlyPriceFen int64   `json:"hourly_price_fen"`
	ProfileText   string   `json:"profile_text"`
}

func New(baseURL, apiKey, model string, mock bool) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		APIKey:  apiKey,
		Model:   model,
		Mock:    mock,
		HTTP:    &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *Client) ParseIntent(ctx context.Context, text string) (*ParsedIntent, error) {
	if c.Mock || c.APIKey == "" {
		return mockParse(text), nil
	}
	sys := `你是本地生活「兴趣搭子」匹配助手。将用户诉求解析为 JSON，字段：
city, interests([]string), gender_pref(male/female/any), budget_fen_max(int, 单位分, 未知为0),
duration_hours(float), personality(string), date_hint(string), raw_summary(string)。
只输出 JSON。`
	out, err := c.chat(ctx, sys, text)
	if err != nil {
		return mockParse(text), nil
	}
	var p ParsedIntent
	if err := json.Unmarshal([]byte(extractJSON(out)), &p); err != nil {
		return mockParse(text), nil
	}
	if p.GenderPref == "" {
		p.GenderPref = "any"
	}
	if p.DurationHours == 0 {
		p.DurationHours = 2
	}
	return &p, nil
}

func (c *Client) Rank(ctx context.Context, intent *ParsedIntent, candidates []PartnerBrief) ([]RankItem, error) {
	if len(candidates) == 0 {
		return nil, nil
	}
	if c.Mock || c.APIKey == "" {
		return mockRank(intent, candidates), nil
	}
	payload, _ := json.Marshal(map[string]any{"intent": intent, "candidates": candidates})
	sys := `根据用户诉求对候选兴趣搭子排序，输出 JSON 数组，每项：partner_id, score(0-100), reason(一句话中文推荐理由)。最多返回 5 个。只输出 JSON。`
	out, err := c.chat(ctx, sys, string(payload))
	if err != nil {
		return mockRank(intent, candidates), nil
	}
	var items []RankItem
	if err := json.Unmarshal([]byte(extractJSON(out)), &items); err != nil {
		return mockRank(intent, candidates), nil
	}
	if len(items) > 5 {
		items = items[:5]
	}
	return items, nil
}

func (c *Client) chat(ctx context.Context, system, user string) (string, error) {
	body := map[string]any{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": user},
		},
		"temperature": 0.2,
	}
	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.APIKey)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("llm status %d: %s", resp.StatusCode, string(raw))
	}
	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty llm choices")
	}
	return parsed.Choices[0].Message.Content, nil
}

func mockParse(text string) *ParsedIntent {
	p := &ParsedIntent{
		City:          "上海",
		Interests:     []string{},
		GenderPref:    "any",
		BudgetFenMax:  0,
		DurationHours: 2,
		Personality:   "",
		DateHint:      "",
		RawSummary:    text,
	}
	cities := []string{"上海", "北京", "深圳", "广州", "杭州", "成都", "南京", "武汉", "重庆", "苏州"}
	for _, city := range cities {
		if strings.Contains(text, city) {
			p.City = city
			break
		}
	}
	interestMap := map[string]string{
		"徒步": "徒步", "爬山": "徒步", "咖啡": "咖啡", "桌游": "桌游",
		"羽毛球": "羽毛球", "网球": "网球", "摄影": "摄影", "美食": "美食",
		"电影": "电影", "剧本杀": "剧本杀", "骑行": "骑行", "跑步": "跑步",
		"展览": "展览", "密室": "密室", "滑雪": "滑雪", "游泳": "游泳",
	}
	for k, v := range interestMap {
		if strings.Contains(text, k) {
			p.Interests = append(p.Interests, v)
		}
	}
	if strings.Contains(text, "女生") || strings.Contains(text, "女搭子") {
		p.GenderPref = "female"
	}
	if strings.Contains(text, "男生") || strings.Contains(text, "男搭子") {
		p.GenderPref = "male"
	}
	re := regexp.MustCompile(`(\d+)\s*小时`)
	if m := re.FindStringSubmatch(text); len(m) == 2 {
		fmt.Sscanf(m[1], "%f", &p.DurationHours)
	}
	reBudget := regexp.MustCompile(`(\d+)\s*元`)
	if m := reBudget.FindStringSubmatch(text); len(m) == 2 {
		var yuan int64
		fmt.Sscanf(m[1], "%d", &yuan)
		p.BudgetFenMax = yuan * 100
	}
	return p
}

func mockRank(intent *ParsedIntent, candidates []PartnerBrief) []RankItem {
	type scored struct {
		item RankItem
	}
	var list []scored
	interestSet := map[string]struct{}{}
	for _, i := range intent.Interests {
		interestSet[i] = struct{}{}
	}
	for _, c := range candidates {
		score := 40.0
		matched := []string{}
		for _, t := range c.Tags {
			if _, ok := interestSet[t]; ok {
				score += 12
				matched = append(matched, t)
			}
		}
		if intent.City != "" && c.City == intent.City {
			score += 15
		}
		if intent.GenderPref == "male" && c.Gender == "male" {
			score += 8
		}
		if intent.GenderPref == "female" && c.Gender == "female" {
			score += 8
		}
		if intent.BudgetFenMax > 0 && c.HourlyPriceFen <= intent.BudgetFenMax {
			score += 10
		}
		reason := c.Highlight
		if reason == "" {
			reason = c.Bio
		}
		if len(matched) > 0 {
			reason = fmt.Sprintf("兴趣契合：%s。%s", strings.Join(matched, "、"), reason)
		} else if intent.City != "" && c.City == intent.City {
			reason = fmt.Sprintf("同城%s推荐。%s", c.City, reason)
		}
		if len([]rune(reason)) > 80 {
			reason = string([]rune(reason)[:80])
		}
		list = append(list, scored{RankItem{PartnerID: c.ID, Score: score, Reason: reason}})
	}
	// simple sort
	for i := 0; i < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			if list[j].item.Score > list[i].item.Score {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
	n := 5
	if len(list) < n {
		n = len(list)
	}
	out := make([]RankItem, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, list[i].item)
	}
	return out
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```json")
		s = strings.TrimPrefix(s, "```")
		s = strings.TrimSuffix(s, "```")
		s = strings.TrimSpace(s)
	}
	startObj := strings.Index(s, "{")
	startArr := strings.Index(s, "[")
	start := startObj
	if startArr >= 0 && (start < 0 || startArr < start) {
		start = startArr
	}
	if start < 0 {
		return s
	}
	endObj := strings.LastIndex(s, "}")
	endArr := strings.LastIndex(s, "]")
	end := endObj
	if endArr > end {
		end = endArr
	}
	if end < start {
		return s
	}
	return s[start : end+1]
}
