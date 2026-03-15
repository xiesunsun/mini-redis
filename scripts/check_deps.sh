#!/bin/bash

ERRORS=0

check() {
  local pkg=$1
  local forbidden=$2
  local reason=$3
  local fix=$4
  local result
  result=$(grep -rn "$forbidden" "internal/$pkg/" 2>/dev/null | grep ".go")
  if [ -n "$result" ]; then
    echo "❌ 违规：internal/$pkg 不能引用 $forbidden"
    echo "   原因：$reason"
    echo "   修复：$fix"
    echo "   参考：docs/architecture.md"
    echo "$result"
    echo ""
    ERRORS=$((ERRORS + 1))
  fi
}

# types 不能引用任何内部包
check "types" "internal/store" \
  "types 是最底层，不能依赖任何层" \
  "把这个逻辑移到合适的上层包中"

check "types" "internal/expiry" \
  "types 是最底层，不能依赖任何层" \
  "把这个逻辑移到合适的上层包中"

check "types" "internal/persistence" \
  "types 是最底层，不能依赖任何层" \
  "把这个逻辑移到合适的上层包中"

check "types" "internal/command" \
  "types 是最底层，不能依赖任何层" \
  "把这个逻辑移到合适的上层包中"

check "types" "internal/network" \
  "types 是最底层，不能依赖任何层" \
  "把这个逻辑移到合适的上层包中"

# store 不能引用 command / network
check "store" "internal/command" \
  "store 是存储层，位于 command 层下方" \
  "把这个逻辑移到 internal/command 中"

check "store" "internal/network" \
  "store 是存储层，位于 network 层下方" \
  "把这个逻辑移到 internal/network 中"

# expiry 不能引用 command / network
check "expiry" "internal/command" \
  "expiry 是过期层，位于 command 层下方" \
  "把这个逻辑移到 internal/command 中"

check "expiry" "internal/network" \
  "expiry 是过期层，位于 network 层下方" \
  "把这个逻辑移到 internal/network 中"

# persistence 不能引用 command / network
check "persistence" "internal/command" \
  "persistence 是持久化层，位于 command 层下方" \
  "把这个逻辑移到 internal/command 中"

check "persistence" "internal/network" \
  "persistence 是持久化层，位于 network 层下方" \
  "把这个逻辑移到 internal/network 中"

# command 不能引用 network
check "command" "internal/network" \
  "command 是命令层，位于 network 层下方" \
  "把这个逻辑移到 internal/network 中"

if [ $ERRORS -eq 0 ]; then
  echo "✅ 依赖检查通过"
else
  echo "共发现 $ERRORS 个违规"
  exit 1
fi