package utils

import (
	"bytes"
	"encoding/base64"
	"errors"
	"mime"
	"net/smtp"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type MailSender struct {
	User    string
	Pwd     string
	Host    string
	Port    string
	WorkDir string
}
type Attachment struct {
	Fid         string
	Name        string
	ContentType string
	Data        []byte
}
type Message struct {
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	Body        string
	ContentType string
	Attachments []Attachment
	sender      *MailSender
}

const boundary = "WiOSBoundary"

func writeKV(buf *bytes.Buffer, k, v string) {
	buf.WriteString(k)
	buf.WriteString(": ")
	buf.WriteString(v)
	buf.WriteString("\r\n")
}
func writeBlock(buf *bytes.Buffer, block string) {
	buf.WriteString("\r\n")
	buf.WriteString(block)
	buf.WriteString("\r\n")
}
func writeStartBoundary(buf *bytes.Buffer) {
	buf.WriteString("\r\n--")
	buf.WriteString(boundary)
	buf.WriteString("\r\n")
}
func writeEndBoundary(buf *bytes.Buffer) {
	buf.WriteString("\r\n--")
	buf.WriteString(boundary)
	buf.WriteString("--")
}
func writeHeader(buf *bytes.Buffer, msg *Message) bool {
	haseFile := false
	writeKV(buf, "From", msg.From)
	writeKV(buf, "To", strings.Join(msg.To, ";"))
	if msg.Cc != nil && len(msg.Cc) > 0 {
		writeKV(buf, "Cc", strings.Join(msg.Cc, ";"))
	}
	if msg.Bcc != nil && len(msg.Bcc) > 0 {
		writeKV(buf, "Bcc", strings.Join(msg.Bcc, ";"))
	}
	encodedSubject := mime.QEncoding.Encode("UTF-8", msg.Subject)
	writeKV(buf, "Subject", encodedSubject)
	if msg.Attachments != nil && len(msg.Attachments) > 0 {
		writeKV(buf, "Content-Type", "multipart/mixed; boundary="+boundary)
		haseFile = true
	} else {
		writeKV(buf, "Content-Type", msg.ContentType)
	}
	writeKV(buf, "MIME-Version", "1.0")
	writeKV(buf, "Date", time.Now().Format(time.RFC1123Z))
	buf.WriteString("\r\n")
	return haseFile
}
func (s *MailSender) writeAttachment(buf *bytes.Buffer, attach *Attachment) error {
	writeStartBoundary(buf)
	writeKV(buf, "Content-Type", attach.ContentType+";name="+attach.Name)
	writeKV(buf, "Content-Disposition", "attachment;filename="+attach.Name)
	writeKV(buf, "Content-Transfer-Encoding", "base64")
	buf.WriteString("\r\n")
	if attach.Fid != "" {
		basePath := path.Join(s.WorkDir, GetFPathByFid(attach.Fid))
		path := filepath.Join(basePath, "0")
		_, err := os.Stat(path)
		if err != nil {
			return err
		}
		file, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		payload := make([]byte, base64.StdEncoding.EncodedLen(len(file)))
		base64.StdEncoding.Encode(payload, file)
		for idx, L := 0, len(payload); idx < L; idx++ {
			buf.WriteByte(payload[idx])
			if (idx+1)%76 == 0 {
				buf.WriteString("\r\n")
			}
		}
		return nil
	}
	if attach.Data != nil {
		for idx, L := 0, len(attach.Data); idx < L; idx++ {
			buf.WriteByte(attach.Data[idx])
			if (idx+1)%76 == 0 {
				buf.WriteString("\r\n")
			}
		}
	}
	writeEndBoundary(buf)
	return nil
}
func (s *MailSender) writeAttachments(buf *bytes.Buffer, msg *Message) error {
	for _, attach := range msg.Attachments {
		err := s.writeAttachment(buf, &attach)
		if err != nil {
			return err
		}
	}
	writeEndBoundary(buf)
	return nil
}
func (s *MailSender) Send(msg Message) error {
	auth := smtp.PlainAuth("", s.User, s.Pwd, s.Host)
	buf := bytes.NewBuffer(nil)
	if msg.From == "" {
		msg.From = s.User
	}
	hasFile := writeHeader(buf, &msg)
	//write body
	if hasFile {
		writeStartBoundary(buf)
		writeKV(buf, "Content-Type", msg.ContentType)
	}
	writeBlock(buf, msg.Body)
	//write attachments
	if hasFile {
		err := s.writeAttachments(buf, &msg)
		if err != nil {
			return err
		}
	}
	return smtp.SendMail(s.Host+":"+s.Port, auth, msg.From, msg.To, buf.Bytes())
}
func (s *MailSender) NewMail() *Message {
	return &Message{sender: s, ContentType: "text/html; charset=\"UTF-8\""}
}
func (m *Message) SetTo(to ...string) *Message {
	m.To = append(m.To, to...)
	return m
}
func (m *Message) SetCc(cc ...string) *Message {
	m.Cc = append(m.Cc, cc...)
	return m
}
func (m *Message) SetBcc(bcc ...string) *Message {
	m.Bcc = append(m.Bcc, bcc...)
	return m
}
func (m *Message) SetSubject(subject string) *Message {
	m.Subject = subject
	return m
}
func (m *Message) SetBody(body string) *Message {
	m.Body = body
	return m
}
func (m *Message) SetContentType(contentType string) *Message {
	m.ContentType = contentType
	return m
}
func (m *Message) SetAttachments(attachments []Attachment) *Message {
	m.Attachments = attachments
	return m
}
func (m *Message) Send() error {
	if m.sender == nil {
		return errors.New("sender is nil")
	}
	return m.sender.Send(*m)
}
