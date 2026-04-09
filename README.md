# docker-exec-sql

起動中のDockerコンテナに対して、指定フォルダ内の複数のSQLファイルをファイル名順に一括実行するシェルスクリプトです。

## 対応データベース

| DB種別 | オプション値 |
|--------|------------|
| MySQL  | `mysql`（デフォルト） |
| PostgreSQL | `postgres` |

## 使い方

```bash
./exec-sql.sh -c <コンテナ名> -d <DB名> -u <ユーザー> -s <SQLフォルダ> [-p <パスワード>] [-t <DB種別>]
```

### オプション

| オプション | 説明 | 必須 |
|-----------|------|------|
| `-c` | DockerコンテナIDまたはコンテナ名 | ✅ |
| `-d` | データベース名 | ✅ |
| `-u` | データベースユーザー名 | ✅ |
| `-s` | SQLファイルが格納されたフォルダパス | ✅ |
| `-p` | データベースパスワード | - |
| `-t` | DB種別 `mysql` / `postgres`（デフォルト: `mysql`） | - |
| `-h` | ヘルプ表示 | - |

## 実行例

### MySQL

```bash
./exec-sql.sh -c my_mysql_container -d mydb -u root -p secret -s ./sql
```

### PostgreSQL

```bash
./exec-sql.sh -c my_postgres_container -d mydb -u postgres -p secret -s ./sql -t postgres
```

### パスワードなし

```bash
./exec-sql.sh -c my_mysql_container -d mydb -u root -s ./sql
```

## SQLファイルの配置例

```
sql/
├── 001_create_tables.sql
├── 002_insert_master.sql
└── 003_insert_data.sql
```

ファイル名の昇順（辞書順）に実行されます。  
ファイル名の先頭に連番をつけることで実行順序を制御できます。

## 動作の概要

1. 指定フォルダ内の `*.sql` ファイルをファイル名順に取得
2. 各SQLファイルをコンテナの `/tmp/` へコピー（`docker cp`）
3. コンテナ内でDBクライアントを使って実行
4. 実行後、`/tmp/` 内の一時ファイルを削除
5. 全件処理後、成功・失敗件数を表示

## 前提条件

- `docker` コマンドが使用可能なこと
- 対象のDockerコンテナが起動済みであること
- コンテナ内に `mysql` または `psql` コマンドがインストール済みであること

## 実行権限の付与

初回実行前にスクリプトへ実行権限を付与してください。

```bash
chmod +x exec-sql.sh
```
