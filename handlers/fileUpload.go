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
	"path"
	"path/filepath"
	"unsafe"
	"wios_server/conf"
	"wios_server/entity"
	"wios_server/middlewares"
	"wios_server/service"
	"wios_server/utils"
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

type FileInfo struct {
	Id   string `json:"id"`
	Size int64  `json:"size"`
	Type string `json:"type"`
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
	cmdMap    map[uint64]*UploadCommand
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
	if fa.Complete {
		//TODO
		return fa
	}
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
	fa.cmdMap[uint64(cmd.Start)] = cmd
	return fa
}
func (fa *FileAgent) CommandComplete(cmd *UploadCommand, data []byte) {
	fa.AgentFile.WriteAt(data, cmd.Start)
	fa.Uploaded += cmd.End - cmd.Start
	SetBSetPositionBit1(fa.BSet, cmd.Idx)
	log.Println("Received file data:",
		float64(fa.Uploaded)/float64(fa.Size)*100, "%")
	if fa.Uploaded >= fa.Size {
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
func (fa *FileAgent) Dispose() {
	if len(fa.Uploader) == 0 {
		if fa.AgentFile != nil {
			fa.AgentFile.Close()
		}
		delete(agentMap, fa.hexId())
	}
	fa.SaveToMetaData()
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
			cmd.Total)...)
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
			cmdMap:    make(map[uint64]*UploadCommand),
		}, nil
	}
	fa := &FileAgent{
		Size:      size,
		BasePath:  basePath,
		AgentFile: agentFile,
		Complete:  false,
		Uploader:  make(map[int64]bool),
		cmdMap:    make(map[uint64]*UploadCommand),
	}
	fa.LoadMetaData(metaData)
	return fa, nil
}
func PushFinishedCmd(bid []byte, cnn *websocket.Conn) {
	idLen := len(bid)
	bb := make([]byte, 4+idLen)
	binary.BigEndian.PutUint32(bb, uint32(idLen))
	copy(bb[4:], bid)
	pushBinMsg(bb, cnn)
}
func CheckExist(workDir string, fileInfo *FileInfo, cnn *websocket.Conn, cnnAddress int64, uploadServer *service.UploadService) {
	fid := fileInfo.Id
	bid, _ := hex.DecodeString(fid)
	basePath := path.Join(workDir, utils.GetFPathByFid(fid))
	if _, err := os.Stat(filepath.Join(basePath, "0")); err == nil {
		PushFinishedCmd(bid, cnn)
		//start save update file info
		fi := entity.Upload{
			FileType: &fileInfo.Type,
			FileName: &fileInfo.Name,
			FileSize: &fileInfo.Size,
		}
		fi.Id = new(string)
		*fi.Id = fileInfo.Id
		fi.UploadSize = new(int64)
		*fi.UploadSize = fileInfo.Size
		fi.DelFlag = new(bool)
		*fi.DelFlag = false
		uploadServer.SaveUpdate(&fi)
		//end save update file info
		return
	}
	fa := agentMap[fid]
	var err error
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

	//start save file info
	fi := entity.Upload{
		FileType: &fileInfo.Type,
		FileName: &fileInfo.Name,
		FileSize: &fileInfo.Size,
	}
	fi.Id = new(string)
	*fi.Id = fileInfo.Id
	fi.UploadSize = new(int64)
	*fi.UploadSize = fa.Uploaded
	uploadServer.SaveUpdate(&fi)
	//end save file info

	agentMap[fid] = fa
	cmd := &UploadCommand{Id: bid}
	fa.NextCommand(cmd)
	pushBinMsg(cmd.toBytes(), cnn)
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

func HandleFileUpload(app *gin.Engine, config *conf.Config, db *sql.DB, ut *utils.Utils) {
	path := config.FileServer.Path
	uploadServer := service.NewUploadService(db)
	app.GET(path, func(c *gin.Context) {
		// 获取客户端发送的协议参数
		protocols := c.Request.Header["Sec-Websocket-Protocol"]
		// 升级HTTP连接为WebSocket连接
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				if len(protocols) > 0 && "wstore" == protocols[0] {
					return middlewares.GetStoreUserInfo(c, ut) != nil
				}
				return middlewares.GetUserInfo(c, ut) != nil
			},
			Subprotocols: []string{"wstore", "PhotoZen"},
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			http.Error(c.Writer, err.Error(), http.StatusUnauthorized)
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
					break
				}
				onStrMsg(config.FileServer.SaveDir, &fileInfo, conn, address, uploadServer)
				continue
			}
			if msgType == websocket.BinaryMessage {
				onBinMsg(msg, conn, uploadServer)
				continue
			}
		}
	})
}
func onOpen(cnn *websocket.Conn, cnnAddress int64) {

}
func onClose(cnn *websocket.Conn, cnnAddress int64) {
	cnn.Close()
	fa := uploaderMap[cnnAddress]
	if fa != nil {
		delete(fa.Uploader, cnnAddress)
		delete(uploaderMap, cnnAddress)
		fa.Dispose()
	}
}
func onStrMsg(workDir string, fileInfo *FileInfo, cnn *websocket.Conn, cnnAddress int64, uploadServer *service.UploadService) {
	CheckExist(workDir, fileInfo, cnn, cnnAddress, uploadServer)
}

func onBinMsg(msg []byte, cnn *websocket.Conn, uploadServer *service.UploadService) {
	start := binary.BigEndian.Uint64(msg[:8])
	bidLen := binary.BigEndian.Uint32(msg[8 : 8+4])
	bid := msg[8+4 : 8+4+bidLen]
	id := hex.EncodeToString(bid)
	fa := agentMap[id]
	if fa == nil {
		return
	}
	cmd := fa.cmdMap[start]
	if cmd == nil {
		cmd = &UploadCommand{Id: bid}
	} else {
		fa.CommandComplete(cmd, msg[8+4+bidLen:])
	}
	//update upload size
	uploadServer.UpdateUploadSize(&id, &fa.Uploaded)
	if fa.Complete {
		PushFinishedCmd(bid, cnn)
		return
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
