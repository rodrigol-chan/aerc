package account

import (
	"errors"
	"strconv"
	"strings"

	"git.sr.ht/~rjarry/aerc/widgets"
)

type Split struct{}

func init() {
	register(Split{})
}

func (Split) Aliases() []string {
	return []string{"split", "vsplit"}
}

func (Split) Complete(aerc *widgets.Aerc, args []string) []string {
	return nil
}

func (Split) Execute(aerc *widgets.Aerc, args []string) error {
	if len(args) > 2 {
		return errors.New("Usage: [v]split n")
	}
	acct := aerc.SelectedAccount()
	if acct == nil {
		return errors.New("No account selected")
	}
	n := 0
	if acct.SplitSize() == 0 {
		if args[0] == "split" {
			n = aerc.SelectedAccount().Messages().Height() / 4
		} else {
			n = aerc.SelectedAccount().Messages().Width() / 2
		}
	}

	var err error
	if len(args) > 1 {
		delta := false
		if strings.HasPrefix(args[1], "+") || strings.HasPrefix(args[1], "-") {
			delta = true
		}
		n, err = strconv.Atoi(args[1])
		if err != nil {
			return errors.New("Usage: [v]split n")
		}
		if delta {
			n = acct.SplitSize() + n
			acct.SetSplitSize(n)
			return nil
		}
	}
	if n == acct.SplitSize() {
		// Repeated commands of the same size have the effect of
		// toggling the split
		n = 0
	}
	if n < 0 {
		// Don't allow split to go negative
		n = 1
	}
	switch args[0] {
	case "split":
		return acct.Split(n)
	case "vsplit":
		return acct.Vsplit(n)
	}
	return nil
}
