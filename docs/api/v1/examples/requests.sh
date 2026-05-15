#!/bin/bash

# Примеры консольных запросов к API InfraTopoBuilder
# Использование: ./requests.sh

BASE_URL="http://localhost:8080"

# Тестовые UUID (замените на реальные после вызова POST /parse/)
LOG_ID="e4b86bc6-a7ad-4e25-a6e8-13d43cfa074d"
NODE_ID="d9e1f57d-b8e0-5c68-bd57-38329bda343f"

echo "1. Healthcheck (/ping)"
curl -s -X GET "$BASE_URL/ping" | jq . || curl -s -X GET "$BASE_URL/ping"
echo -e "\n"

echo "2. Инициация парсинга (POST /api/v1/parse/)"
curl -s -X POST "$BASE_URL/api/v1/parse/" \
  -H "Content-Type: application/json" \
  -d '{
    "filepath": "data/log.zip"
  }' | jq . || echo -e "Выполнено\n"
echo -e "\n"

echo "3. Проверка статуса сессии (GET /api/v1/log/{log_id})"
curl -s -X GET "$BASE_URL/api/v1/log/$LOG_ID" | jq . || curl -s -X GET "$BASE_URL/api/v1/log/$LOG_ID"
echo -e "\n"

echo "4. Получение графа топологии без портов (GET /api/v1/topology/{log_id})"
curl -s -X GET "$BASE_URL/api/v1/topology/$LOG_ID" | jq . || curl -s -X GET "$BASE_URL/api/v1/topology/$LOG_ID"
echo -e "\n"

echo "5. Получение деталей конкретного узла (GET /api/v1/node/{node_id})"
curl -s -X GET "$BASE_URL/api/v1/node/$NODE_ID" | jq . || curl -s -X GET "$BASE_URL/api/v1/node/$NODE_ID"
echo -e "\n"

echo "6. Ленивая загрузка портов узла (GET /api/v1/port/{node_id})"
curl -s -X GET "$BASE_URL/api/v1/port/$NODE_ID" | jq . || curl -s -X GET "$BASE_URL/api/v1/port/$NODE_ID"
echo -e "\n"

echo "7. Сбор метрик для VictoriaMetrics (GET /metrics)"
curl -s -X GET "$BASE_URL/metrics" | head -n 10
echo -e "\n...\n"