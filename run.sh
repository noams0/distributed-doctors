#!/bin/bash

# Nettoyage des anciens pipes
rm -f /tmp/in_* /tmp/out_*

# Création des pipes
mkfifo /tmp/in_A1 /tmp/out_A1 /tmp/in_C1 /tmp/out_C1
mkfifo /tmp/in_A2 /tmp/out_A2 /tmp/in_C2 /tmp/out_C2
mkfifo /tmp/in_A3 /tmp/out_A3 /tmp/in_C3 /tmp/out_C3

# Liste des PIDs à tuer
pids=()

# Fonction de nettoyage à la fin
cleanup() {
  echo "Arrêt des processus..."
  for pid in "${pids[@]}"; do
    kill "$pid" 2>/dev/null
  done
  rm -f /tmp/in_* /tmp/out_*
  exit
}

# Capture CTRL+C ou fermeture
trap cleanup SIGINT SIGTERM EXIT

# Lancement des processus + stockage des PIDs
go run app/app.go -n 1 < /tmp/in_A1 > /tmp/out_A1 & pids+=($!)
go run ctl/ctl.go -n 1 < /tmp/in_C1 > /tmp/out_C1 & pids+=($!)

go run app/app.go -n 2 < /tmp/in_A2 > /tmp/out_A2 & pids+=($!)
go run ctl/ctl.go -n 2 < /tmp/in_C2 > /tmp/out_C2 & pids+=($!)

go run app/app.go -n 3 < /tmp/in_A3 > /tmp/out_A3 & pids+=($!)
go run ctl/ctl.go -n 3 < /tmp/in_C3 > /tmp/out_C3 & pids+=($!)

# Connexions des flux
cat /tmp/out_A1 > /tmp/in_C1 & pids+=($!)
cat /tmp/out_C1 | tee /tmp/in_A1 > /tmp/in_C2 & pids+=($!)

cat /tmp/out_A2 > /tmp/in_C2 & pids+=($!)
cat /tmp/out_C2 | tee /tmp/in_A2 > /tmp/in_C3 & pids+=($!)

cat /tmp/out_A3 > /tmp/in_C3 & pids+=($!)
cat /tmp/out_C3 | tee /tmp/in_A3 > /tmp/in_C1 & pids+=($!)

# Attente infinie pour garder le script vivant
# (CTRL+C déclenchera `cleanup`)
wait
