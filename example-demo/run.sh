#!/bin/bash

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Go Analysis Client Demo Runner${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 检查服务器是否运行
echo -e "${YELLOW}Checking if go-analysis-server is running...${NC}"
if curl -s http://localhost:8097/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Server is running at http://localhost:8097${NC}"
else
    echo -e "${RED}✗ Server is not running!${NC}"
    echo -e "${YELLOW}Please start go-analysis-server first:${NC}"
    echo -e "  cd ../go-analysis-server"
    echo -e "  go run main.go"
    exit 1
fi

# 检查前端是否运行
echo -e "${YELLOW}Checking if frontend is running...${NC}"
if curl -s http://localhost:3000 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Frontend is running at http://localhost:3000${NC}"
else
    echo -e "${YELLOW}⚠ Frontend is not running (optional)${NC}"
    echo -e "  You can view results in the database directly"
    echo -e "  To start frontend:"
    echo -e "    cd ../go-analysis-frontend"
    echo -e "    npm run dev"
fi

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Starting Demo Client...${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# 进入示例目录
cd "$(dirname "$0")"

# 下载依赖
echo -e "${YELLOW}Installing dependencies...${NC}"
go mod tidy

# 运行示例
echo ""
echo -e "${GREEN}Running demo...${NC}"
echo ""
go run main.go

echo ""
echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Demo completed!${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "📊 View your analytics:"
echo -e "  • Frontend: ${GREEN}http://localhost:3000${NC}"
echo -e "  • Check the 'DemoApp' product in dashboard"
echo -e "  • View events, device info, and statistics"
echo ""
