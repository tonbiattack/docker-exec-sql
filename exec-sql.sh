#!/bin/bash
# exec-sql.sh
# 起動中のDockerコンテナで指定フォルダ内のSQLファイルを順番に実行するスクリプト

set -euo pipefail

usage() {
  cat <<EOF
使い方:
  $0 -c <コンテナ名> -d <DBデータベース名> -u <DBユーザー> -s <SQLフォルダ> [-p <DBパスワード>] [-t <DBタイプ>]

オプション:
  -c  Dockerコンテナ名またはID (必須)
  -d  データベース名 (必須)
  -u  データベースユーザー名 (必須)
  -s  SQLファイルが格納されたフォルダパス (必須)
  -p  データベースパスワード (省略時は空)
  -t  データベース種別: mysql | postgres (デフォルト: mysql)
  -h  ヘルプを表示

例:
  $0 -c my_mysql -d mydb -u root -p secret -s ./sql
  $0 -c my_postgres -d mydb -u postgres -s ./sql -t postgres
EOF
  exit 1
}

CONTAINER=""
DB_NAME=""
DB_USER=""
DB_PASS=""
SQL_DIR=""
DB_TYPE="mysql"

while getopts "c:d:u:p:s:t:h" opt; do
  case $opt in
    c) CONTAINER="$OPTARG" ;;
    d) DB_NAME="$OPTARG" ;;
    u) DB_USER="$OPTARG" ;;
    p) DB_PASS="$OPTARG" ;;
    s) SQL_DIR="$OPTARG" ;;
    t) DB_TYPE="$OPTARG" ;;
    h) usage ;;
    *) usage ;;
  esac
done

# 必須パラメータチェック
if [[ -z "$CONTAINER" || -z "$DB_NAME" || -z "$DB_USER" || -z "$SQL_DIR" ]]; then
  echo "[ERROR] 必須オプションが不足しています。"
  usage
fi

# SQLフォルダの存在確認
if [[ ! -d "$SQL_DIR" ]]; then
  echo "[ERROR] SQLフォルダが見つかりません: $SQL_DIR"
  exit 1
fi

# コンテナの起動確認
if ! docker inspect --format '{{.State.Running}}' "$CONTAINER" 2>/dev/null | grep -q "true"; then
  echo "[ERROR] コンテナが起動していません: $CONTAINER"
  exit 1
fi

# SQLファイルの一覧取得 (ファイル名順にソート)
SQL_FILES=($(ls "$SQL_DIR"/*.sql 2>/dev/null | sort))

if [[ ${#SQL_FILES[@]} -eq 0 ]]; then
  echo "[ERROR] SQLファイルが見つかりません: $SQL_DIR/*.sql"
  exit 1
fi

echo "=========================================="
echo " Docker SQL 実行"
echo "=========================================="
echo "  コンテナ : $CONTAINER"
echo "  DB種別   : $DB_TYPE"
echo "  DB名     : $DB_NAME"
echo "  ユーザー : $DB_USER"
echo "  SQLフォルダ: $SQL_DIR"
echo "  ファイル数: ${#SQL_FILES[@]} 件"
echo "=========================================="

SUCCESS=0
FAIL=0

for SQL_FILE in "${SQL_FILES[@]}"; do
  FILENAME=$(basename "$SQL_FILE")
  echo ""
  echo "[実行中] $FILENAME"

  # ファイルをコンテナへコピーして実行
  docker cp "$SQL_FILE" "$CONTAINER:/tmp/$FILENAME"

  case "$DB_TYPE" in
    mysql)
      if [[ -n "$DB_PASS" ]]; then
        docker exec "$CONTAINER" bash -c \
          "mysql -u${DB_USER} -p${DB_PASS} ${DB_NAME} < /tmp/${FILENAME} && rm /tmp/${FILENAME}"
      else
        docker exec "$CONTAINER" bash -c \
          "mysql -u${DB_USER} ${DB_NAME} < /tmp/${FILENAME} && rm /tmp/${FILENAME}"
      fi
      ;;
    postgres)
      if [[ -n "$DB_PASS" ]]; then
        docker exec -e "PGPASSWORD=${DB_PASS}" "$CONTAINER" bash -c \
          "psql -U ${DB_USER} -d ${DB_NAME} -f /tmp/${FILENAME} && rm /tmp/${FILENAME}"
      else
        docker exec "$CONTAINER" bash -c \
          "psql -U ${DB_USER} -d ${DB_NAME} -f /tmp/${FILENAME} && rm /tmp/${FILENAME}"
      fi
      ;;
    *)
      echo "[ERROR] 未対応のDB種別: $DB_TYPE (mysql または postgres を指定してください)"
      exit 1
      ;;
  esac

  if [[ $? -eq 0 ]]; then
    echo "[OK] $FILENAME"
    ((SUCCESS++))
  else
    echo "[FAIL] $FILENAME"
    ((FAIL++))
  fi
done

echo ""
echo "=========================================="
echo " 完了: 成功=${SUCCESS}件 / 失敗=${FAIL}件"
echo "=========================================="

[[ $FAIL -eq 0 ]] && exit 0 || exit 1
