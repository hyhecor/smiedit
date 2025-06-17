package smiedit

import (
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slog"
)

type SMI struct {
}

func (format SMI) Sync(w io.Writer, Text func() string, opt SyncOption) error {

	delta := int64(opt.timestamp / time.Millisecond)

	exp, err := regexp.Compile(`(?i)<SYNC Start=\d+>`)
	if err != nil {
		slog.Error("regexp Compile ", "error", err)
		return err
	}

	fieldStart := "Start="

	expN, err := regexp.Compile(`(?i)Start=\d+`)
	if err != nil {
		slog.Error("regexp Compile ", "error", err)
		return err
	}

	if !exp.MatchString(Text()) {
		slog.Info("not matched", "text", Text())

		fmt.Fprintln(w, Text())
		return nil
	}

	slog.Info("matched", "text", Text())

	indexs := expN.FindAllIndex([]byte(Text()), -1)

	begin := []byte(Text())[:indexs[0][0]]
	end := []byte(Text())[indexs[0][1]:]
	start := []byte(Text())[indexs[0][0]+len(fieldStart) : indexs[0][1]]

	ts, err := strconv.ParseInt(string(start), 10, 0)
	if err != nil {
		slog.Error("time ParseDuration ", "error", err)
		return err
	}

	fmt.Fprintln(w, strings.Join([]string{
		string(begin),
		fmt.Sprintf("%s%d", fieldStart, ts+delta),
		string(end),
	}, ""))

	return nil
}

type SRT struct {
}

func (format SRT) Sync(w io.Writer, Text func() string, opt SyncOption) error {
	// panic("not implemented")

	delta := opt.timestamp

	exp, err := regexp.Compile(`\d+:\d+:\d+,\d+ --> \d+:\d+:\d+,\d+`)
	if err != nil {
		slog.Error("regexp Compile ", "error", err)
		return err
	}

	expN, err := regexp.Compile(`\d+:\d+:\d+,\d+`)
	if err != nil {
		slog.Error("regexp Compile ", "error", err)
		return err
	}

	if !exp.MatchString(Text()) {
		slog.Info("not matched", "text", Text())

		fmt.Fprintln(w, Text())
		return nil
	}

	slog.Info("matched", "text", Text())

	shift := func(s string) string {
		const layout = "15:04:05,000"

		t, err := time.Parse(layout, s)
		if err != nil {
			panic(fmt.Errorf("time parse: %w", err))
		}

		t = t.Add(time.Duration(delta))

		slog.Info("shift", "old", s, "new", t.Format(layout))

		return t.Format(layout)
	}

	indexs := expN.FindAllIndex([]byte(Text()), -1)

	// begin := []byte(Text())[:indexs[0][0]]
	// end := []byte(Text())[indexs[0][1]:]
	// start := []byte(Text())[indexs[0][0]+len(fieldStart) : indexs[0][1]]

	s1 := shift(Text()[indexs[0][0]:indexs[0][1]])
	s2 := shift(Text()[indexs[1][0]:indexs[1][1]])

	fmt.Fprintf(w, "%s --> %s\n", s1, s2)

	return nil
}
