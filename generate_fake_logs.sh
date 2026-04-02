#!/bin/bash
mkdir -p data
touch data/application.log

echo "Injetando logs de teste continuamente em data/application.log..."
echo "(Pressione Ctrl+C para parar)"

while true; do
  TIMESTAMP=$(date +"%Y-%m-%d %H:%M:%S")
  
  # Aleatoriza os níveis e origens
  LEVELS=("INFO" "WARN" "ERROR" "DEBUG")
  APPS=("auth-service" "payment-gateway" "nginx" "database")
  
  LEVEL=${LEVELS[$RANDOM % ${#LEVELS[@]}]}
  APP=${APPS[$RANDOM % ${#APPS[@]}]}
  
  if [ "$LEVEL" == "ERROR" ]; then
    echo "$TIMESTAMP [ERROR] $APP - Runtime Exception occurred: Connection timeout" >> data/application.log
    echo "    at com.corelog.auth.SessionManager.validate(SessionManager.java:42)" >> data/application.log
    echo "    at com.corelog.server.RequestHandler.handle(RequestHandler.java:108)" >> data/application.log
    echo "    ... 12 more" >> data/application.log
  elif [ "$LEVEL" == "WARN" ]; then
    echo "$TIMESTAMP [WARN] $APP - Token bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIi... is about to expire" >> data/application.log
  else
    echo "$TIMESTAMP [$LEVEL] $APP - User login request payload: { \"email\": \"admin@teste.com\", \"password\": \"secreta123\" }" >> data/application.log
  fi
  
  sleep 1.5
done
