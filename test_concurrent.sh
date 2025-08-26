#!/bin/bash

echo "=== Test du WordMon Go - Mode Concurrent ==="
echo "Lancement d'une démo de 8 secondes avec spawning toutes les 3 secondes"
echo ""

cd "/c/Users/samue/Documents/SAMUEL/ForeachAcademy/MASTER/Golang/exerice1/wordmon-go"

# Lancer en arrière-plan et capturer la sortie
./wordmon.exe --concurrent --duration 8 &
PID=$!

# Attendre 9 secondes pour capturer la sortie complète
sleep 9

# Vérifier si le processus est encore en cours
if kill -0 $PID 2>/dev/null; then
    echo "Arrêt forcé du processus..."
    kill $PID
fi

echo ""
echo "=== Test terminé ==="
