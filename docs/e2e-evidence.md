# E2E Evidence 2026-09-05T14:09:31Z

## Login OK
## Match
{
    "code": 0,
    "message": "ok",
    "data": {
        "match_request_id": 4,
        "parsed": {
            "city": "\u4e0a\u6d77",
            "interests": [
                "\u5496\u5561"
            ],
            "gender_pref": "female",
            "budget_fen_max": 0,
            "duration_hours": 2,
            "personality": "",
            "date_hint": "",
            "raw_summary": "\u4e0a\u6d77\u627e\u5973\u642d\u5b50\u559d\u5496\u55612\u5c0f\u65f6"
        },
        "candidates": [
            {
                "partner_id": 1,
                "rank_no": 1,
                "score": 75,
                "reason": "\u5174\u8da3\u5951\u5408\uff1a\u5496\u5561\u3002\u540c\u57ce\u5496\u5561\u6f2b\u6b65\u642d\u5b50",
                "selected": true,
                "partner": {
                    "id": 1,
                    "union_id": 1,
                    "nickname": "\u963f\u79be",
                    "gender": "female",
                    "city": "\u4e0a\u6d77",
                    "bio": "\u5496\u5561\u4e0e\u80f6\u7247\u6444\u5f71\u7231\u597d\u8005\uff0c\u64c5\u957f\u5e26\u4f60\u53d1\u73b0\u57ce\u5e02\u89d2\u843d\u3002",
                    "highlight": "\u540c\u57ce\u5496\u5561\u6f2b\u6b65\u642d\u5b50",
                    "avatar_url": "https://picsum.photos/seed/ahe/800/1000",
                    "gallery": [
                        "https://picsum.photos/seed/ahe1/800/1000"
                    ],
                    "status": "online",
                    "hourly_price_fen": 12800,
                    "min_hours": 2,
                    "weekend_surcharge_rate": 1000,
                    "night_surcharge_rate": 1500,
                    "profile_text": "\u4e0a\u6d77 \u5973 \u5496\u5561 \u6444\u5f71 \u6f2b\u6b65",
                    "created_at": "2026-09-05T14:03:27.946Z",
                    "updated_at": "2026-09-05T14:03:27.946Z",
                    "tags": [
                        {
                            "id": 1,
                            "partner_id": 1,
                            "tag": "\u5496\u5561"
                        },
                        {
                            "id": 2,
                            "partner_id": 1,
                            "tag": "\u6444\u5f71"
                        },
                        {
                            "id": 3,
                            "partner_id": 1,
                            "tag": "\u6f2b\u6b65"
                        }
## Confirm
{
    "code": 0,
    "message": "ok",
    "data": {
        "sid": "q7eywgrdby",
        "mini_program_url": "https://example.com/mp/pages/discover/index?sid=q7eywgrdby",
        "mini_program_path": "/pages/discover/index?sid=q7eywgrdby",
        "expires_at": "2026-09-12T14:09:31.189690963Z"
    }
}
## Browse cards
[('阿禾', 12800, '兴趣契合：咖啡。同城咖啡漫步搭子'), ('Momo', 8800, '同城上海推荐。桌游局开黑搭子'), ('小满', 10800, '同城上海推荐。美食探店搭子')]
## Order after pay
{'order_no': 'P202609051409310004', 'status': 'notified', 'total_amount_fen': 25600, 'commission_rate': 1500, 'commission_amount_fen': 3840, 'start_time': '16:00:00'}
## Commission report
{
    "code": 0,
    "message": "ok",
    "data": [
        {
            "amount_fen": 7680,
            "order_count": 2,
            "pending_fen": 7680,
            "union_id": 1,
            "union_name": "\u6d66\u4e1c\u5174\u8da3\u5de5\u4f1a"
        }
    ]
}
## Notifications
新订单 P202609051409310004：预约 2026-09-07 16:00:00，时长 2.0 小时，用户电话 137****1111，请尽快联系。
