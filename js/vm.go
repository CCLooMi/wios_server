package js

import (
	"context"
	"errors"
	"github.com/robertkrimen/otto"
	"log"
	"sync"
	"time"
	"wios_server/entity"
	"wios_server/utils"
)

var halt = errors.New("Stahp")

type VmManager struct {
	mu  sync.RWMutex
	vms map[string]*Vm
}

func NewVmManager() *VmManager {
	return &VmManager{
		vms: make(map[string]*Vm),
	}
}
func (mgr *VmManager) Add(vm *Vm) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	mgr.vms[vm.ID] = vm
}
func (mgr *VmManager) Get(id string) *Vm {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	return mgr.vms[id]
}
func (mgr *VmManager) Delete(id string) {
	mgr.mu.Lock()
	defer mgr.mu.Unlock()
	delete(mgr.vms, id)
}
func (mgr *VmManager) List() []map[string]interface{} {
	mgr.mu.RLock()
	defer mgr.mu.RUnlock()
	data := make([]map[string]interface{}, 0)
	for _, vm := range mgr.vms {
		data = append(data, map[string]interface{}{
			"id":        vm.ID,
			"title":     vm.Title,
			"user":      vm.User.Nickname,
			"startTime": vm.StartTime.Format("2006-01-02 15:04:05"),
		})
	}
	return data
}

var vmManager = NewVmManager()

type Vm struct {
	ID           string
	Title        *string
	User         *entity.User
	StartTime    time.Time
	otto         *otto.Otto
	cleanupOnce  sync.Once
	cleanupFuncs []func()
}

func NewVm(title *string, u *entity.User) *Vm {
	otto := otto.New()
	otto.Interrupt = make(chan func(), 1)
	ctx, cancel := context.WithCancel(context.Background())
	vm := &Vm{
		ID:           utils.UUID(),
		Title:        title,
		User:         u,
		StartTime:    time.Now(),
		otto:         otto,
		cleanupFuncs: []func(){cancel},
	}
	vmManager.Add(vm)
	ApplyExportsTo(otto)
	otto.Set("self", vm)
	otto.Set("ctx", ctx)
	return vm
}
func (vm *Vm) Set(key string, value interface{}) error {
	return vm.otto.Set(key, value)
}
func (vm *Vm) Finally(f func()) {
	if f == nil {
		return
	}
	vm.cleanupFuncs = append(vm.cleanupFuncs, safeFunc(f))
}
func (vm *Vm) cleanup() {
	vm.cleanupOnce.Do(func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Cleanup failed: %v", r)
			}
		}()
		for _, f := range vm.cleanupFuncs {
			f()
		}
		vmManager.Delete(vm.ID)
	})
}
func (vm *Vm) Exit() string {
	vm.otto.Interrupt <- func() {
		panic(halt)
	}
	vm.cleanup()
	return "vm[" + vm.ID + "] exited."
}
func (vm *Vm) Execute(script string) (otto.Value, error) {
	start := time.Now()
	defer func() {
		vm.cleanup()
		duration := time.Since(start)
		if caught := recover(); caught != nil {
			if caught == halt {
				log.Printf("JSVM[%s] Stopped after: %s", vm.ID, duration.String())
				return
			}
			log.Printf("JSVM[%s] Stopped after: %s,caught: %v", vm.ID, duration.String(), caught)
			return
		}
	}()
	return vm.otto.Run(script)
}
func StopVM(id string) string {
	vm := vmManager.Get(id)
	if vm == nil {
		return "vm[" + id + "] not found."
	}
	return vm.Exit()
}
func VMList() []map[string]interface{} {
	return vmManager.List()
}
func safeFunc(f func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("SafeFunc failed: %v", r)
			}
		}()
		f()
	}
}
