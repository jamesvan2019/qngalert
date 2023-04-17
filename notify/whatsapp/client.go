package whatsapp

import (
	"context"
	"errors"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	qrterminal "github.com/mdp/qrterminal/v3"
	"go.mau.fi/whatsmeow"
	waBinary "go.mau.fi/whatsmeow/binary"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"qngalert/config"
	"strings"
	"sync/atomic"
	"time"
)

type WhatsappBot struct {
	api      *whatsmeow.Client
	Cfg      *config.Config
	lastSend int64
}

var logLevel = "INFO"
var debugLogs = flag.Bool("debug", false, "Enable debug logs?")
var dbDialect = flag.String("db-dialect", "sqlite3", "Database dialect (sqlite3 or postgres)")
var dbAddress = flag.String("db-address", "file:mdtest.db?_foreign_keys=on", "Database address")
var requestFullSync = flag.Bool("request-full-sync", false, "Request full (1 year) history sync when logging in?")
var pairRejectChan = make(chan bool, 1)

func (t *WhatsappBot) Init(cfg *config.Config) error {
	waBinary.IndentXML = true
	flag.Parse()

	if *debugLogs {
		logLevel = "DEBUG"
	}
	if *requestFullSync {
		store.DeviceProps.RequireFullSync = proto.Bool(true)
	}
	//log = waLog.Stdout("Main", logLevel, true)

	dbLog := waLog.Stdout("Database", logLevel, true)
	storeContainer, err := sqlstore.New(*dbDialect, *dbAddress, dbLog)
	if err != nil {
		log.Printf("\nFailed to connect to database: %v", err)
		return errors.New("error")
	}
	device, err := storeContainer.GetFirstDevice()
	if err != nil {
		log.Printf("\nFailed to get device: %v", err)
		return errors.New("error")
	}
	t.api = whatsmeow.NewClient(device, waLog.Stdout("Client", logLevel, true))
	var isWaitingForPair atomic.Bool
	t.api.PrePairCallback = func(jid types.JID, platform, businessName string) bool {
		isWaitingForPair.Store(true)
		defer isWaitingForPair.Store(false)
		log.Printf("\nPairing %s (platform: %q, business name: %q). Type r within 3 seconds to reject pair", jid, platform, businessName)
		select {
		case reject := <-pairRejectChan:
			if reject {
				log.Printf("\nRejecting pair")
				return false
			}
		case <-time.After(3 * time.Second):
		}
		log.Printf("\nAccepting pair")
		return true
	}

	ch, err := t.api.GetQRChannel(context.Background())
	if err != nil {
		// This error means that we're already logged in, so ignore it.
		if !errors.Is(err, whatsmeow.ErrQRStoreContainsID) {
			log.Printf("\nFailed to get QR channel: %v", err)
		}
	} else {
		go func() {
			for evt := range ch {
				if evt.Event == "code" {
					qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
				} else {
					log.Printf("\nQR channel result: %s", evt.Event)
				}
			}
		}()
	}

	t.api.AddEventHandler(t.handler)
	err = t.api.Connect()
	if err != nil {
		log.Printf("\nFailed to connect: %v", err)
		return errors.New("error")
	}
	t.Cfg = cfg
	return nil
}

func (t *WhatsappBot) Stop() {
	t.api.Disconnect()
}
func (t *WhatsappBot) Login() error {
	//t.ListGroup()
	return nil
}

func parseJID(arg string) (types.JID, bool) {
	if arg[0] == '+' {
		arg = arg[1:]
	}
	if !strings.ContainsRune(arg, '@') {
		return types.NewJID(arg, types.DefaultUserServer), true
	} else {
		recipient, err := types.ParseJID(arg)
		if err != nil {
			log.Printf("Invalid JID %s: %v\n", arg, err)
			return recipient, false
		} else if recipient.User == "" {
			log.Printf("Invalid JID %s: no server specified\n", arg)
			return recipient, false
		}
		return recipient, true
	}
}

func (t *WhatsappBot) Notify(title, content string) error {
	return nil
	if time.Now().Unix()-t.lastSend < 600 { // 10分钟内不重发
		return nil
	}
	t.lastSend = time.Now().Unix()
	str := fmt.Sprintf("%s\n%s", title, content)
	msg := &waProto.Message{Conversation: proto.String(str)}
	recipient, ok := parseJID("120363120359891292@g.us")
	if !ok {
		return errors.New("JID not exist")
	}
	resp, err := t.api.SendMessage(context.Background(), recipient, msg)
	if err != nil {
		log.Println(err)
		return err
	}
	log.Println("Send Whatsapp message", str, "send id:", resp.ID)
	return nil
}

func (t *WhatsappBot) ListGroup() {
	groups, err := t.api.GetJoinedGroups()
	if err != nil {
		log.Printf("Failed to get group list: %v\n", err)
	} else {
		for _, group := range groups {
			log.Printf("%+v\n", group)
		}
	}
}
