package js

import (
	"github.com/eatmoreapple/openwechat"
	"math"
	"regexp"
	"sync"
)

type Webot struct {
	bot      *openwechat.Bot
	handlers []func(msg *openwechat.Message)
	mu       sync.Mutex
}

func NewWebot(bot *openwechat.Bot) *Webot {
	return &Webot{bot,
		make([]func(msg *openwechat.Message), 0),
		sync.Mutex{},
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
func (w *Webot) OnMsg(handler func(msg *openwechat.Message)) func() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers = append(w.handlers, handler)
	index := len(w.handlers) - 1
	return func() {
		w.mu.Lock()
		defer w.mu.Unlock()
		if index < len(w.handlers) && w.handlers[index] != nil {
			w.handlers[index] = nil
		}
		if len(w.handlers) > 0 && float64(w.nilCount())/float64(len(w.handlers)) > 0.3 {
			w.compactHandlers()
		}
	}
}
func (w *Webot) HandleMessage(msg *openwechat.Message) {
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
