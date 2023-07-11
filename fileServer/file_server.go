package fileserver

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

var saveDir string

func SaveDir() string {
	return saveDir
}
func CheckExist(fileInfo map[string]interface{}) (bool, interface{}) {
	fid := fileInfo["id"].(string)
	bid, _ := hex.DecodeString(fid)
	a := uint(bid[0]) >> 6
	b := uint(bid[1]) >> 6
	basePath := filepath.Join(SaveDir(), fmt.Sprintf("%d/%d/%s", a, b, fid))
	if _, err := os.Stat(filepath.Join(basePath, "0")); err == nil {
		return true, map[string]interface{}{
			"complete": 1,
		}
	} else {
		return false, StartFileAgent(fileInfo, basePath)
	}
}

func StartFileAgent(fileInfo map[string]interface{}, basePath string) interface{} {
	fid := fileInfo["id"].(string)
	bid, _ := hex.DecodeString(fid)
	pid := Lookup(bid)
	if pid == nil {
		// start FileAgentSupervisor and register the child process
		// assuming FileAgentSupervisor is defined and StartChild function is available
		// e.g., `pid, err := FileAgentSupervisor.StartChild(fileInfo)`
		// register(bid, pid)
		// return pid
		return nil
	} else {
		return pid
	}
}

func NextCommand(agentPid interface{}) map[string]interface{} {
	// assuming GenServer is defined and Call function is available
	// e.g., `result, err := GenServer.Call(agentPid, "next_command")`
	// handle error and return appropriate response
	return map[string]interface{}{
		"complete": -1,
	}
}

func WriteData(agentPid interface{}, i int, position int, data []byte) error {
	// assuming GenServer is defined and Cast function is available
	// e.g., `GenServer.Cast(agentPid, []interface{}{"write_data", i, position, data})`
	// handle error and return appropriate response
	return nil
}

func Register(name string, pid interface{}) {
	// assuming Syn is defined and Register function is available
	// e.g., `Syn.Register(SynTable(), name, pid)`
}

func RegisterWithMeta(name string, pid interface{}, meta interface{}) {
	// assuming Syn is defined and Register function is available
	// e.g., `Syn.Register(SynTable(), name, pid, meta)`
}

func Unregister(name string) {
	// assuming Syn is defined and Unregister function is available
	// e.g., `Syn.Unregister(SynTable(), name)`
}

func Lookup(name []byte) interface{} {
	// assuming Syn is defined and Lookup function is available
	// e.g., `result := Syn.Lookup(SynTable(), name)`
	// return result
	return nil
}
