package service

import (
	"fmt"

	"github.com/slack-go/slack"
	"seeft-slack-notification/internal/config"
	"seeft-slack-notification/internal/model"
)

type SlackService struct {
	client    *slack.Client
	channelID string
}

func NewSlackService(cfg *config.Config) *SlackService {
	return &SlackService{
		client:    slack.New(cfg.SlackBotToken),
		channelID: cfg.SlackChannelID,
	}
}

// SendNotification DMとチャンネルに通知を送信
func (s *SlackService) SendNotification(notification *model.Notification, userName, slackUserID string) error {
	// Block Kitメッセージを構築
	blocks := s.buildMessageBlocks(notification, userName)
	
	// チャンネル送信
	if err := s.sendToChannel(blocks); err != nil {
		return fmt.Errorf("failed to send to channel: %w", err)
	}

	// DM送信
	if err := s.SendDMToUser(slackUserID, blocks); err != nil {
		return fmt.Errorf("failed to send DM: %w", err)
	}

	return nil
}

// sendToChannel チャンネルに送信
func (s *SlackService) sendToChannel(blocks []slack.Block) error {
	_, _, err := s.client.PostMessage(
		s.channelID,
		slack.MsgOptionBlocks(blocks...),
	)
	return err
}

// buildMessageBlocks Block Kitメッセージを構築
func (s *SlackService) buildMessageBlocks(notification *model.Notification, userName string) []slack.Block {
	// 時刻をtimeIDから算出（例: timeID 25 = 6:00）
	timeStr := s.timeIDToTimeString(notification.TimeID)

	// 変更タイプを判定
	changeType := "新規作成"
	if notification.OldTaskName != "" {
		changeType = "変更"
	}

	// ヘッダーセクション
	headerText := fmt.Sprintf("シフト%s通知", changeType)
	headerBlock := slack.NewHeaderBlock(slack.NewTextBlockObject("plain_text", headerText, false, false))

	// 情報セクション
	fields := []*slack.TextBlockObject{
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*ユーザー:*\n%s", userName), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*日付:*\n%s", notification.Date), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*天気:*\n%s", notification.Weather), false, false),
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*時刻:*\n%s", timeStr), false, false),
	}

	if notification.OldTaskName != "" {
		fields = append(fields,
			slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*変更前:*\n%s", notification.OldTaskName), false, false),
		)
	}

	fields = append(fields,
		slack.NewTextBlockObject("mrkdwn", fmt.Sprintf("*変更後:*\n%s", notification.NewTaskName), false, false),
	)

	sectionBlock := slack.NewSectionBlock(nil, fields, nil)

	return []slack.Block{headerBlock, sectionBlock}
}

// timeIDToTimeString timeIDを時刻文字列に変換
// timeID 25 = 6:00 を基準とする
func (s *SlackService) timeIDToTimeString(timeID int) string {
	// timeID 25 = 6:00 なので、timeID - 25 = 0時からの経過時間（30分単位）
	hoursFromBase := (timeID - 25) / 2
	minutesFromBase := ((timeID - 25) % 2) * 30

	hours := 6 + hoursFromBase
	minutes := minutesFromBase

	return fmt.Sprintf("%02d:%02d", hours, minutes)
}

// SendDMToUser Slack User IDを使用してDMを送信
func (s *SlackService) SendDMToUser(slackUserID string, blocks []slack.Block) error {
	_, _, err := s.client.PostMessage(
		slackUserID,
		slack.MsgOptionBlocks(blocks...),
	)
	return err
}

