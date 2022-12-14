package config

import (
	"fmt"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.sr.ht/~rjarry/aerc/lib/templates"
	"git.sr.ht/~rjarry/aerc/log"
	"github.com/gdamore/tcell/v2"
	"github.com/go-ini/ini"
	"github.com/imdario/mergo"
)

type UIConfig struct {
	IndexColumns    []*ColumnDef `ini:"-"`
	ColumnSeparator string       `ini:"column-separator"`
	// deprecated
	IndexFormat string `ini:"index-format"`

	AutoMarkRead                  bool          `ini:"auto-mark-read"`
	TimestampFormat               string        `ini:"timestamp-format"`
	ThisDayTimeFormat             string        `ini:"this-day-time-format"`
	ThisWeekTimeFormat            string        `ini:"this-week-time-format"`
	ThisYearTimeFormat            string        `ini:"this-year-time-format"`
	MessageViewTimestampFormat    string        `ini:"message-view-timestamp-format"`
	MessageViewThisDayTimeFormat  string        `ini:"message-view-this-day-time-format"`
	MessageViewThisWeekTimeFormat string        `ini:"message-view-this-week-time-format"`
	MessageViewThisYearTimeFormat string        `ini:"message-view-this-year-time-format"`
	ShowHeaders                   []string      `delim:","`
	RenderAccountTabs             string        `ini:"render-account-tabs"`
	PinnedTabMarker               string        `ini:"pinned-tab-marker"`
	SidebarWidth                  int           `ini:"sidebar-width"`
	PreviewHeight                 int           `ini:"preview-height"`
	EmptyMessage                  string        `ini:"empty-message"`
	EmptyDirlist                  string        `ini:"empty-dirlist"`
	MouseEnabled                  bool          `ini:"mouse-enabled"`
	ThreadingEnabled              bool          `ini:"threading-enabled"`
	ForceClientThreads            bool          `ini:"force-client-threads"`
	ClientThreadsDelay            time.Duration `ini:"client-threads-delay"`
	FuzzyComplete                 bool          `ini:"fuzzy-complete"`
	NewMessageBell                bool          `ini:"new-message-bell"`
	Spinner                       string        `ini:"spinner"`
	SpinnerDelimiter              string        `ini:"spinner-delimiter"`
	IconUnencrypted               string        `ini:"icon-unencrypted"`
	IconEncrypted                 string        `ini:"icon-encrypted"`
	IconSigned                    string        `ini:"icon-signed"`
	IconSignedEncrypted           string        `ini:"icon-signed-encrypted"`
	IconUnknown                   string        `ini:"icon-unknown"`
	IconInvalid                   string        `ini:"icon-invalid"`
	IconAttachment                string        `ini:"icon-attachment"`
	DirListFormat                 string        `ini:"dirlist-format"`
	DirListDelay                  time.Duration `ini:"dirlist-delay"`
	DirListTree                   bool          `ini:"dirlist-tree"`
	DirListCollapse               int           `ini:"dirlist-collapse"`
	Sort                          []string      `delim:" "`
	NextMessageOnDelete           bool          `ini:"next-message-on-delete"`
	CompletionDelay               time.Duration `ini:"completion-delay"`
	CompletionMinChars            int           `ini:"completion-min-chars"`
	CompletionPopovers            bool          `ini:"completion-popovers"`
	StyleSetDirs                  []string      `ini:"stylesets-dirs" delim:":"`
	StyleSetName                  string        `ini:"styleset-name"`
	style                         StyleSet
	// customize border appearance
	BorderCharVertical   rune `ini:"-"`
	BorderCharHorizontal rune `ini:"-"`

	ReverseOrder       bool `ini:"reverse-msglist-order"`
	ReverseThreadOrder bool `ini:"reverse-thread-order"`
	SortThreadSiblings bool `ini:"sort-thread-siblings"`

	// private
	contextualUis    []*UiConfigContext
	contextualCounts map[uiContextType]int
	contextualCache  map[uiContextKey]*UIConfig
}

type uiContextType int

const (
	uiContextFolder uiContextType = iota
	uiContextAccount
	uiContextSubject
)

type UiConfigContext struct {
	ContextType uiContextType
	Regex       *regexp.Regexp
	UiConfig    *UIConfig
}

type uiContextKey struct {
	ctxType uiContextType
	value   string
}

