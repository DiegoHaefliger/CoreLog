#!/bin/bash
set -e

echo "[1/4] Building Frontend (React/Vite)..."
cd frontend
npm install
npm run build
cd ..

echo "[2/4] Injecting UI into Backend..."
rm -rf backend/ui
mv frontend/dist backend/ui

echo "[3/4] Building Go CoreLog Binary (Linux AMD64 minimal footprint)..."
cd backend
GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o ../corelog-server cmd/server/main.go
cd ..

echo "[4/4] Build Complete! Executable produced at: ./corelog-linux-amd64"
echo "You can now SCP 'corelog-linux-amd64' and 'install.sh' to your target server."
