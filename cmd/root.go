package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
	container string
	dbName    string
	dbUser    string
	dbPass    string
	sqlDir    string
)

var rootCmd = &cobra.Command{
	Use:   "docker-exec-sql",
	Short: "ローカルのSQLファイルを起動中のDockerコンテナ(MySQL)へ一括実行する",
	Example: `  docker-exec-sql -c my_mysql -d mydb -u root -p secret -s ./sql
  docker-exec-sql -c my_mysql -d mydb -u root -s ./sql`,
	RunE: run,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&container, "container", "c", "", "DockerコンテナIDまたはコンテナ名 (必須)")
	rootCmd.Flags().StringVarP(&dbName, "database", "d", "", "データベース名 (必須)")
	rootCmd.Flags().StringVarP(&dbUser, "user", "u", "", "データベースユーザー名 (必須)")
	rootCmd.Flags().StringVarP(&dbPass, "password", "p", "", "データベースパスワード")
	rootCmd.Flags().StringVarP(&sqlDir, "sql-dir", "s", "", "SQLファイルが格納されたローカルフォルダパス (必須)")

	rootCmd.MarkFlagRequired("container")
	rootCmd.MarkFlagRequired("database")
	rootCmd.MarkFlagRequired("user")
	rootCmd.MarkFlagRequired("sql-dir")
}

func run(cmd *cobra.Command, args []string) error {
	// SQLフォルダの存在確認
	if _, err := os.Stat(sqlDir); os.IsNotExist(err) {
		return fmt.Errorf("SQLフォルダが見つかりません: %s", sqlDir)
	}

	// コンテナの起動確認
	if err := checkContainer(container); err != nil {
		return err
	}

	// SQLファイル取得
	sqlFiles, err := findSQLFiles(sqlDir)
	if err != nil {
		return err
	}
	if len(sqlFiles) == 0 {
		return fmt.Errorf("SQLファイルが見つかりません: %s/*.sql", sqlDir)
	}

	fmt.Println("==========================================")
	fmt.Println(" Docker MySQL SQL 実行")
	fmt.Println("==========================================")
	fmt.Printf("  コンテナ   : %s\n", container)
	fmt.Printf("  DB名       : %s\n", dbName)
	fmt.Printf("  ユーザー   : %s\n", dbUser)
	fmt.Printf("  SQLフォルダ: %s\n", sqlDir)
	fmt.Printf("  ファイル数 : %d 件\n", len(sqlFiles))
	fmt.Println("==========================================")

	success, fail := 0, 0

	for _, sqlFile := range sqlFiles {
		filename := filepath.Base(sqlFile)
		fmt.Printf("\n[実行中] %s\n", filename)

		if err := execSQL(sqlFile, filename); err != nil {
			fmt.Printf("[FAIL] %s: %v\n", filename, err)
			fail++
		} else {
			fmt.Printf("[OK] %s\n", filename)
			success++
		}
	}

	fmt.Println("\n==========================================")
	fmt.Printf(" 完了: 成功=%d件 / 失敗=%d件\n", success, fail)
	fmt.Println("==========================================")

	if fail > 0 {
		return fmt.Errorf("%d件のSQLファイルが失敗しました", fail)
	}
	return nil
}

func checkContainer(name string) error {
	out, err := exec.Command("docker", "inspect", "--format", "{{.State.Running}}", name).Output()
	if err != nil {
		return fmt.Errorf("コンテナが見つかりません: %s", name)
	}
	if strings.TrimSpace(string(out)) != "true" {
		return fmt.Errorf("コンテナが起動していません: %s", name)
	}
	return nil
}

func findSQLFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("フォルダの読み取りに失敗しました: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.ToLower(filepath.Ext(e.Name())) == ".sql" {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)
	return files, nil
}

func execSQL(sqlFile, filename string) error {
	// ローカルファイルをコンテナの /tmp へコピー
	dest := container + ":/tmp/" + filename
	if err := exec.Command("docker", "cp", sqlFile, dest).Run(); err != nil {
		return fmt.Errorf("docker cp 失敗: %w", err)
	}

	// mysqlコマンドを組み立て
	mysqlCmd := fmt.Sprintf("mysql -u%s %s %s < /tmp/%s; rm /tmp/%s",
		dbUser, passwordFlag(), dbName, filename, filename)

	out, err := exec.Command("docker", "exec", container, "bash", "-c", mysqlCmd).CombinedOutput()
	if len(out) > 0 {
		// Warning行を除いた出力のみ表示
		printFilteredOutput(out)
	}
	if err != nil {
		return fmt.Errorf("mysql 実行失敗: %w", err)
	}
	return nil
}

func passwordFlag() string {
	if dbPass == "" {
		return ""
	}
	return "-p" + dbPass
}

func printFilteredOutput(out []byte) {
	for _, line := range strings.Split(string(out), "\n") {
		// mysqlのパスワード警告は表示しない
		if strings.Contains(line, "Using a password on the command line interface can be insecure") {
			continue
		}
		if strings.TrimSpace(line) != "" {
			fmt.Println(line)
		}
	}
}
