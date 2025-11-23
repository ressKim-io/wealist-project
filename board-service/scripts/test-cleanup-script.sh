#!/bin/bash

# 정리 스크립트 테스트
# 이 스크립트는 정리 유틸리티가 올바르게 컴파일되고 실행되는지 확인합니다

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

echo "정리 스크립트 테스트"
echo "===================="
echo ""

# 1. 컴파일 테스트
echo "1. Go 스크립트 컴파일 확인 중..."
cd "$PROJECT_ROOT"
if go build -o /tmp/cleanup-test scripts/cleanup-nil-author-ids.go 2>/dev/null; then
    echo "   ✓ 컴파일 성공"
    rm -f /tmp/cleanup-test
else
    echo "   ✗ 컴파일 실패"
    exit 1
fi

# 2. Dry-run 모드 테스트
echo ""
echo "2. Dry-run 모드 테스트 중..."
if go run scripts/cleanup-nil-author-ids.go --dry-run > /tmp/cleanup-output.txt 2>&1; then
    echo "   ✓ Dry-run 모드 실행 성공"
    
    # 출력 확인
    if grep -q "nil UUID author_id를 가진 보드" /tmp/cleanup-output.txt; then
        echo "   ✓ 예상된 출력 확인됨"
    else
        echo "   ✗ 예상된 출력이 없음"
        cat /tmp/cleanup-output.txt
        exit 1
    fi
else
    echo "   ✗ Dry-run 모드 실패"
    cat /tmp/cleanup-output.txt
    exit 1
fi

# 3. 셸 스크립트 실행 권한 확인
echo ""
echo "3. 셸 스크립트 실행 권한 확인 중..."
if [ -x "$SCRIPT_DIR/cleanup-nil-author-ids.sh" ]; then
    echo "   ✓ 실행 권한 있음"
else
    echo "   ✗ 실행 권한 없음"
    echo "   다음 명령으로 권한 부여: chmod +x scripts/cleanup-nil-author-ids.sh"
    exit 1
fi

# 4. README 파일 존재 확인
echo ""
echo "4. 문서 파일 확인 중..."
if [ -f "$SCRIPT_DIR/CLEANUP_NIL_AUTHOR_IDS.md" ]; then
    echo "   ✓ 영문 README 존재"
else
    echo "   ✗ 영문 README 없음"
    exit 1
fi

if [ -f "$SCRIPT_DIR/CLEANUP_NIL_AUTHOR_IDS_KR.md" ]; then
    echo "   ✓ 한글 README 존재"
else
    echo "   ✗ 한글 README 없음"
    exit 1
fi

# 정리
rm -f /tmp/cleanup-output.txt

echo ""
echo "===================="
echo "✓ 모든 테스트 통과!"
echo ""
echo "다음 단계:"
echo "1. 데이터베이스에 연결하여 실제 실행:"
echo "   ./scripts/cleanup-nil-author-ids.sh"
echo ""
echo "2. 또는 dry-run 모드로 데모 보기:"
echo "   go run scripts/cleanup-nil-author-ids.go --dry-run"
