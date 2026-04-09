# docker-exec-sql

ローカルの指定フォルダ内にある複数のSQLファイルを、起動中のDockerコンテナ（MySQL）へファイル名順に一括実行するツールです。

シェルスクリプト版（`exec-sql.sh`）と Go CLI版（Cobra）の2種類があります。

---

## Go CLI版（推奨）

### 前提条件

- Go 1.18以上
- `docker` コマンドが使用可能なこと
- 対象のDockerコンテナが起動済みであること
- コンテナ内に `mysql` コマンドがインストール済みであること

### セットアップ

```bash
go mod tidy
```

### 使い方

```bash
# go run で実行
go run . -c <コンテナ名> -d <DB名> -u <ユーザー> -s <SQLフォルダ> [-p <パスワード>]

# バイナリをビルドして実行
go build -o docker-exec-sql .
./docker-exec-sql -c <コンテナ名> -d <DB名> -u <ユーザー> -s <SQLフォルダ>
```

### フラグ

| フラグ | 長形式 | 説明 | 必須 |
|-------|--------|------|------|
| `-c` | `--container` | DockerコンテナIDまたはコンテナ名 | ✅ |
| `-d` | `--database` | データベース名 | ✅ |
| `-u` | `--user` | データベースユーザー名 | ✅ |
| `-s` | `--sql-dir` | SQLファイルが格納されたローカルフォルダパス | ✅ |
| `-p` | `--password` | データベースパスワード | - |
| `-h` | `--help` | ヘルプ表示 | - |

### 実行例

```bash
# パスワードあり
go run . -c my_mysql_container -d mydb -u root -p secret -s ./sql

# パスワードなし
go run . -c my_mysql_container -d mydb -u root -s ./sql
```

---

## シェルスクリプト版

### 前提条件

- `docker` コマンドが使用可能なこと
- 対象のDockerコンテナが起動済みであること
- コンテナ内に `mysql` コマンドがインストール済みであること

### 実行権限の付与

```bash
chmod +x exec-sql.sh
```

### 使い方

```bash
./exec-sql.sh -c <コンテナ名> -d <DB名> -u <ユーザー> -s <SQLフォルダ> [-p <パスワード>]
```

### オプション

| オプション | 説明 | 必須 |
|-----------|------|------|
| `-c` | DockerコンテナIDまたはコンテナ名 | ✅ |
| `-d` | データベース名 | ✅ |
| `-u` | データベースユーザー名 | ✅ |
| `-s` | SQLファイルが格納されたローカルフォルダパス | ✅ |
| `-p` | データベースパスワード | - |
| `-h` | ヘルプ表示 | - |

### 実行例

```bash
# パスワードあり
./exec-sql.sh -c my_mysql_container -d mydb -u root -p secret -s ./sql

# パスワードなし
./exec-sql.sh -c my_mysql_container -d mydb -u root -s ./sql
```

---

## 検証用MySQL環境（docker-compose）

`docker-compose.yml` で検証用のMySQLコンテナを手軽に起動できます。

| 項目 | 値 |
|------|----|
| コンテナ名 | `test-mysql` |
| データベース | `testdb` |
| ユーザー | `root` |
| パスワード | `secret` |
| ポート | `3306` |

```bash
# 起動
docker compose up -d

# 停止・削除
docker compose down
```

起動後、以下のコマンドで動作確認できます。

```bash
go run . -c test-mysql -d testdb -u root -p secret -s ./sql
```

---

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

1. ローカルの指定フォルダ内の `*.sql` ファイルをファイル名順に取得
2. 各SQLファイルを `docker cp` でコンテナの `/tmp/` へコピー
3. コンテナ内で `mysql` コマンドを使って実行
4. 実行後、コンテナ内の `/tmp/` から一時ファイルを削除
5. 全件処理後、成功・失敗件数を表示
