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
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/unicode"
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
	filename         string
	timestamp        time.Duration
	output           string
	decoderEncoding  string
	encodingEncoding string
}

var sync = Sync{
	timestamp:        time.Duration(0),
	output:           "-",
	decoderEncoding:  "UTF8",
	encodingEncoding: "UTF8",
}

func init() {
	syncCmd.PersistentFlags().DurationVarP(&sync.timestamp, "timestamp", "t", sync.timestamp, "sync timestamp")
	syncCmd.PersistentFlags().StringVarP(&sync.output, "output", "o", "-", "out filename")
	syncCmd.PersistentFlags().StringVar(&sync.decoderEncoding, "decoder-encoding", sync.decoderEncoding, `decoder encoding enum("UTF8", "EUCKR")`)
	syncCmd.PersistentFlags().StringVar(&sync.encodingEncoding, "encoder-encoding", sync.encodingEncoding, `encoder encoding enum("UTF8", "EUCKR")`)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (sync Sync) Exec() error {
	file, err := os.Open(sync.filename)
	if err != nil {
		slog.Error("cannot file open", "error", err, "filename", sync.filename)
		return err
	}

	defer file.Close()

	delta := int64(sync.timestamp / time.Millisecond)

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

	var buf = &bytes.Buffer{}
	if _, err := io.Copy(buf, file); err != nil {
		slog.Error("io Copy ", "error", err)
		return err
	}

	w := os.Stdout
	if sync.output != "-" {
		w, err = os.Create(sync.output)
		if err != nil {
			slog.Error("io Copy ", "error", err)
			return err
		}

		defer w.Close()
	}

	decoder := transform.NewReader(buf, Encoding(sync.decoderEncoding).NewDecoder())
	encoder := transform.NewWriter(w, Encoding(sync.encodingEncoding).NewEncoder())

	fileScanner := bufio.NewScanner(decoder)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {

		if !exp.MatchString(fileScanner.Text()) {
			slog.Info("not matched", fileScanner.Text())

			fmt.Fprintln(encoder, fileScanner.Text())
			continue
		}

		slog.Info("matched", fileScanner.Text())

		indexs := expN.FindAllIndex([]byte(fileScanner.Text()), -1)

		begin := []byte(fileScanner.Text())[:indexs[0][0]]
		end := []byte(fileScanner.Text())[indexs[0][1]:]
		start := []byte(fileScanner.Text())[indexs[0][0]+len(fieldStart) : indexs[0][1]]

		ts, err := strconv.ParseInt(string(start), 10, 0)
		if err != nil {
			slog.Error("time ParseDuration ", "error", err)
			return err
		}

		fmt.Fprintln(encoder, strings.Join([]string{
			string(begin),
			fmt.Sprintf("%s%d", fieldStart, ts+delta),
			string(end),
		}, ""))
	}

	return nil
}

func Encoding(encoding string) encoding.Encoding {
	switch encoding {
	case "UTF8":
		return unicode.UTF8
	case "EUCKR":
		return korean.EUCKR
	default:
		panic(fmt.Errorf("unsupported encoding string=%q", encoding))
	}
}
