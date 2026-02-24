#!/bin/bash
# ─────────────────────────────────────────────────────────────
# VPS Setup Script for Master-Slave Server
# Run this ONCE on a fresh Ubuntu 22.04 VPS as root
#
# Usage: ssh root@31.97.61.182
#        Then paste & run this whole script
# ─────────────────────────────────────────────────────────────

set -e

echo "══════════════════════════════════════════════════════════"
echo "  Master-Slave Server — VPS Setup"
echo "══════════════════════════════════════════════════════════"

# ─── Step 1: Update system ────────────────────────────────────
echo ""
echo "📦 [1/5] Updating system packages..."
apt-get update -y && apt-get upgrade -y

# ─── Step 2: Install Docker ──────────────────────────────────
echo ""
echo "🐳 [2/5] Installing Docker..."
if ! command -v docker &> /dev/null; then
    apt-get install -y ca-certificates curl gnupg
    install -m 0755 -d /etc/apt/keyrings
    curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
    chmod a+r /etc/apt/keyrings/docker.gpg
    echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu \
      $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
      tee /etc/apt/sources.list.d/docker.list > /dev/null
    apt-get update -y
    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
    echo "✅ Docker installed!"
else
    echo "✅ Docker already installed."
fi

docker --version
docker compose version

# ─── Step 3: Install Git ─────────────────────────────────────
echo ""
echo "📂 [3/5] Installing Git..."
apt-get install -y git

# ─── Step 4: Clone the repo ──────────────────────────────────
echo ""
echo "📥 [4/5] Cloning repository..."
APP_DIR="/opt/master-slave-server"
if [ -d "$APP_DIR" ]; then
    echo "Directory exists, pulling latest..."
    cd "$APP_DIR" && git pull
else
    git clone https://github.com/ckdash-git/Master_Slave_Server.git "$APP_DIR"
    cd "$APP_DIR"
fi

# ─── Step 5: Create .env and start ───────────────────────────
echo ""
echo "🔐 [5/5] Setting up environment & starting services..."

# Create .env if it doesn't exist
if [ ! -f "$APP_DIR/.env" ]; then
    cp "$APP_DIR/.env.example" "$APP_DIR/.env"
    # Generate a new random JWT secret for this VPS
    NEW_SECRET=$(openssl rand -hex 32)
    sed -i "s|JWT_SECRET=.*|JWT_SECRET=${NEW_SECRET}|" "$APP_DIR/.env"
    echo "✅ Created .env with new JWT secret"
else
    echo "✅ .env already exists, keeping it"
fi

# Build and start
cd "$APP_DIR"
docker compose down 2>/dev/null || true
docker compose up --build -d

# Wait for services
echo ""
echo "⏳ Waiting for services to start..."
sleep 15

# Health check
echo ""
echo "🏥 Running health check..."
curl -s http://localhost:8080/health | python3 -m json.tool 2>/dev/null || echo "⏳ Server still starting, try: curl http://localhost:8080/health"

echo ""
echo "══════════════════════════════════════════════════════════"
echo "  ✅ DEPLOYMENT COMPLETE!"
echo "══════════════════════════════════════════════════════════"
echo ""
echo "  📡 Endpoints:"
echo "     Health:         http://31.97.61.182:8080/health"
echo "     Login:          POST http://31.97.61.182:8080/auth/login"
echo "     Verify:         GET  http://31.97.61.182:8080/auth/verify"
echo "     Refresh:        POST http://31.97.61.182:8080/auth/refresh"
echo "     Exchange Code:  POST http://31.97.61.182:8080/auth/exchange-code"
echo "     Claim Token:    POST http://31.97.61.182:8080/auth/claim-token"
echo ""
echo "  🧪 Test login:"
echo "     curl -X POST http://31.97.61.182:8080/auth/login \\"
echo "       -H 'Content-Type: application/json' \\"
echo "       -d '{\"email\":\"admin@cachatto.click\",\"password\":\"password123\"}'"
echo ""
echo "  📋 Logs:  docker compose -f $APP_DIR/docker-compose.yml logs -f"
echo "  🔄 Restart: docker compose -f $APP_DIR/docker-compose.yml restart"
echo "══════════════════════════════════════════════════════════"
