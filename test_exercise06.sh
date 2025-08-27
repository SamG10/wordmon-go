#!/bin/bash

echo "=== Tests de l'Exercice 06 - Configuration & Données ==="
echo ""

cd "/c/Users/samue/Documents/SAMUEL/ForeachAcademy/MASTER/Golang/exerice1/wordmon-go"

echo "✅ Test 1: Chargement de base (YAML)"
WORDMON_CHALLENGES_PATH="./configs/challenges.yaml" timeout 3 ./wordmon.exe --test | head -10

echo ""
echo "✅ Test 2: Override par ENV (intervalle spawn = 2s)"
WORDMON_CHALLENGES_PATH="./configs/challenges.yaml" WORDMON_SPAWN_INTERVAL=2 timeout 6 ./wordmon.exe --concurrent --duration 4 | grep -E "(override|Spawning|spawn toutes)"

echo ""
echo "✅ Test 3: Format TOML"
WORDMON_CONFIG_PATH="configs/game.toml" WORDMON_CHALLENGES_PATH="./configs/challenges.yaml" timeout 3 ./wordmon.exe --test | grep -E "(chargement|game:|version)"

echo ""
echo "✅ Test 4: Validation erreur (somme ≠ 100)"
WORDMON_CONFIG_PATH="configs/game_invalid.yaml" WORDMON_CHALLENGES_PATH="./configs/challenges.yaml" ./wordmon.exe --test 2>&1 | grep -E "(validation|somme)"

echo ""
echo "✅ Test 5: Validation erreur (rareté inconnue)"
WORDMON_WORDS_PATH="configs/words_invalid.json" WORDMON_CHALLENGES_PATH="./configs/challenges.yaml" ./wordmon.exe --test 2>&1 | grep -E "(validation|rareté inconnue)"

echo ""
echo "✅ Test 6: Snapshot sauvegardé"
if [ -f "data/snapshot.json" ]; then
    echo "   Snapshot trouvé: data/snapshot.json"
    echo "   Contenu:"
    cat data/snapshot.json | head -5
else
    echo "   ❌ Snapshot non trouvé"
fi

echo ""
echo "=== Tous les tests de l'exercice 06 terminés ==="
