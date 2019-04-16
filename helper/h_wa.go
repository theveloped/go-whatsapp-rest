package helper

import (
	"encoding/base64"
	"encoding/gob"
	"errors"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"fmt"

	"bytes"
	"net/http"
	"encoding/json"

	svc "github.com/theveloped/go-whatsapp-rest/service"
	whatsapp "github.com/dimaskiddo/go-whatsapp"
	qrcode "github.com/skip2/go-qrcode"
)

type responseHandler struct{
	webhook 	string
    created 	uint64
}

type messageTextResponse struct {
    whatsapp.TextMessage
    Response DialogResponse
}

type messageImageResponse struct {
    whatsapp.ImageMessage
    Response DialogResponse
}

func (wh responseHandler) HandleError(err error) {
	fmt.Fprintf(os.Stderr, "[!] %v\n", err)
}

func (wh responseHandler) HandleTextMessage(message whatsapp.TextMessage) {
	if !message.Info.FromMe {

		remoteJid := strings.Split(message.Info.RemoteJid, "@")[0]
		dialogResponse, err := DetectIntentText(svc.Config.GetString("DIALOGFLOW_PROJECT_ID"), remoteJid, message.Text, "en")

		if err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}

		responseMessage := messageTextResponse{TextMessage: message, Response: dialogResponse}

		jsonStr, _ := json.Marshal(responseMessage)
		_, _ = http.Post(wh.webhook, "application/json", bytes.NewBuffer(jsonStr))
	}
}

func (wh responseHandler) HandleImageMessage(message whatsapp.ImageMessage) {
	if !message.Info.FromMe {
		fmt.Printf("[+] Handling image\n")

		remoteJid := strings.Split(message.Info.RemoteJid, "@")[0]
		dialogResponse, err := DetectIntentText(svc.Config.GetString("DIALOGFLOW_PROJECT_ID"), remoteJid, message.Caption, "en")

		if err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}

		responseMessage := messageImageResponse{ImageMessage: message, Response: dialogResponse}

		jsonStr, _ := json.Marshal(responseMessage)
		_, _ = http.Post(wh.webhook, "application/json", bytes.NewBuffer(jsonStr))

		data, err := message.Download()
		if err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}

		filename := fmt.Sprintf("%v/%v.%v", svc.Config.GetString("SERVER_UPLOAD_PATH"), message.Info.Id, strings.Split(message.Type, "/")[1])
		file, err := os.Create(filename)
		defer file.Close()
		if err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}

		_, err = file.Write(data)
		if err != nil {
			fmt.Printf("[!] %v\n", err)
			return
		}

		fmt.Printf("[!] stored: %v\n", filename)
	}
}

func (wh responseHandler) HandleJsonMessage(message string) {
	fmt.Printf("[+] %v\n", message)
}

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

func WAConnect(jid string, webhook string, timeout int, file string, qrstr chan<- string, errmsg chan<- error) {
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

		if len(webhook) > 0 {
			// fmt.Printf("[!] removing webhooks\n")
			// wac[jid].RemoveHandlers()

			fmt.Printf("[!] adding webhook: %v\n", webhook)
			wac[jid].AddHandler(responseHandler{webhook, uint64(time.Now().Unix())})
		}

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

		_, _ = wac[jid].Presence(jidDest + jidPrefix, "composing")

		<-time.After(time.Duration(msgDelay) * time.Second)

		err := wac[jid].Send(content)
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

		_, _ = wac[jid].Presence(jidDest + jidPrefix, "composing")

		<-time.After(time.Duration(msgDelay) * time.Second)

		err := wac[jid].Send(content)
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
