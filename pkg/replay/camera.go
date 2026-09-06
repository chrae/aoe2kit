package replay

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"sort"
)

// Camera surfaces the op=3 viewlock stream: per-event camera x/y over game time.
// Honesty boundary: the op=3 payload is 12 bytes. Bytes 0..8 decode as two
// float32 map coordinates (verified against plausible map ranges). Bytes 8..12
// are an UNVERIFIED tail; if its u32 value coincides with a roster player
// number we report that as an inferred hint, never as proven ownership.

type CameraReport struct {
	Path         string              `json:"path,omitempty"`
	Method       string              `json:"method"`
	Verification string              `json:"verification"`
	Summary      CameraSummary       `json:"summary"`
	Streams      []CameraStream      `json:"streams,omitempty"`
	Events       []CameraEvent       `json:"events,omitempty"`
	Warnings     []string            `json:"warnings,omitempty"`
}

type CameraSummary struct {
	Events              int    `json:"events"`
	EmittedEvents       int    `json:"emitted_events"`
	DurationMS          int    `json:"duration_ms"`
	Duration            string `json:"duration"`
	DistinctTails       int    `json:"distinct_tail_values"`
	TailsMatchRoster    bool   `json:"tail_values_coincide_with_player_numbers"`
	TailMappingNote     string `json:"tail_mapping_note"`
}

type CameraStream struct {
	TailU32                uint32  `json:"tail_u32"`
	TailHex                string  `json:"tail_hex"`
	PlayerNumberIfTail     int     `json:"player_number_if_tail_is_player,omitempty"`
	PlayerNameIfTail       string  `json:"player_name_if_tail_is_player,omitempty"`
	MappingConfidence      string  `json:"mapping_confidence"`
	Events                 int     `json:"events"`
	FirstTimeMS            int     `json:"first_time_ms"`
	FirstTime              string  `json:"first_time"`
	LastTimeMS             int     `json:"last_time_ms"`
	LastTime               string  `json:"last_time"`
	MinX                   float64 `json:"min_x"`
	MaxX                   float64 `json:"max_x"`
	MinY                   float64 `json:"min_y"`
	MaxY                   float64 `json:"max_y"`
	DistanceTraveled       float64 `json:"distance_traveled"`
	MeanStepDistance       float64 `json:"mean_step_distance"`
	MaxStepDistance        float64 `json:"max_step_distance"`
	LargeJumps             int     `json:"large_jumps_over_20"`
}

type CameraEvent struct {
	TimeMS       int     `json:"time_ms"`
	Time         string  `json:"time"`
	X            float64 `json:"x"`
	Y            float64 `json:"y"`
	TailU32      uint32  `json:"tail_u32"`
	SourceOffset int     `json:"source_offset"`
}

type CameraOptions struct {
	// Limit caps emitted event rows (0 = all). Summaries always cover every event.
	Limit int
	// Tail filters events/streams to a single tail_u32 value when >= 0.
	Tail int
}

