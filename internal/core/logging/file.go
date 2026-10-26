package logging

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"sync"
	"time"
)

// 日志文件设置的缺省值（CONTRACTS §14.6）。
const (
	DefaultMaxFileSize  = 100 << 20 // 100 MiB
	DefaultMaxTotalSize = 1 << 30   // 1 GiB
	DefaultMaxAgeDays   = 30
)

// 日志文件名为 alonabot-<日期>.log，同日续写的文件为 alonabot-<日期>.<序号>.log（CONTRACTS §14.6）。
const dateLayout = "2006-01-02"

var logFileName = regexp.MustCompile(`^alonabot-(\d{4}-\d{2}-\d{2})(?:\.([1-9]\d*))?\.log$`)

// fileName 返回 day 当天第 index 个文件的文件名；index 0 不带序号。
func fileName(day string, index int) string {
	if index == 0 {
		return "alonabot-" + day + ".log"
	}
	return "alonabot-" + day + "." + strconv.Itoa(index) + ".log"
}

// FileOptions 描述日志文件的位置与上限。
type FileOptions struct {
	Dir        string // <data_dir>/logs
	MaxSize    int64  // 单个文件的字节上限，达到后同日续写下一个文件
	MaxTotal   int64  // 全部日志文件合计的字节上限
	MaxAgeDays int    // 日期早于这么多天之前的文件被删除
}

// fileWriter 把控制台格式的记录追加到当天的日志文件：跨过零点或达到单文件上限时换文件，
// 换文件时按合计与天数上限删除最旧的文件。写入失败时丢弃记录并只报告一次，到下一次换日期时重新打开。
type fileWriter struct {
	report func(error)      // 报告文件错误；不能写回本文件
	now    func() time.Time // 取当前本地时间

	// mu 保护以下字段。
	mu       sync.Mutex
	opt      FileOptions // Dir 为空表示尚未启用，写入被丢弃
	f        *os.File    // 当前文件；打开或写入失败后为 nil
	day      string      // 当前文件的日期
	index    int         // 当前文件的同日序号
	size     int64       // 当前文件的大小
	reported bool        // 本次失败已经报告过；成功打开文件后清除
}

// enable 校验上限、创建目录并打开当天的文件。
func (w *fileWriter) enable(opt FileOptions) error {
	if opt.MaxSize <= 0 || opt.MaxTotal <= 0 || opt.MaxAgeDays <= 0 {
		return fmt.Errorf("log file limits must be positive: max_size=%d max_total=%d max_age_days=%d", opt.MaxSize, opt.MaxTotal, opt.MaxAgeDays)
	}
	if opt.MaxSize > opt.MaxTotal {
		return fmt.Errorf("log.file.max_size (%d bytes) exceeds log.file.max_total (%d bytes)", opt.MaxSize, opt.MaxTotal)
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.now == nil {
		w.now = time.Now
	}
	w.opt = opt
	if err := os.MkdirAll(opt.Dir, 0o700); err != nil {
		w.fail(err)
		return nil
	}
	w.openDay(w.now().Format(dateLayout))
	return nil
}

// Write 追加一条完整的记录；总是报告成功，文件错误由 report 处理，不影响 stderr 上的输出。
func (w *fileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.opt.Dir == "" {
		return len(p), nil
	}
	// 1. 跨过零点换到新日期的文件；同日写满单文件上限时续写下一个序号（单条超限的记录独占一个文件）。
	if day := w.now().Format(dateLayout); day != w.day {
		w.openDay(day)
	} else if w.f != nil && w.size > 0 && w.size+int64(len(p)) > w.opt.MaxSize {
		w.open(day, w.index+1)
	}
	if w.f == nil {
		return len(p), nil
	}
	// 2. 写入；失败时关闭文件，之后的记录丢弃到下一次换日期。
	n, err := w.f.Write(p)
	w.size += int64(n)
	if err != nil {
		w.fail(err)
	}
	return len(p), nil
}

// openDay 打开 day 当天序号最大的已有文件继续写；它已写满时续写下一个序号。
func (w *fileWriter) openDay(day string) {
	files, err := w.list()
	if err != nil {
		w.closeCurrent()
		w.day = day
		w.fail(err)
		return
	}
	index := 0
	for _, f := range files {
		if f.day == day {
			index = max(index, f.index)
		}
	}
	w.open(day, index)
	if w.f != nil && w.size >= w.opt.MaxSize {
		w.open(day, index+1)
	}
}

// open 关闭当前文件，以追加方式打开 day 的第 index 个文件，然后按上限清理旧文件。
func (w *fileWriter) open(day string, index int) {
	w.closeCurrent()
	w.day, w.index = day, index
	f, err := os.OpenFile(filepath.Join(w.opt.Dir, fileName(day, index)), os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o600)
	if err != nil {
		w.fail(err)
		return
	}
	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		w.fail(err)
		return
	}
	w.f, w.size, w.reported = f, fi.Size(), false
	w.prune()
}

// closeCurrent 关闭当前文件；关闭失败按写入失败报告。
func (w *fileWriter) closeCurrent() {
	if w.f == nil {
		return
	}
	if err := w.f.Close(); err != nil {
		w.fail(err)
	}
	w.f = nil
}

// fail 关闭当前文件并在本次失败第一次发生时报告。
func (w *fileWriter) fail(err error) {
	if w.f != nil {
		_ = w.f.Close()
		w.f = nil
	}
	if !w.reported {
		w.reported = true
		w.report(err)
	}
}

// logFile 是目录中的一个日志文件。
type logFile struct {
	name  string
	day   string
	index int
	size  int64
}

// list 按日期与序号从旧到新列出目录中的日志文件。
func (w *fileWriter) list() ([]logFile, error) {
	entries, err := os.ReadDir(w.opt.Dir)
	if err != nil {
		return nil, err
	}
	var files []logFile
	for _, e := range entries {
		m := logFileName.FindStringSubmatch(e.Name())
		if m == nil || !e.Type().IsRegular() {
			continue
		}
		fi, err := e.Info()
		if errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, err
		}
		index, _ := strconv.Atoi(m[2]) // 正则已保证是数字；无序号时为 0
		files = append(files, logFile{name: e.Name(), day: m[1], index: index, size: fi.Size()})
	}
	slices.SortFunc(files, func(a, b logFile) int { return cmp.Or(cmp.Compare(a.day, b.day), cmp.Compare(a.index, b.index)) })
	return files, nil
}

// prune 删除日期早于 MaxAgeDays 天之前的文件，再从最旧的开始删除，直到合计不超过 MaxTotal；
// 当前文件不删除。删除失败的文件单独报告，不影响写入。
func (w *fileWriter) prune() {
	files, err := w.list()
	if err != nil {
		w.report(err)
		return
	}
	current := fileName(w.day, w.index)
	cutoff := w.now().AddDate(0, 0, -w.opt.MaxAgeDays).Format(dateLayout)
	var total int64
	for _, f := range files {
		total += f.size
	}
	for _, f := range files {
		if f.name == current || (f.day >= cutoff && total <= w.opt.MaxTotal) {
			continue
		}
		if err := os.Remove(filepath.Join(w.opt.Dir, f.name)); err != nil {
			w.report(err)
			continue
		}
		total -= f.size
	}
}

// close 关闭当前文件并停止写入。
func (w *fileWriter) close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.opt.Dir = ""
	if w.f == nil {
		return nil
	}
	err := w.f.Close()
	w.f = nil
	return err
}
