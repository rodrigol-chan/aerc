package account

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"git.sr.ht/~rjarry/aerc/log"
	"git.sr.ht/~rjarry/aerc/widgets"
	mboxer "git.sr.ht/~rjarry/aerc/worker/mbox"
	"git.sr.ht/~rjarry/aerc/worker/types"
)

type ExportMbox struct{}

func init() {
	register(ExportMbox{})
}

func (ExportMbox) Aliases() []string {
	return []string{"export-mbox"}
}

func (ExportMbox) Complete(aerc *widgets.Aerc, args []string) []string {
	if acct := aerc.SelectedAccount(); acct != nil {
		if path := acct.SelectedDirectory(); path != "" {
			if f := filepath.Base(path); f != "" {
				return []string{f + ".mbox"}
			}
		}
	}
	return nil
}

func (ExportMbox) Execute(aerc *widgets.Aerc, args []string) error {
	if len(args) != 2 {
		return exportFolderUsage(args[0])
	}
	filename := args[1]

	acct := aerc.SelectedAccount()
	if acct == nil {
		return errors.New("No account selected")
	}
	store := acct.Store()
	if store == nil {
		return errors.New("No message store selected")
	}

	aerc.PushStatus("Exporting to "+filename, 10*time.Second)

	go func() {
		defer log.PanicHandler()
		file, err := os.Create(filename)
		if err != nil {
			log.Errorf("failed to create file: %v", err)
			aerc.PushError(err.Error())
			return
		}
		defer file.Close()

		var mu sync.Mutex
		var ctr uint32
		var retries int

		done := make(chan bool)
		uids := make([]uint32, len(store.Uids()))
		copy(uids, store.Uids())
		t := time.Now()

		for len(uids) > 0 {
			if retries > 0 {
				if retries > 10 {
					errorMsg := fmt.Sprintf("too many retries: %d; stopping export", retries)
					log.Errorf(errorMsg)
					aerc.PushError(args[0] + " " + errorMsg)
					break
				}
				sleeping := time.Duration(retries * 1e9 * 2)
				log.Debugf("sleeping for %s before retrying; retries: %d", sleeping, retries)
				time.Sleep(sleeping)
			}

			log.Debugf("fetching %d for export", len(uids))
			acct.Worker().PostAction(&types.FetchFullMessages{
				Uids: uids,
			}, func(msg types.WorkerMessage) {
				switch msg := msg.(type) {
				case *types.Done:
					done <- true
				case *types.Error:
					log.Errorf("failed to fetch message: %v", msg.Error)
					aerc.PushError(args[0] + " error encountered: " + msg.Error.Error())
					done <- false
				case *types.FullMessage:
					mu.Lock()
					err := mboxer.Write(file, msg.Content.Reader, "", t)
					if err != nil {
						log.Warnf("failed to write mbox: %v", err)
					}
					for i, uid := range uids {
						if uid == msg.Content.Uid {
							uids = append(uids[:i], uids[i+1:]...)
							break
						}
					}
					ctr++
					mu.Unlock()
				}
			})
			if ok := <-done; ok {
				break
			}
			retries++
		}
		statusInfo := fmt.Sprintf("Exported %d of %d messages to %s.", ctr, len(store.Uids()), filename)
		aerc.PushStatus(statusInfo, 10*time.Second)
		log.Debugf(statusInfo)
	}()

	return nil
}

func exportFolderUsage(cmd string) error {
	return fmt.Errorf("Usage: %s <filename>", cmd)
}
