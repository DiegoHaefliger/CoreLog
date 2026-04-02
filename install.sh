#!/bin/bash
set -e

# CoreLog Auto-Deploy Script
# Run this on your target server as root

if [ "$EUID" -ne 0 ]; then
  echo "Por favor, execute este script como root (sudo)."
  exit 1
fi

BINARY_NAME="corelog-server"
INSTALL_DIR="/usr/local/bin"
DATA_DIR="/var/lib/corelog"
SERVICE_FILE="/etc/systemd/system/corelog.service"

# Verificar se binário existe na pasta local
if [ ! -f "./$BINARY_NAME" ]; then
    echo "Erro: Binário $BINARY_NAME não encontrado no diretório atual."
    echo "Execute './build.sh' na sua máquina de desenvolvimento e faça upload do binário '$BINARY_NAME' para cá."
    exit 1
fi

echo "1. Instalando binário em $INSTALL_DIR/corelog..."
cp "./$BINARY_NAME" "$INSTALL_DIR/corelog"
chmod +x "$INSTALL_DIR/corelog"

echo "2. Criando diretório de dados em $DATA_DIR..."
mkdir -p "$DATA_DIR"

# Definir Caminho do Log Alvo (pode ser ajustado)
TARGET_LOG="/var/log/application.log"
read -p "Qual é o caminho do arquivo de log que deseja monitorar? [Default: $TARGET_LOG]: " USER_LOG
TARGET_LOG=${USER_LOG:-$TARGET_LOG}

# Garantir que o log alvo exista preventivamente para evitar erros de leitura
touch "$TARGET_LOG"

echo "3. Criando serviço SystemD Daemon..."
cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=CoreLog - Centralized Log Monitoring Daemon
After=network.target

[Service]
Type=simple
User=root
# Running from data dir so corelog.db is safely stored there
WorkingDirectory=$DATA_DIR
ExecStart=$INSTALL_DIR/corelog --log-path=$TARGET_LOG --db-path=$DATA_DIR/corelog.db --docker
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF

echo "4. Recarregando SystemD e iniciando CoreLog..."
systemctl daemon-reload
systemctl enable corelog
systemctl restart corelog

echo "======================================"
echo "Deploy finalizado!"
echo "O CoreLog está rodando no background."
echo "- DB Local: $DATA_DIR/corelog.db"
echo "- Alvo de Logs: $TARGET_LOG"
echo "- Porta Web: 9091 (verifique Firewall/VPC)"
echo "Use 'systemctl status corelog' para checar a saúde do sistema."
