#!/bin/bash
# 티스토리 봇 스케줄러 실행 스크립트
# Go 버전 충돌 방지를 위해 GOTOOLCHAIN=local 설정

export GOTOOLCHAIN=local
cd "$(dirname "$0")"

echo "🚀 티스토리 스케줄러 시작..."
./tistory-bot.exe schedule

