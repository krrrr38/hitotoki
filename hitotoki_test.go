package hitotoki

import (
	"bytes"
	"github.com/PuerkitoBio/goquery"
	"os"
	"reflect"
	"testing"
)

func Test_parseHistory(t *testing.T) {
	type args struct {
		area string
		file string
	}
	tests := []struct {
		name string
		args args
		want HitotokiAreaHistory
	}{
		{
			name: "ok",
			args: args{
				area: "豊洲みずべ",
				file: "has_ok.html",
			},
			want: HitotokiAreaHistory{
				Area: "豊洲みずべ",
				Months: []HitotokiMonthlyHistory{
					{
						YearMonth: "2023年11月",
						Days: []HitotokiDailyHistory{
							{Day: 8, AM: 2, PM: 0},
							{Day: 9, AM: 0, PM: 0},
							{Day: 10, AM: 0, PM: 0},
							{Day: 13, AM: 0, PM: 0},
							{Day: 14, AM: 0, PM: 0},
							{Day: 15, AM: 0, PM: 0},
							{Day: 16, AM: 0, PM: 0},
							{Day: 17, AM: 0, PM: 0},
							{Day: 20, AM: 0, PM: 0},
							{Day: 21, AM: 0, PM: 0},
							{Day: 22, AM: 0, PM: 0},
							{Day: 24, AM: 0, PM: 0},
							{Day: 27, AM: 0, PM: 0},
							{Day: 28, AM: 0, PM: 0},
							{Day: 29, AM: 0, PM: 0},
							{Day: 30, AM: 0, PM: 0},
						},
					},
					{
						YearMonth: "2023年12月",
						Days: []HitotokiDailyHistory{
							{Day: 1, AM: 0, PM: 0},
							{Day: 4, AM: 0, PM: 0},
							{Day: 5, AM: 0, PM: 0},
							{Day: 6, AM: 1, PM: 0},
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file, err := os.ReadFile("testdata/" + tt.args.file)
			if err != nil {
				t.Errorf("testdata file cannot read: %v", err)
			}
			doc, err := goquery.NewDocumentFromReader(bytes.NewReader(file))
			if got := parseHistory(tt.args.area, doc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseHistory() = %v, want %v", got, tt.want)
			}
		})
	}
}
