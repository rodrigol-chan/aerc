package commands

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"git.sr.ht/~sircmpwn/aerc/widgets"
)

type ChangeTab struct{}

func init() {
	register(ChangeTab{})
}

func (ChangeTab) Aliases() []string {
	return []string{"ct", "change-tab"}
}

func (ChangeTab) Complete(aerc *widgets.Aerc, args []string) []string {
	if len(args) == 0 {
		return aerc.TabNames()
	}
	out := make([]string, 0)
	for _, tab := range aerc.TabNames() {
		if strings.HasPrefix(tab, args[0]) {
			out = append(out, tab)
		}
	}
	return out
}

func (ChangeTab) Execute(aerc *widgets.Aerc, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Usage: %s <tab>", args[0])
	}
	if args[1] == "-" {
		ok := aerc.SelectPreviousTab()
		if !ok {
			return errors.New("No previous tab to return to")
		}
	} else {
		n, err := strconv.Atoi(args[1])
		if err == nil {
			if strings.HasPrefix(args[1], "+") {
				for ; n > 0; n-- {
					aerc.NextTab()
				}
			} else if strings.HasPrefix(args[1], "-") {
				for ; n < 0; n++ {
					aerc.PrevTab()
				}
			} else {
				ok := aerc.SelectTabIndex(n)
				if !ok {
					return errors.New(
						"No tab with that index")
				}
			}
		} else {
			ok := aerc.SelectTab(args[1])
			if !ok {
				return errors.New("No tab with that name")
			}
		}
	}
	return nil
}
