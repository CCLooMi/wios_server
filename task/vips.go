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
	"strings"
	"time"
	"wios_server/entity"
	"wios_server/utils"
)

func startDZIServer(lc fx.Lifecycle, ut *utils.Utils) {
	flagId := utils.UUID()
	doFlag := func(flagId string) int64 {
		um := mysql.UPDATE(entity.Files{}, "f").
			SET("f.flag_id = ?", flagId).
			SET("f.flag_exp = DATE_ADD(NOW(6), INTERVAL ? SECOND)", 20).
			WHERE("f.status IS NULL").
			AND("f.file_type LIKE 'image/%'").
			AND("(f.flag_exp < NOW(6) OR f.flag_exp IS NULL)").
			LIMIT(3)
		um.LOGSQL(false)
		r := um.Execute(ut.Db).Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		return n
	}
	keepLock := func(flagId string) int64 {
		um := mysql.UPDATE(entity.Files{}, "f").
			SET("f.flag_exp = DATE_ADD(NOW(6), INTERVAL ? SECOND)", 20).
			WHERE("f.status IS NULL").
			AND("f.flag_id = ?", flagId).
			LIMIT(3)
		um.LOGSQL(false)
		r := um.Execute(ut.Db).Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		return n
	}
	setStatus := func(fid string, status string) int64 {
		um := mysql.UPDATE(entity.Files{}, "f").
			SET("f.status = ?", status).
			WHERE("f.file_id=?", fid)
		um.LOGSQL(false)
		r := um.Execute(ut.Db).Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		return n
	}
	var n int64
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			mainCtx, cancel := context.WithCancel(context.Background())
			sm := mysql.SELECT("f.file_id").
				FROM(entity.Files{}, "f").
				WHERE("f.flag_id =?", flagId).
				AND("f.status IS NULL").
				LIMIT(3)
			sm.LOGSQL(false)
			go func() {
				defer cancel()
				for {
					//execute core task
					if n = doFlag(flagId); n > 0 {
						fids := sm.Execute(ut.Db).GetResultAsList()
						for _, fid := range fids {
							fidStr := fid.(**string)
							fpath := path.Join(ut.Config.FileServer.SaveDir, utils.GetFPathByFid(**fidStr), "0")
							absPath, err := filepath.Abs(fpath)
							if err != nil {
								log.Printf("get absolute path error: %v", err)
								setStatus(**fidStr, "error: abs")
								continue
							}

							if err := genDzi(absPath); err != nil {
								log.Printf("create dzi error: %v", err)
								setStatus(**fidStr, "error: dzi")
								continue
							}

							setStatus(**fidStr, "done")
						}
					}

					//breakable sleep
					select {
					case <-mainCtx.Done():
						log.Println("DZI main task exit")
						return
					case <-time.After(10 * time.Second):
					}
				}
			}()

			go func() {
				for {
					if n > 0 {
						keepLock(flagId)
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
func genDzi(inputPath string) error {
	dir := filepath.Dir(inputPath)
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	baseName := strings.TrimSuffix(base, ext)
	outputBase := filepath.Join(dir, baseName)
	cmd := exec.Command(GetVipsPath(), "dzsave", inputPath, outputBase,
		"--layout", "dz",
		//"--depth", "onetile",
		"--tile-size", "512",
		"--suffix", ".jpg[Q=90]",
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("执行 vips 命令失败: %v, 输出: %s", err, string(out))
	}
	return nil
}
func GetVipsPath() string {
	// check default path
	if path, err := exec.LookPath("vips"); err == nil {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "vips"
	}
	basePath := path.Join(home, "/scoop/apps/libvips/current/bin")
	vipsPath := filepath.Join(basePath, "vips")
	_, err = os.Stat(vipsPath)
	if err != nil {
		return "vips"
	}
	return vipsPath
}

var Module = fx.Options(
	fx.Invoke(startDZIServer),
	fx.Invoke(startVideoProcessor),
)
