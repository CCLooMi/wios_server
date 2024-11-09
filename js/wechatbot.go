package js

import (
	"github.com/eatmoreapple/openwechat"
	"math"
	"regexp"
)

type Webot struct {
	bot *openwechat.Bot
}

func NewWebot(bot *openwechat.Bot) *Webot {
	return &Webot{bot}
}

func (w *Webot) GetCurrentUser() (*openwechat.Self, error) {
	return w.bot.GetCurrentUser()
}
func (w *Webot) SendText(msg string, to string, isGroup bool) ([]*openwechat.SentMessage, error) {
	self, err := w.bot.GetCurrentUser()
	if err != nil {
		return nil, err
	}
	var ss = make([]*openwechat.SentMessage, 0)
	reg, err := regexp.Compile(to)
	if err != nil {
		return nil, err
	}
	if isGroup {
		gs, err := self.Groups(true)
		if err != nil {
			return nil, err
		}
		gs.Search(math.MaxInt, func(group *openwechat.Group) bool {
			if reg.Match([]byte(group.NickName)) || reg.Match([]byte(group.RemarkName)) {
				sm, err := group.SendText(msg)
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
	fs, _ := self.Friends(true)
	fs.Search(math.MaxInt, func(friend *openwechat.Friend) bool {
		if reg.Match([]byte(friend.NickName)) || reg.Match([]byte(friend.RemarkName)) {
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
