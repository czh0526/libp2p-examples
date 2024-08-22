package main

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"io"
	"time"
)

type ChatUI struct {
	chatroom  *ChatRoom
	app       *tview.Application
	peersList *tview.TextView

	msgW    io.Writer
	inputCh chan string
	doneCh  chan struct{}
}

func NewChatUI(chatroom *ChatRoom) *ChatUI {
	app := tview.NewApplication()

	msgBox := tview.NewTextView()
	msgBox.SetDynamicColors(true)
	msgBox.SetBorder(true)
	msgBox.SetTitle(fmt.Sprintf("Room: %s", chatroom.roomName))

	peersList := tview.NewTextView()
	peersList.SetBorder(true)
	peersList.SetTitle("Peers")

	chatPanel := tview.NewFlex().
		AddItem(msgBox, 0, 1, false).
		AddItem(peersList, 35, 1, false)

	// 用户输入框
	input := tview.NewInputField().
		SetLabel(chatroom.nick + " > ").
		SetFieldWidth(0).
		SetFieldBackgroundColor(tcell.ColorBlack)
	// 缓存用户输入
	inputCh := make(chan string, 32)
	// 处理用户输入
	input.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}
		line := input.GetText()
		if len(line) == 0 {
			return
		}

		if line == "/quit" {
			app.Stop()
			return
		}
		inputCh <- line
		input.SetText("")
	})

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(chatPanel, 0, 1, false).
		AddItem(input, 1, 1, true)
	app.SetRoot(flex, true)

	return &ChatUI{
		chatroom: chatroom,
		app:      app,
		msgW:     msgBox,
		inputCh:  inputCh,
	}
}

func (ui *ChatUI) Run() error {
	go ui.handleEvents()
	defer ui.end()

	return ui.app.Run()
}

func (ui *ChatUI) end() {
	ui.doneCh <- struct{}{}
}

func (ui *ChatUI) handleEvents() {

	peerRefreshTicker := time.NewTicker(time.Second)
	defer peerRefreshTicker.Stop()

	for {
		select {
		case input := <-ui.inputCh: // 读取用户输入
			{
				err := ui.chatroom.Publish(input)
				if err != nil {
					printErr("publish error: %s", err)
				}
				ui.displaySelfMessage(input)
			}
		case m := <-ui.chatroom.Messages: // 读取聊天室内容
			{
				ui.displayChatMessage(m)
			}
		case <-peerRefreshTicker.C:
			{
				ui.refreshPeers()
			}
		case <-ui.chatroom.ctx.Done(): // 聊天室结束
			{
				return
			}
		case <-ui.doneCh: // 用户退出
			{
				return
			}
		}
	}
}

func (ui *ChatUI) displayChatMessage(cm *ChatMessage) {
	prompt := withColor("green", fmt.Sprintf("<%s>:", cm.SenderNick))
	fmt.Fprintf(ui.msgW, "%s %s\n", prompt, cm.Message)
}

func (ui *ChatUI) displaySelfMessage(m string) {
	prompt := withColor("yellow", fmt.Sprintf("<%s>:", ui.chatroom.nick))
	fmt.Fprintf(ui.msgW, "%s %s\n", prompt, m)
}

func withColor(color, msg string) string {
	return fmt.Sprintf("[%s]%s[-]", color, msg)
}

func (ui *ChatUI) refreshPeers() {
	peers := ui.chatroom.ListPeers()

	ui.peersList.Clear()
	for _, peer := range peers {
		fmt.Fprintf(ui.peersList, shortID(peer))
	}
	ui.app.Draw()
}
