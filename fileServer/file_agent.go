package fileserver

import (
	"encoding/binary"
	"io/ioutil"
	"os"
	"path/filepath"
)

type FileAgent struct {
	Size      int
	BasePath  string
	BSet      []byte
	IStart    int
	AgentFile *os.File
	Complete  bool
	UpCount   int
}

const BlobSize = 524288

func StartLink(arg interface{}) error {
	// assuming GenServer is defined and StartLink function is available
	// e.g., `GenServer.StartLink(__MODULE__, arg)`
	return nil
}

func Init(state interface{}) (FileAgent, error) {
	fa := FileAgent{}
	fileInfo := state.(map[string]interface{})
	// id := fileInfo["id"].(string)
	size := fileInfo["size"].(int)
	basePath := fileInfo["base_path"].(string)

	err := os.MkdirAll(basePath, os.ModePerm)
	if err != nil {
		return fa, err
	}

	agentFile, err := os.OpenFile(filepath.Join(basePath, "0.td"), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fa, err
	}

	bSet, err := ioutil.ReadFile(filepath.Join(basePath, "up.meta"))
	if err == nil {
		fa = FileAgent{
			Size:      size,
			BasePath:  basePath,
			BSet:      bSet,
			IStart:    getIStart(bSet),
			AgentFile: agentFile,
			Complete:  true,
		}
	} else {
		bSet = newBSet(size)
		fa = FileAgent{
			Size:      size,
			BasePath:  basePath,
			BSet:      bSet,
			IStart:    0,
			AgentFile: agentFile,
			Complete:  false,
		}
	}

	return fa, nil
}

func getIStart(bSet []byte) int {
	max := int(bSet[0]) >> 6
	return getIStartHelper(bSet, 0, max, len(bSet)*8)
}

func getIStartHelper(bSet []byte, start, max, bSetBitSize int) int {
	if start >= max || max <= 0 {
		return -1
	} else {
		i := bSet[start] & 0x01
		if i == 1 {
			return getIStartHelper(bSet, start+1, max, bSetBitSize)
		} else {
			return start
		}
	}
}

func setPositionValue(bset []byte, position int, value byte) []byte {
	byteIndex := position / 8
	bitOffset := uint(position % 8)

	if value == 0 {
		bset[byteIndex] &^= 1 << bitOffset
	} else {
		bset[byteIndex] |= 1 << bitOffset
	}
	return bset
}

func newBSet(fSize int) []byte {
	asize := fSize >> 19
	if (fSize & 1048575) > 0 {
		sz := (asize + 1) >> 3
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(asize+1))
		if (asize+1)&7 > 0 {
			return append(buf, make([]byte, sz+1)...)
		} else {
			return append(buf, make([]byte, sz)...)
		}
	} else {
		sz := asize >> 3
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(asize))
		if asize&7 > 0 {
			return append(buf, make([]byte, sz+1)...)
		} else {
			return append(buf, make([]byte, sz)...)
		}
	}
}

func nextCommand(state FileAgent) (map[string]interface{}, FileAgent) {
	// fSize := state.Size
	bSet := state.BSet
	iStart := state.IStart
	blobSize := BlobSize
	i := getIStart(bSet)

	if i == -1 {
		ii := getIStart(bSet)
		if ii == i {
			return map[string]interface{}{
					"i":        iStart,
					"complete": 1,
				}, FileAgent{
					Size:      state.Size,
					BasePath:  state.BasePath,
					BSet:      state.BSet,
					IStart:    iStart,
					AgentFile: state.AgentFile,
					Complete:  true,
					UpCount:   state.UpCount,
				}
		} else {
			indexStart := ii * blobSize
			indexEnd := indexStart + blobSize
			if indexEnd > state.Size {
				indexEnd = state.Size
			}
			complete := float32(indexStart) / float32(state.Size)
			return map[string]interface{}{
					"i":        ii,
					"start":    indexStart,
					"end":      indexEnd,
					"complete": complete,
				}, FileAgent{
					Size:      state.Size,
					BasePath:  state.BasePath,
					BSet:      state.BSet,
					IStart:    ii + 1,
					AgentFile: state.AgentFile,
					Complete:  state.Complete,
					UpCount:   state.UpCount,
				}
		}
	} else {
		indexStart := i * blobSize
		indexEnd := indexStart + blobSize
		if indexEnd > state.Size {
			indexEnd = state.Size
		}
		complete := float32(indexStart) / float32(state.Size)
		return map[string]interface{}{
				"i":        i,
				"start":    indexStart,
				"end":      indexEnd,
				"complete": complete,
			}, FileAgent{
				Size:      state.Size,
				BasePath:  state.BasePath,
				BSet:      state.BSet,
				IStart:    i + 1,
				AgentFile: state.AgentFile,
				Complete:  state.Complete,
				UpCount:   state.UpCount,
			}
	}
}

func HandleInfo(msg interface{}, state FileAgent) (FileAgent, error) {
	switch msg := msg.(type) {
	case string:
		if msg == "shutdown" {
			err := state.AgentFile.Close()
			if err != nil {
				return state, err
			}
			basePath := state.BasePath
			if state.Complete {
				err = os.Rename(filepath.Join(basePath, "0.td"), filepath.Join(basePath, "0"))
				if err != nil {
					return state, err
				}
				err = os.Remove(filepath.Join(basePath, "up.meta"))
				if err != nil {
					return state, err
				}
			} else {
				err = ioutil.WriteFile(filepath.Join(basePath, "up.meta"), state.BSet, 0644)
				if err != nil {
					return state, err
				}
			}

			return state, nil
		}
	case []interface{}:
		if len(msg) > 0 {
			if atom, ok := msg[0].(string); ok && atom == "DOWN" {
				// pid := msg[3].(string)
				upCount := state.UpCount
				if upCount-1 < 1 {
					err := state.AgentFile.Close()
					if err != nil {
						return state, err
					}

					return state, nil
				} else {
					return FileAgent{
							Size:      state.Size,
							BasePath:  state.BasePath,
							BSet:      state.BSet,
							IStart:    state.IStart,
							AgentFile: state.AgentFile,
							Complete:  state.Complete,
							UpCount:   upCount - 1,
						},
						nil
				}
			}
		}
	}
	return state, nil
}
