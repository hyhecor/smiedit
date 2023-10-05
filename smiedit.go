package smiedit

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slog"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/transform"
)

var rootCmd = &cobra.Command{
	Use:     "smiedit",
	Short:   "smi file editor",
	Version: Ver,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("root")
		fmt.Println("args:", args)
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
	// rootCmd.AddCommand(versionCmd)
}

var syncCmd = &cobra.Command{
	Use:        "sync",
	Short:      "edit smi sync",
	Args:       cobra.MinimumNArgs(1),
	ArgAliases: []string{"filename"},
	Run: func(cmd *cobra.Command, args []string) {
		// fmt.Println("sync")
		// fmt.Println("args:", args)

		sync.filename = args[0]

		// fmt.Println("timestamp:", sync.timestamp)
		// fmt.Println("filename:", sync.filename)

		sync.Exec()
	},
}

type Sync struct {
	filename  string
	timestamp time.Duration
	output    string
}

var sync Sync

func init() {
	syncCmd.PersistentFlags().DurationVarP(&sync.timestamp, "timestamp", "t", time.Duration(0), "sync timestamp")
	syncCmd.PersistentFlags().StringVarP(&sync.output, "output", "o", "-", "out filename")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (cmd Sync) Exec() error {
	file, err := os.Open(cmd.filename)
	if err != nil {
		slog.Error("cannot file open", "error", err, "filename", cmd.filename)
		return err
	}

	defer file.Close()

	delta := int64(cmd.timestamp / time.Millisecond)

	exp, err := regexp.Compile(`<SYNC Start=\d+>`)
	if err != nil {
		slog.Error("regexp Compile ", "error", err)
		return err
	}

	fieldStart := "Start="

	expN, err := regexp.Compile(`Start=\d+`)
	if err != nil {
		slog.Error("regexp Compile ", "error", err)
		return err
	}

	var buf = &bytes.Buffer{}
	if _, err := io.Copy(buf, file); err != nil {
		slog.Error("io Copy ", "error", err)
		return err
	}

	w := os.Stdout
	if cmd.output != "-" {
		w, err = os.Create(cmd.output)
		if err != nil {
			slog.Error("io Copy ", "error", err)
			return err
		}

		defer w.Close()
	}

	reader := transform.NewReader(buf, korean.EUCKR.NewDecoder())
	// writer := transform.NewWriter(os.Stdout, unicode.UTF8.NewEncoder())
	writer := transform.NewWriter(w, korean.EUCKR.NewEncoder())

	fileScanner := bufio.NewScanner(reader)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		if !exp.MatchString(fileScanner.Text()) {
			fmt.Fprintln(writer, fileScanner.Text())
			continue
		}

		indexs := expN.FindAllIndex([]byte(fileScanner.Text()), -1)

		begin := []byte(fileScanner.Text())[:indexs[0][0]]
		end := []byte(fileScanner.Text())[indexs[0][1]:]
		start := []byte(fileScanner.Text())[indexs[0][0]+len(fieldStart) : indexs[0][1]]

		ts, err := strconv.ParseInt(string(start), 10, 0)
		if err != nil {
			slog.Error("time ParseDuration ", "error", err)
			return err
		}

		fmt.Fprintln(writer, strings.Join([]string{
			string(begin),
			fmt.Sprintf("%s%d", fieldStart, ts+delta),
			string(end),
		}, ""))
	}

	return nil
}