func defaultUiConfig() *UIConfig {
	date, _ := templates.ParseTemplate("column-date", "{{.DateAutoFormat .Date.Local}}")
	name, _ := templates.ParseTemplate("column-name", "{{index (.From | names) 0}}")
	flags, _ := templates.ParseTemplate("column-flags", `{{.Flags | join ""}}`)
	subject, _ := templates.ParseTemplate("column-subject", "{{.Subject}}")
	return &UIConfig{
		IndexFormat: "", // deprecated
		IndexColumns: []*ColumnDef{
			{
				Name:     "date",
				Width:    20,
				Flags:    ALIGN_LEFT | WIDTH_EXACT,
				Template: date,
			},
			{
				Name:     "name",
				Width:    17,
				Flags:    ALIGN_LEFT | WIDTH_EXACT,
				Template: name,
			},
			{
				Name:     "flags",
				Width:    4,
				Flags:    ALIGN_RIGHT | WIDTH_EXACT,
				Template: flags,
			},
			{
				Name:     "subject",
				Flags:    ALIGN_LEFT | WIDTH_AUTO,
				Template: subject,
			},
		},
		ColumnSeparator:    "  ",
		AutoMarkRead:       true,
		TimestampFormat:    "2006-01-02 03:04 PM",
		ThisDayTimeFormat:  "",
		ThisWeekTimeFormat: "",
		ThisYearTimeFormat: "",
		ShowHeaders: []string{
			"From", "To", "Cc", "Bcc", "Subject", "Date",
		},
		RenderAccountTabs:   "auto",
		PinnedTabMarker:     "`",
		SidebarWidth:        20,
		PreviewHeight:       12,
		EmptyMessage:        "(no messages)",
		EmptyDirlist:        "(no folders)",
		MouseEnabled:        false,
		ClientThreadsDelay:  50 * time.Millisecond,
		NewMessageBell:      true,
		FuzzyComplete:       false,
		Spinner:             "[..]    , [..]   ,  [..]  ,   [..] ,    [..],   [..] ,  [..]  , [..]   ",
		SpinnerDelimiter:    ",",
		IconUnencrypted:     "",
		IconSigned:          "[s]",
		IconEncrypted:       "[e]",
		IconSignedEncrypted: "",
		IconUnknown:         "[s?]",
		IconInvalid:         "[s!]",
		IconAttachment:      "a",
		DirListFormat:       "%n %>r",
		DirListDelay:        200 * time.Millisecond,
		NextMessageOnDelete: true,
		CompletionDelay:     250 * time.Millisecond,
		CompletionMinChars:  1,
		CompletionPopovers:  true,
		StyleSetDirs:        []string{},
		StyleSetName:        "default",
		// border defaults
		BorderCharVertical:   ' ',
		BorderCharHorizontal: ' ',
		// private
		contextualCache:  make(map[uiContextKey]*UIConfig),
		contextualCounts: make(map[uiContextType]int),
	}
}

var Ui = defaultUiConfig()

func parseUi(file *ini.File) error {
	if ui, err := file.GetSection("ui"); err == nil {
		if err := Ui.parse(ui); err != nil {
			return err
		}
	}

	for _, sectionName := range file.SectionStrings() {
		if !strings.Contains(sectionName, "ui:") {
			continue
		}

		uiSection, err := file.GetSection(sectionName)
		if err != nil {
			return err
		}
		uiSubConfig := UIConfig{}
		if err := uiSubConfig.parse(uiSection); err != nil {
			return err
		}
		contextualUi := UiConfigContext{
			UiConfig: &uiSubConfig,
		}

		var index int
		switch {
		case strings.Contains(sectionName, "~"):
			index = strings.Index(sectionName, "~")
			regex := string(sectionName[index+1:])
			contextualUi.Regex, err = regexp.Compile(regex)
			if err != nil {
				return err
			}
		case strings.Contains(sectionName, "="):
			index = strings.Index(sectionName, "=")
			value := string(sectionName[index+1:])
			contextualUi.Regex, err = regexp.Compile(regexp.QuoteMeta(value))
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("Invalid Ui Context regex in %s", sectionName)
		}

		switch sectionName[3:index] {
		case "account":
			contextualUi.ContextType = uiContextAccount
		case "folder":
			contextualUi.ContextType = uiContextFolder
		case "subject":
			contextualUi.ContextType = uiContextSubject
		default:
			return fmt.Errorf("Unknown Contextual Ui Section: %s", sectionName)
		}
		Ui.contextualUis = append(Ui.contextualUis, &contextualUi)
		Ui.contextualCounts[contextualUi.ContextType]++
	}

	// append default paths to styleset-dirs
	for _, dir := range SearchDirs {
		Ui.StyleSetDirs = append(
			Ui.StyleSetDirs, path.Join(dir, "stylesets"),
		)
	}

	if err := Ui.loadStyleSet(Ui.StyleSetDirs); err != nil {
		return err
	}

	for _, contextualUi := range Ui.contextualUis {
		if contextualUi.UiConfig.StyleSetName == "" &&
			len(contextualUi.UiConfig.StyleSetDirs) == 0 {
			continue // no need to do anything if nothing is overridden
		}
		// fill in the missing part from the base
		if contextualUi.UiConfig.StyleSetName == "" {
			contextualUi.UiConfig.StyleSetName = Ui.StyleSetName
		} else if len(contextualUi.UiConfig.StyleSetDirs) == 0 {
			contextualUi.UiConfig.StyleSetDirs = Ui.StyleSetDirs
		}
		// since at least one of them has changed, load the styleset
		if err := contextualUi.UiConfig.loadStyleSet(
			contextualUi.UiConfig.StyleSetDirs); err != nil {
			return err
		}
	}

	log.Debugf("aerc.conf: [ui] %#v", Ui)

	return nil
}

