package task

import (
	"context"
	"encoding/json"
	"github.com/CCLooMi/sql-mak/mysql"
	"go.uber.org/fx"
	"log"
	"time"
	"wios_server/entity"
	"wios_server/utils"
)

func startUploadCleaner(lc fx.Lifecycle, ut *utils.Utils) {
	flagId := utils.UUID()
	batchSize := 1000
	doFlag := func(flagId string) int64 {
		um := mysql.UPDATE(entity.Upload{}, "u").
			SET("u.flag_id=?", flagId).
			SET("u.flag_exp = DATE_ADD(NOW(6), INTERVAL ? SECOND)", 20).
			WHERE("u.flag_exp < NOW(6) OR u.flag_exp IS NULL").
			LIMIT(batchSize)
		um.LOGSQL(false)
		r := um.Execute(ut.Db).Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		return n
	}
	keepLock := func(flagId string) int64 {
		um := mysql.UPDATE(entity.Upload{}, "u").
			SET("u.flag_exp = DATE_ADD(NOW(6), INTERVAL ? SECOND)", 20).
			WHERE("u.flag_id = ?", flagId).
			LIMIT(batchSize)
		um.LOGSQL(false)
		r := um.Execute(ut.Db).Update()
		n, err := r.RowsAffected()
		if err != nil {
			log.Println(err)
		}
		return n
	}
	delRecord := func(fids ...interface{}) int64 {
		um := mysql.DELETE().FROM(entity.Upload{}).
			WHERE_IN("id", fids...)
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
			sm := mysql.SELECT("u.id").
				FROM(entity.Upload{}, "u").
				WHERE("u.flag_id =?", flagId).
				LIMIT(batchSize)
			sm.LOGSQL(false)
			go func() {
				defer cancel()
				for {
					//execute core task
					if n = doFlag(flagId); n > 0 {
						fids := sm.Execute(ut.Db).GetResultAsList()
						delList := make([]interface{}, 0)
						for _, fid := range fids {
							fidStr := fid.(**string)
							if !ut.CheckFileExistByFid(**fidStr) {
								delList = append(delList, **fidStr)
							}
						}
						if len(delList) > 0 {
							n = delRecord(delList...)
							jsn, _ := json.Marshal(delList)
							log.Printf("Clearned upload record: %s", jsn)
						}
					}

					//breakable sleep
					select {
					case <-mainCtx.Done():
						log.Println("Upload cleaner task exit")
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
