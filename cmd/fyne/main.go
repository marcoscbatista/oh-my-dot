package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/marcoscbatista/oh-my-dot/internal/dotfiles"
)

type Logger struct {
	verbose bool
}

func (l *Logger) Verbosef(format string, args ...any) {
	if l.verbose {
		fmt.Fprintf(os.Stderr, "[VERBOSE] "+format+"\n", args...)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "[ERROR] "+format+"\n", args...)
}

func main() {
	isVerbose := flag.Bool("v", false, "Activate verbose mode")
	flag.Parse()

	log := Logger{verbose: *isVerbose}

	log.Verbosef("Starting oh-my-dot (verbose=%v)", *isVerbose)

	home, err := os.UserHomeDir()
	if err != nil {
		log.Errorf("Failed to resolve home directory: %s", err)
		os.Exit(1)
	}

	store, err := dotfiles.NewDotFileStore("oh-my-dot/db.json")
	if err != nil {
		log.Errorf("Failed to initialize dotfiles store: %s", err)
		os.Exit(1)
	}

	service := dotfiles.DotFilesService{
		Store: store,
	}

	handler := dotfiles.DotFilesHandler{
		Service:     &service,
		DotfilesDir: filepath.Join(home, "oh-my-dot"),
		ConfigPath:  filepath.Join(home, ".fake-config"),
	}

	a := app.New()

	w := a.NewWindow("oh-my-dot")
	w.SetIcon(theme.FolderIcon())
	w.Resize(fyne.NewSize(820, 560))
	w.CenterOnScreen()

	data, err := handler.GetAll()
	if err != nil {
		log.Errorf("Failed to fetch dotfiles list: %s", err)
		dialog.ShowError(err, w)
		os.Exit(1)
	}

	selectedID := -1
	selectedName := ""

	var list *widget.List
	var listCard *widget.Card
	var btnSwitch *widget.Button
	var actions *fyne.Container

	countLabel := widget.NewLabel("")
	countLabel.Importance = widget.LowImportance

	activeLabel := widget.NewLabel("")
	activeLabel.Importance = widget.LowImportance

	statusLabel := widget.NewLabel("Ready")
	statusLabel.Importance = widget.LowImportance

	configLabel := widget.NewLabel(fmt.Sprintf("Config: %s", handler.ConfigPath))
	configLabel.Importance = widget.LowImportance

	title := widget.NewLabelWithStyle(
		"oh-my-dot",
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	subtitle := widget.NewLabel("Manage, add and switch dotfiles packages")
	subtitle.Importance = widget.MediumImportance

	header := container.NewBorder(
		nil,
		nil,
		container.NewHBox(
			widget.NewIcon(theme.FolderIcon()),
			container.NewVBox(title, subtitle),
		),
		container.NewVBox(countLabel, activeLabel),
		nil,
	)

	updateMetaLabels := func() {
		countLabel.SetText(fmt.Sprintf("%d package(s)", len(data)))

		active := "No active package"

		for _, dot := range data {
			if dot.InUse {
				active = "Active: " + dot.Name
				break
			}
		}

		activeLabel.SetText(active)
	}

	emptyTitle := widget.NewLabelWithStyle(
		"No dotfiles packages yet",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	emptySubtitle := widget.NewLabel(
		"Add a Git repository to start managing your dotfiles",
	)
	emptySubtitle.Alignment = fyne.TextAlignCenter
	emptySubtitle.Importance = widget.LowImportance

	emptyState := container.NewCenter(
		container.NewVBox(
			container.NewCenter(widget.NewIcon(theme.FolderIcon())),
			emptyTitle,
			emptySubtitle,
		),
	)

	refreshCardContent := func() {
		if len(data) == 0 {
			listCard.SetContent(emptyState)
		} else {
			listCard.SetContent(list)
		}

		listCard.Refresh()
	}

	refreshData := func() {
		log.Verbosef("Refreshing dotfiles list from store")

		var err error

		data, err = handler.GetAll()
		if err != nil {
			log.Errorf("Failed to refresh dotfiles lit: %s", err)
			dialog.ShowError(err, w)
			return
		}

		selectedID = -1
		selectedName = ""

		if btnSwitch != nil {
			btnSwitch.Hide()
		}

		updateMetaLabels()

		if list != nil {
			list.UnselectAll()
			list.Refresh()
		}

		refreshCardContent()

		statusLabel.SetText("Package list refreshed")
	}

	list = widget.NewList(
		func() int {
			return len(data)
		},
		func() fyne.CanvasObject {
			icon := widget.NewIcon(theme.FolderIcon())

			name := widget.NewLabel("Package name")
			name.TextStyle = fyne.TextStyle{Bold: true}

			idLabel := widget.NewLabel("ID #0")
			idLabel.Importance = widget.LowImportance

			return container.NewHBox(
				icon,
				name,
				layout.NewSpacer(),
				idLabel,
			)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)

			icon := row.Objects[0].(*widget.Icon)
			nameLabel := row.Objects[1].(*widget.Label)
			idLabel := row.Objects[3].(*widget.Label)

			dot := data[i]

			nameLabel.SetText(dot.Name)
			idLabel.SetText(fmt.Sprintf("ID #%d", dot.ID))

			if dot.InUse {
				icon.SetResource(theme.ConfirmIcon())
			} else {
				icon.SetResource(theme.FolderIcon())
			}

			icon.Refresh()
			nameLabel.Refresh()
			idLabel.Refresh()
		},
	)

	list.OnSelected = func(id widget.ListItemID) {
		dot := data[id]

		selectedID = dot.ID
		selectedName = dot.Name

		log.Verbosef("Selected package: id=%d name=%q inUse=%v", dot.ID, dot.Name, dot.InUse)

		if dot.InUse {
			btnSwitch.Hide()
			statusLabel.SetText(fmt.Sprintf("%s is already active", dot.Name))

			if actions != nil {
				actions.Refresh()
			}

			return
		}

		btnSwitch.SetText("Activate " + selectedName)
		btnSwitch.Show()

		if actions != nil {
			actions.Refresh()
		}

		statusLabel.SetText(fmt.Sprintf("Selected: %s", selectedName))
	}

	list.OnUnselected = func(id widget.ListItemID) {
		if id >= 0 && id < len(data) {
			log.Verbosef("Unselected package id=%d", data[id].ID)
		}

		selectedID = -1
		selectedName = ""

		btnSwitch.Hide()

		if actions != nil {
			actions.Refresh()
		}

		statusLabel.SetText("Ready")
	}

	btnCreate := widget.NewButtonWithIcon("Add Package", theme.ContentAddIcon(), func() {
		log.Verbosef("Opening 'Add Package' dialog")

		name := widget.NewEntry()
		name.SetPlaceHolder("personal, work, minimal...")

		remoteAddr := widget.NewEntry()
		remoteAddr.SetPlaceHolder("https://github.com/user/dotfiles.git")

		formItems := []*widget.FormItem{
			widget.NewFormItem("Name", name),
			widget.NewFormItem("Remote URL", remoteAddr),
		}

		formDialog := dialog.NewForm(
			"Add dotfiles package",
			"Add",
			"Cancel",
			formItems,
			func(ok bool) {
				if !ok {
					log.Verbosef("'Add Package' dialog cancelled by user")
					return
				}

				packageName := strings.TrimSpace(name.Text)
				remoteURL := strings.TrimSpace(remoteAddr.Text)

				log.Verbosef("Submitting new package: name=%q remote=%q", packageName, remoteURL)

				if packageName == "" {
					log.Errorf("Validation failed: name is empty")
					dialog.ShowError(fmt.Errorf("name cannot be empty"), w)
					return
				}

				if remoteURL == "" {
					log.Errorf("Validation failed: remote URL is empty")
					dialog.ShowError(fmt.Errorf("remote URL cannot be empty"), w)
					return
				}

				statusLabel.SetText("Adding package...")

				if err := handler.Create(packageName, remoteURL, false); err != nil {
					log.Errorf("Failed to create package %q: %s", packageName, err)
					statusLabel.SetText("Failed to add package")
					dialog.ShowError(err, w)
					return
				}

				log.Verbosef("Package %q created successfully", packageName)

				refreshData()

				statusLabel.SetText(fmt.Sprintf("%s added successfully", packageName))

				dialog.ShowInformation(
					"Package added",
					fmt.Sprintf("%s was added successfully.", packageName),
					w,
				)
			},
			w,
		)

		formDialog.Resize(fyne.NewSize(640, 300))
		formDialog.Show()
	})
	btnCreate.Importance = widget.HighImportance

	btnRefresh := widget.NewButtonWithIcon("Refresh", theme.ViewRefreshIcon(), func() {
		log.Verbosef("Manual refresh requested by user")
		refreshData()
	})

	btnSwitch = widget.NewButtonWithIcon("Activate", theme.ConfirmIcon(), func() {
		if selectedID == -1 {
			log.Verbosef("Activate clicked with no package selected, ignoring")
			return
		}

		activatedID := selectedID
		activatedName := selectedName

		message := fmt.Sprintf(
			"Activate %s?\n\nYour current config path will be replaced if it is not already managed by oh-my-dot.\nA backup will be created in:\n%s",
			activatedName,
			handler.DotfilesDir,
		)

		dialog.ShowConfirm(
			"Activate dotfiles",
			message,
			func(ok bool) {
				if !ok {
					log.Verbosef("Activation cancelled for package %q", activatedName)
					statusLabel.SetText("Activation cancelled")
					return
				}

				log.Verbosef("Activating package: id=%d name=%q", activatedID, activatedName)

				statusLabel.SetText(fmt.Sprintf("Activating %s...", activatedName))

				err := service.Switch(activatedID, handler.ConfigPath, handler.DotfilesDir)
				if err != nil {
					log.Errorf("Failed to activate package %q (id=%d): %s", activatedName, activatedID, err)
					statusLabel.SetText("Activation failed")
					dialog.ShowError(err, w)
					return
				}

				log.Verbosef("Package %q (id=%d) activated successfully", activatedName, activatedID)

				refreshData()

				statusLabel.SetText(fmt.Sprintf("%s is now active", activatedName))

				dialog.ShowInformation(
					"Package activated",
					fmt.Sprintf("%s is now active.", activatedName),
					w,
				)
			},
			w,
		)
	})
	btnSwitch.Importance = widget.SuccessImportance
	btnSwitch.Hide()

	actions = container.NewHBox(
		btnCreate,
		btnRefresh,
		layout.NewSpacer(),
		btnSwitch,
	)

	topBar := container.NewVBox(
		header,
		widget.NewSeparator(),
		actions,
	)

	listCard = widget.NewCard(
		"Packages",
		"Select an inactive package to activate it.",
		list,
	)

	footer := container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(
			statusLabel,
			layout.NewSpacer(),
			configLabel,
		),
	)

	content := container.NewBorder(
		topBar,
		footer,
		nil,
		nil,
		listCard,
	)

	updateMetaLabels()
	refreshCardContent()

	w.SetContent(container.NewPadded(content))

	log.Verbosef("UI ready, showing main window")
	w.ShowAndRun()

	log.Verbosef("Application window closed, exiting")
}
