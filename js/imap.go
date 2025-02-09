package js

import (
	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
	"github.com/emersion/go-message/mail"
	"go.uber.org/fx"
	"io"
	"log"
	"wios_server/utils"
)

type Attachment struct {
	Name string
	Data []byte
}
type Email struct {
	Id           string
	Uid          uint32
	SeqNum       uint32
	PersonalName string
	From         string
	Subject      string
	Body         map[string]string
	Attachments  []*Attachment
}

func (e *Email) Reply(msender *utils.MailSender, body string) error {
	return msender.Send(utils.Message{
		To:          []string{e.From},
		Body:        body,
		Subject:     e.Subject,
		ContentType: "text/html; charset=\"UTF-8\"",
	})
}

func newImapClient(imapServer string, username string, password string) (*client.Client, error) {
	c, err := client.DialTLS(imapServer, nil)
	if err != nil {
		return nil, err
	}
	if err := c.Login(username, password); err != nil {
		return nil, err
	}
	return c, nil
}
func fetchMail(boxname string, n uint32, c *client.Client) ([]*Email, error) {
	mbox, err := c.Select(boxname, false)
	if err != nil {
		return nil, err
	}
	if mbox.Messages == 0 {
		return []*Email{}, nil
	}
	from := uint32(1)
	to := mbox.Messages
	if mbox.Messages > n {
		from = mbox.Messages - n + 1
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddRange(from, to)
	messages := make(chan *imap.Message, n)
	done := make(chan error, 1)
	go func() {
		done <- c.Fetch(seqSet,
			[]imap.FetchItem{
				imap.FetchEnvelope,
				imap.FetchFlags,
				imap.FetchInternalDate,
				imap.FetchUid,
			}, messages)
	}()
	ems := make([]*Email, 0)
	for msg := range messages {
		if em, err := parseMail(msg); err == nil {
			ems = append(ems, em)
		}
	}
	if err := <-done; err != nil {
		return nil, err
	}
	return ems, nil
}
func fetchMailByUid(boxname string, c *client.Client, uid ...uint32) ([]*Email, error) {
	mbox, err := c.Select(boxname, false)
	if err != nil {
		return nil, err
	}
	if mbox.Messages == 0 {
		return []*Email{}, nil
	}
	seqSet := new(imap.SeqSet)
	seqSet.AddNum(uid...)
	messages := make(chan *imap.Message, len(uid))
	done := make(chan error, 1)
	go func() {
		done <- c.UidFetch(seqSet,
			[]imap.FetchItem{
				imap.FetchEnvelope,
				imap.FetchFlags,
				imap.FetchInternalDate,
				imap.FetchUid,
				imap.FetchBodyStructure,
				imap.FetchBody,
				imap.FetchRFC822,
				imap.FetchRFC822Size,
				imap.FetchRFC822Text,
				imap.FetchRFC822Header,
			}, messages)
	}()
	ems := make([]*Email, 0)
	for msg := range messages {
		if em, err := parseMail(msg); err == nil {
			ems = append(ems, em)
		}
	}
	if err := <-done; err != nil {
		return nil, err
	}
	return ems, nil
}

func parseMail(msg *imap.Message) (*Email, error) {
	// 解析邮件内容
	f0 := msg.Envelope.From[0]
	from := f0.MailboxName + "@" + f0.HostName
	em := &Email{
		Id:           msg.Envelope.MessageId,
		Uid:          msg.Uid,
		SeqNum:       msg.SeqNum,
		From:         from,
		PersonalName: f0.PersonalName,
		Subject:      msg.Envelope.Subject,
	}
	section := &imap.BodySectionName{}
	r := msg.GetBody(section)
	if r == nil {
		return em, nil
	}
	// 创建MIME解析器
	mr, err := mail.CreateReader(r)
	if err != nil {
		return nil, err
	}
	em.Body = make(map[string]string, 2)
	em.Attachments = make([]*Attachment, 0)
	for i := 0; true; i++ {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("内容解析错误: %v", err)
			continue
		}
		var bs *imap.BodyStructure
		if msg.BodyStructure.Parts != nil {
			bs = msg.BodyStructure.Parts[i]
		} else {
			bs = msg.BodyStructure
		}
		switch part.Header.(type) {
		case *mail.InlineHeader:
			em.Body[bs.MIMEType+"/"+bs.MIMESubType] = processTextContent(part)
		case *mail.AttachmentHeader:
			em.Attachments = append(em.Attachments, processAttachment(part))
		}
	}
	return em, nil
}

// 处理文本内容
func processTextContent(part *mail.Part) string {
	body, err := io.ReadAll(part.Body)
	if err != nil {
		return ""
	}
	return string(body)
}

// 处理附件
func processAttachment(part *mail.Part) *Attachment {
	header := part.Header.(*mail.AttachmentHeader)
	filename, err := header.Filename()
	if err != nil {
		return nil
	}
	data, err := io.ReadAll(part.Body)
	if err != nil {
		return nil
	}
	return &Attachment{Name: filename, Data: data}
	//// 创建保存目录
	//_ = os.Mkdir("attachments", 0755)
	//
	//// 保存文件
	//filePath := fmt.Sprintf("attachments/%s", filename)
	//file, err := os.Create(filePath)
	//if err != nil {
	//	log.Printf("文件创建失败: %v", err)
	//	return
	//}
	//defer file.Close()
	//
	//if _, err := io.Copy(file, part.Body); err != nil {
	//	log.Printf("附件保存失败: %v", err)
	//}
}

func newMailSender(user string, pwd string, host string, port string, workdir string) *utils.MailSender {
	return &utils.MailSender{
		User:    user,
		Pwd:     pwd,
		Host:    host,
		Port:    port,
		WorkDir: workdir,
	}
}

var imapModule = fx.Options(
	fx.Invoke(func() {
		RegExport("newImapClient", newImapClient)
		RegExport("fetchMail", fetchMail)
		RegExport("fetchMailByUid", fetchMailByUid)
		RegExport("newMailSender", newMailSender)
	}),
)
