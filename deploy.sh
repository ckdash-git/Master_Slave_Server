#!/bin/bash
# ─────────────────────────────────────────────────────────────
# Deploy Master-Slave Server to VPS
# Usage: ./deploy.sh
# ─────────────────────────────────────────────────────────────

set -e

VPS_HOST="root@31.97.61.182"
REMOTE_DIR="/opt/master-slave-server"
DOMAIN="cachatto.click"

echo "🚀 Deploying Master-Slave Server to ${VPS_HOST}..."

# ─── Step 1: Setup VPS directory ──────────────────────────────
echo "📁 Setting up remote directory..."
ssh ${VPS_HOST} "mkdir -p ${REMOTE_DIR}/migrations"

# ─── Step 2: Upload files ────────────────────────────────────
echo "📤 Uploading project files..."
scp docker-compose.yml Dockerfile .env.example ${VPS_HOST}:${REMOTE_DIR}/
scp -r migrations/ ${VPS_HOST}:${REMOTE_DIR}/
scp -r cmd/ internal/ go.mod go.sum ${VPS_HOST}:${REMOTE_DIR}/

# ─── Step 3: Upload .env (only if it doesn't exist on server)
echo "🔐 Uploading .env..."
ssh ${VPS_HOST} "test -f ${REMOTE_DIR}/.env || echo 'No .env found on server, uploading default...'"
scp .env ${VPS_HOST}:${REMOTE_DIR}/.env

# ─── Step 4: Build and start on VPS ──────────────────────────
echo "🔨 Building and starting services on VPS..."
ssh ${VPS_HOST} "cd ${REMOTE_DIR} && docker compose down 2>/dev/null || true && docker compose up --build -d"

# ─── Step 5: Verify ──────────────────────────────────────────
echo "⏳ Waiting for services to start..."
sleep 10

echo "🏥 Health check..."
ssh ${VPS_HOST} "curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || echo 'Health check pending...'"

echo ""
echo "✅ Deployment complete!"
echo ""
echo "📡 Endpoints available at:"
echo "   Health:         http://${DOMAIN}:8080/health"
echo "   Login:          POST http://${DOMAIN}:8080/auth/login"
echo "   Verify:         GET  http://${DOMAIN}:8080/auth/verify"
echo "   Refresh:        POST http://${DOMAIN}:8080/auth/refresh"
echo "   Exchange Code:  POST http://${DOMAIN}:8080/auth/exchange-code"
echo "   Claim Token:    POST http://${DOMAIN}:8080/auth/claim-token"
echo ""
echo "🧪 Test with:"
echo "   curl -X POST http://${DOMAIN}:8080/auth/login \\"
echo "     -H 'Content-Type: application/json' \\"
echo "     -d '{\"email\":\"admin@cachatto.click\",\"password\":\"password123\"}'"
