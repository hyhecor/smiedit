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
	filename       string
	timestamp      time.Duration
	output         string
	readerEncoding string
	writerEncoding string
}

var sync = Sync{
	timestamp:      time.Duration(0),
	output:         "-",
	readerEncoding: Encoding_UTF8,
	writerEncoding: Encoding_UTF8,
}

func init() {
	syncCmd.Flags().DurationVarP(&sync.timestamp, "timestamp", "t", sync.timestamp, "sync timestamp")
	syncCmd.Flags().StringVarP(&sync.output, "output", "o", "-", "out filename")
	syncCmd.Flags().StringVarP(&sync.readerEncoding, "reader-encoding", "R", sync.readerEncoding, `decoder encoding`)
	syncCmd.Flags().StringVarP(&sync.writerEncoding, "writer-encoding", "W", sync.writerEncoding, `encoder encoding`)
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

	var input = &bytes.Buffer{}
	if _, err := io.Copy(input, file); err != nil {
		slog.Error("io Copy ", "error", err)
		return err
	}

	output := os.Stdout
	if sync.output != "-" {
		output, err = os.Create(sync.output)
		if err != nil {
			slog.Error("io Copy ", "error", err)
			return err
		}

		defer output.Close()
	}

	r := transform.NewReader(input, NewEncoding(sync.readerEncoding).NewDecoder())
	w := transform.NewWriter(output, NewEncoding(sync.writerEncoding).NewEncoder())

	fileScanner := bufio.NewScanner(r)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {

		Text := fileScanner.Text

		if !exp.MatchString(Text()) {
			slog.Info("not matched", Text())

			fmt.Fprintln(w, Text())
			continue
		}

		slog.Info("matched", Text())

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
	}

	return nil
}

func NewEncoding(encoding string) encoding.Encoding {
	switch encoding {
	case Encoding_UTF8:
		return unicode.UTF8
	case Encoding_UTF8BOM:
		return unicode.UTF8BOM
	case Encoding_UTF16LE:
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
	case Encoding_UTF16LEBOM:
		return unicode.UTF16(unicode.LittleEndian, unicode.UseBOM)
	case Encoding_UTF16BE:
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM)
	case Encoding_UTF16BEBOM:
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM)
	case Encoding_EUCKR:
		return korean.EUCKR
	default:
		panic(fmt.Errorf("unsupported encoding string=%v enum=[%v]", encoding, strings.Join(Encodings(), ", ")))
	}
}

const (
	Encoding_UTF8       = "UTF8"
	Encoding_UTF8BOM    = "UTF8BOM"
	Encoding_UTF16LE    = "UTF16LE"
	Encoding_UTF16LEBOM = "UTF16LEBOM"
	Encoding_UTF16BE    = "UTF16BE"
	Encoding_UTF16BEBOM = "UTF16BEBOM"
	Encoding_EUCKR      = "EUCKR"
)

func Encodings() []string {
	return []string{
		Encoding_UTF8,
		Encoding_UTF8BOM,
		Encoding_UTF16LE,
		Encoding_UTF16LEBOM,
		Encoding_UTF16BE,
		Encoding_UTF16BEBOM,
		Encoding_EUCKR,
	}
}