func BuildCamera(path string, opts CameraOptions) (*CameraReport, error) {
	data, err := ReadRecordBytes(path)
	if err != nil {
		return nil, err
	}
	rec, err := Parse(data)
	if err != nil {
		return nil, err
	}
	if rec.HeaderLength > len(data) {
		return nil, fmt.Errorf("record header length %d exceeds file length %d", rec.HeaderLength, len(data))
	}
	report := &CameraReport{
		Path:         path,
		Method:       "body_op3_viewlock_stream",
		Verification: "structure_verified_xy_decoded_tail_ownership_unverified",
	}
	body := data[rec.HeaderLength:]
	reader := bytes.NewReader(body)
	if err := readReplayMeta(reader); err != nil {
		return nil, fmt.Errorf("action-stream meta parse failed: %w", err)
	}

	type tailState struct {
		stream CameraStream
		lastX  float64
		lastY  float64
		seen   bool
	}
	states := map[uint32]*tailState{}
	timeMS := 0
	raw := make([]byte, 12)

walk:
	for {
		opOffset := len(body) - reader.Len()
		op, err := readU32(reader)
		if err != nil {
			if err == io.EOF {
				break walk
			}
			report.Warnings = append(report.Warnings, "body walk stopped: "+err.Error())
			break walk
		}
		switch op {
		case 1:
			if _, _, _, err := readReplayAction(reader); err != nil {
				report.Warnings = append(report.Warnings, "action parse stopped: "+err.Error())
				break walk
			}
		case 2:
			delta, err := readReplaySync(reader)
			if err != nil {
				report.Warnings = append(report.Warnings, "sync parse stopped: "+err.Error())
				break walk
			}
			timeMS += int(delta)
		case 3:
			if _, err := io.ReadFull(reader, raw); err != nil {
				report.Warnings = append(report.Warnings, "viewlock parse stopped: "+err.Error())
				break walk
			}
			x := f32(raw[0:4])
			y := f32(raw[4:8])
			tail := uint32(raw[8]) | uint32(raw[9])<<8 | uint32(raw[10])<<16 | uint32(raw[11])<<24
			report.Summary.Events++
			if opts.Tail >= 0 && tail != uint32(opts.Tail) {
				continue
			}
			st, ok := states[tail]
			if !ok {
				st = &tailState{stream: CameraStream{
					TailU32:     tail,
					TailHex:     hex.EncodeToString(raw[8:12]),
					FirstTimeMS: timeMS,
					FirstTime:   FormatTime(timeMS),
					MinX:        x, MaxX: x, MinY: y, MaxY: y,
				}}
				states[tail] = st
			}
			s := &st.stream
			s.Events++
			s.LastTimeMS = timeMS
			s.LastTime = FormatTime(timeMS)
			s.MinX = math.Min(s.MinX, x)
			s.MaxX = math.Max(s.MaxX, x)
			s.MinY = math.Min(s.MinY, y)
			s.MaxY = math.Max(s.MaxY, y)
			if st.seen {
				step := math.Hypot(x-st.lastX, y-st.lastY)
				s.DistanceTraveled += step
				if step > s.MaxStepDistance {
					s.MaxStepDistance = step
				}
				if step > 20 {
					s.LargeJumps++
				}
			}
			st.lastX, st.lastY, st.seen = x, y, true
			if opts.Limit == 0 || len(report.Events) < opts.Limit {
				report.Events = append(report.Events, CameraEvent{
					TimeMS: timeMS, Time: FormatTime(timeMS),
					X: x, Y: y, TailU32: tail, SourceOffset: opOffset,
				})
			}
		case 4:
			if err := skipN(reader, 4); err != nil {
				report.Warnings = append(report.Warnings, "chat prefix parse stopped: "+err.Error())
				break walk
			}
			length, err := readU32(reader)
			if err != nil {
				report.Warnings = append(report.Warnings, "chat length parse stopped: "+err.Error())
				break walk
			}
			if length > uint32(reader.Len()) {
				report.Warnings = append(report.Warnings, fmt.Sprintf("chat length %d exceeds remaining body %d", length, reader.Len()))
				break walk
			}
			if err := skipN(reader, int(length)); err != nil {
				report.Warnings = append(report.Warnings, "chat payload parse stopped: "+err.Error())
				break walk
			}
		case 5:
			result, err := skipStart(reader)
			if err != nil {
				report.Warnings = append(report.Warnings, "start op parse stopped: "+err.Error())
				break walk
			}
			if result.Terminal {
				break walk
			}
		case 6:
			break walk
		default:
			if err := skipReplaySaveChapter(reader, len(body)); err != nil {
				report.Warnings = append(report.Warnings, fmt.Sprintf("unknown operation id %d at body offset %d: %v", op, opOffset, err))
				break walk
			}
		}
	}

	report.Summary.EmittedEvents = len(report.Events)
	report.Summary.DurationMS = timeMS
	report.Summary.Duration = FormatTime(timeMS)

	// Roster coincidence check (inferred hint, never proven ownership).
	rosterNames := map[int]string{}
	for _, player := range rec.Players {
		if player.Active && player.Number > 0 {
			rosterNames[player.Number] = player.Name
		}
	}
	allMatch := len(states) > 0
	for tail := range states {
		if _, ok := rosterNames[int(tail)]; !ok {
			allMatch = false
			break
		}
	}
	report.Summary.DistinctTails = len(states)
	report.Summary.TailsMatchRoster = allMatch
	if allMatch {
		report.Summary.TailMappingNote = "every distinct tail_u32 equals an active roster player number; player fields below are INFERRED from that coincidence, not proven ownership"
	} else {
		report.Summary.TailMappingNote = "tail_u32 values do not all match roster player numbers; no player mapping inferred"
	}

	for _, st := range states {
		s := st.stream
		if s.Events > 1 {
			s.MeanStepDistance = s.DistanceTraveled / float64(s.Events-1)
		}
		if allMatch {
			if name, ok := rosterNames[int(s.TailU32)]; ok {
				s.PlayerNumberIfTail = int(s.TailU32)
				s.PlayerNameIfTail = name
				s.MappingConfidence = "inferred_tail_matches_roster_number_unverified"
			}
		} else {
			s.MappingConfidence = "unmapped_raw_tail"
		}
		report.Streams = append(report.Streams, s)
	}
	sort.Slice(report.Streams, func(i, j int) bool { return report.Streams[i].TailU32 < report.Streams[j].TailU32 })
	return report, nil
}
