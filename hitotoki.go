package hitotoki

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	ua = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/118.0.0.0 Safari/537.36 hitotoki/0.0.1"
)

var (
	areas = map[string]string{
		"豊洲みずべ": "https://koto-kosodate-portal.jp/mizube/general/refresh_cal_50.html",
		"有明みずべ": "https://koto-kosodate-portal.jp/mizube/general/refresh_cal_60.html",
	}
	hitotokiBusinessDayRegex = regexp.MustCompile(`(\d+)(?:\s+)AM (\d+)(?:\s+)PM (\d+)`)
)

type MiscHitotokiService struct {
	lineNotifyClient *LineNotifyClient
	storage          Storage
}

func NewMiscHitotokiService(lineNotifyClient *LineNotifyClient, storage Storage) *MiscHitotokiService {
	return &MiscHitotokiService{
		lineNotifyClient: lineNotifyClient,
		storage:          storage,
	}
}

func (s *MiscHitotokiService) SendCanceledNotifies(ctx context.Context) error {
	prev, err := s.storage.GetPrevRecord(ctx)
	if err != nil {
		return fmt.Errorf("failed to get previous record: %v", err)
	}
	log.Printf("previous record: %v\n", prev)

	current, err := s.getLatestRecord()
	if err != nil {
		return fmt.Errorf("failed to get latest record: %v", err)
	}
	log.Printf("current record: %v\n", current)

	if s.isSame(prev, current) {
		log.Println("hitotoki previous and current records are same, no need to notify")
		return nil
	}

	err = s.storage.SetCurrentRecord(ctx, current)
	if err != nil {
		return err
	}
	return s.postDeltaHistories(prev, current)
}

func (s *MiscHitotokiService) getLatestRecord() (*HitotokiRecord, error) {
	var histories []HitotokiAreaHistory
	for area, url := range areas {
		doc, err := fetchDocument(url)
		if err != nil {
			return nil, err
		}
		histories = append(histories, parseHistory(area, doc))
	}

	sumOfDaysRecords := 0
	for _, history := range histories {
		for _, month := range history.Months {
			sumOfDaysRecords += len(month.Days)
		}
	}

	if sumOfDaysRecords == 0 {
		log.Println("fetch hitotoki records with zero, it might that our requests were blocked and so on")
	} else {
		sumOfAvailableCount := 0
		for _, history := range histories {
			for _, month := range history.Months {
				for _, day := range month.Days {
					sumOfAvailableCount += day.AM + day.PM
				}
			}
		}
		log.Printf("collect available hitotoki seats with count %d", sumOfAvailableCount)
	}

	return &HitotokiRecord{AreaHistories: histories}, nil
}

func fetchDocument(url string) (*goquery.Document, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", ua)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch document: %s", resp.Status)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}

func parseHistory(area string, doc *goquery.Document) HitotokiAreaHistory {
	var months []HitotokiMonthlyHistory
	monthEls := doc.Find("#formBody .col_general")
	tableEls := doc.Find("#formBody table")

	monthEls.Each(func(i int, s *goquery.Selection) {
		yearMonth := s.Text()
		table := goquery.NewDocumentFromNode(tableEls.Nodes[i])
		var days []HitotokiDailyHistory
		table.Find("td").Each(func(j int, td *goquery.Selection) {
			txt := td.Text()
			converted := toHalfWidth(strings.ReplaceAll(txt, "×", "0"))
			matched := hitotokiBusinessDayRegex.FindStringSubmatch(converted)
			if matched != nil {
				day := atoi(matched[1])
				am := atoi(matched[2])
				pm := atoi(matched[3])
				days = append(days, HitotokiDailyHistory{Day: day, AM: am, PM: pm})
			}
		})
		months = append(months, HitotokiMonthlyHistory{YearMonth: yearMonth, Days: days})
	})
	return HitotokiAreaHistory{Area: area, Months: months}
}

