package js

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"go.uber.org/fx"
	"log"
	"math"
	"regexp"
	"sync"
	"time"
	"wios_server/conf"
)

type Webot struct {
	bot      *openwechat.Bot
	handlers []func(msg *openwechat.Message)
	mu       sync.Mutex
}

func NewWebot(bot *openwechat.Bot) *Webot {
	wb := &Webot{bot,
		make([]func(msg *openwechat.Message), 0),
		sync.Mutex{},
	}
	bot.MessageHandler = wb.handleMessage
	return wb
}
func (w *Webot) Login(uname *string, ts int64) (string, error) {
	timeout := time.Duration(ts) * time.Second
	urlChan := make(chan string, 1)
	errChan := make(chan error, 1)
	w.bot.UUIDCallback = func(uuid string) {
		select {
		case urlChan <- "https://login.weixin.qq.com/qrcode/" + uuid:
		default: // prevent deadlock
		}
	}
	go func() {
		defer close(urlChan)
		defer close(errChan)
		u, err := w.GetCurrentUser()
		if err != nil {
			if loginErr := w.bot.Login(); loginErr != nil {
				errChan <- loginErr
			}
			return
		}
		if uname == nil || *uname == "" || u.NickName == *uname {
			urlChan <- ""
			return
		}
		if logoutErr := w.bot.Logout(); logoutErr != nil {
			errChan <- logoutErr
			return
		}
		if loginErr := w.bot.Login(); loginErr != nil {
			errChan <- loginErr
		}
	}()
	select {
	case url := <-urlChan:
		return url, nil
	case err := <-errChan:
		return "", err
	case <-time.After(timeout):
		return "", fmt.Errorf("timeout waiting for login URL")
	}
}
func (w *Webot) GetCurrentUser() (*openwechat.Self, error) {
	return w.bot.GetCurrentUser()
}
func (w *Webot) SendText(msg string, to string, update bool) ([]*openwechat.SentMessage, error) {
	self, err := w.bot.GetCurrentUser()
	if err != nil {
		return nil, err
	}
	var ss = make([]*openwechat.SentMessage, 0)
	reg, err := regexp.Compile(to)
	if err != nil {
		return nil, err
	}
	gs, err := self.Groups(update)
	if err != nil {
		return nil, err
	}
	gs.Search(math.MaxInt, func(group *openwechat.Group) bool {
		if reg.Match([]byte(group.NickName)) ||
			reg.Match([]byte(group.RemarkName)) ||
			group.UserName == to {
			sm, err := group.SendText(msg)
			if err != nil {
				return false
			}
			ss = append(ss, sm)
			return true
		}
		return false
	})
	fs, err := self.Friends(update)
	if err != nil {
		return nil, err
	}
	fs.Search(math.MaxInt, func(friend *openwechat.Friend) bool {
		if reg.Match([]byte(friend.NickName)) ||
			reg.Match([]byte(friend.RemarkName)) ||
			friend.UserName == to {
			sm, err := friend.SendText(msg)
			if err != nil {
				return false
			}
			ss = append(ss, sm)
			return true
		}
		return false
	})
	return ss, nil
}
func (w *Webot) OnMsg(handler func(msg *openwechat.Message), vm *Vm) {
	if handler == nil || vm == nil {
		return
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers = append(w.handlers, handler)
	index := len(w.handlers) - 1
	vm.Finally(func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		if index < len(w.handlers) && w.handlers[index] != nil {
			w.handlers[index] = nil
		}
		if len(w.handlers) > 0 && float64(w.nilCount())/float64(len(w.handlers)) > 0.3 {
			w.compactHandlers()
		}
	})
}
func (w *Webot) handleMessage(msg *openwechat.Message) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for _, handler := range w.handlers {
		if handler != nil {
			handler(msg)
		}
	}
}

func (w *Webot) nilCount() int {
	count := 0
	for _, handler := range w.handlers {
		if handler == nil {
			count++
		}
	}
	return count
}

func (w *Webot) compactHandlers() {
	if len(w.handlers) == 0 {
		return
	}
	newHandlers := make([]func(msg *openwechat.Message), 0, len(w.handlers))
	for _, handler := range w.handlers {
		if handler != nil {
			newHandlers = append(newHandlers, handler)
		}
	}
	w.handlers = newHandlers
}
func newWechatBot(config *conf.Config) *openwechat.Bot {
	bot := openwechat.DefaultBot(openwechat.Desktop)
	bot.MessageHandler = func(msg *openwechat.Message) {
		jss, err := json.Marshal(msg)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(string(jss))
		}
		if msg.IsText() && msg.Content == "ping" {
			msg.ReplyText("pong")
		}
	}
	bot.UUIDCallback = openwechat.PrintlnQrcodeUrl
	return bot
}
func startWechatBot(lc fx.Lifecycle, bot *openwechat.Bot, config *conf.Config) *openwechat.Bot {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				err := bot.Login()
				if err != nil {
					return
				}
				bot.Block()
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			go func() {
				bot.Exit()
				log.Println("WechatBot stopped.")
			}()
			return nil
		},
	})
	return bot
}

var webotModule = fx.Options(
	fx.Provide(newWechatBot),
	fx.Invoke(
		startWechatBot,
		func(bot *openwechat.Bot) {
			RegExport("webot", NewWebot(bot))
		},
	),
)
