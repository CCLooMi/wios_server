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
	"unsafe"
	"wios_server/conf"
)

const BlobSize = 524288

var agentMap = make(map[string]*FileAgent)
var uploaderMap = make(map[int64]*FileAgent)
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
var cmdMap = make(map[int64]*UploadCommand)

type FileInfo struct {
	Id   string `json:"id"`
	Size int64  `json:"size"`
	Name string `json:"name"`
}
type FileAgent struct {
	Id        []byte
	Size      int64
	BasePath  string
	BSet      []byte
	IStart    int64
	AgentFile *os.File
	Complete  bool
	Uploaded  int64
	Uploader  map[int64]bool
}
type UploadCommand struct {
	Id       []byte
	Start    int64
	End      int64
	Uploaded int64
	Total    int64
	Idx      int64
}

func (fa *FileAgent) NextCommand(cmd *UploadCommand) *FileAgent {
	i := GetIStart(fa.BSet, fa.IStart)
	fa.IStart = i + 1
	cmd.Start = i * BlobSize
	cmd.End = cmd.Start + BlobSize
	if cmd.End > fa.Size {
		cmd.End = fa.Size
	}
	cmd.Uploaded = fa.Uploaded
	cmd.Total = fa.Size
	cmd.Idx = i
	return fa
}
func (fa *FileAgent) CommandComplete(cmd *UploadCommand, data []byte) {
	fa.AgentFile.WriteAt(data, cmd.Start)
	fa.Uploaded += cmd.End - cmd.Start
	SetBSetPositionBit1(fa.BSet, cmd.Idx)
	log.Println("Received file data:",
		float64(fa.Uploaded)/float64(fa.Size)*100, "%")
	if fa.Uploaded == fa.Size {
		//rename file
		fa.AgentFile.Close()
		newName := filepath.Join(fa.BasePath, "0")
		os.Rename(filepath.Join(fa.BasePath, "0.td"), newName)
		fa.AgentFile = nil
		//remove up.meta
		os.Remove(fa.GetBSetFileName())
		fa.Complete = true
		delete(agentMap, cmd.hexId())
		log.Println(fa.BSet)
	}
}
func (fa *FileAgent) GetBSetFileName() string {
	return filepath.Join(fa.BasePath, "up.meta")
}
func (fa *FileAgent) SaveToMetaData() {
	data := make([]byte, 8+len(fa.BSet))
	binary.BigEndian.PutUint64(data, uint64(fa.Uploaded))
	copy(data[8:], fa.BSet)
	os.WriteFile(fa.GetBSetFileName(), data, 0644)
}
func (fa *FileAgent) LoadMetaData(metaData []byte) {
	fa.Uploaded = int64(binary.BigEndian.Uint64(metaData[:8]))
	fa.BSet = metaData[8:]
	fa.IStart = GetIStart(fa.BSet, 0)
}
func (fa *FileAgent) hexId() string {
	return hex.EncodeToString(fa.Id)
}
func (cmd *UploadCommand) toBytes() []byte {
	idLen := len(cmd.Id)
	bb := make([]byte, 4+idLen)
	binary.BigEndian.PutUint32(bb, uint32(idLen))
	copy(bb[4:], cmd.Id)
	return append(
		bb,
		int64ToBytes(cmd.Start,
			cmd.End,
			cmd.Uploaded,
			cmd.Total,
			cmd.Idx)...)
}
func (cmd *UploadCommand) hexId() string {
	return hex.EncodeToString(cmd.Id)
}
func int64ToBytes(values ...int64) []byte {
	var result []byte
	for _, value := range values {
		bytesBuffer := make([]byte, 8)
		binary.BigEndian.PutUint64(bytesBuffer, uint64(value))
		result = append(result, bytesBuffer...)
	}
	return result
}
func NewFileAgen(fileInfo *FileInfo, basePath string, bid []byte) (*FileAgent, error) {
	size := fileInfo.Size
	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	agentFile, err := os.OpenFile(filepath.Join(basePath, "0.td"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	metaData, err := os.ReadFile(filepath.Join(basePath, "up.meta"))
	if err != nil {
		bSet := NewBSet(size)
		return &FileAgent{
			Id:        bid,
			Size:      size,
			BasePath:  basePath,
			BSet:      bSet,
			IStart:    0,
			AgentFile: agentFile,
			Complete:  false,
			Uploaded:  0,
			Uploader:  make(map[int64]bool),
		}, nil
	}
	fa := &FileAgent{
		Size:      size,
		BasePath:  basePath,
		AgentFile: agentFile,
		Complete:  false,
		Uploader:  make(map[int64]bool),
	}
	fa.LoadMetaData(metaData)
	return fa, nil
}
func CheckExist(fileInfo *FileInfo, cnn *websocket.Conn, cnnAddress int64) {
	fid := fileInfo.Id
	bid, _ := hex.DecodeString(fid)
	a := uint(bid[0]) >> 6
	b := uint(bid[1]) >> 6
	basePath := filepath.Join(conf.Cfg.FileServer.SaveDir, fmt.Sprintf("%d/%d/%s", a, b, fid))
	if _, err := os.Stat(filepath.Join(basePath, "0")); err == nil {
		pushBinMsg(bid, cnn)
	} else {
		fa := agentMap[fid]
		if fa == nil {
			fa, err = NewFileAgen(fileInfo, basePath, bid)
			if err != nil {
				panic(err)
			}
		}
		refFa := uploaderMap[cnnAddress]
		if refFa != nil {
			delete(refFa.Uploader, cnnAddress)
		}
		uploaderMap[cnnAddress] = fa
		fa.Uploader[cnnAddress] = true

		agentMap[fid] = fa
		cmd := &UploadCommand{Id: bid}
		fa.NextCommand(cmd)
		cmdMap[cnnAddress] = cmd
		pushBinMsg(cmd.toBytes(), cnn)
	}
}
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
	bSet[i>>3] |= 1 << (7 - (i & 7))
}
func SetBSetPositionBit1(bSet []byte, i int64) {
	SetPositionValue(bSet, 64+i)
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
		address := int64(uintptr(unsafe.Pointer(conn)))
		defer onClose(conn, address)
		onOpen(conn, address)
		for {
			// 读取消息
			msgType, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Failed to read message from WebSocket:", err)
				break
			}
			// 处理消息
			if msgType == websocket.TextMessage {
				log.Println("Received string message:", string(msg))
				var fileInfo = FileInfo{}
				err := json.Unmarshal(msg, &fileInfo)
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				onStrMsg(&fileInfo, conn, address)
			}
			if msgType == websocket.BinaryMessage {
				onBinMsg(msg, conn, address)
			}
		}
	}
}
func onOpen(cnn *websocket.Conn, cnnAddress int64) {

}
func onClose(cnn *websocket.Conn, cnnAddress int64) {
	cnn.Close()
	fa := uploaderMap[cnnAddress]
	if fa != nil {
		delete(fa.Uploader, cnnAddress)
		delete(uploaderMap, cnnAddress)
		if len(fa.Uploader) == 0 {
			if fa.AgentFile != nil {
				fa.AgentFile.Close()
			}
			delete(agentMap, fa.hexId())
		}
		fa.SaveToMetaData()
	}
	//cancel command
	delete(cmdMap, cnnAddress)
}
func onStrMsg(fileInfo *FileInfo, cnn *websocket.Conn, cnnAddress int64) {
	CheckExist(fileInfo, cnn, cnnAddress)
}
func onBinMsg(msg []byte, cnn *websocket.Conn, cnnAddress int64) {
	bidLen := binary.BigEndian.Uint32(msg[:4])
	bid := msg[4 : 4+bidLen]
	id := hex.EncodeToString(bid)
	fa := agentMap[id]

	cmd := cmdMap[cnnAddress]

	if cmd == nil {
		cmd = &UploadCommand{Id: bid}
		cmdMap[cnnAddress] = cmd
	} else {
		fa.CommandComplete(cmd, msg[4+bidLen:])
	}
	fa.NextCommand(cmd)
	pushBinMsg(cmd.toBytes(), cnn)
}
func pushStrMsg(msg string, cnn *websocket.Conn) error {
	return cnn.WriteMessage(websocket.TextMessage, []byte(msg))
}
func pushBinMsg(msg []byte, cnn *websocket.Conn) error {
	return cnn.WriteMessage(websocket.BinaryMessage, msg)
}
