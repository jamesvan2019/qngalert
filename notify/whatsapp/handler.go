package whatsapp

import (
	"encoding/json"
	"fmt"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"log"
	"mime"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var historySyncID int32
var startupTime = time.Now().Unix()

func (w *WhatsappBot) handler(rawEvt interface{}) {
	switch evt := rawEvt.(type) {
	case *events.AppStateSyncComplete:
		if len(w.api.Store.PushName) > 0 && evt.Name == appstate.WAPatchCriticalBlock {
			err := w.api.SendPresence(types.PresenceAvailable)
			if err != nil {
				log.Printf("\nFailed to send available presence: %v", err)
			} else {
				log.Printf("\nMarked self as available")
			}
		}
	case *events.Connected, *events.PushNameSetting:
		if len(w.api.Store.PushName) == 0 {
			return
		}
		// Send presence available when connecting and when the pushname is changed.
		// This makes sure that outgoing messages always have the right pushname.
		err := w.api.SendPresence(types.PresenceAvailable)
		if err != nil {
			log.Printf("\nFailed to send available presence: %v", err)
		} else {
			log.Printf("\nMarked self as available")
		}
	case *events.StreamReplaced:
		os.Exit(0)
	case *events.Message:
		return
		metaParts := []string{fmt.Sprintf("pushname: %s", evt.Info.PushName), fmt.Sprintf("timestamp: %s", evt.Info.Timestamp)}
		if evt.Info.Type != "" {
			metaParts = append(metaParts, fmt.Sprintf("type: %s", evt.Info.Type))
		}
		if evt.Info.Category != "" {
			metaParts = append(metaParts, fmt.Sprintf("category: %s", evt.Info.Category))
		}
		if evt.IsViewOnce {
			metaParts = append(metaParts, "view once")
		}
		if evt.IsViewOnce {
			metaParts = append(metaParts, "ephemeral")
		}
		if evt.IsViewOnceV2 {
			metaParts = append(metaParts, "ephemeral (v2)")
		}
		if evt.IsDocumentWithCaption {
			metaParts = append(metaParts, "document with caption")
		}
		if evt.IsEdit {
			metaParts = append(metaParts, "edit")
		}

		log.Printf("\nReceived message %s from %s (%s): %+v", evt.Info.ID, evt.Info.SourceString(), strings.Join(metaParts, ", "), evt.Message)

		if evt.Message.GetPollUpdateMessage() != nil {
			decrypted, err := w.api.DecryptPollVote(evt)
			if err != nil {
				log.Printf("\nFailed to decrypt vote: %v", err)
			} else {
				log.Printf("\nSelected options in decrypted vote:")
				for _, option := range decrypted.SelectedOptions {
					log.Printf("\n- %X", option)
				}
			}
		} else if evt.Message.GetEncReactionMessage() != nil {
			decrypted, err := w.api.DecryptReaction(evt)
			if err != nil {
				log.Printf("\nFailed to decrypt encrypted reaction: %v", err)
			} else {
				log.Printf("\nDecrypted reaction: %+v", decrypted)
			}
		}

		img := evt.Message.GetImageMessage()
		if img != nil {
			data, err := w.api.Download(img)
			if err != nil {
				log.Printf("\nFailed to download image: %v", err)
				return
			}
			exts, _ := mime.ExtensionsByType(img.GetMimetype())
			path := fmt.Sprintf("%s%s", evt.Info.ID, exts[0])
			err = os.WriteFile(path, data, 0600)
			if err != nil {
				log.Printf("\nFailed to save image: %v", err)
				return
			}
			log.Printf("\nSaved image in message to %s", path)
		}
	case *events.Receipt:
		return
		if evt.Type == events.ReceiptTypeRead || evt.Type == events.ReceiptTypeReadSelf {
			log.Printf("\n%v was read by %s at %s", evt.MessageIDs, evt.SourceString(), evt.Timestamp)
		} else if evt.Type == events.ReceiptTypeDelivered {
			log.Printf("\n%s was delivered to %s at %s", evt.MessageIDs[0], evt.SourceString(), evt.Timestamp)
		}
	case *events.Presence:
		if evt.Unavailable {
			if evt.LastSeen.IsZero() {
				log.Printf("\n%s is now offline", evt.From)
			} else {
				log.Printf("\n%s is now offline (last seen: %s)", evt.From, evt.LastSeen)
			}
		} else {
			log.Printf("\n%s is now online", evt.From)
		}
	case *events.HistorySync:
		id := atomic.AddInt32(&historySyncID, 1)
		fileName := fmt.Sprintf("history-%d-%d.json", startupTime, id)
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			log.Printf("\nFailed to open file to write history sync: %v", err)
			return
		}
		enc := json.NewEncoder(file)
		enc.SetIndent("", "  ")
		err = enc.Encode(evt.Data)
		if err != nil {
			log.Printf("\nFailed to write history sync: %v", err)
			return
		}
		log.Printf("\nWrote history sync to %s", fileName)
		_ = file.Close()
	case *events.AppState:
		log.Printf("\nApp state event: %+v / %+v", evt.Index, evt.SyncActionValue)
	case *events.KeepAliveTimeout:
		log.Printf("\nKeepalive timeout event: %+v", evt)
		if evt.ErrorCount > 3 {
			log.Printf("\nGot >3 keepalive timeouts, forcing reconnect")
			go func() {
				w.api.Disconnect()
				err := w.api.Connect()
				if err != nil {
					log.Printf("\nError force-reconnecting after keepalive timeouts: %v", err)
				}
			}()
		}
	case *events.KeepAliveRestored:
		log.Printf("\nKeepalive restored")
	}
}
