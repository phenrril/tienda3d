package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/phenrril/tienda3d/internal/adapters/telegram"
	"github.com/phenrril/tienda3d/internal/domain"
)

// RunWorkshopDigestLoop envía un resumen diario por Telegram de pedidos con entrega en los próximos 5 días.
func (a *App) RunWorkshopDigestLoop(ctx context.Context) {
	if a.WorkshopAdmin == nil {
		return
	}
	tz := strings.TrimSpace(os.Getenv("WORKSHOP_DIGEST_TZ"))
	if tz == "" {
		tz = "America/Argentina/Buenos_Aires"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		log.Warn().Err(err).Str("tz", tz).Msg("WORKSHOP_DIGEST_TZ inválida, usando local")
		loc = time.Local
	}
	hour := 9
	if h := os.Getenv("WORKSHOP_DIGEST_HOUR"); h != "" {
		if v, err := strconv.Atoi(h); err == nil && v >= 0 && v < 24 {
			hour = v
		}
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				a.maybeSendWorkshopDigest(context.Background(), loc, hour)
			}
		}
	}()
}

func (a *App) maybeSendWorkshopDigest(ctx context.Context, loc *time.Location, hour int) {
	wa := a.WorkshopAdmin
	if wa == nil {
		return
	}
	now := time.Now().In(loc)
	if now.Hour() != hour || now.Minute() > 2 {
		return
	}
	todayStr := now.Format("2006-01-02")
	last, _ := wa.Settings.Get(ctx, domain.SettingWorkshopDigestLast)
	if last == todayStr {
		return
	}
	end := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc).AddDate(0, 0, 5)
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	list, err := wa.Orders.ListUpcomingForDigest(ctx, start, end)
	if err != nil {
		log.Warn().Err(err).Msg("workshop digest list")
		return
	}
	if len(list) == 0 {
		_ = wa.Settings.Set(ctx, domain.SettingWorkshopDigestLast, todayStr)
		return
	}
	var b strings.Builder
	b.WriteString("Pedidos próximos (5 días):\n")
	for _, o := range list {
		fmt.Fprintf(&b, "• %s entrega %s estado %s\n", o.ClientSlug, o.DeliveryDate.Format("2006-01-02"), o.Status)
	}
	if err := telegram.SendPlain(b.String()); err != nil {
		log.Warn().Err(err).Msg("workshop digest telegram")
		return
	}
	_ = wa.Settings.Set(ctx, domain.SettingWorkshopDigestLast, todayStr)
}
