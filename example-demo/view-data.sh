#!/bin/bash

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Go Analysis Data Viewer${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 服务器URL
SERVER_URL="http://localhost:8097"

# 1. 查询事件统计
echo -e "${GREEN}📊 Event Statistics for DemoApp:${NC}"
echo ""
curl -s "${SERVER_URL}/api/events/query?product=DemoApp&limit=100" | \
  jq -r '.data.events[] | "\(.name) - \(.timestamp) - Device: \(.device_id[0:8])..."' | \
  sort | uniq -c | sort -rn || echo "Failed to fetch events"

echo ""
echo -e "${BLUE}----------------------------------------${NC}"

# 2. 查询最近的事件
echo -e "${GREEN}📋 Recent Events (Last 10):${NC}"
echo ""
curl -s "${SERVER_URL}/api/events/query?product=DemoApp&limit=10" | \
  jq -r '.data.events[] | "[\(.timestamp | strftime("%H:%M:%S"))] \(.name) - \(.properties | to_entries | map("\(.key)=\(.value)") | join(", "))"' \
  2>/dev/null || echo "Failed to fetch recent events"

echo ""
echo -e "${BLUE}----------------------------------------${NC}"

# 3. 查询设备信息
echo -e "${GREEN}💻 Device Information:${NC}"
echo ""
DEVICE_ID=$(curl -s "${SERVER_URL}/api/events/query?product=DemoApp&limit=1" | jq -r '.data.events[0].device_id' 2>/dev/null)
if [ -n "$DEVICE_ID" ] && [ "$DEVICE_ID" != "null" ]; then
  echo -e "Device ID: ${YELLOW}${DEVICE_ID}${NC}"
  
  # 查询该设备的安装信息
  curl -s "${SERVER_URL}/api/events/query?product=DemoApp&limit=1" | \
    jq -r '.data.events[0] | "Session ID: \(.session_id)\nUser ID: \(.user_id // "N/A")"' \
    2>/dev/null || echo "Failed to fetch device info"
else
  echo "No device information found"
fi

echo ""
echo -e "${BLUE}----------------------------------------${NC}"

# 4. 总体统计
echo -e "${GREEN}📈 Overall Statistics:${NC}"
echo ""
curl -s "${SERVER_URL}/api/stats?product=DemoApp" | jq '.' 2>/dev/null || echo "Failed to fetch stats"

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}View in browser:${NC} http://localhost:3000"
echo -e "${BLUE}========================================${NC}"
