package js

import (
	"errors"
	"github.com/robertkrimen/otto"
	"log"
	"time"
	"wios_server/entity"
	"wios_server/utils"
)

var halt = errors.New("Stahp")

var vmMap = make(map[string]*Vm)

type Vm struct {
	ID           string       `json:"id"`
	Title        *string      `json:"title"`
	User         *entity.User `json:"user"`
	StartTime    time.Time    `json:"start_time"`
	otto         *otto.Otto
	cleanupFuncs []func()
}

func (vm *Vm) SetTitle(t *string) {
	vm.Title = t
}
func (vm *Vm) Cleanup() {
	for _, f := range vm.cleanupFuncs {
		f()
	}
}
func (vm *Vm) Finally(f func()) {
	vm.cleanupFuncs = append(vm.cleanupFuncs, makFunc(f))
}
func (vm *Vm) Exit() string {
	vm.otto.Interrupt <- func() {
		panic(halt)
	}
	delete(vmMap, vm.ID)
	closeChannel(vm.otto.Interrupt)
	return "vm[" + vm.ID + "] exited."
}
func (vm *Vm) Set(key string, value interface{}) error {
	return vm.otto.Set(key, value)
}
func (vm *Vm) Execute(script string) (otto.Value, error) {
	start := time.Now()
	defer func() {
		duration := time.Since(start)
		if caught := recover(); caught != nil {
			if caught == halt {
				log.Fatalf("JSVM[%s] Stopping after: %s", vm.ID, duration.String())
				return
			}
			// Something else happened, repanic!
			panic(caught)
		}
	}()
	return vm.otto.Run(script)
}
func NewVm(title *string, u *entity.User) *Vm {
	otto := otto.New()
	otto.Interrupt = make(chan func(), 1)
	vm := &Vm{
		Title:        title,
		ID:           utils.UUID(),
		User:         u,
		StartTime:    time.Now(),
		otto:         otto,
		cleanupFuncs: make([]func(), 0),
	}
	vmMap[vm.ID] = vm
	ApplyExportsTo(otto)
	otto.Set("self", vm)
	return vm
}
func StopVM(id string) string {
	vm := vmMap[id]
	if vm == nil {
		return "vm[" + id + "] not found."
	}
	return vm.Exit()
}
func VMList() []map[string]interface{} {
	data := make([]map[string]interface{}, 0)
	for k, v := range vmMap {
		data = append(data, map[string]interface{}{
			"id":        k,
			"title":     v.Title,
			"user":      v.User.Nickname,
			"startTime": v.StartTime.Format("2006-01-02 15:04:05"),
		})
	}
	return data
}
func closeChannel(c chan func()) {
	select {
	case c <- func() {}:
		close(c)
	default:
	}
}
func makFunc(f func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("Run Cleanup func failed: %v\n", r)
			}
		}()
		f()
	}
}
