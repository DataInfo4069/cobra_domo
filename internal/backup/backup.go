package backup

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"cobra_domo/config"
)

type BackupResult struct {
	Filepath  string
	Size      int64
	Duration  time.Duration
	Success   bool
	Error     error
	Timestamp time.Time
}

func ExecuteBackup(ctx context.Context) (*BackupResult, error) {
	start := time.Now()
	result := &BackupResult{
		Timestamp: time.Now(),
	}

	defer func() {
		result.Duration = time.Since(start)
	}()

	// 确保备份目录存在
	backupDir := config.Config.Backup.Path
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		result.Error = fmt.Errorf("创建备份目录失败: %w", err)
		return result, result.Error
	}

	// 生成备份文件名
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("backup_%s_%s", config.Config.Database.Name, timestamp)

	if config.Config.Backup.Compress {
		filename += ".sql.gz"
	} else {
		filename += ".sql"
	}

	filepath := filepath.Join(backupDir, filename)

	// 执行备份
	var cmd *exec.Cmd

	dbConfig := config.Config.Database
	switch strings.ToLower(dbConfig.Type) {
	case "mysql", "mariadb":
		cmd = createMySQLBackupCommand(dbConfig, filepath)
	case "postgres", "postgresql":
		cmd = createPostgreSQLBackupCommand(dbConfig, filepath)
	default:
		result.Error = fmt.Errorf("不支持的数据库类型: %s", dbConfig.Type)
		return result, result.Error
	}

	// 执行命令
	slog.Info("开始执行数据库备份", "database", dbConfig.Name, "type", dbConfig.Type)
	output, err := cmd.CombinedOutput()

	if err != nil {
		result.Error = fmt.Errorf("备份失败: %w\n输出: %s", err, output)
		slog.Error("备份失败", "error", err, "output", string(output))
		return result, result.Error
	}

	// 获取文件信息
	if info, err := os.Stat(filepath); err == nil {
		result.Size = info.Size()
	}

	result.Filepath = filepath
	result.Success = true

	slog.Info("备份成功",
		"file", filepath,
		"size", FormatFileSize(result.Size),
		"duration", result.Duration)

	// 清理旧备份
	if config.Config.Backup.Retention > 0 {
		cleaned := cleanupOldBackups(backupDir)
		if cleaned > 0 {
			slog.Info("清理旧备份", "count", cleaned)
		}
	}

	return result, nil
}

func createMySQLBackupCommand(db config.DatabaseConfig, outputPath string) *exec.Cmd {
	args := []string{
		"-h", db.Host,
		"-P", fmt.Sprintf("%d", db.Port),
		"-u", db.User,
		fmt.Sprintf("-p%s", db.Password),
		db.Name,
	}

	if config.Config.Backup.Compress {
		// 使用管道：mysqldump | gzip > 文件
		cmdStr := fmt.Sprintf("mysqldump %s | gzip > %s",
			strings.Join(args[1:], " "), // 跳过 mysqldump 命令本身
			outputPath)
		return exec.Command("bash", "-c", cmdStr)
	}

	args = append([]string{"mysqldump"}, args...)
	args = append(args, "-r", outputPath)
	return exec.Command(args[0], args[1:]...)
}

func createPostgreSQLBackupCommand(db config.DatabaseConfig, outputPath string) *exec.Cmd {
	os.Setenv("PGPASSWORD", db.Password)

	args := []string{
		"-h", db.Host,
		"-p", fmt.Sprintf("%d", db.Port),
		"-U", db.User,
		"-d", db.Name,
		"-f", outputPath,
	}

	if config.Config.Backup.Compress {
		// PostgreSQL 支持内置压缩
		args = append(args, "-Z", "9")
	}

	return exec.Command("pg_dump", args...)
}

func cleanupOldBackups(backupDir string) int {
	retentionDays := config.Config.Backup.Retention
	if retentionDays <= 0 {
		return 0
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		slog.Error("读取备份目录失败", "error", err)
		return 0
	}

	cleaned := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// 检查是否是备份文件
		if !strings.HasPrefix(info.Name(), "backup_") {
			continue
		}

		// 检查是否过期
		if info.ModTime().Before(cutoffTime) {
			filePath := filepath.Join(backupDir, info.Name())
			if err := os.Remove(filePath); err == nil {
				cleaned++
				slog.Debug("删除旧备份文件", "file", info.Name())
			}
		}
	}

	return cleaned
}

func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
