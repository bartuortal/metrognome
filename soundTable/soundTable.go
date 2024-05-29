package soundTable

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
)

var VolumeChangeAmount = 5

type SoundRow struct {
	Id       int
	Filename string
	Freq     int
	Volume   int
	table    *SoundTable
}

func (sr *SoundRow) ChangeVolumeBy(amount int) {
	sr.Volume += amount
	sr.applyToRow()
}

func (sr *SoundRow) IncrementVolume() {
	sr.ChangeVolumeBy(VolumeChangeAmount)
}

func (sr *SoundRow) DecrementVolume() {
	sr.ChangeVolumeBy(-VolumeChangeAmount)
}

func (sr *SoundRow) applyToRow() {
	newRow := []string{fmt.Sprint(sr.Id), sr.Filename, fmt.Sprint(sr.Freq), fmt.Sprint(sr.Volume)}
	rows := sr.table.Rows()
	for i := 0; i < len(rows); i++ {
		if rows[i][1] == sr.Filename {
			rows[i] = newRow
		}
	}
}

func NewFromTable(table table.Model) SoundTable {
	return SoundTable{table, nil}
}

type SoundTable struct {
	table.Model
	rows []SoundRow
}

func (st *SoundTable) AddRow(filename string) {
	st.AddRowWithFreq(filename, 0)
}

func (st *SoundTable) AddRowWithFreq(filename string, freq int) {
	rows := st.Rows()
	rows = append(rows, []string{fmt.Sprint(len(rows)), filename, fmt.Sprint(freq), "100"})
	st.SetRows(rows)
}

func (st SoundTable) SelectedSoundRow() SoundRow {
	selected := st.SelectedRow()
	id, _ := strconv.Atoi(selected[0])
	filename := selected[1]
	freq, _ := strconv.Atoi(selected[2])
	volume, _ := strconv.Atoi(selected[3])
	return SoundRow{
		Id:       id,
		Filename: filename,
		Freq:     freq,
		Volume:   volume,
		table:    &st,
	}
}
