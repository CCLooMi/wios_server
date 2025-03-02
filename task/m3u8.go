package task

import (
	"context"
	"fmt"
	"github.com/CCLooMi/sql-mak/mysql"
	"go.uber.org/fx"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"
	"wios_server/entity"
	"wios_server/utils"
)

func startVideoProcessor(lc fx.Lifecycle, ut *utils.Utils) {
	flagID := utils.UUID()

	// 标记需要处理的视频文件
	doFlag := func(flagID string) int64 {
		um := mysql.UPDATE(entity.Files{}, "f").
			SET("f.flag_id = ?", flagID).
			SET("f.flag_exp = DATE_ADD(NOW(6), INTERVAL ? SECOND)", 20). // 锁定10秒
			WHERE("f.status IS NULL").
			AND("f.file_type LIKE 'video/%'"). // 只处理视频文件
			AND("(f.flag_exp < NOW(6) OR f.flag_exp IS NULL)").
			LIMIT(3) // 每次处理3个
		um.LOGSQL(false)
		r := um.Execute(ut.Db).Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Printf("标记视频文件失败: %v", err)
		}
		return n
	}

	// 保持处理锁
	keepLock := func(flagID string) int64 {
		um := mysql.UPDATE(entity.Files{}, "f").
			SET("f.flag_exp = DATE_ADD(NOW(6), INTERVAL ? SECOND)", 20).
			WHERE("f.status IS NULL").
			AND("f.flag_id = ?", flagID).
			LIMIT(3)
		um.LOGSQL(false)
		r := um.Execute(ut.Db).Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Printf("保持视频处理锁失败: %v", err)
		}
		return n
	}

	// 更新文件状态
	setStatus := func(fid string, status string) int64 {
		um := mysql.UPDATE(entity.Files{}, "f").
			SET("f.status = ?", status).
			WHERE("f.file_id=?", fid)
		um.LOGSQL(false)
		r := um.Execute(ut.Db).Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Printf("更新状态失败: %v", err)
		}
		return n
	}

	var n int64
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mainCtx, cancel := context.WithCancel(context.Background())
			sm := mysql.SELECT("f.file_id").
				FROM(entity.Files{}, "f").
				WHERE("f.flag_id =?", flagID).
				AND("f.status IS NULL").
				LIMIT(3)
			sm.LOGSQL(false)
			go func() {
				defer cancel()
				for {
					if n = doFlag(flagID); n > 0 {
						results := sm.Execute(ut.Db).GetResultAsList()
						for _, fid := range results {
							fidStr := fid.(**string)
							fpath := path.Join(ut.Config.FileServer.SaveDir, utils.GetFPathByFid(**fidStr), "0")
							absPath, err := filepath.Abs(fpath)
							if err != nil {
								log.Printf("get absolute path error: %v", err)
								setStatus(**fidStr, "error: abs")
								continue
							}

							if err := convertToM3U8(mainCtx, absPath); err != nil {
								log.Printf("convert to m3u8 error: %v", err)
								setStatus(**fidStr, "error: vdo")
								continue
							}

							setStatus(**fidStr, "done")
						}
					}

					select {
					case <-mainCtx.Done():
						log.Println("m3u8 task stopped")
						return
					case <-time.After(10 * time.Second):
					}
				}
			}()

			go func() {
				for {
					if n > 0 {
						keepLock(flagID)
					}
					select {
					case <-mainCtx.Done():
						return
					case <-time.After(10 * time.Second):
					}
				}
			}()

			lc.Append(fx.Hook{
				OnStop: func(ctx context.Context) error {
					cancel()
					return nil
				},
			})

			return nil
		},
	})
}

func convertToM3U8(ctx context.Context, inputPath string) error {
	dir := filepath.Dir(inputPath)
	outputDir := filepath.Join(dir, "hls")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir error: %v", err)
	}
	outputPath := path.Join(outputDir, "0.m3u8")
	cmd := exec.CommandContext(ctx,
		getFFmpegPath(),
		"-i", inputPath,
		"-codec:", "copy", // 保持原始编码
		"-start_number", "0", // 分片从0开始
		"-hls_time", "10", // 每个分片10秒
		"-hls_list_size", "0", // 保留所有分片记录
		"-f", "hls", // 输出格式
		outputPath,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Execute FFmpeg command error: %v\n output: %s", err, string(output))
	}

	// check output file
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("expected output file not found: %s", outputPath)
	}

	return nil
}

func getFFmpegPath() string {
	// check default path
	if path, err := exec.LookPath("ffmpeg"); err == nil {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "ffmpeg"
	}
	basePath := path.Join(home, "/scoop/apps/ffmpeg/current/bin")
	ffmpegPath := filepath.Join(basePath, "ffmpeg")
	_, err = os.Stat(ffmpegPath)
	if err != nil {
		return "ffmpeg"
	}
	return ffmpegPath
}
