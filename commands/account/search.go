package account

import (
	"errors"
	"fmt"

	"git.sr.ht/~rjarry/aerc/widgets"
)

type SearchFilter struct{}

func init() {
	register(SearchFilter{})
}

func (SearchFilter) Aliases() []string {
	return []string{"search", "filter"}
}

func (SearchFilter) Complete(aerc *widgets.Aerc, args []string) []string {
	return nil
}

func (SearchFilter) Execute(aerc *widgets.Aerc, args []string) error {
	acct := aerc.SelectedAccount()
	if acct == nil {
		return errors.New("No account selected")
	}
	store := acct.Store()
	if store == nil {
		return errors.New("Cannot perform action. Messages still loading")
	}

	var cb func([]uint32)
	if args[0] == "filter" {
		aerc.SetExtraStatus("Filtering...")
		cb = func(uids []uint32) {
			aerc.SetExtraStatus(fmt.Sprintf("%s", args))
			acct.Logger().Printf("Filter results: %v", uids)
			store.ApplyFilter(uids)
		}
	} else {
		aerc.SetExtraStatus("Searching...")
		cb = func(uids []uint32) {
			aerc.SetExtraStatus(fmt.Sprintf("%s", args))
			acct.Logger().Printf("Search results: %v", uids)
			store.ApplySearch(uids)
			// TODO: Remove when stores have multiple OnUpdate handlers
			acct.Messages().Invalidate()
		}
	}
	store.Search(args, cb)
	return nil
}
