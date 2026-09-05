#!/usr/bin/env bash
set -euo pipefail
BASE="${BASE_URL:-http://127.0.0.1:8080}"

echo "== health =="
curl -sf "$BASE/healthz" | tee /tmp/health.json
echo

echo "== admin login =="
LOGIN=$(curl -sf -X POST "$BASE/admin/login" -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}')
TOKEN=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['data']['token'])" "$LOGIN")

echo "== match =="
MATCH=$(curl -sf -X POST "$BASE/admin/match/chat" -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"text":"上海找女搭子周末喝咖啡拍照，预算200，2小时"}')
MID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['data']['match_request_id'])" "$MATCH")
PIDS=$(python3 -c "import json,sys; d=json.loads(sys.argv[1])['data']; print(__import__('json').dumps([c['partner_id'] for c in d['candidates']]))" "$MATCH")

echo "== confirm =="
CONF=$(curl -sf -X POST "$BASE/admin/match/$MID/confirm" -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' -d "{\"partner_ids\":$PIDS}")
SID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['data']['sid'])" "$CONF")
echo "sid=$SID"

echo "== wx login + browse =="
WX=$(curl -sf -X POST "$BASE/api/wx/login" -H 'Content-Type: application/json' \
  -d '{"code":"e2e_script","nickname":"脚本用户"}')
UTOKEN=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['data']['token'])" "$WX")
BROWSE=$(curl -sf "$BASE/api/browse/$SID" -H "Authorization: Bearer $UTOKEN")
PID=$(python3 -c "import json,sys; print(json.load(sys.stdin)['data']['cards'][0]['partner_id'])" <<<"$BROWSE")

DATE=$(python3 -c "from datetime import date,timedelta; print((date.today()+timedelta(days=2)).isoformat())")
ORDER=$(curl -sf -X POST "$BASE/api/orders" -H "Authorization: Bearer $UTOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"sid\":\"$SID\",\"partner_id\":$PID,\"schedule_date\":\"$DATE\",\"start_time\":\"15:00\",\"duration_hours\":2,\"contact_phone\":\"13800138000\"}")
OID=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['data']['id'])" "$ORDER")
PAY=$(curl -sf -X POST "$BASE/api/orders/$OID/pay" -H "Authorization: Bearer $UTOKEN")
OUT=$(python3 -c "import json,sys; print(json.loads(sys.argv[1])['data']['out_trade_no'])" "$PAY")
curl -sf -X POST "$BASE/api/pay/mock-notify" -H 'Content-Type: application/json' \
  -d "{\"out_trade_no\":\"$OUT\"}" >/dev/null

echo "== result =="
curl -sf "$BASE/admin/orders/$OID" -H "Authorization: Bearer $TOKEN" | python3 -m json.tool | head -30
curl -sf "$BASE/admin/commission/report" -H "Authorization: Bearer $TOKEN" | python3 -m json.tool
echo OK