func (config *UIConfig) parse(section *ini.Section) error {
	if err := section.MapTo(config); err != nil {
		return err
	}

	if key, err := section.GetKey("border-char-vertical"); err == nil {
		chars := []rune(key.String())
		if len(chars) != 1 {
			return fmt.Errorf("%v must be one and only one character", key)
		}
		config.BorderCharVertical = chars[0]
	}
	if key, err := section.GetKey("border-char-horizontal"); err == nil {
		chars := []rune(key.String())
		if len(chars) != 1 {
			return fmt.Errorf("%v must be one and only one character", key)
		}
		config.BorderCharHorizontal = chars[0]
	}

	// Values with type=time.Duration must be explicitly set. If these
	// values are given a default in the struct passed to ui.MapTo, which
	// they are, a zero-value in the config won't overwrite the default.
	if key, err := section.GetKey("dirlist-delay"); err == nil {
		dur, err := key.Duration()
		if err != nil {
			return err
		}
		config.DirListDelay = dur
	}
	if key, err := section.GetKey("completion-delay"); err == nil {
		dur, err := key.Duration()
		if err != nil {
			return err
		}
		config.CompletionDelay = dur
	}

	if config.MessageViewTimestampFormat == "" {
		config.MessageViewTimestampFormat = config.TimestampFormat
	}
	if config.MessageViewThisDayTimeFormat == "" {
		config.MessageViewThisDayTimeFormat = config.TimestampFormat
	}
	if config.MessageViewThisWeekTimeFormat == "" {
		config.MessageViewThisWeekTimeFormat = config.TimestampFormat
	}
	if config.MessageViewThisDayTimeFormat == "" {
		config.MessageViewThisDayTimeFormat = config.TimestampFormat
	}

	if config.IndexFormat != "" {
		log.Warnf("%s %s",
			"The index-format setting has been replaced by index-columns.",
			"index-format will be removed in aerc 0.17.")
	}
	if key, err := section.GetKey("index-columns"); err == nil {
		columns, err := ParseColumnDefs(key, section)
		if err != nil {
			return err
		}
		config.IndexColumns = columns
		config.IndexFormat = "" // to silence popup at startup
	} else if config.IndexFormat != "" {
		columns, err := convertIndexFormat(config.IndexFormat)
		if err != nil {
			return err
		}
		config.IndexColumns = columns
	}

	return nil
}

var indexFmtRegexp = regexp.MustCompile(`%(-?\d+)?(\.\d+)?([A-Za-z%])`)

