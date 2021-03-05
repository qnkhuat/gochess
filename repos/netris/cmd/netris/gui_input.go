package main

import (
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/tslocum/tview"

	"git.sr.ht/~tslocum/netris/pkg/event"
	"github.com/gdamore/tcell"
)

type Keybinding struct {
	k tcell.Key
	r rune
	m tcell.ModMask

	a event.GameAction
}

var keybindings = []*Keybinding{
	{r: 'z', a: event.ActionRotateCCW},
	{r: 'Z', a: event.ActionRotateCCW},
	{r: 'x', a: event.ActionRotateCW},
	{r: 'X', a: event.ActionRotateCW},
	{k: tcell.KeyLeft, a: event.ActionMoveLeft},
	{r: 'h', a: event.ActionMoveLeft},
	{r: 'H', a: event.ActionMoveLeft},
	{k: tcell.KeyDown, a: event.ActionSoftDrop},
	{r: 'j', a: event.ActionSoftDrop},
	{r: 'J', a: event.ActionSoftDrop},
	{k: tcell.KeyUp, a: event.ActionHardDrop},
	{r: 'k', a: event.ActionHardDrop},
	{r: 'K', a: event.ActionHardDrop},
	{k: tcell.KeyRight, a: event.ActionMoveRight},
	{r: 'l', a: event.ActionMoveRight},
	{r: 'L', a: event.ActionMoveRight},
}

var draftKeybindings []*Keybinding

func scrollMessages(direction int) {
	var scroll int
	if showLogLines > 3 {
		scroll = (showLogLines - 2) * direction
	} else {
		scroll = showLogLines * direction
	}

	r, _ := recent.GetScrollOffset()
	r += scroll
	if r < 0 {
		r = 0
	}
	recent.ScrollTo(r, 0)

	draw <- event.DrawMessages
}

