#!/bin/bash
ERRORS=0
check(){
    local pkg=$1
    local forbidden=$2
    local result
    result=$(grep -rn "$forbidden" "internal/$pkg/" 2>/dev/null | grep ".go")
    if [ -n "$result" ]; then
        echo "❌ 违规：internal/$pkg 不能引用 $forbidden"
        echo "$result"
        echo ""
        ERRORS=$((ERRORS + 1))
  fi
}

#types 不能依赖任何层
check "types" "internal/store"
check "types" "internal/expiry"
check "types" "internal/persistence"
check "types" "internal/command"
check "types" "internal/network"


#store 不能引用 command /network
check "store" "internal/command"
check "store" "internal/network"

#expiry不能引用 command /network
check "expiry" "internal/command"
check "expiry" "internal/network" 

#persistence 不能引用 command /network
check "persistence" "internal/command"
check "persistence" "internal/network"

#command 不能引用 network

check "command" "internal/network"

if [ "$ERRORS" -eq 0 ]; then
    echo "✅ 依赖检查通过"
else
    echo "共发现 $ERRORS 个违规"
    exit 1
fi