func convertIndexFormat(indexFormat string) ([]*ColumnDef, error) {
	matches := indexFmtRegexp.FindAllStringSubmatch(indexFormat, -1)
	if matches == nil {
		return nil, fmt.Errorf("invalid index-format")
	}

	var columns []*ColumnDef

	for _, m := range matches {
		alignWidth := m[1]
		verb := m[3]

		var f string
		var width float64 = 0
		var flags ColumnFlags = ALIGN_LEFT
		name := ""

		switch verb {
		case "%":
			f = verb
		case "a":
			f = `{{(index .From 0).Address}}`
			name = "sender"
		case "A":
			f = `{{if eq (len .ReplyTo) 0}}{{(index .From 0).Address}}{{else}}{{(index .ReplyTo 0).Address}}{{end}}`
			name = "reply-to"
		case "C":
			f = "{{.Number}}"
			name = "num"
		case "d", "D":
			f = "{{.DateAutoFormat .Date.Local}}"
			name = "date"
		case "f":
			f = `{{index (.From | persons) 0}}`
			name = "from"
		case "F":
			f = `{{.Peer | names | join ", "}}`
			name = "peers"
		case "g":
			f = `{{.Labels | join ", "}}`
			name = "labels"
		case "i":
			f = "{{.MessageId}}"
			name = "msg-id"
		case "n":
			f = `{{index (.From | names) 0}}`
			name = "name"
		case "r":
			f = `{{.To | persons | join ", "}}`
			name = "to"
		case "R":
			f = `{{.Cc | persons | join ", "}}`
			name = "cc"
		case "s":
			f = "{{.Subject}}"
			name = "subject"
		case "t":
			f = "{{(index .To 0).Address}}"
			name = "to0"
		case "T":
			f = "{{.Account}}"
			name = "account"
		case "u":
			f = "{{index (.From | mboxes) 0}}"
			name = "mboxes"
		case "v":
			f = "{{index (.From | names) 0}}"
			name = "name"
		case "Z":
			f = `{{.Flags | join ""}}`
			name = "flags"
			width = 4
			flags = ALIGN_RIGHT
		case "l":
			f = "{{.Size}}"
			name = "size"
		default:
			f = "%" + verb
		}
		if name == "" {
			name = "wtf"
		}

		t, err := templates.ParseTemplate(fmt.Sprintf("column-%s", name), f)
		if err != nil {
			return nil, err
		}

		if alignWidth != "" {
			width, err = strconv.ParseFloat(alignWidth, 64)
			if err != nil {
				return nil, err
			}
			if width < 0 {
				width = -width
			} else {
				flags = ALIGN_RIGHT
			}
		}
		if width == 0 {
			flags |= WIDTH_AUTO
		} else {
			flags |= WIDTH_EXACT
		}

		columns = append(columns, &ColumnDef{
			Name:     name,
			Width:    width,
			Flags:    flags,
			Template: t,
		})
	}

	return columns, nil
}

func (ui *UIConfig) loadStyleSet(styleSetDirs []string) error {
	ui.style = NewStyleSet()
	err := ui.style.LoadStyleSet(ui.StyleSetName, styleSetDirs)
	if err != nil {
		return fmt.Errorf("Unable to load %q styleset: %w",
			ui.StyleSetName, err)
	}

	return nil
}

func (base *UIConfig) mergeContextual(
	contextType uiContextType, s string,
) *UIConfig {
	for _, contextualUi := range base.contextualUis {
		if contextualUi.ContextType != contextType {
			continue
		}
		if !contextualUi.Regex.Match([]byte(s)) {
			continue
		}
		// Try to make this as lightweight as possible and avoid copying
		// the base UIConfig object unless necessary.
		ui := *base
		err := mergo.Merge(&ui, contextualUi.UiConfig, mergo.WithOverride)
		if err != nil {
			log.Warnf("merge ui failed: %v", err)
		}
		ui.contextualCache = make(map[uiContextKey]*UIConfig)
		if contextualUi.UiConfig.StyleSetName != "" {
			ui.style = contextualUi.UiConfig.style
		}
		return &ui
	}
	return base
}

func (uiConfig *UIConfig) GetStyle(so StyleObject) tcell.Style {
	return uiConfig.style.Get(so)
}

func (uiConfig *UIConfig) GetStyleSelected(so StyleObject) tcell.Style {
	return uiConfig.style.Selected(so)
}

func (uiConfig *UIConfig) GetComposedStyle(base StyleObject,
	styles []StyleObject,
) tcell.Style {
	return uiConfig.style.Compose(base, styles)
}

func (uiConfig *UIConfig) GetComposedStyleSelected(
	base StyleObject, styles []StyleObject,
) tcell.Style {
	return uiConfig.style.ComposeSelected(base, styles)
}

func (base *UIConfig) contextual(
	ctxType uiContextType, value string, useCache bool,
) *UIConfig {
	if base.contextualCounts[ctxType] == 0 {
		// shortcut if no contextual ui for that type
		return base
	}
	if !useCache {
		return base.mergeContextual(ctxType, value)
	}
	key := uiContextKey{ctxType: ctxType, value: value}
	c, found := base.contextualCache[key]
	if !found {
		c = base.mergeContextual(ctxType, value)
		base.contextualCache[key] = c
	}
	return c
}

func (base *UIConfig) ForAccount(account string) *UIConfig {
	return base.contextual(uiContextAccount, account, true)
}

func (base *UIConfig) ForFolder(folder string) *UIConfig {
	return base.contextual(uiContextFolder, folder, true)
}

func (base *UIConfig) ForSubject(subject string) *UIConfig {
	// TODO: this [ui:subject] contextual config should be dropped and
	// replaced by another solution. Possibly something in the stylesets.
	// Do not use a cache for contextual subject config as this
	// could consume all available memory given enough time and
	// enough messages.
	return base.contextual(uiContextSubject, subject, false)
}
