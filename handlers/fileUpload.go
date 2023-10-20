package handlers

import (
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"wios_server/conf"
)

func HandleFileUpload(db *sql.DB) func(ctx *gin.Context) {
	return func(c *gin.Context) {
		// 升级HTTP连接为WebSocket连接
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				//允许所有来源
				return true
			},
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Println("Failed to upgrade connection to WebSocket:", err)
			return
		}
		defer conn.Close()

		for {
			// 读取消息
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failed to read message from WebSocket:", err)
				break
			}
			// 处理消息
			log.Println("Received message:", string(msg))
			if msgType == websocket.TextMessage {
				var fileInfo = FileInfo{}
				err := json.Unmarshal(msg, &fileInfo)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				go processStrMsg(&fileInfo, conn)
			}
			if msgType == websocket.BinaryMessage {
				go processBinMsg(msg, conn)
			}
		}
	}
}
func pushStrMsg(msg string, cnn *websocket.Conn) error {
	return cnn.WriteMessage(websocket.TextMessage, []byte(msg))
}
func pushBinMsg(msg []byte, cnn *websocket.Conn) error {
	return cnn.WriteMessage(websocket.BinaryMessage, msg)
}
func processStrMsg(fileInfo *FileInfo, cnn *websocket.Conn) {
	isExist, data := CheckExist(fileInfo)
	if isExist {
		pushBinMsg(data, cnn)
		return
	}
	pushBinMsg(data, cnn)
}
func processBinMsg(msg []byte, cnn *websocket.Conn) {
	bid := msg[:8]
	id := hex.EncodeToString(bid)
	fa := agentMap[id]

	cmd := &UploadCommand{Id: bid}
	fa.NextCommand(cmd)
	pushBinMsg(cmd.toBytes(), cnn)
}

type FileInfo struct {
	Id   string `json:"id"`
	Size int64  `json:"size"`
	Name string `json:"name"`
}
type FileAgent struct {
	Size      int64
	BasePath  string
	BSet      []byte
	IStart    int64
	AgentFile *os.File
	Complete  bool
	UpCount   int64
}

func (fa *FileAgent) NextCommand(cmd *UploadCommand) *FileAgent {
	i := GetIStart(fa.BSet, fa.IStart)
	fa.IStart = i + 1
	cmd.Start = i * BlobSize
	cmd.End = cmd.Start + BlobSize
	if cmd.End > fa.Size {
		cmd.End = fa.Size
	}
	cmd.Uploaded = fa.UpCount
	cmd.Total = fa.Size
	return fa
}
func (fa *FileAgent) CommandComplete(cmd *UploadCommand, data []byte) {
	fa.Complete = true

}

type UploadCommand struct {
	Id       []byte
	Start    int64
	End      int64
	Uploaded int64
	Total    int64
}

func (cmd *UploadCommand) toBytes() []byte {
	return append(cmd.Id,
		Int64ToBytes(cmd.Start,
			cmd.End,
			cmd.Uploaded,
			cmd.Total)...)
}
func Int64ToBytes(values ...int64) []byte {
	var result []byte
	for _, value := range values {
		bytesBuffer := make([]byte, 8)
		binary.BigEndian.PutUint64(bytesBuffer, uint64(value))
		result = append(result, bytesBuffer...)
	}
	return result
}

const BlobSize = 524288

var agentMap = make(map[string]*FileAgent)

func CheckExist(fileInfo *FileInfo) (bool, []byte) {
	fid := fileInfo.Id
	bid, _ := hex.DecodeString(fid)
	a := uint(bid[0]) >> 6
	b := uint(bid[1]) >> 6
	basePath := filepath.Join(conf.Cfg.FileServer.SaveDir, fmt.Sprintf("%d/%d/%s", a, b, fid))
	if _, err := os.Stat(filepath.Join(basePath, "0")); err == nil {
		return true, bid
	} else {
		fa := agentMap[fid]
		if fa == nil {
			fa, err = NewFileAgen(fileInfo, basePath)
			if err != nil {
				panic(err)
			}
		}
		cmd := UploadCommand{Id: bid}
		fa.NextCommand(&cmd)
		return false, cmd.toBytes()
	}
}
func NewFileAgen(fileInfo *FileInfo, basePath string) (*FileAgent, error) {
	size := fileInfo.Size
	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	agentFile, err := os.OpenFile(filepath.Join(basePath, "0.td"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	bSet, err := os.ReadFile(filepath.Join(basePath, "up.meta"))
	if err == nil {
		return &FileAgent{
			Size:      size,
			BasePath:  basePath,
			BSet:      bSet,
			IStart:    GetIStart(bSet, 0),
			AgentFile: agentFile,
			Complete:  true,
		}, nil
	}
	bSet = NewBSet(size)
	return &FileAgent{
		Size:      size,
		BasePath:  basePath,
		BSet:      bSet,
		IStart:    0,
		AgentFile: agentFile,
		Complete:  false,
	}, nil
}

var iSetMap = []byte{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2,
	3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3, 3,
	4, 4, 4, 4, 4, 4, 4, 4, 5, 5, 5, 5, 6, 6, 7, 255}

func GetIStart(bSet []byte, iStart int64) int64 {
	max := int64(binary.BigEndian.Uint64(bSet[:8]))
	if max <= 0 {
		return -1
	}
	bSet = bSet[8:]
	bSetSize := int64(len(bSet))
	var i int64
	for i = iStart >> 3; i < bSetSize; i++ {
		if i >= max {
			return -1
		}
		bi := bSet[i] | (255 << (8 - (iStart & 7)))
		if bi != 255 {
			return int64((i << 3) + int64(iSetMap[bi]))
		}
	}
	return -1
}
func SetPositionValue(bSet []byte, i int64) {
	bSet[i>>3] |= 1 << (8 - (i & 7))
}

func NewBSet(fSize int64) []byte {
	aSize := fSize >> 19
	buf := make([]byte, 8)
	//1048575 = (1<<20) -1
	if (fSize & 1048575) > 0 {
		sz := (aSize + 1) >> 3
		binary.BigEndian.PutUint64(buf, uint64(aSize+1))
		if (aSize+1)&7 > 0 {
			return append(buf, make([]byte, sz+1)...)
		} else {
			return append(buf, make([]byte, sz)...)
		}
	} else {
		sz := aSize >> 3
		binary.BigEndian.PutUint64(buf, uint64(aSize))
		if aSize&7 > 0 {
			return append(buf, make([]byte, sz+1)...)
		} else {
			return append(buf, make([]byte, sz)...)
		}
	}
}
