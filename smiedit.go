package smiedit

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
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

		syncOption.filename = args[0]

		// fmt.Println("timestamp:", sync.timestamp)
		// fmt.Println("filename:", sync.filename)

		syncOption.Exec()
	},
}

type SyncOption struct {
	filename       string
	timestamp      time.Duration
	outFilename    string
	fileFormat     string
	readerEncoding string
	writerEncoding string
}

var syncOption = SyncOption{
	timestamp:      time.Duration(0),
	outFilename:    "-",
	fileFormat:     FileFormat_SMI,
	readerEncoding: Encoding_UTF8,
	writerEncoding: Encoding_UTF8,
}

func init() {
	syncCmd.Flags().DurationVarP(&syncOption.timestamp, "timestamp", "t", syncOption.timestamp, "sync timestamp")
	syncCmd.Flags().StringVarP(&syncOption.outFilename, "out", "o", syncOption.outFilename, "out filename")
	syncCmd.Flags().StringVarP(&syncOption.fileFormat, "file-format", "f", syncOption.fileFormat, "file format")
	syncCmd.Flags().StringVarP(&syncOption.readerEncoding, "reader-encoding", "R", syncOption.readerEncoding, `decoder encoding`)
	syncCmd.Flags().StringVarP(&syncOption.writerEncoding, "writer-encoding", "W", syncOption.writerEncoding, `encoder encoding`)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (opt SyncOption) Exec() error {
	file, err := os.Open(opt.filename)
	if err != nil {
		slog.Error("cannot file open", "error", err, "filename", opt.filename)
		return err
	}

	defer file.Close()

	var input = &bytes.Buffer{}
	if _, err := io.Copy(input, file); err != nil {
		slog.Error("io Copy ", "error", err)
		return err
	}

	output := os.Stdout
	if opt.outFilename != "-" {
		output, err = os.Create(opt.outFilename)
		if err != nil {
			slog.Error("io Copy ", "error", err)
			return err
		}

		defer output.Close()
	}

	formater := NewFileFormat(opt.fileFormat)

	r := transform.NewReader(input, NewEncoding(opt.readerEncoding).NewDecoder())
	w := transform.NewWriter(output, NewEncoding(opt.writerEncoding).NewEncoder())

	fileScanner := bufio.NewScanner(r)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		err := formater.Sync(w, fileScanner.Text, opt)
		if err != nil {
			return err
		}
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

type FileFormater interface {
	Sync(w io.Writer, Text func() string, opt SyncOption) error
}

func NewFileFormat(fileFormat string) FileFormater {
	switch fileFormat {
	case FileFormat_SMI:
		return &SMI{}
	case FileFormat_SRT:
		return &SRT{}
	default:
		panic(fmt.Errorf("unsupported file format string=%v enum=[%v]", fileFormat, strings.Join(FileFormats(), ", ")))
	}
}

const (
	FileFormat_SMI = "smi"
	FileFormat_SRT = "srt"
)

func FileFormats() []string {
	return []string{
		FileFormat_SMI,
		FileFormat_SRT,
	}
}
