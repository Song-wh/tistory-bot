# 🤖 티스토리 자동 포스팅 봇

티스토리에 자동으로 글을 포스팅하는 Go 기반 봇입니다.

## ✨ 주요 기능

- 📝 **다중 계정 지원** - 여러 블로그 동시 관리
- ⏰ **자동 스케줄링** - cron 표현식으로 예약 포스팅
- 🖼️ **썸네일 자동 생성** - 카테고리별 맞춤 디자인
- 📊 **성과 분석** - 카테고리별/시간대별 분석 & 최적화 제안
- 🔄 **실시간 API 연동** - 구글 트렌드, 스포츠 API 등

## 📦 지원 콘텐츠

| 카테고리 | 설명 |
|----------|------|
| `crypto` | 암호화폐 시세 + AI 추천 |
| `trend` | 실시간 인기 검색어 (구글 트렌드 연동) |
| `tech` | IT/테크 뉴스 |
| `movie` | 영화/드라마 + OTT 링크 |
| `sports` | 스포츠 뉴스 + 경기 결과 (실시간 API) |
| `lotto` | 로또 당첨번호 |
| `lotto-predict` | AI 로또 예측 |
| `fortune` | 띠별 오늘의 운세 + 쿠팡 연동 |
| `golf` | 내일 골프 날씨 예보 |
| `golf-tips` | 골프 레슨 팁 + 용품 추천 |
| `error` | 프로그래밍 에러 해결 아카이브 |
| `coupang` | 쿠팡 특가 상품 |

---

## 🚀 설치 및 실행

### 1. 빌드

```bash
git clone https://github.com/Song-wh/tistory-bot.git
cd tistory-bot
go build -o tistory-bot.exe ./cmd/tistory-bot/
```

### 2. 설정

`config.yaml` 파일을 수정하세요:

```yaml
accounts:
  - name: "my-blog"
    enabled: true
    tistory:
      email: "your@email.com"
      password: "your_password"
      blog_name: "your-blog-name"
    coupang:
      partner_id: "AFXXXXXXX"
    schedule:
      enabled: true
      jobs:
        - category: trend
          cron: "0 12 * * *"
        - category: crypto
          cron: "0 10,18 * * *"
```

### 3. 실행

```bash
# 단일 포스팅
./tistory-bot.exe post trend

# 자동 스케줄러 실행
./tistory-bot.exe schedule
```

---

## 📋 명령어

### 기본 명령어

```bash
# 로그인 테스트
./tistory-bot.exe login

# 단일 포스팅
./tistory-bot.exe post [category]
./tistory-bot.exe post crypto --account my-blog

# 계정 목록 조회
./tistory-bot.exe accounts

# 자동 스케줄러 실행
./tistory-bot.exe schedule
```

---

## 📊 콘텐츠 성과 분석

어떤 콘텐츠가 잘 되는지 분석하고, 잘 되는 콘텐츠 비중을 높일 수 있습니다.

### 분석 리포트 보기

```bash
./tistory-bot.exe analytics report
```

**결과:**
```
════════════════════════════════════════════════════════════
📊 콘텐츠 성과 분석 리포트 - my-blog
════════════════════════════════════════════════════════════
📅 생성일: 2025-12-23 16:32
📝 총 포스트: 50개 | 총 조회수: 17125

🏆 카테고리별 성과 순위
────────────────────────────────────────────────────────────
🥇 에러-해결      평균 조회수: 535  → 빈도 증가 권장!
🥈 트렌드-실검    평균 조회수: 605
🥉 주식-코인      평균 조회수: 450
...

⏰ 최적 발행 시간대
  1위: 16:00 (평균 조회수: 342)

💡 추천 사항
  🏆 '에러-해결' 포스팅 빈도 증가 권장
  ⏰ 16시 발행이 가장 효과적!
════════════════════════════════════════════════════════════
```

### 실제 데이터 수집 (선택)

```bash
./tistory-bot.exe analytics collect
```

티스토리 관리자 페이지에서 실제 조회수/댓글/공감 수집합니다.

### 스케줄 최적화 제안

```bash
./tistory-bot.exe analytics optimize
```

**결과:**
```
🎯 최적화된 스케줄 제안
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
config.yaml에 복붙하세요:

- category: error
  cron: "0 9,15,21 * * *"   # 성과 좋음 → 하루 3회

- category: trend
  cron: "0 10,18 * * *"     # 성과 좋음 → 하루 2회

- category: movie
  cron: "0 12 * * 1,4"      # 성과 낮음 → 주 2회만
━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

### 추천 워크플로우

1. **주 1회** `analytics report` 실행
2. 추천에 따라 `config.yaml` 스케줄 조정
3. 성과 좋은 카테고리 빈도 증가, 낮은 것은 축소
4. 스케줄러 재시작

---

## 🖼️ 썸네일 자동 생성

카테고리별로 맞춤 썸네일이 자동 생성됩니다.

### 설정 (`config.yaml`)

```yaml
thumbnail:
  enabled: true
  output_dir: "./thumbnails"
```

### 카테고리별 디자인

| 카테고리 | 색상 | 아이콘 |
|----------|------|--------|
| crypto | 🟡 골드→오렌지 | BTC |
| trend | 🌸 핑크→레드 | HOT |
| tech | 🔵 블루→퍼플 | TECH |
| movie | 🔴 크림슨→마젠타 | MOVIE |
| sports | 🩵 그린→시안 | SPORTS |
| fortune | ✨ 골드→오렌지 | FORTUNE |
| error | ⬛ 다크그레이 | DEBUG |

---

## ⚙️ 고급 설정

### 랜덤 딜레이

봇같아 보이지 않게 포스팅 시간에 0~45분 랜덤 딜레이가 적용됩니다.

### 태그 최적화

- 최대 10개 태그 자동 제한 (티스토리 규정)
- 중복 태그 자동 제거
- 동적 태그 생성 (실제 콘텐츠 기반)

### API 연동

```yaml
# 스포츠 API (선택)
football_data:
  api_key: "YOUR_API_KEY"  # football-data.org

# 영화 API
tmdb:
  api_key: "YOUR_API_KEY"
```

---

## 🖥️ 백그라운드 실행 (Windows)

노트북 덮개 닫아도 계속 실행:

```powershell
# 관리자 권한으로 실행
powercfg /setdcvalueindex SCHEME_CURRENT SUB_BUTTONS LIDACTION 0
powercfg /setacvalueindex SCHEME_CURRENT SUB_BUTTONS LIDACTION 0
powercfg /setactive SCHEME_CURRENT
```

---

## 📝 라이선스

MIT License

---

## 🙏 기여

이슈와 PR 환영합니다!

