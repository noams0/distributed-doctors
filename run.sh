#!/bin/bash

rm -f /tmp/in_* /tmp/out_*

mkfifo /tmp/in_A1 /tmp/out_A1 /tmp/in_C1 /tmp/out_C1
mkfifo /tmp/in_A2 /tmp/out_A2 /tmp/in_C2 /tmp/out_C2
mkfifo /tmp/in_A3 /tmp/out_A3 /tmp/in_C3 /tmp/out_C3

# Lancer les sites
go run app/app.go -n 1 < /tmp/in_A1 > /tmp/out_A1 &
go run ctl/ctl.go -n 1 < /tmp/in_C1 > /tmp/out_C1 &

go run app/app.go -n 2 < /tmp/in_A2 > /tmp/out_A2 &
go run ctl/ctl.go -n 2 < /tmp/in_C2 > /tmp/out_C2 &

go run app/app.go -n 3 < /tmp/in_A3 > /tmp/out_A3 &
go run ctl/ctl.go -n 3 < /tmp/in_C3 > /tmp/out_C3 &

# Connexion des flux
cat /tmp/out_A1 > /tmp/in_C1 &
cat /tmp/out_C1 | tee /tmp/in_A1 > /tmp/in_C2 &

cat /tmp/out_A2 > /tmp/in_C2 &
cat /tmp/out_C2 | tee /tmp/in_A2 > /tmp/in_C3 &

cat /tmp/out_A3 > /tmp/in_C3 &
cat /tmp/out_C3 | tee /tmp/in_A3 > /tmp/in_C1 &