func handleKeypress(ev *tcell.EventKey) *tcell.EventKey {
	k := ev.Key()
	r := ev.Rune()

	if capturingKeybind {
		capturingKeybind = false
		if k == tcell.KeyEscape {
			draftKeybindings = nil

			app.SetRoot(gameSettingsContainerGrid, true)
			updateGameSettings()

			return nil
		}

		for i, bind := range draftKeybindings {
			if (bind.k != 0 && bind.k != k) || (bind.r != 0 && bind.r != r) || (bind.m != 0 && bind.m != ev.Modifiers()) {
				continue
			}

			draftKeybindings = append(draftKeybindings[:i], draftKeybindings[i+1:]...)
			break
		}

		var action event.GameAction
		switch gameSettingsSelectedButton {
		case 0:
			action = event.ActionRotateCCW
		case 1:
			action = event.ActionRotateCW
		case 2:
			action = event.ActionMoveLeft
		case 3:
			action = event.ActionMoveRight
		case 4:
			action = event.ActionSoftDrop
		case 5:
			action = event.ActionHardDrop
		default:
			log.Fatal("setting keybind for unknown action")
		}

		draftKeybindings = append(draftKeybindings, &Keybinding{k: k, r: r, m: ev.Modifiers(), a: action})

		app.SetRoot(gameSettingsContainerGrid, true)
		updateGameSettings()
		return nil
	} else if titleVisible {
		if titleScreen > 1 {
			switch k {
			case tcell.KeyEscape:
				titleScreen = 1
				titleSelectedButton = 0

				app.SetRoot(titleContainerGrid, true)
				updateTitle()
				return nil
			}

			if titleScreen == 3 {
				switch k {
				case tcell.KeyTab:
					gameSettingsSelectedButton++
					if gameSettingsSelectedButton > 7 {
						gameSettingsSelectedButton = 7
					}

					updateGameSettings()
					return nil
				case tcell.KeyBacktab:
					gameSettingsSelectedButton--
					if gameSettingsSelectedButton < 0 {
						gameSettingsSelectedButton = 0
					}

					updateGameSettings()
					return nil
				case tcell.KeyEnter:
					if gameSettingsSelectedButton == 6 || gameSettingsSelectedButton == 7 {
						if gameSettingsSelectedButton == 7 {
							keybindings = make([]*Keybinding, len(draftKeybindings))
							copy(keybindings, draftKeybindings)
						}
						draftKeybindings = nil

						titleScreen = 1
						titleSelectedButton = 0

						app.SetRoot(titleContainerGrid, true)
						updateTitle()
						return nil
					}

					modal := tview.NewModal().SetText("Press desired key(s) to set keybinding or press Escape to cancel.").ClearButtons()
					app.SetRoot(modal, true)

					capturingKeybind = true
					return nil
				}
			}

			return ev
		}

		switch k {
		case tcell.KeyEnter:
			if titleScreen == 1 {
				switch titleSelectedButton {
				case 0:
					resetPlayerSettingsForm()

					titleScreen = 2
					titleSelectedButton = 0

					app.SetRoot(playerSettingsContainerGrid, true)
					app.SetFocus(playerSettingsForm)
					app.Draw()
					return nil
				case 1:
					titleScreen = 3
					titleSelectedButton = 0
					gameSettingsSelectedButton = 0

					draftKeybindings = make([]*Keybinding, len(keybindings))
					copy(draftKeybindings, keybindings)

					app.SetRoot(gameSettingsContainerGrid, true)
					app.SetFocus(buttonKeybindRotateCCW)
					app.Draw()
					return nil
				case 2:
					titleScreen = 0
					titleSelectedButton = 0

					updateTitle()
					return nil
				}
			} else {
				if joinedGame {
					switch titleSelectedButton {
					case 0:
						setTitleVisible(false)
						return nil
					case 1:
						titleScreen = 1
						titleSelectedButton = 0

						updateTitle()
						return nil
					case 2:
						done <- true
						return nil
					}
				} else {
					switch titleSelectedButton {
					case 0:
						selectMode <- event.ModePlayOnline
						return nil
					case 1:
						selectMode <- event.ModePractice
						return nil
					case 2:
						titleScreen = 1
						titleSelectedButton = 0

						updateTitle()
						return nil
					}
				}
			}
			return nil
		case tcell.KeyUp, tcell.KeyBacktab:
			previousTitleButton()
			updateTitle()
			return nil
		case tcell.KeyDown, tcell.KeyTab:
			nextTitleButton()
			updateTitle()
			return nil
		case tcell.KeyEscape:
			if titleScreen == 1 {
				titleScreen = 0
				titleSelectedButton = 0
				updateTitle()
			} else if joinedGame {
				setTitleVisible(false)
			} else {
				done <- true
			}
			return nil
		default:
			switch r {
			case 'k', 'K':
				previousTitleButton()
				updateTitle()
				return nil
			case 'j', 'J':
				nextTitleButton()
				updateTitle()
				return nil
			}
		}

		return ev
	}

	if inputActive {
		switch k {
		case tcell.KeyEnter:
			msg := inputView.GetText()
			if msg != "" {
				if strings.HasPrefix(msg, "/cpu") {
					if profileCPU == nil {
						if len(msg) < 5 {
							logMessage("Profile name must be specified")
						} else {
							profileName := strings.TrimSpace(msg[5:])

							var err error
							profileCPU, err = os.Create(profileName)
							if err != nil {
								log.Fatal(err)
							}

							err = pprof.StartCPUProfile(profileCPU)
							if err != nil {
								log.Fatal(err)
							}

							logMessage(fmt.Sprintf("Started profiling CPU usage as %s", profileName))
						}
					} else {
						pprof.StopCPUProfile()
						profileCPU.Close()
						profileCPU = nil

						logMessage("Stopped profiling CPU usage")
					}
				} else {
					if activeGame != nil {
						activeGame.Event <- &event.MessageEvent{Message: msg}
					} else {
						logMessage("Message not sent - not currently connected to any game")
					}
				}
			}

			setInputStatus(false)
			return nil
		case tcell.KeyPgUp:
			scrollMessages(-1)
			return nil
		case tcell.KeyPgDn:
			scrollMessages(1)
			return nil
		case tcell.KeyEscape:
			setInputStatus(false)
			return nil
		}

		return ev
	}

	switch k {
	case tcell.KeyEnter:
		setInputStatus(!inputActive)
		return nil
	case tcell.KeyTab:
		setShowDetails(!showDetails)
		return nil
	case tcell.KeyPgUp:
		scrollMessages(-1)
		return nil
	case tcell.KeyPgDn:
		scrollMessages(1)
		return nil
	case tcell.KeyEscape:
		setTitleVisible(true)
		return nil
	}

	for _, bind := range keybindings {
		if (bind.k != 0 && bind.k != k) || (bind.r != 0 && bind.r != r) || (bind.m != 0 && bind.m != ev.Modifiers()) {
			continue
		} else if activeGame == nil {
			break
		}

		activeGame.ProcessAction(bind.a)
		return nil
	}

	return ev
}
