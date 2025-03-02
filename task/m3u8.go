package task

import (
	"bytes"
	"context"
	"fmt"
	"github.com/CCLooMi/sql-mak/mysql"
	"go.uber.org/fx"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
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
func getCodec(inputPath string) (string, string, error) {
	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)
	cmd2 := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "a:0",
		"-show_entries", "stream=codec_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputPath,
	)
	vCode, err := exeCmd(cmd)
	if err != nil {
		return "", "", err
	}
	aCode, err := exeCmd(cmd2)
	if err != nil {
		return "", "", err
	}
	return vCode, aCode, nil
}
func exeCmd(cmd *exec.Cmd) (string, error) {
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out.String()), nil
}

var vMap = map[string]bool{"h264": true, "hevc": true, "vp8": true, "vp9": true, "av1": true}
var aMap = map[string]bool{"aac": true, "ac3": true, "opus": true, "vorbis": true, "flac": true, "mp3": true, "pcm": true}

func convertToM3U8(ctx context.Context, inputPath string) error {
	vCode, aCode, err := getCodec(inputPath)
	if err != nil {
		return err
	}

	dir := filepath.Dir(inputPath)
	outputDir := filepath.Join(dir, "hls")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("create output dir error: %v", err)
	}
	outputPath := path.Join(outputDir, "0.m3u8")

	// 动态构建 FFmpeg 参数
	cmdArgs := []string{
		"-hwaccel", "auto",
		"-i", inputPath,
		// 映射所有流
		"-map", "0",
		// 字幕强制转 WebVTT（HLS 唯一支持的字幕格式）
		"-c:s", "webvtt",
	}

	// 视频编码处理（支持则复制，否则转 H.264）
	if vMap[vCode] {
		cmdArgs = append(cmdArgs, "-c:v", "copy")
		log.Printf("视频流使用原始编码: %s", vCode)
	} else {
		//vCodec := hasNvidiaNVENC()?"h264_nvenc":
		if hasNvidiaNVENC() {
			cmdArgs = append(cmdArgs, "-c:v", "h264_nvenc")
		} else {
			cmdArgs = append(cmdArgs, "-c:v", "libx264")
		}
		cmdArgs = append(cmdArgs,
			"-preset", "fast", // 平衡速度与质量
			"-sc_threshold", "0", // 强制场景切换生成关键帧
		)
		log.Printf("视频流转码为 H.264 (原编码: %s)", vCode)
	}

	// 音频编码处理（支持则复制，否则转 AAC）
	if aMap[aCode] {
		cmdArgs = append(cmdArgs, "-c:a", "copy")
		log.Printf("音频流使用原始编码: %s", aCode)
	} else {
		cmdArgs = append(cmdArgs,
			"-c:a", "aac",
		)
		log.Printf("音频流转码为 AAC (原编码: %s)", aCode)
	}
	cmdArgs = append(cmdArgs,
		// HLS 通用参数
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_flags", "independent_segments",
		"-f", "hls",
		outputPath,
	)
	cmd := exec.CommandContext(ctx, getFFmpegPath(), cmdArgs...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("Execute FFmpeg command error: %v\n output: %s", err, string(output))
	}

	// check output file
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return fmt.Errorf("expected output file not found: %s", outputPath)
	}

	return nil
}
func hasNvidiaNVENC() bool {
	// Windows: 检查 nvidia-smi 是否存在
	// Linux: 检查 /dev/nvidia* 设备或运行 nvidia-smi
	cmd := exec.Command("nvidia-smi")
	if err := cmd.Run(); err == nil {
		return true
	}
	return false
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
