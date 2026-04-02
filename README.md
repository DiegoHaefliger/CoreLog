# CoreLog

> Monitoramento centralizado de logs para a infraestrutura Haevo.

Sistema que coleta, processa e exibe logs de arquivos e containers Docker em tempo real, com busca full-text (SQLite FTS5), filtros por level, serviço e intervalo de data/hora.

---

## Stack

| Camada | Tecnologia |
|--------|-----------|
| Backend | Go (net/http, SQLite via `modernc.org/sqlite`) |
| Frontend | React + TypeScript + Vite + TailwindCSS |
| Banco | SQLite com FTS5 (embutido no binário) |
| Deploy | Binário único — frontend embutido via `embed.FS` |

---

## Pré-requisitos

| Ferramenta | Versão mínima |
|------------|--------------|
| Go | 1.21+ |
| Node.js | 18+ |
| npm | 9+ |

---

## Rodando localmente (desenvolvimento)

### 1. Backend

```bash
cd backend
go run cmd/server/main.go --log-path=../data/application.log --db-path=../data/corelog.db
```

O servidor sobe na porta **9091**.

### 2. Frontend (dev server com hot-reload)

Em outro terminal:

```bash
cd frontend
npm install
npm run dev
```

O Vite sobe na porta **5173** e faz proxy para o backend em `9091` (configurado em `vite.config.ts`).

Acesse: `http://localhost:5173`

### 3. Gerar logs de teste

```bash
./generate_fake_logs.sh
```

Injeta logs aleatórios em `data/application.log` a cada 1,5 segundos. Pressione `Ctrl+C` para parar.

---

## Build para produção

O build compila o frontend e o embute dentro do binário Go, gerando um único executável.

```bash
./build.sh
```

Isso executa:
1. `npm run build` — gera os assets do frontend em `frontend/dist`
2. Move o `dist` para `backend/ui` (onde o `embed.FS` espera)
3. Compila o binário Go: `corelog-server` na raiz do projeto

O binário gerado é **autossuficiente** — sem dependências externas além do sistema operacional.

---

## Deploy em produção

### Primeira instalação no servidor

Após o `./build.sh` na máquina local:

```bash
# Copiar binário e script de instalação para o servidor
scp ./corelog-server ./install.sh root@SEU_SERVIDOR:/tmp/

# No servidor — executar o script de instalação (como root)
ssh root@SEU_SERVIDOR "cd /tmp && bash install.sh"
```

O `install.sh` irá:
- Instalar o binário em `/usr/local/bin/corelog`
- Criar o diretório de dados em `/var/lib/corelog`
- Criar e habilitar o serviço systemd `corelog.service`
- Iniciar o servidor na porta **9091**

### Atualização (deploy de nova versão)

**Na máquina local:**

```bash
# 1. Build com as novas alterações
./build.sh

# 2. Enviar binário para o servidor
# Via SCP:
scp ./corelog-server root@SEU_SERVIDOR:/var/www/api/corelog/corelog-server
# Ou via FileZilla: /var/www/api/corelog/corelog-server
```

**No servidor:**

```bash
systemctl restart corelog
systemctl status corelog
```

> **Nota:** o `install.sh` padrão instala em `/usr/local/bin/corelog`. Em produção o caminho foi customizado para `/var/www/api/corelog/corelog-server`.

### Verificar se a versão foi atualizada

O número de versão aparece abaixo de "CoreLog" na sidebar da interface. Se mudar após o deploy, o binário foi atualizado com sucesso.

### Configuração do serviço systemd (referência)

```ini
[Unit]
Description=CoreLog - Centralized Log Monitoring Daemon
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/var/www/api/corelog
ExecStart=/var/www/api/corelog/corelog-server \
  --log-path=/var/log/application.log \
  --db-path=/var/www/api/corelog/corelog.db \
  --docker
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Flags do binário

| Flag | Padrão | Descrição |
|------|--------|-----------|
| `--log-path` | `data/application.log` | Arquivo de log a monitorar |
| `--db-path` | `data/corelog.db` | Caminho do banco SQLite |
| `--docker` | `false` | Coleta logs de todos os containers Docker automaticamente |

---

## Dados e banco

| Caminho | Descrição |
|---------|-----------|
| `/var/www/api/corelog/corelog.db` | Banco SQLite (produção) |
| `data/corelog.db` | Banco SQLite (desenvolvimento local) |

O banco é criado automaticamente na primeira execução. **Não commitar o `.db` no git.**

---

## Credenciais padrão

| Usuário | Senha |
|---------|-------|
| `admin` | `admin` |

Na primeira entrada, o sistema exige troca de senha. Guarde bem — a recuperação exige acesso direto ao banco.

---

## Comandos úteis no servidor

```bash
# Parar
systemctl stop corelog

# Status do serviço
systemctl status corelog

# Logs do próprio CoreLog (stdout/stderr)
journalctl -u corelog -f

# Reiniciar após atualização de binário
systemctl restart corelog


