package helper

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"mime/multipart"
	"os"
	"strings"
	"time"

	whatsapp "github.com/Rhymen/go-whatsapp"
	qrcode "github.com/skip2/go-qrcode"
)

var wac = make(map[string]*whatsapp.Conn)

func WAInit(jid string, timeout int) error {
	if wac[jid] == nil {
		conn, err := whatsapp.NewConn(time.Duration(timeout) * time.Second)
		if err != nil {
			return err
		}
		conn.SetClientName("Go WhatsApp REST", "Go WhatsApp")
		wac[jid] = conn
	}

	return nil
}

func WASessionLoad(file string) (whatsapp.Session, error) {
	session := whatsapp.Session{}

	buffer, err := os.Open(file)
	if err != nil {
		return session, err
	}
	defer buffer.Close()

	err = gob.NewDecoder(buffer).Decode(&session)
	if err != nil {
		return session, err
	}

	return session, nil
}

func WASessionSave(file string, session whatsapp.Session) error {
	buffer, err := os.Create(file)
	if err != nil {
		return err
	}
	defer buffer.Close()

	err = gob.NewEncoder(buffer).Encode(session)
	if err != nil {
		return err
	}

	return nil
}

func WASessionLogin(jid string, file string, qr chan<- string) error {
	if wac[jid] != nil {
		_, err := os.Stat(file)
		if err == nil {
			err = os.Remove(file)
			if err != nil {
				return err
			}
		}

		session, err := wac[jid].Login(qr)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "already logged in":
				return nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				delete(wac, jid)
				return errors.New("connection is invalid")
			default:
				return err
			}
		}

		err = WASessionSave(file, session)
		if err != nil {
			return err
		}
	} else {
		return errors.New("connection is invalid")
	}

	return nil
}

func WASessionRestore(jid string, file string, sess whatsapp.Session) error {
	if wac[jid] != nil {
		session, err := wac[jid].RestoreWithSession(sess)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "already logged in":
				return nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				delete(wac, jid)
				return errors.New("connection is invalid")
			default:
				errLogout := wac[jid].Logout()
				if errLogout != nil {
					return errLogout
				}

				delete(wac, jid)
				return err
			}
		}

		err = WASessionSave(file, session)
		if err != nil {
			return err
		}
	} else {
		return errors.New("connection is invalid")
	}

	return nil
}

func WASessionLogout(jid string, file string) error {
	if wac[jid] != nil {
		err := wac[jid].Logout()
		if err != nil {
			return err
		}

		_, err = os.Stat(file)
		if err == nil {
			err = os.Remove(file)
			if err != nil {
				return err
			}
		}

		delete(wac, jid)
	} else {
		return errors.New("connection is invalid")
	}

	return nil
}

func WAConnect(jid string, timeout int, file string, qrstr chan<- string, errmsg chan<- error) {
	if wac[jid] != nil {
		chanqr := make(chan string)
		go func() {
			select {
			case tmp := <-chanqr:
				png, errPNG := qrcode.Encode(tmp, qrcode.Medium, 256)
				if errPNG != nil {
					errmsg <- errPNG
					return
				}

				qrstr <- base64.StdEncoding.EncodeToString(png)
			case <-time.After(time.Duration(timeout) * time.Second):
				errmsg <- errors.New("qr code generate timed out")
			}
		}()

		session, err := WASessionLoad(file)
		if err != nil {
			err = WASessionLogin(jid, file, chanqr)
			if err != nil {
				errmsg <- err
				return
			}
		} else {
			err = WASessionRestore(jid, file, session)
			if err != nil {
				err := WAInit(jid, timeout)
				if err != nil {
					errmsg <- err
					return
				}

				err = WASessionLogin(jid, file, chanqr)
				if err != nil {
					errmsg <- err
					return
				}
			}
		}
	} else {
		errmsg <- errors.New("connection is invalid")
		return
	}

	errmsg <- errors.New("")
	return
}

func WAMessageText(jid string, jidDest string, msgText string, msgDelay int) error {
	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.TextMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			Text: msgText,
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		_, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out":
				return nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				delete(wac, jid)
				return errors.New("connection is invalid")
			default:
				return err
			}
		}
	} else {
		return errors.New("connection is invalid")
	}

	return nil
}

func WAMessageImage(jid string, jidDest string, msgImageStream multipart.File, msgImageType string, msgCaption string, msgDelay int) error {
	if wac[jid] != nil {
		jidPrefix := "@s.whatsapp.net"
		if len(strings.SplitN(jidDest, "-", 2)) == 2 {
			jidPrefix = "@g.us"
		}

		content := whatsapp.ImageMessage{
			Info: whatsapp.MessageInfo{
				RemoteJid: jidDest + jidPrefix,
			},
			Content: msgImageStream,
			Type:    msgImageType,
			Caption: msgCaption,
		}

		<-time.After(time.Duration(msgDelay) * time.Second)

		_, err := wac[jid].Send(content)
		if err != nil {
			switch strings.ToLower(err.Error()) {
			case "sending message timed out":
				return nil
			case "could not send proto: failed to write message: error writing to websocket: websocket: close sent":
				delete(wac, jid)
				return errors.New("connection is invalid")
			default:
				return err
			}
		}
	} else {
		return errors.New("connection is invalid")
	}

	return nil
}
