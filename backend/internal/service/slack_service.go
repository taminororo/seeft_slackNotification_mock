package service

import (
	"fmt"
	"log"

	"seeft-slack-notification/internal/config"

	"github.com/slack-go/slack"
)

// NotificationPayload 通知に必要なデータの塊
type NotificationPayload struct {
	ActionType  string // "CREATE", "UPDATE", "DELETE"
	UserName    string
	SlackUserID string
	Date        string
	TimeID      int
	Weather     string
	TaskName    string // 新しいタスク名（削除の場合は空）
	OldTaskName string // 古いタスク名（新規の場合は空）
}

type SlackService struct {
	client            *slack.Client
	channelID         string
	notificationQueue chan NotificationPayload // ★これが「キュー」です
}

const (
	BaseTimeID  = 25
	BaseHour    = 6
	MinutesStep = 30
	QueueSize   = 100 // 一度に溜められる通知の数
)

func NewSlackService(cfg *config.Config) *SlackService {
	s := &SlackService{
		client:            slack.New(cfg.SlackBotToken),
		channelID:         cfg.SlackChannelID,
		notificationQueue: make(chan NotificationPayload, QueueSize),
	}

	// ★裏で動く「配送係」を起動する
	go s.runWorker()

	return s
}

// EnqueueNotification 通知をキューに追加する（呼び出し元は待たされない）
func (s *SlackService) EnqueueNotification(payload NotificationPayload) {
	select {
	case s.notificationQueue <- payload:
		// キューに入れたらすぐ戻る
	default:
		// キューが満杯ならログを出して諦める（ブロッキング防止）
		log.Println("Error: Slack notification queue is full, dropping message")
	}
}

// runWorker キューから取り出して送信する（裏方）
func (s *SlackService) runWorker() {
	for payload := range s.notificationQueue {
		if err := s.send(payload); err != nil {
			log.Printf("Failed to send slack notification: %v", err)
		}
	}
}

// send 実際にSlackに送信する内部関数
func (s *SlackService) send(p NotificationPayload) error {
	blocks := s.buildMessageBlocks(p)

	// 1. チャンネルに送信
	_, _, err := s.client.PostMessage(
		s.channelID,
		slack.MsgOptionBlocks(blocks...),
	)
	if err != nil {
		return fmt.Errorf("channel send error: %w", err)
	}

	// 2. 本人にDM送信 (IDがある場合のみ)
	if p.SlackUserID != "" {
		_, _, err = s.client.PostMessage(
			p.SlackUserID,
			slack.MsgOptionBlocks(blocks...),
		)
		if err != nil {
			// DM失敗はログだけ出してエラーにはしない（チャンネルには届いているため）
			log.Printf("dm send error for user %s: %v", p.UserName, err)
		}
	}

	return nil
}

// buildMessageBlocks リッチなメッセージを作成
func (s *SlackService) buildMessageBlocks(p NotificationPayload) []slack.Block {
	timeStr := s.timeIDToTimeString(p.TimeID)

	// アクションごとの色とタイトル設定
	var title, emoji string
	// SlackのBlockKitでは直接色は指定できないが、Attachmentを使うか、絵文字で表現する
	switch p.ActionType {
	case "CREATE":
		title = "シフト追加"
		emoji = ":sparkles:" // キラキラ
	case "UPDATE":
		title = "シフト変更"
		emoji = ":pencil2:" // 鉛筆
	case "DELETE":
		title = "シフト削除"
		emoji = ":wastebasket:" // ゴミ箱
	default:
		title = "お知らせ"
		emoji = ":mega:"
	}

	headerText := fmt.Sprintf("%s %s通知", emoji, title)
	headerBlock := slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", headerText, false, false))

	// 基本情報
	fields := []*slack.TextBlockObject{
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*ユーザー:*\n%s", p.UserName), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*日付:*\n%s", p.Date), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*時刻:*\n%s", timeStr), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*天気:*\n%s", p.Weather), false, false),
	}

	// 差分情報
	if p.ActionType == "UPDATE" {
		fields = append(fields,
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*変更前:*\n~%s~", p.OldTaskName), false, false),
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*変更後:*\n*%s*", p.TaskName), false, false),
		)
	} else if p.ActionType == "CREATE" {
		fields = append(fields,
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*タスク:*\n%s", p.TaskName), false, false),
		)
	} else if p.ActionType == "DELETE" {
		fields = append(fields,
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*削除されたタスク:*\n~%s~", p.OldTaskName), false, false),
		)
	}

	sectionBlock := slack.NewSectionBlock(nil, fields, nil)
	dividerBlock := slack.NewDividerBlock() // 区切り線

	return []slack.Block{headerBlock, sectionBlock, dividerBlock}
}

func (s *SlackService) timeIDToTimeString(timeID int) string {
	hoursFromBase := (timeID - BaseTimeID) / 2
	minutesFromBase := ((timeID - BaseTimeID) % 2) * 30
	hours := 6 + hoursFromBase
	minutes := minutesFromBase
	return fmt.Sprintf("%02d:%02d", hours, minutes)
}