func toHalfWidth(s string) string {
	var sb strings.Builder
	for _, c := range s {
		if c >= 0xFF10 && c <= 0xFF19 {
			sb.WriteRune(c - 0xFEE0)
		} else {
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

func (s *MiscHitotokiService) postDeltaHistories(prev, current *HitotokiRecord) error {
	canceled := parseCanceledRecord(prev, current)

	if canceled == nil {
		log.Println("no hitotoki canceled records")
		return nil
	}

	var message strings.Builder
	message.WriteString("🍙 ひととき保育キャンセル枠通知 🍙\n\n")
	for _, areaHistory := range canceled.AreaHistories {
		message.WriteString(fmt.Sprintf("エリア: %s\n空き枠:\n", areaHistory.Area))
		for _, month := range areaHistory.Months {
			for _, day := range month.Days {
				message.WriteString(fmt.Sprintf("・%s%d日 AM %d PM %d\n", month.YearMonth, day.Day, day.AM, day.PM))
			}
		}
		message.WriteString(fmt.Sprintf("URL: %s\n", areas[areaHistory.Area]))
		message.WriteString("* * * * * *\n")
	}

	log.Println("send hitotoki canceled delta records")
	err := s.lineNotifyClient.PostMessage(message.String())
	if err != nil {
		return fmt.Errorf("failed to send line notify: %v", err)
	}
	return nil
}

func (s *MiscHitotokiService) isSame(prev, current *HitotokiRecord) bool {
	if prev == nil && current == nil {
		return true
	}
	if (prev == nil && current != nil) || (prev != nil && current == nil) {
		return false
	}

	prevBytes, _ := json.Marshal(prev)
	currentBytes, _ := json.Marshal(current)
	return string(prevBytes) == string(currentBytes)
}

func parseCanceledRecord(prev, current *HitotokiRecord) *HitotokiRecord {
	var canceledAreas []HitotokiAreaHistory
	for _, currentArea := range current.AreaHistories {
		for _, prevArea := range prev.AreaHistories {
			if currentArea.Area == prevArea.Area {
				canceledArea := currentArea.canceled(prevArea)
				if canceledArea != nil {
					canceledAreas = append(canceledAreas, *canceledArea)
				}
			}
		}
	}
	if len(canceledAreas) == 0 {
		return nil
	}
	return &HitotokiRecord{AreaHistories: canceledAreas}
}

type HitotokiRecord struct {
	AreaHistories []HitotokiAreaHistory `json:"areaHistories"`
}

type HitotokiAreaHistory struct {
	Area   string                   `json:"area"`
	Months []HitotokiMonthlyHistory `json:"months"`
}

func (h *HitotokiAreaHistory) canceled(prev HitotokiAreaHistory) *HitotokiAreaHistory {
	var canceledMonths []HitotokiMonthlyHistory
	for _, currentMonth := range h.Months {
		for _, prevMonth := range prev.Months {
			if currentMonth.YearMonth == prevMonth.YearMonth {
				canceledMonth := currentMonth.canceled(prevMonth)
				if canceledMonth != nil {
					canceledMonths = append(canceledMonths, *canceledMonth)
				}
			}
		}
	}
	if len(canceledMonths) == 0 {
		return nil
	}
	return &HitotokiAreaHistory{Area: h.Area, Months: canceledMonths}
}

type HitotokiMonthlyHistory struct {
	YearMonth string                 `json:"yearMonth"`
	Days      []HitotokiDailyHistory `json:"days"`
}

func (m *HitotokiMonthlyHistory) canceled(prev HitotokiMonthlyHistory) *HitotokiMonthlyHistory {
	var canceledDays []HitotokiDailyHistory
	for _, currentDay := range m.Days {
		for _, prevDay := range prev.Days {
			if currentDay.Day == prevDay.Day && currentDay.hasCanceled(prevDay) {
				canceledDays = append(canceledDays, currentDay)
			}
		}
	}
	if len(canceledDays) == 0 {
		return nil
	}
	return &HitotokiMonthlyHistory{YearMonth: m.YearMonth, Days: canceledDays}
}

type HitotokiDailyHistory struct {
	Day int `json:"day"`
	AM  int `json:"am"`
	PM  int `json:"pm"`
}

func (d *HitotokiDailyHistory) hasCanceled(prev HitotokiDailyHistory) bool {
	return (prev.AM == 0 && prev.PM == 0) && (d.AM > 0 || d.PM > 0)
}

func atoi(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}
