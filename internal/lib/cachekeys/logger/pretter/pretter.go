package slogpretty

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	stdLog "log"
	"log/slog"

	"github.com/fatih/color"
)

type PrettyHandlerOptions struct {
	SlogOpts *slog.HandlerOptions
}

type PrettyHandler struct {
	opts PrettyHandlerOptions
	slog.Handler
	l     *stdLog.Logger
	attrs []slog.Attr
}

func (opts PrettyHandlerOptions) NewPrettyHandler(
	out io.Writer,
) *PrettyHandler {
	h := &PrettyHandler{
		Handler: slog.NewJSONHandler(out, opts.SlogOpts),
		l:       stdLog.New(out, "", 0),
	}

	return h
}

func (h *PrettyHandler) Handle(_ context.Context, r slog.Record) error {
	level := r.Level.String() + ":"

	switch r.Level {
	case slog.LevelDebug:
		level = color.MagentaString(level)
	case slog.LevelInfo:
		level = color.BlueString(level)
	case slog.LevelWarn:
		level = color.YellowString(level)
	case slog.LevelError:
		level = color.RedString(level)
	}

	fields := make(map[string]interface{}, r.NumAttrs())

	// Проход по атрибутам записи
	r.Attrs(func(a slog.Attr) bool {
		fields[a.Key] = h.convertValue(a.Value)
		return true
	})

	// Добавление дополнительных атрибутов (если есть)
	for _, a := range h.attrs {
		fields[a.Key] = h.convertValue(a.Value)
	}

	// Форматирование данных в JSON
	var b []byte
	var err error

	if len(fields) > 0 {
		b, err = json.MarshalIndent(fields, "", "  ")
		if err != nil {
			return err
		}
	}

	timeStr := r.Time.Format("[15:05:05.000]")
	msg := color.CyanString(r.Message)

	// Логирование с форматированием
	h.l.Println(
		timeStr,
		level,
		msg,
		color.WhiteString(string(b)),
	)

	return nil
}

// convertValue - безопасное преобразование значений в интерфейс
func (h *PrettyHandler) convertValue(v slog.Value) interface{} {
	switch v.Kind() {
	case slog.KindString:
		return v.String()
	case slog.KindInt64:
		return v.Int64()
	case slog.KindUint64:
		return v.Uint64()
	case slog.KindFloat64:
		return v.Float64()
	case slog.KindBool:
		return v.Bool()
	case slog.KindTime:
		return v.Time()
	case slog.KindDuration:
		return v.Duration()
	case slog.KindAny:
		// Проверка на тип ошибки
		if err, ok := v.Any().(error); ok {
			return err.Error()
		}
		return v.Any()
	default:
		return fmt.Sprintf("unsupported value type: %v", v.Kind())
	}
}

func (h *PrettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	// Добавление новых атрибутов к текущим
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	copy(newAttrs[len(h.attrs):], attrs)

	return &PrettyHandler{
		Handler: h.Handler,
		l:       h.l,
		attrs:   newAttrs,
	}
}

func (h *PrettyHandler) WithGroup(name string) slog.Handler {
	// Добавление группировки атрибутов
	groupAttrs := append(h.attrs, slog.Group(name))
	return &PrettyHandler{
		Handler: h.Handler.WithGroup(name),
		l:       h.l,
		attrs:   groupAttrs,
	}
}
