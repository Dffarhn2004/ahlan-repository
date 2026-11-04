package httpapi

import (
	"bytes"
	"time"

	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"ponpes-saas/internal/repository"
)

type telegramExportTurn struct {
	Username   string
	ChatID     int64
	CreatedAt  time.Time
	Question   string
	Answer     string
	AnsweredAt *time.Time
}

// pairTelegramExportRows matches each bot answer to the oldest unanswered question of its conversation.
func pairTelegramExportRows(rows []repository.TelegramExportRow) []telegramExportTurn {
	turns := []telegramExportTurn{}
	pending := map[uuid.UUID][]int{}
	for _, row := range rows {
		switch row.SenderType {
		case "USER":
			turns = append(turns, telegramExportTurn{Username: row.Username, ChatID: row.ChatID, CreatedAt: row.CreatedAt, Question: row.MessageText})
			pending[row.ConversationID] = append(pending[row.ConversationID], len(turns)-1)
		case "BOT":
			queue := pending[row.ConversationID]
			if len(queue) == 0 {
				turns = append(turns, telegramExportTurn{Username: row.Username, ChatID: row.ChatID, CreatedAt: row.CreatedAt, Answer: row.MessageText, AnsweredAt: &row.CreatedAt})
				continue
			}
			index := queue[0]
			pending[row.ConversationID] = queue[1:]
			turns[index].Answer = row.MessageText
			turns[index].AnsweredAt = &row.CreatedAt
		}
	}
	return turns
}

func telegramWorkbook(turns []telegramExportTurn) ([]byte, error) {
	book := excelize.NewFile()
	defer func() { _ = book.Close() }()
	sheet := "Log Telegram"
	_ = book.SetSheetName("Sheet1", sheet)
	bold, err := book.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true, Color: "000000"}, Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1}, Border: []excelize.Border{{Type: "bottom", Color: "B7C5BC", Style: 1}}})
	if err != nil {
		return nil, err
	}
	wrap, err := book.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{Vertical: "top", WrapText: true}})
	if err != nil {
		return nil, err
	}
	for column, width := range map[string]float64{"A": 22, "B": 24, "C": 48, "D": 70, "E": 22} {
		if err := book.SetColWidth(sheet, column, column, width); err != nil {
			return nil, err
		}
	}
	headers := []string{"Waktu", "Pengguna", "Pertanyaan pengguna", "Balasan Hermes", "Waktu balasan"}
	for index, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(index+1, 1)
		if err := book.SetCellValue(sheet, cell, header); err != nil {
			return nil, err
		}
		if err := book.SetCellStyle(sheet, cell, cell, bold); err != nil {
			return nil, err
		}
	}
	for rowIndex, turn := range turns {
		row := rowIndex + 2
		person := "Pengguna Telegram"
		if turn.Username != "" {
			person = "@" + turn.Username
		}
		for column, value := range []any{turn.CreatedAt, person, turn.Question, turn.Answer, turn.AnsweredAt} {
			cell, _ := excelize.CoordinatesToCellName(column+1, row)
			if err := book.SetCellValue(sheet, cell, value); err != nil {
				return nil, err
			}
			if err := book.SetCellStyle(sheet, cell, cell, wrap); err != nil {
				return nil, err
			}
		}
	}
	if err := book.SetPanes(sheet, &excelize.Panes{Freeze: true, YSplit: 1}); err != nil {
		return nil, err
	}
	var output bytes.Buffer
	if err := book.Write(&output); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
